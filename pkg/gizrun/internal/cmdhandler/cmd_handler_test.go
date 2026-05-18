package cmdhandler

import (
	"context"
	"strings"
	"testing"
)

func TestCmdHandlerHandleAndMatch(t *testing.T) {
	mux := New()
	handler := HandleFunc(func(context.Context, []string, []string) error { return nil })

	if err := mux.Handle("admin/play", handler); err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	got, ok := mux.Lookup("admin/play")
	if !ok {
		t.Fatal("Lookup ok = false, want true")
	}
	if got == nil {
		t.Fatal("Lookup handler = nil")
	}
}

func TestCmdHandlerHandleRejectsNilHandler(t *testing.T) {
	mux := New()
	err := mux.Handle("admin/play", nil)
	if err == nil || !strings.Contains(err.Error(), "nil handler") {
		t.Fatalf("Handle err = %v, want nil handler error", err)
	}
}

func TestCmdHandlerHandleRejectsDuplicate(t *testing.T) {
	mux := New()
	handler := HandleFunc(func(context.Context, []string, []string) error { return nil })
	if err := mux.Handle("admin/play", handler); err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if err := mux.Handle("admin/play", handler); err == nil {
		t.Fatal("Handle duplicate err = nil, want error")
	}
}

func TestCmdHandlerLookupMiss(t *testing.T) {
	mux := New()
	if handler, ok := mux.Lookup("missing"); ok || handler != nil {
		t.Fatalf("Lookup missing = (%v, %v), want nil false", handler, ok)
	}
}
