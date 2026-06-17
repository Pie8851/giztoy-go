package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/GizClaw/gizclaw-go/pkg/audio/codec/opus"
	"github.com/GizClaw/gizclaw-go/pkg/audio/codecconv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

type opStats struct {
	Name      string
	Total     time.Duration
	First     time.Duration
	FirstName string
	Events    int
	Bytes     int
	Chars     int
}

type chainStats struct {
	Name  string
	Total time.Duration
	Ops   []opStats
}

type chainResult struct {
	Stats         chainStats
	Completion    string
	SpeechPath    string
	SpeechBytes   int
	Transcription string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "openai-compat: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := loadConfig(args)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	httpClient := &http.Client{Timeout: cfg.Timeout}
	client := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL),
		option.WithHTTPClient(httpClient),
	)

	var chains []chainStats
	fail := func(err error) error {
		if len(chains) > 0 {
			printChains(chains)
		}
		return err
	}

	voiceID := cfg.VoiceID
	if voiceID == "" {
		var err error
		voiceID, err = firstVoiceID(ctx, httpClient, cfg)
		if err != nil {
			return fail(err)
		}
	}

	nonStream, err := runNonStreamingChain(ctx, client, cfg, voiceID)
	chains = append(chains, nonStream.Stats)
	if err != nil {
		return fail(err)
	}

	stream, err := runStreamingChain(ctx, httpClient, client, cfg, voiceID)
	chains = append(chains, stream.Stats)
	if err != nil {
		return fail(err)
	}

	fmt.Printf("base_url=%s\n", cfg.BaseURL)
	fmt.Printf("model=%s\n", cfg.ModelID)
	fmt.Printf("tts_model=%s\n", cfg.TTSModelID)
	fmt.Printf("voice=%s\n", voiceID)
	fmt.Printf("non_stream_chat=%q\n", strings.TrimSpace(nonStream.Completion))
	fmt.Printf("stream_chat=%q\n", strings.TrimSpace(stream.Completion))
	fmt.Printf("non_stream_speech=%s bytes=%d\n", nonStream.SpeechPath, nonStream.SpeechBytes)
	fmt.Printf("stream_speech=%s bytes=%d\n", stream.SpeechPath, stream.SpeechBytes)
	if cfg.ASRModelID != "" {
		fmt.Printf("asr_model=%s\n", cfg.ASRModelID)
		fmt.Printf("non_stream_transcription=%q\n", strings.TrimSpace(nonStream.Transcription))
		fmt.Printf("stream_transcription=%q\n", strings.TrimSpace(stream.Transcription))
	}
	printChains(chains)
	return nil
}

func runNonStreamingChain(ctx context.Context, client openai.Client, cfg config, voiceID string) (chainResult, error) {
	start := time.Now()
	result := chainResult{Stats: chainStats{Name: "non_stream"}}

	completion, stat, err := runChat(ctx, client, cfg.ModelID, cfg.Thinking)
	result.Stats.Ops = append(result.Stats.Ops, stat)
	if err != nil {
		result.Stats.Total = time.Since(start)
		return result, err
	}
	result.Completion = completion

	speechPath, speechBytes, stat, err := runSpeech(ctx, client, cfg.OutputDir, cfg.TTSModelID, voiceID, "speech.mp3", completion)
	result.Stats.Ops = append(result.Stats.Ops, stat)
	if err != nil {
		result.Stats.Total = time.Since(start)
		return result, err
	}
	result.SpeechPath = speechPath
	result.SpeechBytes = speechBytes

	if cfg.ASRModelID != "" {
		transcription, stat, err := runTranscription(ctx, client, cfg.ASRModelID, speechPath)
		result.Stats.Ops = append(result.Stats.Ops, stat)
		if err != nil {
			result.Stats.Total = time.Since(start)
			return result, err
		}
		if err := assertTranscriptionMatches("non-stream", completion, transcription); err != nil {
			result.Stats.Total = time.Since(start)
			return result, err
		}
		result.Transcription = transcription
	}

	result.Stats.Total = time.Since(start)
	return result, nil
}

func runStreamingChain(ctx context.Context, httpClient *http.Client, client openai.Client, cfg config, voiceID string) (chainResult, error) {
	start := time.Now()
	result := chainResult{Stats: chainStats{Name: "stream"}}

	completion, stat, err := runStreamingChat(ctx, client, cfg.ModelID, cfg.Thinking)
	result.Stats.Ops = append(result.Stats.Ops, stat)
	if err != nil {
		result.Stats.Total = time.Since(start)
		return result, err
	}
	result.Completion = completion

	speechPath, speechBytes, stat, err := runSpeechStream(ctx, client, cfg.OutputDir, cfg.TTSModelID, voiceID, "speech-stream.mp3", completion)
	result.Stats.Ops = append(result.Stats.Ops, stat)
	if err != nil {
		result.Stats.Total = time.Since(start)
		return result, err
	}
	result.SpeechPath = speechPath
	result.SpeechBytes = speechBytes

	if cfg.ASRModelID != "" {
		transcription, stat, err := runTranscriptionStream(ctx, httpClient, cfg, speechPath)
		result.Stats.Ops = append(result.Stats.Ops, stat)
		if err != nil {
			result.Stats.Total = time.Since(start)
			return result, err
		}
		if err := assertTranscriptionSimilar("stream", completion, transcription, 0.85); err != nil {
			result.Stats.Total = time.Since(start)
			return result, err
		}
		result.Transcription = transcription
	}

	result.Stats.Total = time.Since(start)
	return result, nil
}

func runChat(ctx context.Context, client openai.Client, modelID string, thinking thinkingConfig) (string, opStats, error) {
	start := time.Now()
	params := openai.ChatCompletionNewParams{
		Model: shared.ChatModel(modelID),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a smoke test endpoint. Follow the user instruction exactly."),
			openai.UserMessage("Reply with this exact Chinese sentence only, no punctuation: 小猫今天开心跑步"),
		},
	}
	applyThinking(&params, thinking)
	completion, err := client.Chat.Completions.New(ctx, params)
	stat := opStats{Name: "chat", Total: time.Since(start)}
	if err != nil {
		return "", stat, fmt.Errorf("chat completion with model %q: %w", modelID, err)
	}
	if len(completion.Choices) == 0 {
		return "", stat, fmt.Errorf("chat completion with model %q returned no choices", modelID)
	}
	text := strings.TrimSpace(completion.Choices[0].Message.Content)
	if text == "" {
		return "", stat, fmt.Errorf("chat completion with model %q returned empty content", modelID)
	}
	stat.Chars = utf8.RuneCountInString(text)
	return text, stat, nil
}

func runStreamingChat(ctx context.Context, client openai.Client, modelID string, thinking thinkingConfig) (string, opStats, error) {
	start := time.Now()
	params := openai.ChatCompletionNewParams{
		Model: shared.ChatModel(modelID),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a smoke test endpoint. Follow the user instruction exactly."),
			openai.UserMessage("Reply with this exact Chinese sentence only, no punctuation: 小猫今天开心跑步"),
		},
	}
	applyThinking(&params, thinking)
	stream := client.Chat.Completions.NewStreaming(ctx, params)
	defer stream.Close()

	var completion strings.Builder
	stat := opStats{Name: "chat_stream", FirstName: "first_token"}
	for stream.Next() {
		stat.Events++
		chunk := stream.Current()
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta != "" && stat.First == 0 {
			stat.First = time.Since(start)
		}
		completion.WriteString(delta)
	}
	stat.Total = time.Since(start)
	if err := stream.Err(); err != nil {
		return "", stat, fmt.Errorf("streaming chat completion with model %q: %w", modelID, err)
	}
	if strings.TrimSpace(completion.String()) == "" {
		return "", stat, fmt.Errorf("streaming chat completion with model %q returned no content", modelID)
	}
	stat.Chars = utf8.RuneCountInString(completion.String())
	return completion.String(), stat, nil
}

func applyThinking(params *openai.ChatCompletionNewParams, thinking thinkingConfig) {
	if params == nil {
		return
	}
	body := map[string]any{}
	if thinking.Enabled != nil {
		body["enabled"] = *thinking.Enabled
	}
	if strings.TrimSpace(thinking.Level) != "" {
		body["level"] = strings.TrimSpace(thinking.Level)
	}
	if len(body) == 0 {
		return
	}
	params.SetExtraFields(map[string]any{"thinking": body})
}

func runSpeech(ctx context.Context, client openai.Client, outputDir, modelID, voiceID, filename, input string) (string, int, opStats, error) {
	start := time.Now()
	speech, err := client.Audio.Speech.New(ctx, openai.AudioSpeechNewParams{
		Input:          input,
		Model:          openai.SpeechModel(modelID),
		Voice:          openai.AudioSpeechNewParamsVoice(voiceID),
		ResponseFormat: openai.AudioSpeechNewParamsResponseFormatMP3,
	})
	stat := opStats{Name: "speech", FirstName: "first_byte"}
	if err != nil {
		stat.Total = time.Since(start)
		return "", 0, stat, fmt.Errorf("speech with model %q voice %q: %w", modelID, voiceID, err)
	}
	defer speech.Body.Close()

	audio, firstByte, err := readAllMeasured(speech.Body, start)
	stat.Total = time.Since(start)
	stat.First = firstByte
	stat.Bytes = len(audio)
	if err != nil {
		return "", 0, stat, fmt.Errorf("read speech audio: %w", err)
	}
	if len(audio) == 0 {
		return "", 0, stat, fmt.Errorf("speech with model %q voice %q returned empty audio", modelID, voiceID)
	}
	filename = audioFilename(filename, speech.Header.Get("Content-Type"), audio)
	path := filepath.Join(outputDir, filename)
	if err := os.WriteFile(path, audio, 0o644); err != nil {
		return "", 0, stat, fmt.Errorf("write speech audio: %w", err)
	}
	return path, len(audio), stat, nil
}

func runSpeechStream(ctx context.Context, client openai.Client, outputDir, modelID, voiceID, filename, input string) (string, int, opStats, error) {
	start := time.Now()
	speech, err := client.Audio.Speech.New(ctx, openai.AudioSpeechNewParams{
		Input:          input,
		Model:          openai.SpeechModel(modelID),
		Voice:          openai.AudioSpeechNewParamsVoice(voiceID),
		ResponseFormat: openai.AudioSpeechNewParamsResponseFormatMP3,
		StreamFormat:   openai.AudioSpeechNewParamsStreamFormatSSE,
	})
	stat := opStats{Name: "speech_stream", FirstName: "first_audio_delta"}
	if err != nil {
		stat.Total = time.Since(start)
		return "", 0, stat, fmt.Errorf("streaming speech with model %q voice %q: %w", modelID, voiceID, err)
	}
	defer speech.Body.Close()
	if got := speech.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		stat.Total = time.Since(start)
		return "", 0, stat, fmt.Errorf("streaming speech content type = %q, want text/event-stream", got)
	}
	audio, events, firstDelta, err := readSpeechSSE(speech.Body, start)
	stat.Total = time.Since(start)
	stat.First = firstDelta
	stat.Events = events
	stat.Bytes = len(audio)
	if err != nil {
		return "", 0, stat, err
	}
	filename = audioFilename(filename, "", audio)
	path := filepath.Join(outputDir, filename)
	if err := os.WriteFile(path, audio, 0o644); err != nil {
		return "", 0, stat, fmt.Errorf("write streamed speech audio: %w", err)
	}
	return path, len(audio), stat, nil
}

func runTranscription(ctx context.Context, client openai.Client, modelID, audioPath string) (string, opStats, error) {
	file, err := os.Open(audioPath)
	if err != nil {
		return "", opStats{Name: "transcription"}, fmt.Errorf("open transcription audio: %w", err)
	}
	defer file.Close()

	start := time.Now()
	transcription, err := client.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		File:  file,
		Model: openai.AudioModel(modelID),
	})
	stat := opStats{Name: "transcription", Total: time.Since(start)}
	if err != nil {
		return "", stat, fmt.Errorf("transcription with model %q: %w", modelID, err)
	}
	if strings.TrimSpace(transcription.Text) == "" {
		return "", stat, fmt.Errorf("transcription with model %q returned empty text", modelID)
	}
	stat.Chars = utf8.RuneCountInString(transcription.Text)
	return transcription.Text, stat, nil
}

func runTranscriptionStream(ctx context.Context, httpClient *http.Client, cfg config, audioPath string) (string, opStats, error) {
	start := time.Now()
	stat := opStats{Name: "transcription_stream", FirstName: "first_delta"}
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	writeDone := make(chan transcriptionUploadResult, 1)
	go writeStreamingTranscriptionRequest(pw, writer, cfg.ASRModelID, audioPath, writeDone)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.BaseURL+"/audio/transcriptions", pr)
	if err != nil {
		_ = pr.CloseWithError(err)
		stat.Total = time.Since(start)
		return "", stat, fmt.Errorf("create streaming transcription request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := httpClient.Do(req)
	if err != nil {
		_ = pr.CloseWithError(err)
		stat.Total = time.Since(start)
		return "", stat, fmt.Errorf("streaming transcription with model %q: %w", cfg.ASRModelID, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		stat.Total = time.Since(start)
		return "", stat, fmt.Errorf("streaming transcription with model %q status %d: %s", cfg.ASRModelID, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		stat.Total = time.Since(start)
		return "", stat, fmt.Errorf("streaming transcription content type = %q, want text/event-stream", got)
	}

	text, events, firstDelta, err := readTranscriptionSSE(resp.Body, start)
	upload := <-writeDone
	stat.Total = time.Since(start)
	stat.First = firstDelta
	stat.Events = events
	stat.Bytes = upload.Bytes
	if upload.Err != nil {
		return "", stat, fmt.Errorf("upload streaming transcription audio: %w", upload.Err)
	}
	if err != nil {
		return "", stat, err
	}
	if strings.TrimSpace(text) == "" {
		return "", stat, fmt.Errorf("streaming transcription with model %q returned empty text", cfg.ASRModelID)
	}
	stat.Chars = utf8.RuneCountInString(text)
	return text, stat, nil
}

type transcriptionUploadResult struct {
	Bytes int
	Err   error
}

func writeStreamingTranscriptionRequest(pw *io.PipeWriter, writer *multipart.Writer, modelID, audioPath string, done chan<- transcriptionUploadResult) {
	finish := func(bytes int, err error) {
		if err != nil {
			_ = pw.CloseWithError(err)
		} else {
			_ = pw.Close()
		}
		done <- transcriptionUploadResult{Bytes: bytes, Err: err}
	}
	if err := writer.WriteField("model", modelID); err != nil {
		finish(0, err)
		return
	}
	if err := writer.WriteField("stream", "true"); err != nil {
		finish(0, err)
		return
	}
	audio, contentType, err := openStreamingTranscriptionAudio(audioPath)
	if err != nil {
		finish(0, err)
		return
	}
	defer audio.Close()
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filepath.Base(audioPath)))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		finish(0, err)
		return
	}
	n, err := io.Copy(part, audio)
	if err != nil {
		finish(int(n), err)
		return
	}
	if err := writer.Close(); err != nil {
		finish(int(n), err)
		return
	}
	finish(int(n), nil)
}

func openStreamingTranscriptionAudio(path string) (io.ReadCloser, string, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ogg", ".opus":
		file, err := os.Open(path)
		if err != nil {
			return nil, "", err
		}
		defer file.Close()
		var pcm bytes.Buffer
		if _, err := codecconv.OggToPCM(&pcm, file, opus.SampleRate16K); err != nil {
			return nil, "", fmt.Errorf("decode streamed speech ogg to pcm: %w", err)
		}
		return io.NopCloser(bytes.NewReader(pcm.Bytes())), "audio/pcm", nil
	default:
		file, err := os.Open(path)
		if err != nil {
			return nil, "", err
		}
		return file, audioContentType(path), nil
	}
}

func audioContentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ogg", ".opus":
		return "audio/ogg"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".flac":
		return "audio/flac"
	case ".m4a", ".mp4":
		return "audio/mp4"
	default:
		return "application/octet-stream"
	}
}

func audioFilename(filename, contentType string, data []byte) string {
	ext := audioExtension(contentType, data)
	if ext == "" {
		return filename
	}
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	if base == "" {
		base = "audio"
	}
	return base + ext
}

func audioExtension(contentType string, data []byte) string {
	contentType = strings.TrimSpace(strings.Split(contentType, ";")[0])
	switch contentType {
	case "audio/ogg", "application/ogg":
		return ".ogg"
	case "audio/mpeg", "audio/mp3":
		return ".mp3"
	case "audio/wav", "audio/wave", "audio/x-wav":
		return ".wav"
	case "audio/flac":
		return ".flac"
	case "audio/mp4", "audio/m4a":
		return ".m4a"
	}
	switch {
	case len(data) >= 4 && string(data[:4]) == "OggS":
		return ".ogg"
	case len(data) >= 3 && string(data[:3]) == "ID3":
		return ".mp3"
	case len(data) >= 2 && data[0] == 0xff && data[1]&0xe0 == 0xe0:
		return ".mp3"
	case len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WAVE":
		return ".wav"
	case len(data) >= 4 && string(data[:4]) == "fLaC":
		return ".flac"
	case len(data) >= 12 && string(data[4:8]) == "ftyp":
		return ".m4a"
	default:
		return ""
	}
}

func assertTranscriptionMatches(name, expected, actual string) error {
	expectedNorm := normalizeTranscript(expected)
	actualNorm := normalizeTranscript(actual)
	if expectedNorm == "" {
		return fmt.Errorf("%s transcription expected text is empty after normalization", name)
	}
	if actualNorm == "" {
		return fmt.Errorf("%s transcription actual text is empty after normalization", name)
	}
	if expectedNorm == actualNorm {
		return nil
	}
	return fmt.Errorf("%s transcription mismatch: expected %q normalized %q, got %q normalized %q", name, expected, expectedNorm, actual, actualNorm)
}

func assertTranscriptionSimilar(name, expected, actual string, minRatio float64) error {
	expectedNorm := normalizeTranscript(expected)
	actualNorm := normalizeTranscript(actual)
	if expectedNorm == "" {
		return fmt.Errorf("%s transcription expected text is empty after normalization", name)
	}
	if actualNorm == "" {
		return fmt.Errorf("%s transcription actual text is empty after normalization", name)
	}
	ratio := lcsRatio(expectedNorm, actualNorm)
	if ratio >= minRatio {
		return nil
	}
	return fmt.Errorf("%s transcription mismatch: similarity %.2f below %.2f: expected %q normalized %q, got %q normalized %q", name, ratio, minRatio, expected, expectedNorm, actual, actualNorm)
}

func lcsRatio(a, b string) float64 {
	ar := []rune(a)
	br := []rune(b)
	if len(ar) == 0 || len(br) == 0 {
		return 0
	}
	prev := make([]int, len(br)+1)
	curr := make([]int, len(br)+1)
	for i := range ar {
		for j := range br {
			if ar[i] == br[j] {
				curr[j+1] = prev[j] + 1
			} else if curr[j] > prev[j+1] {
				curr[j+1] = curr[j]
			} else {
				curr[j+1] = prev[j+1]
			}
		}
		prev, curr = curr, prev
		clear(curr)
	}
	return float64(prev[len(br)]) / float64(max(len(ar), len(br)))
}

func normalizeTranscript(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || (r >= '\u4e00' && r <= '\u9fff') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func readSpeechSSE(r io.Reader, start time.Time) ([]byte, int, time.Duration, error) {
	var audio []byte
	var deltaCount int
	var firstDelta time.Duration
	var done bool
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		var event struct {
			Audio string `json:"audio"`
			Done  bool   `json:"done"`
			Type  string `json:"type"`
		}
		if err := json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data:"))), &event); err != nil {
			return nil, deltaCount, firstDelta, fmt.Errorf("decode streaming speech event %q: %w", line, err)
		}
		switch event.Type {
		case "speech.audio.delta":
			if firstDelta == 0 {
				firstDelta = time.Since(start)
			}
			chunk, err := base64.StdEncoding.DecodeString(event.Audio)
			if err != nil {
				return nil, deltaCount, firstDelta, fmt.Errorf("decode streaming speech audio delta: %w", err)
			}
			if len(chunk) == 0 {
				return nil, deltaCount, firstDelta, fmt.Errorf("streaming speech audio delta is empty")
			}
			audio = append(audio, chunk...)
			deltaCount++
		case "speech.audio.done":
			done = true
		default:
			return nil, deltaCount, firstDelta, fmt.Errorf("unexpected streaming speech event type %q", event.Type)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, deltaCount, firstDelta, fmt.Errorf("read streaming speech events: %w", err)
	}
	if deltaCount == 0 {
		return nil, deltaCount, firstDelta, fmt.Errorf("streaming speech returned no audio delta events")
	}
	if !done {
		return nil, deltaCount, firstDelta, fmt.Errorf("streaming speech did not return speech.audio.done")
	}
	if len(audio) == 0 {
		return nil, deltaCount, firstDelta, fmt.Errorf("streaming speech decoded audio is empty")
	}
	return audio, deltaCount, firstDelta, nil
}

func readTranscriptionSSE(r io.Reader, start time.Time) (string, int, time.Duration, error) {
	var delta strings.Builder
	var doneText string
	var eventCount int
	var firstDelta time.Duration
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}
		var event struct {
			Delta string `json:"delta"`
			Text  string `json:"text"`
			Type  string `json:"type"`
		}
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			return "", eventCount, firstDelta, fmt.Errorf("decode streaming transcription event %q: %w", line, err)
		}
		eventCount++
		switch event.Type {
		case "transcript.text.delta":
			if event.Delta != "" && firstDelta == 0 {
				firstDelta = time.Since(start)
			}
			delta.WriteString(event.Delta)
		case "transcript.text.done":
			doneText = event.Text
		default:
			return "", eventCount, firstDelta, fmt.Errorf("unexpected transcription stream event type %q", event.Type)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", eventCount, firstDelta, fmt.Errorf("read streaming transcription events: %w", err)
	}
	if strings.TrimSpace(doneText) != "" {
		return doneText, eventCount, firstDelta, nil
	}
	return delta.String(), eventCount, firstDelta, nil
}

func readAllMeasured(r io.Reader, start time.Time) ([]byte, time.Duration, error) {
	var out []byte
	var firstByte time.Duration
	buf := make([]byte, 32*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if firstByte == 0 {
				firstByte = time.Since(start)
			}
			out = append(out, buf[:n]...)
		}
		if err == io.EOF {
			return out, firstByte, nil
		}
		if err != nil {
			return out, firstByte, err
		}
	}
}

func printChains(chains []chainStats) {
	fmt.Println("chains:")
	printChainSummary(chains)
	for _, chain := range chains {
		if len(chain.Ops) == 0 {
			continue
		}
		fmt.Printf("%s:\n", chain.Name)
		printStatsTable(chain.Ops)
	}
}

func printChainSummary(chains []chainStats) {
	headers := []string{"chain", "total", "operations"}
	rows := make([][]string, 0, len(chains))
	for _, chain := range chains {
		rows = append(rows, []string{
			chain.Name,
			chain.Total.Round(time.Millisecond).String(),
			formatCount(len(chain.Ops)),
		})
	}
	printTable(headers, rows)
}

func printStatsTable(stats []opStats) {
	headers := []string{"operation", "total", "first", "events", "bytes", "chars"}
	rows := make([][]string, 0, len(stats))
	for _, stat := range stats {
		first := "-"
		if stat.FirstName != "" && stat.First > 0 {
			first = stat.First.Round(time.Millisecond).String()
		}
		rows = append(rows, []string{
			stat.Name,
			stat.Total.Round(time.Millisecond).String(),
			first,
			formatCount(stat.Events),
			formatCount(stat.Bytes),
			formatCount(stat.Chars),
		})
	}
	printTable(headers, rows)
}

func printTable(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = utf8.RuneCountInString(header)
	}
	for _, row := range rows {
		for i, value := range row {
			if w := utf8.RuneCountInString(value); w > widths[i] {
				widths[i] = w
			}
		}
	}

	printTableBorder("┌", "┬", "┐", widths)
	printTableRow(headers, widths)
	printTableBorder("├", "┼", "┤", widths)
	for _, row := range rows {
		printTableRow(row, widths)
	}
	printTableBorder("└", "┴", "┘", widths)
}

func printTableBorder(left, mid, right string, widths []int) {
	fmt.Print(left)
	for i, width := range widths {
		if i > 0 {
			fmt.Print(mid)
		}
		fmt.Print(strings.Repeat("─", width+2))
	}
	fmt.Println(right)
}

func printTableRow(row []string, widths []int) {
	fmt.Print("│")
	for i, width := range widths {
		value := ""
		if i < len(row) {
			value = row[i]
		}
		fmt.Printf(" %s%s │", value, strings.Repeat(" ", width-utf8.RuneCountInString(value)))
	}
	fmt.Println()
}

func formatCount(n int) string {
	if n <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d", n)
}

func firstVoiceID(ctx context.Context, client *http.Client, cfg config) (string, error) {
	voicesURL, err := url.Parse(cfg.BaseURL + "/voices")
	if err != nil {
		return "", fmt.Errorf("parse voices url: %w", err)
	}
	q := voicesURL.Query()
	q.Set("providerKind", "volc-tenant")
	q.Set("limit", "20")
	voicesURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, voicesURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create voices request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("list voices through %s: %w", voicesURL.String(), err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read voices response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("list voices status = %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed struct {
		Data []struct {
			ID       string `json:"id"`
			Provider struct {
				Kind string `json:"kind"`
			} `json:"provider"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("decode voices response: %w", err)
	}
	for _, voice := range parsed.Data {
		if voice.ID != "" && voice.Provider.Kind == "volc-tenant" {
			return voice.ID, nil
		}
	}
	return "", fmt.Errorf("no volc voice returned by %s: %.512s", voicesURL.String(), string(body))
}
