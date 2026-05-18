package gizrun

import (
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/configfile"
	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/log/volclog"
	"github.com/goccy/go-yaml"
)

func TestVolcLogHandlerRequiresConfig(t *testing.T) {
	if _, err := volclog.NewHandler(volclog.Config{}); err == nil {
		t.Fatal("volclog.NewHandler err = nil, want validation error")
	}
}

func TestVolcLogConfigParser(t *testing.T) {
	registry := configfile.NewYamlParser()
	registry.Register("volc_log", configfile.ParseFunc(func(data []byte) (any, error) {
		var config volcLogConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, err
		}
		return config, nil
	}))
	value, err := registry.Parse([]byte(`
volc_log:
  enabled: true
  endpoint: tls-cn.example.com
  topic_id: topic
`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	config, ok := value.(configfile.ConfigFile)
	if !ok {
		t.Fatalf("Parse value type = %T, want configfile.ConfigFile", value)
	}
	section, ok := config.Config("volc_log")
	if !ok {
		t.Fatal("volc_log section not found")
	}
	logSection, ok := section.(volcLogConfig)
	if !ok {
		t.Fatalf("volc_log section type = %T, want volcLogConfig", section)
	}
	if !logSection.Enabled {
		t.Fatal("volc enabled = false, want true")
	}
	if logSection.Endpoint != "tls-cn.example.com" {
		t.Fatalf("volc endpoint = %q, want tls-cn.example.com", logSection.Endpoint)
	}
	if logSection.TopicID != "topic" {
		t.Fatalf("volc topic id = %q, want topic", logSection.TopicID)
	}
}
