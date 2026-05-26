// User story: As a Play UI user, I can start a WebRTC call, open the RPC data
// channel, and run peer RPC commands through the local proxy.
package ui_test

import (
	"testing"
)

func playAllActionsStories() []Story {
	return []Story{{
		Name: "202-play-all-actions",
		Run: func(_ testing.TB, page *Page) {
			page.GotoPlay("/")
			page.ClickRole("button", "Start Video Call")
			page.ExpectText("Connected")
			page.ClickRole("button", "Logs")
			page.ExpectText("webrtc.state")
			page.ExpectText("peer.ping")
			page.ExpectText("rpc.response")
			page.ClickRole("button", "Close RPC logs")
			page.ClickRole("button", "Get Info")
			page.ClickRole("button", "Logs")
			page.ExpectText("peer.info.get")
			page.ExpectText("Seeded UI Device")
			page.ClickRole("button", "Close RPC logs")
			page.ClickRole("button", "Get Runtime")
			page.ClickRole("button", "Logs")
			page.ExpectText("peer.runtime.get")
			page.ExpectText("online")
		},
	}}
}
