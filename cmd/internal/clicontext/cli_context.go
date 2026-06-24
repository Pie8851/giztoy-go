package clicontext

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GizClaw/gizclaw-go/cmd/internal/identity"
	"github.com/GizClaw/gizclaw-go/pkg/giznet"
	"github.com/goccy/go-yaml"
)

// ServerConfig holds the connection info for a remote server.
type ServerConfig struct {
	Address    string            `yaml:"address"`
	PublicKey  giznet.PublicKey  `yaml:"-"`
	CipherMode giznet.CipherMode `yaml:"cipher-mode,omitempty"`
}

// Config is the per-cli-context configuration stored in config.yaml.
type Config struct {
	Server ServerConfig `yaml:"server"`
}

// CLIContext represents a loaded CLI context directory.
type CLIContext struct {
	Name    string
	Dir     string
	Config  Config
	KeyPair *giznet.KeyPair
}

// Load reads a CLI context from its directory.
func Load(dir string) (*CLIContext, error) {
	name := filepath.Base(dir)

	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err != nil {
		return nil, fmt.Errorf("clicontext: read config: %w", err)
	}
	var keyCheck struct {
		Server map[string]any `yaml:"server"`
	}
	if err := yaml.Unmarshal(data, &keyCheck); err != nil {
		return nil, fmt.Errorf("clicontext: parse config: %w", err)
	}
	if _, ok := keyCheck.Server["public-key"]; ok {
		return nil, fmt.Errorf("clicontext: server.public-key is no longer supported; use server.private-key or server.identity-key")
	}
	var raw struct {
		Server struct {
			Address     string            `yaml:"address"`
			PrivateKey  *giznet.Key       `yaml:"private-key"`
			IdentityKey string            `yaml:"identity-key"`
			CipherMode  giznet.CipherMode `yaml:"cipher-mode"`
		} `yaml:"server"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("clicontext: parse config: %w", err)
	}
	serverPublicKey, err := resolveServerPublicKey(dir, raw.Server.PrivateKey, raw.Server.IdentityKey)
	if err != nil {
		return nil, err
	}
	if err := validateCipherMode(raw.Server.CipherMode); err != nil {
		return nil, err
	}
	cfg := Config{
		Server: ServerConfig{
			Address:    raw.Server.Address,
			PublicKey:  serverPublicKey,
			CipherMode: raw.Server.CipherMode,
		},
	}

	kp, err := identity.Load(filepath.Join(dir, "identity.key"))
	if err != nil {
		return nil, fmt.Errorf("clicontext: load identity: %w", err)
	}

	return &CLIContext{Name: name, Dir: dir, Config: cfg, KeyPair: kp}, nil
}

func resolveServerPublicKey(configDir string, privateKey *giznet.Key, identityPath string) (giznet.PublicKey, error) {
	configured := 0
	if privateKey != nil {
		configured++
	}
	if strings.TrimSpace(identityPath) != "" {
		configured++
	}
	if configured == 0 {
		return giznet.PublicKey{}, fmt.Errorf("clicontext: parse server public key: missing private-key or identity-key")
	}
	if configured > 1 {
		return giznet.PublicKey{}, fmt.Errorf("clicontext: configure only one of server.private-key or server.identity-key")
	}
	if privateKey != nil {
		if privateKey.IsZero() {
			return giznet.PublicKey{}, fmt.Errorf("clicontext: parse server private key: zero key")
		}
		kp, err := giznet.NewKeyPair(*privateKey)
		if err != nil {
			return giznet.PublicKey{}, fmt.Errorf("clicontext: derive server public key: %w", err)
		}
		return kp.Public, nil
	}
	identityPath = strings.TrimSpace(identityPath)
	if !filepath.IsAbs(identityPath) {
		identityPath = filepath.Join(configDir, identityPath)
	}
	kp, err := identity.Load(identityPath)
	if err != nil {
		return giznet.PublicKey{}, fmt.Errorf("clicontext: load server identity key: %w", err)
	}
	return kp.Public, nil
}

// ServerPublicKey parses and returns the server's public key.
func (c *CLIContext) ServerPublicKey() (giznet.PublicKey, error) {
	return c.Config.Server.PublicKey, nil
}
