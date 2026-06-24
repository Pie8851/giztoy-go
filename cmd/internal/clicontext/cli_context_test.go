package clicontext

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/GizClaw/gizclaw-go/cmd/internal/identity"
	"github.com/GizClaw/gizclaw-go/pkg/giznet"
)

var (
	testServerPrivateKey  = testPrivateKeyText(0xab)
	testServerPrivateKey2 = testPrivateKeyText(0xcd)
	testServerPublicKey   = testPublicKeyText(0xab)
	testServerPublicKey2  = testPublicKeyText(0xcd)
)

func testPrivateKey(fill byte) giznet.Key {
	var key giznet.Key
	for i := range key {
		key[i] = fill
	}
	return key
}

func testPrivateKeyText(fill byte) string {
	return testPrivateKey(fill).String()
}

func testPublicKeyText(fill byte) string {
	kp, err := giznet.NewKeyPair(testPrivateKey(fill))
	if err != nil {
		panic(err)
	}
	return kp.Public.String()
}

func testKeyPair(t *testing.T, fill byte) *giznet.KeyPair {
	t.Helper()
	kp, err := giznet.NewKeyPair(testPrivateKey(fill))
	if err != nil {
		t.Fatalf("NewKeyPair error = %v", err)
	}
	return kp
}

func TestStoreCreateAndLoad(t *testing.T) {
	s := &Store{Root: t.TempDir()}

	if err := s.CreateWithOptions("local", "127.0.0.1:9820", CreateOptions{
		ServerPrivateKey: testServerPrivateKey,
		CipherMode:       giznet.CipherModeAES256GCM,
	}); err != nil {
		t.Fatalf("Create err=%v", err)
	}

	cliCtx, err := Load(filepath.Join(s.Root, "local"))
	if err != nil {
		t.Fatalf("Load err=%v", err)
	}
	if cliCtx.Name != "local" {
		t.Fatalf("Name=%q, want local", cliCtx.Name)
	}
	if cliCtx.Config.Server.Address != "127.0.0.1:9820" {
		t.Fatalf("Address=%q", cliCtx.Config.Server.Address)
	}
	if cliCtx.Config.Server.PublicKey.String() != testServerPublicKey {
		t.Fatalf("PublicKey=%q", cliCtx.Config.Server.PublicKey.String())
	}
	if cliCtx.Config.Server.CipherMode != giznet.CipherModeAES256GCM {
		t.Fatalf("CipherMode=%q, want %q", cliCtx.Config.Server.CipherMode, giznet.CipherModeAES256GCM)
	}
	if cliCtx.KeyPair == nil || cliCtx.KeyPair.Public.IsZero() {
		t.Fatal("KeyPair not loaded")
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Join(s.Root, "local", "config.yaml"))
		if err != nil {
			t.Fatalf("Stat config err=%v", err)
		}
		if mode := info.Mode().Perm(); mode != 0o600 {
			t.Fatalf("config mode=%o, want 600", mode)
		}
	}
}

func TestStoreCreateWithIdentityKey(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	serverKP := testKeyPair(t, 0x24)
	identityPath := filepath.Join(t.TempDir(), "server.identity")
	if err := os.WriteFile(identityPath, serverKP.Private[:], 0o600); err != nil {
		t.Fatal(err)
	}

	if err := s.CreateWithOptions("local", "127.0.0.1:9820", CreateOptions{
		ServerIdentityKey: identityPath,
		CipherMode:        giznet.CipherModeChaChaPoly,
	}); err != nil {
		t.Fatalf("CreateWithOptions err=%v", err)
	}

	cliCtx, err := Load(filepath.Join(s.Root, "local"))
	if err != nil {
		t.Fatalf("Load err=%v", err)
	}
	if cliCtx.Config.Server.PublicKey != serverKP.Public {
		t.Fatalf("PublicKey=%v, want %v", cliCtx.Config.Server.PublicKey, serverKP.Public)
	}
}

func TestStoreCreateRejectsInvalidCipherMode(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	err := s.CreateWithOptions("local", "127.0.0.1:9820", CreateOptions{
		ServerPrivateKey: testServerPrivateKey,
		CipherMode:       giznet.CipherMode("bad"),
	})
	if err == nil {
		t.Fatal("CreateWithOptions should reject invalid cipher mode")
	}
}

func TestStoreCreateDuplicate(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Create("dup", "addr", testServerPrivateKey); err != nil {
		t.Fatal(err)
	}
	if err := s.Create("dup", "addr", testServerPrivateKey); err == nil {
		t.Fatal("duplicate Create should fail")
	}
}

func TestStoreCreateRejectsInvalidName(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	for _, bad := range []string{"", "../escape", "a/b", ".", ".."} {
		if err := s.Create(bad, "addr", testServerPrivateKey); err == nil {
			t.Fatalf("Create(%q) should fail", bad)
		}
	}
}

func TestStoreCurrentAutoSet(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Create("first", "addr", testServerPrivateKey); err != nil {
		t.Fatal(err)
	}

	cliCtx, err := s.Current()
	if err != nil {
		t.Fatalf("Current err=%v", err)
	}
	if cliCtx == nil {
		t.Fatal("Current returned nil after first Create")
	}
	if cliCtx.Name != "first" {
		t.Fatalf("Current Name=%q, want first", cliCtx.Name)
	}
}

func TestStoreCurrentNone(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	cliCtx, err := s.Current()
	if err != nil {
		t.Fatalf("Current err=%v", err)
	}
	if cliCtx != nil {
		t.Fatal("Current should be nil when no context exists")
	}
}

func TestStoreUse(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Create("a", "addr-a", testServerPrivateKey); err != nil {
		t.Fatal(err)
	}
	if err := s.Create("b", "addr-b", testServerPrivateKey2); err != nil {
		t.Fatal(err)
	}

	if err := s.Use("b"); err != nil {
		t.Fatalf("Use err=%v", err)
	}

	cliCtx, err := s.Current()
	if err != nil {
		t.Fatalf("Current err=%v", err)
	}
	if cliCtx.Name != "b" {
		t.Fatalf("Current Name=%q, want b", cliCtx.Name)
	}
}

func TestStoreUseNonExistent(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Use("nope"); err == nil {
		t.Fatal("Use(nonexistent) should fail")
	}
}

func TestStoreDeleteNonCurrent(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Create("first", "addr", testServerPrivateKey); err != nil {
		t.Fatal(err)
	}
	if err := s.Create("second", "addr", testServerPrivateKey); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("second"); err != nil {
		t.Fatalf("Delete err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(s.Root, "second")); !os.IsNotExist(err) {
		t.Fatalf("deleted context still exists or stat failed: %v", err)
	}
	_, current, err := s.List()
	if err != nil {
		t.Fatalf("List err=%v", err)
	}
	if current != "first" {
		t.Fatalf("current=%q, want first", current)
	}
}

func TestStoreDeleteCurrentClearsCurrent(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Create("only", "addr", testServerPrivateKey); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("only"); err != nil {
		t.Fatalf("Delete current err=%v", err)
	}
	if current, err := s.Current(); err != nil || current != nil {
		t.Fatalf("Current after delete = %v, err=%v; want nil", current, err)
	}
	_, currentName, err := s.List()
	if err != nil {
		t.Fatalf("List err=%v", err)
	}
	if currentName != "" {
		t.Fatalf("List current=%q, want empty", currentName)
	}
}

func TestStoreDeleteRejectsInvalidOrMissing(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	for _, bad := range []string{"", "../escape", "a/b", ".", ".."} {
		if err := s.Delete(bad); err == nil {
			t.Fatalf("Delete(%q) should fail", bad)
		}
	}
	if err := s.Delete("missing"); err == nil {
		t.Fatal("Delete(missing) should fail")
	}
}

func TestStoreList(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Create("beta", "addr", testServerPrivateKey); err != nil {
		t.Fatal(err)
	}
	if err := s.Create("alpha", "addr", testServerPrivateKey); err != nil {
		t.Fatal(err)
	}

	names, current, err := s.List()
	if err != nil {
		t.Fatalf("List err=%v", err)
	}
	if len(names) != 2 || names[0] != "alpha" || names[1] != "beta" {
		t.Fatalf("List names=%v, want [alpha beta]", names)
	}
	if current != "beta" {
		t.Fatalf("List current=%q, want beta", current)
	}
}

func TestStoreListEmpty(t *testing.T) {
	s := &Store{Root: filepath.Join(t.TempDir(), "nonexistent")}
	names, current, err := s.List()
	if err != nil {
		t.Fatalf("List err=%v", err)
	}
	if len(names) != 0 {
		t.Fatalf("List names=%v, want empty", names)
	}
	if current != "" {
		t.Fatalf("List current=%q, want empty", current)
	}
}

func TestServerPublicKey(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	pk := testServerPublicKey
	if err := s.Create("spk", "addr", testServerPrivateKey); err != nil {
		t.Fatal(err)
	}
	cliCtx, err := Load(filepath.Join(s.Root, "spk"))
	if err != nil {
		t.Fatal(err)
	}
	key, err := cliCtx.ServerPublicKey()
	if err != nil {
		t.Fatalf("ServerPublicKey err=%v", err)
	}
	if key.String() != pk {
		t.Fatalf("ServerPublicKey=%q, want %q", key.String(), pk)
	}
}

func TestServerPrivateKeyInvalid(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Create("badpk", "addr", "not-a-key"); err == nil {
		t.Fatal("Create(invalid private key) should fail")
	}
}

func TestLoadMissingConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatal("Load(no config) should fail")
	}
}

func TestLoadBadYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(":::"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatal("Load(bad yaml) should fail")
	}
}

func TestLoadRejectsInvalidCipherMode(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
server:
  address: 127.0.0.1:9820
  private-key: `+testServerPrivateKey+`
  cipher-mode: bad
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := identity.LoadOrGenerate(filepath.Join(dir, "identity.key")); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(dir); err == nil {
		t.Fatal("Load should reject invalid cipher mode")
	}
}

func TestLoadDerivesServerPublicKeyFromPrivateKey(t *testing.T) {
	dir := t.TempDir()
	serverKP := testKeyPair(t, 0x21)
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
server:
  address: 127.0.0.1:9820
  private-key: `+serverKP.Private.String()+`
  cipher-mode: chacha_poly
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := identity.LoadOrGenerate(filepath.Join(dir, "identity.key")); err != nil {
		t.Fatal(err)
	}

	cliCtx, err := Load(dir)
	if err != nil {
		t.Fatalf("Load err=%v", err)
	}
	if cliCtx.Config.Server.PublicKey != serverKP.Public {
		t.Fatalf("PublicKey=%v, want %v", cliCtx.Config.Server.PublicKey, serverKP.Public)
	}
}

func TestLoadDerivesServerPublicKeyFromIdentityKey(t *testing.T) {
	dir := t.TempDir()
	serverKP := testKeyPair(t, 0x22)
	if err := os.WriteFile(filepath.Join(dir, "server.identity"), serverKP.Private[:], 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
server:
  address: 127.0.0.1:9820
  identity-key: server.identity
  cipher-mode: chacha_poly
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := identity.LoadOrGenerate(filepath.Join(dir, "identity.key")); err != nil {
		t.Fatal(err)
	}

	cliCtx, err := Load(dir)
	if err != nil {
		t.Fatalf("Load err=%v", err)
	}
	if cliCtx.Config.Server.PublicKey != serverKP.Public {
		t.Fatalf("PublicKey=%v, want %v", cliCtx.Config.Server.PublicKey, serverKP.Public)
	}
}

func TestLoadRejectsConflictingServerKeySources(t *testing.T) {
	dir := t.TempDir()
	serverKP := testKeyPair(t, 0x23)
	if err := os.WriteFile(filepath.Join(dir, "server.identity"), serverKP.Private[:], 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
server:
  address: 127.0.0.1:9820
  private-key: `+serverKP.Private.String()+`
  identity-key: server.identity
  cipher-mode: chacha_poly
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := identity.LoadOrGenerate(filepath.Join(dir, "identity.key")); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "configure only one") {
		t.Fatalf("Load conflict err=%v", err)
	}
}

func TestLoadRejectsLegacyServerPublicKey(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
server:
  address: 127.0.0.1:9820
  public-key: `+testServerPublicKey+`
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := identity.LoadOrGenerate(filepath.Join(dir, "identity.key")); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "server.public-key is no longer supported") {
		t.Fatalf("Load legacy public key err=%v", err)
	}
}

func TestStoreLoadByName(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Create("myctx", "127.0.0.1:9820", testServerPrivateKey); err != nil {
		t.Fatal(err)
	}

	cliCtx, err := s.LoadByName("myctx")
	if err != nil {
		t.Fatalf("LoadByName err=%v", err)
	}
	if cliCtx.Name != "myctx" {
		t.Fatalf("Name=%q, want myctx", cliCtx.Name)
	}
	if cliCtx.Config.Server.Address != "127.0.0.1:9820" {
		t.Fatalf("Address=%q", cliCtx.Config.Server.Address)
	}
}

func TestStoreLoadByNameRejectsTraversal(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	for _, bad := range []string{"", "../escape", "a/b", ".", ".."} {
		if _, err := s.LoadByName(bad); err == nil {
			t.Fatalf("LoadByName(%q) should fail", bad)
		}
	}
}

func TestStoreLoadByNameNotExist(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if _, err := s.LoadByName("nope"); err == nil {
		t.Fatal("LoadByName(nonexistent) should fail")
	}
}

func TestStoreSymlinkIsRelative(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Create("myctx", "addr", testServerPrivateKey); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(s.Root, currentLink)
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("Readlink err=%v", err)
	}
	if filepath.IsAbs(target) {
		t.Fatalf("symlink target is absolute: %q", target)
	}
	if target != "myctx" {
		t.Fatalf("symlink target=%q, want myctx", target)
	}
}

func TestStoreListAbsoluteCurrentSymlink(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Create("alpha", "addr", testServerPrivateKey); err != nil {
		t.Fatal(err)
	}
	if err := s.Create("beta", "addr", testServerPrivateKey); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(s.Root, currentLink)
	if err := os.Remove(link); err != nil {
		t.Fatalf("Remove current symlink error=%v", err)
	}
	if err := os.Symlink(filepath.Join(s.Root, "alpha"), link); err != nil {
		t.Fatalf("Symlink error=%v", err)
	}

	names, current, err := s.List()
	if err != nil {
		t.Fatalf("List err=%v", err)
	}
	if len(names) != 2 || names[0] != "alpha" || names[1] != "beta" {
		t.Fatalf("List names=%v", names)
	}
	if current != "alpha" {
		t.Fatalf("List current=%q, want alpha", current)
	}
}
