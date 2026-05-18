package gizrun

import (
	"flag"
	"log/slog"
	"os"

	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/configfile"
	"github.com/GizClaw/gizclaw-go/pkg/gizrun/internal/log/volclog"
	"github.com/goccy/go-yaml"
)

type volcLogConfig struct {
	Enabled bool `yaml:"enabled"`

	Endpoint        string `yaml:"endpoint"`
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	SecurityToken   string `yaml:"security_token"`
	APIKey          string `yaml:"api_key"`

	TopicID   string `yaml:"topic_id"`
	Source    string `yaml:"source"`
	FileName  string `yaml:"file_name"`
	ShardHash string `yaml:"shard_hash"`

	Level            slog.Leveler
	AddSource        bool `yaml:"add_source"`
	EnableNanosecond bool `yaml:"enable_nanosecond"`
}

func (c volcLogConfig) withLevel(level slog.Leveler) volcLogConfig {
	c.Level = level
	return c
}

func (c volcLogConfig) expandEnv() volcLogConfig {
	return volcLogConfig{
		Enabled:          c.Enabled,
		Endpoint:         os.ExpandEnv(c.Endpoint),
		Region:           os.ExpandEnv(c.Region),
		AccessKeyID:      os.ExpandEnv(c.AccessKeyID),
		AccessKeySecret:  os.ExpandEnv(c.AccessKeySecret),
		SecurityToken:    os.ExpandEnv(c.SecurityToken),
		APIKey:           os.ExpandEnv(c.APIKey),
		TopicID:          os.ExpandEnv(c.TopicID),
		Source:           os.ExpandEnv(c.Source),
		FileName:         os.ExpandEnv(c.FileName),
		ShardHash:        os.ExpandEnv(c.ShardHash),
		Level:            c.Level,
		AddSource:        c.AddSource,
		EnableNanosecond: c.EnableNanosecond,
	}
}

func init() {
	InitAt(initSeqFlags, func() error {
		flag.BoolVar(&runCtx.volcLog.Enabled, "gizrun-volc-log-enabled", runCtx.volcLog.Enabled, "enable Volcengine TLS log sink")
		flag.StringVar(&runCtx.volcLog.Endpoint, "gizrun-volc-log-endpoint", runCtx.volcLog.Endpoint, "Volcengine TLS endpoint")
		flag.StringVar(&runCtx.volcLog.Region, "gizrun-volc-log-region", runCtx.volcLog.Region, "Volcengine TLS region")
		flag.StringVar(&runCtx.volcLog.AccessKeyID, "gizrun-volc-log-access-key-id", runCtx.volcLog.AccessKeyID, "Volcengine access key id")
		flag.StringVar(&runCtx.volcLog.AccessKeySecret, "gizrun-volc-log-access-key-secret", runCtx.volcLog.AccessKeySecret, "Volcengine access key secret")
		flag.StringVar(&runCtx.volcLog.SecurityToken, "gizrun-volc-log-security-token", runCtx.volcLog.SecurityToken, "Volcengine security token")
		flag.StringVar(&runCtx.volcLog.APIKey, "gizrun-volc-log-api-key", runCtx.volcLog.APIKey, "Volcengine API key")
		flag.StringVar(&runCtx.volcLog.TopicID, "gizrun-volc-log-topic-id", runCtx.volcLog.TopicID, "Volcengine TLS topic id")
		flag.StringVar(&runCtx.volcLog.Source, "gizrun-volc-log-source", runCtx.volcLog.Source, "Volcengine TLS log source")
		flag.StringVar(&runCtx.volcLog.FileName, "gizrun-volc-log-file-name", runCtx.volcLog.FileName, "Volcengine TLS log file name")
		flag.StringVar(&runCtx.volcLog.ShardHash, "gizrun-volc-log-shard-hash", runCtx.volcLog.ShardHash, "Volcengine TLS shard hash")
		flag.BoolVar(&runCtx.volcLog.AddSource, "gizrun-volc-log-add-source", runCtx.volcLog.AddSource, "enable Volcengine TLS add source")
		flag.BoolVar(&runCtx.volcLog.EnableNanosecond, "gizrun-volc-log-enable-nanosecond", runCtx.volcLog.EnableNanosecond, "enable Volcengine TLS nanosecond timestamp")
		RegisterConfigParser("volc_log", configfile.ParseFunc(func(data []byte) (any, error) {
			var config volcLogConfig
			if err := yaml.Unmarshal(data, &config); err != nil {
				return nil, err
			}
			return config, nil
		}))
		return nil
	})
	PostInitAt(postInitSeqRuntime, func(ctx *RunContext) error {
		var volcConfig volcLogConfig
		if section, ok := ctx.Config("volc_log"); ok {
			value, ok := section.(volcLogConfig)
			if !ok || !value.Enabled {
				return nil
			}
			volcConfig = value.withLevel(&ctx.logLevel).expandEnv()
		} else {
			if !ctx.volcLog.Enabled {
				return nil
			}
			volcConfig = ctx.volcLog.withLevel(&ctx.logLevel).expandEnv()
		}
		handler, err := volclog.NewHandler(volclog.Config{
			Endpoint:         volcConfig.Endpoint,
			Region:           volcConfig.Region,
			AccessKeyID:      volcConfig.AccessKeyID,
			AccessKeySecret:  volcConfig.AccessKeySecret,
			SecurityToken:    volcConfig.SecurityToken,
			APIKey:           volcConfig.APIKey,
			TopicID:          volcConfig.TopicID,
			Source:           volcConfig.Source,
			FileName:         volcConfig.FileName,
			ShardHash:        volcConfig.ShardHash,
			Level:            volcConfig.Level,
			AddSource:        volcConfig.AddSource,
			EnableNanosecond: volcConfig.EnableNanosecond,
		})
		if err != nil {
			return err
		}
		ctx.volcLogHandler = handler
		ctx.logHandlers = append(ctx.logHandlers, handler)
		return nil
	})
	ExitAt(exitSeqRuntime, func(ctx *RunContext) error {
		if ctx.volcLogHandler != nil {
			_ = ctx.volcLogHandler.Close()
			ctx.volcLogHandler = nil
		}
		return nil
	})
}
