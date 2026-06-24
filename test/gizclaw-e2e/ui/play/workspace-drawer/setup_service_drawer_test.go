//go:build gizclaw_e2e

package playui_test

import (
	. "github.com/GizClaw/gizclaw-go/test/gizclaw-e2e/ui/internal/harness"
	"testing"
)

func TestSetupServicePlayWorkspaceDrawer(t *testing.T) {
	RunPlayStories(t, []Story{{
		Name: "setup-service-play-workspace-drawer",
		Run: func(_ testing.TB, page *Page) {
			page.GotoPlay("/")
			page.ExpectText("OpenAI Gateway")
			page.ClickRole("button", "Workspace")
			page.ExpectText("Realtime Chat")
			page.ExpectText("Push To Talk")
			page.ExpectText("History")
			page.ExpectText("Memory")
			page.ExpectText("Recall")
			page.ExpectText("flowcraft-voice")
		},
	}})
}
