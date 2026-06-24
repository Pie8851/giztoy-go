//go:build gizclaw_e2e

package rpc_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpcapi"
)

func TestServerFirmwareRPC(t *testing.T) {
	env := newServerResourceHarness(t)

	list, err := env.peer.ListFirmwares(env.ctx, "firmware.list.seeded", rpcapi.FirmwareListRequest{})
	if err != nil {
		t.Fatalf("firmware.list seeded: %v", err)
	}
	if !hasFirmware(list.Items, "seed-firmware") {
		t.Fatalf("firmware.list missing seed-firmware: %#v", list.Items)
	}

	got, err := env.peer.GetFirmware(env.ctx, "firmware.get.seeded", rpcapi.FirmwareGetRequest{FirmwareId: "seed-firmware"})
	if err != nil {
		t.Fatalf("firmware.get seeded: %v", err)
	}
	if got.Name != "seed-firmware" {
		t.Fatalf("firmware.get name = %q", got.Name)
	}
	if got.Slots.Stable.Version == nil || *got.Slots.Stable.Version != "1.0.0" {
		t.Fatalf("firmware stable version = %#v", got.Slots.Stable.Version)
	}
	if got.Slots.Stable.Artifacts == nil || len(*got.Slots.Stable.Artifacts) != 1 || (*got.Slots.Stable.Artifacts)[0].Path == nil {
		t.Fatalf("firmware stable artifacts = %#v", got.Slots.Stable.Artifacts)
	}

	var out bytes.Buffer
	download, err := env.peer.DownloadFirmware(env.ctx, "firmware.download.seeded", rpcapi.FirmwareDownloadRequest{
		FirmwareId:   "seed-firmware",
		Channel:      rpcapi.FirmwareChannelNameStable,
		ArtifactName: "main",
	}, &out)
	if err != nil {
		t.Fatalf("firmware.download seeded: %v", err)
	}
	if out.String() != "firmware-payload" {
		t.Fatalf("firmware.download payload = %q", out.String())
	}
	if download.Bytes != int64(len("firmware-payload")) {
		t.Fatalf("firmware.download bytes = %d", download.Bytes)
	}
	if download.Metadata.FirmwareId != "seed-firmware" || download.Metadata.Channel != rpcapi.FirmwareChannelNameStable {
		t.Fatalf("firmware.download metadata = %#v", download.Metadata)
	}
	if download.Metadata.Artifact.Name != "main" || download.Metadata.Artifact.Kind != rpcapi.FirmwareArtifactKindApp {
		t.Fatalf("firmware.download artifact = %#v", download.Metadata.Artifact)
	}

	denied := env.h.ConnectClientFromContext("peer-denied")
	defer denied.Close()
	deniedList, err := denied.ListFirmwares(env.ctx, "firmware.list.denied", rpcapi.FirmwareListRequest{})
	if err != nil {
		t.Fatalf("firmware.list denied peer: %v", err)
	}
	if len(deniedList.Items) != 0 {
		t.Fatalf("firmware.list denied items = %#v", deniedList.Items)
	}
	if _, err := denied.GetFirmware(env.ctx, "firmware.get.denied", rpcapi.FirmwareGetRequest{FirmwareId: "seed-firmware"}); err == nil || !strings.Contains(err.Error(), "acl: denied") {
		t.Fatalf("firmware.get denied error = %v", err)
	}
}

func hasFirmware(items []rpcapi.Firmware, name string) bool {
	for _, item := range items {
		if item.Name == name {
			return true
		}
	}
	return false
}
