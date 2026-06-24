//go:build gizclaw_e2e

// User story: As an admin operator, I can inspect a seeded peer across its
// info, edit, and CLI views.
package adminui_test

import (
	. "github.com/GizClaw/gizclaw-go/test/gizclaw-e2e/ui/internal/harness"
	"net/url"
	"testing"
)

func adminPeerDetailStories() []Story {
	return []Story{{
		Name: "111-admin-peer-detail",
		Run: func(_ testing.TB, page *Page) {
			page.GotoAdmin("/peers/" + url.PathEscape(page.Seed.DevicePublicKey))
			page.ExpectText("Seeded UI Device")
			page.ExpectText(page.Seed.DevicePublicKey)
			page.ExpectText("Configuration")
			page.ExpectText("e2e-client")
			page.ExpectText("Last Address")
			page.ExpectText("Online")

			page.ClickRole("tab", "Edit")
			page.ExpectText("Peer Actions")

			page.ClickRole("tab", "CLI")
			page.ExpectText("PeerConfig Resource Spec")
			page.ExpectText("gizclaw admin peers")
		},
	}}
}
