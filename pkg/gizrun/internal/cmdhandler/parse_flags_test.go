package cmdhandler

import (
	"flag"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	fs := flag.NewFlagSet("cmdhandler-test", flag.ContinueOnError)
	var config string
	var debug bool
	fs.StringVar(&config, "config", "", "config file")
	fs.BoolVar(&debug, "debug", false, "debug mode")

	args, flags, err := Parse(fs, []string{
		"-config", "app.yaml",
		"chat",
		"-model=gpt",
		"--debug",
		"hello",
		"--",
		"-literal",
	})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if config != "app.yaml" {
		t.Fatalf("config = %q, want app.yaml", config)
	}
	if !debug {
		t.Fatal("debug = false, want true")
	}
	if !reflect.DeepEqual(args, []string{"chat", "hello", "-literal"}) {
		t.Fatalf("args = %#v, want chat hello literal", args)
	}
	if !reflect.DeepEqual(flags, []string{"-model=gpt"}) {
		t.Fatalf("flags = %#v, want unknown model flag", flags)
	}
}
