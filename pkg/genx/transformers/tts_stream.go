package transformers

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/GizClaw/gizclaw-go/pkg/genx"
)

const defaultTTSSegmentMaxRunes = 256

type ttsChunkMeta struct {
	Role     genx.Role
	Name     string
	StreamID string
}

type ttsSynthesizer func(context.Context, string, ttsChunkMeta, string, *bufferStream) error

func runTTSTransform(ctx context.Context, input genx.Stream, output *bufferStream, mimeType string, synthesize ttsSynthesizer) {
	defer output.Close()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	segmenter := newTTSSentenceSegmenter(defaultTTSSegmentMaxRunes)
	var meta ttsChunkMeta

	flush := func(all bool) error {
		for _, segment := range segmenter.Segments(all) {
			if strings.TrimSpace(segment) == "" {
				continue
			}
			debugTTSSegment(meta, segment, all)
			if err := synthesize(ctx, segment, meta, mimeType, output); err != nil {
				return err
			}
		}
		return nil
	}

	for {
		chunk, err := input.Next()
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, genx.ErrDone) {
				output.CloseWithError(err)
				return
			}
			if err := flush(true); err != nil {
				output.CloseWithError(err)
			}
			return
		}
		if chunk == nil {
			continue
		}

		updateTTSMeta(&meta, chunk)

		text, isText := chunk.Part.(genx.Text)
		if isText {
			if text != "" {
				segmenter.WriteString(string(text))
				if err := flush(false); err != nil {
					output.CloseWithError(err)
					return
				}
			}
			if chunk.IsEndOfStream() {
				if err := flush(true); err != nil {
					output.CloseWithError(err)
					return
				}
				if err := output.Push(newTTSEOSChunk(meta, mimeType)); err != nil {
					return
				}
				segmenter.Reset()
				meta = ttsChunkMeta{}
			}
			continue
		}

		if err := flush(true); err != nil {
			output.CloseWithError(err)
			return
		}
		if err := output.Push(chunk); err != nil {
			return
		}
	}
}

func updateTTSMeta(meta *ttsChunkMeta, chunk *genx.MessageChunk) {
	meta.Role = chunk.Role
	meta.Name = chunk.Name
	if chunk.Ctrl != nil && chunk.Ctrl.StreamID != "" {
		meta.StreamID = chunk.Ctrl.StreamID
	}
}

func pushTTSAudioChunk(output *bufferStream, meta ttsChunkMeta, mimeType string, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	chunk := &genx.MessageChunk{
		Role: meta.Role,
		Name: meta.Name,
		Part: &genx.Blob{
			MIMEType: mimeType,
			Data:     normalizeTTSAudio(mimeType, data),
		},
	}
	if meta.StreamID != "" {
		chunk.Ctrl = &genx.StreamCtrl{StreamID: meta.StreamID}
	}
	return output.Push(chunk)
}

func newTTSEOSChunk(meta ttsChunkMeta, mimeType string) *genx.MessageChunk {
	return &genx.MessageChunk{
		Role: meta.Role,
		Name: meta.Name,
		Part: &genx.Blob{MIMEType: mimeType},
		Ctrl: &genx.StreamCtrl{StreamID: meta.StreamID, EndOfStream: true},
	}
}
