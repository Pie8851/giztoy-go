//go:build gizclaw_e2e

package adminui_test

import (
	. "github.com/GizClaw/gizclaw-go/test/gizclaw-e2e/ui/internal/harness"
	"testing"
)

func TestACLStories(t *testing.T) {
	RunAdminStories(t, adminACLStories())
}
