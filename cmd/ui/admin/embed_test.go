package adminui

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestFS(t *testing.T) {
	for _, name := range []string{"index.html", "app.js", "app.css"} {
		data, err := fs.ReadFile(FS(), name)
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", name, err)
		}
		if len(data) == 0 {
			t.Fatalf("%s is empty", name)
		}
	}
}

func TestSubFSPanicsWhenSubFails(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("subFS did not panic")
		}
	}()
	subFS(fstest.MapFS{}, "../dist")
}

func TestSubFS(t *testing.T) {
	files := subFS(fstest.MapFS{
		"dist/index.html": {Data: []byte("ok")},
	}, "dist")
	data, err := fs.ReadFile(files, "index.html")
	if err != nil {
		t.Fatalf("ReadFile(index.html) error = %v", err)
	}
	if string(data) != "ok" {
		t.Fatalf("index.html = %q, want ok", data)
	}
}
