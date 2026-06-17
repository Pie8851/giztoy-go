package transformers

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/genx"
)

func TestRunTTSTransformPreservesInputStreamIDAndNormalizesAudio(t *testing.T) {
	input := &testStream{chunks: []*genx.MessageChunk{
		{
			Role: genx.RoleUser,
			Name: "gear",
			Part: genx.Text("你好，世界。"),
			Ctrl: &genx.StreamCtrl{StreamID: "input-stream"},
		},
		{
			Role: genx.RoleUser,
			Name: "gear",
			Part: genx.Text(""),
			Ctrl: &genx.StreamCtrl{StreamID: "input-stream", EndOfStream: true},
		},
	}, doneErr: io.EOF}

	output := newBufferStream(8)
	var texts []string
	runTTSTransform(context.Background(), input, output, "audio/mpeg", func(_ context.Context, text string, meta ttsChunkMeta, mimeType string, out *bufferStream) error {
		texts = append(texts, text)
		data := append(fakeID3Header(), []byte("audio:"+text)...)
		return pushTTSAudioChunk(out, meta, mimeType, data)
	})

	if want := []string{"你好，", "世界。"}; !reflect.DeepEqual(texts, want) {
		t.Fatalf("synthesized texts = %#v, want %#v", texts, want)
	}

	chunks := collectTransformerChunks(t, output)
	if len(chunks) != 3 {
		t.Fatalf("got %d output chunks, want 3", len(chunks))
	}
	for i, chunk := range chunks[:2] {
		if chunk.Ctrl == nil || chunk.Ctrl.StreamID != "input-stream" {
			t.Fatalf("audio chunk %d StreamID = %#v, want input-stream", i, chunk.Ctrl)
		}
		blob, ok := chunk.Part.(*genx.Blob)
		if !ok {
			t.Fatalf("audio chunk %d part = %T, want *genx.Blob", i, chunk.Part)
		}
		if bytes.Contains(blob.Data, []byte("ID3")) {
			t.Fatalf("audio chunk %d still contains ID3 tag", i)
		}
		if chunk.Role != genx.RoleUser || chunk.Name != "gear" {
			t.Fatalf("audio chunk %d meta = role %q name %q", i, chunk.Role, chunk.Name)
		}
	}
	eos := chunks[2]
	if eos.Ctrl == nil || !eos.Ctrl.EndOfStream || eos.Ctrl.StreamID != "input-stream" {
		t.Fatalf("eos ctrl = %#v, want input-stream eos", eos.Ctrl)
	}
}

func TestRunTTSTransformDoesNotCreateStreamID(t *testing.T) {
	input := &testStream{chunks: []*genx.MessageChunk{
		{Part: genx.Text("hello.")},
		genx.NewTextEndOfStream(),
	}, doneErr: io.EOF}

	output := newBufferStream(4)
	runTTSTransform(context.Background(), input, output, "audio/mpeg", func(_ context.Context, _ string, meta ttsChunkMeta, mimeType string, out *bufferStream) error {
		return pushTTSAudioChunk(out, meta, mimeType, []byte("audio"))
	})

	chunks := collectTransformerChunks(t, output)
	if len(chunks) != 2 {
		t.Fatalf("got %d output chunks, want 2", len(chunks))
	}
	if chunks[0].Ctrl != nil {
		t.Fatalf("audio chunk ctrl = %#v, want nil", chunks[0].Ctrl)
	}
	if chunks[1].Ctrl == nil || !chunks[1].Ctrl.EndOfStream || chunks[1].Ctrl.StreamID != "" {
		t.Fatalf("eos ctrl = %#v, want eos without stream id", chunks[1].Ctrl)
	}
}

func collectTransformerChunks(t *testing.T, stream genx.Stream) []*genx.MessageChunk {
	t.Helper()
	var chunks []*genx.MessageChunk
	for {
		chunk, err := stream.Next()
		if err != nil {
			if err == io.EOF || err == genx.ErrDone {
				return chunks
			}
			t.Fatalf("Next() error = %v", err)
		}
		chunks = append(chunks, chunk)
	}
}
