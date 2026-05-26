package peer

import (
	"testing"
	"time"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/apitypes"
	"github.com/GizClaw/gizclaw-go/pkg/giznet"
)

func TestConvertHelpers(t *testing.T) {
	now := time.Unix(1_700_600_000, 0).UTC()
	autoRegistered := true
	deviceName := "convert-device"
	publicKey := giznet.PublicKey{1}
	gear := apitypes.Gear{
		PublicKey:      publicKey.String(),
		Role:           apitypes.GearRoleServer,
		Status:         apitypes.GearStatusActive,
		AutoRegistered: &autoRegistered,
		CreatedAt:      now,
		UpdatedAt:      now,
		Configuration:  apitypes.Configuration{},
		Device: apitypes.DeviceInfo{
			Name: &deviceName,
		},
	}

	adminRegistrations := toAdminRegistrationList([]apitypes.Gear{gear}, false, nil)
	if len(adminRegistrations.Items) != 1 || adminRegistrations.Items[0].PublicKey != gear.PublicKey {
		t.Fatalf("toAdminRegistrationList = %+v", adminRegistrations)
	}
	if adminRegistrations.Items[0].Device == nil || adminRegistrations.Items[0].Device.Name == nil || *adminRegistrations.Items[0].Device.Name != deviceName {
		t.Fatalf("toAdminRegistrationList device = %+v", adminRegistrations.Items[0].Device)
	}

	convertedDevice, err := toGearDeviceInfo(gear.Device)
	if err != nil {
		t.Fatalf("toGearDeviceInfo error: %v", err)
	}
	if convertedDevice.Name == nil || *convertedDevice.Name != *gear.Device.Name {
		t.Fatalf("toGearDeviceInfo = %+v", convertedDevice)
	}

	adminDevice, err := toAdminDeviceInfo(apitypes.DeviceInfo{
		Name: gear.Device.Name,
		Sn:   gear.Device.Sn,
	})
	if err != nil {
		t.Fatalf("toAdminDeviceInfo error: %v", err)
	}
	if adminDevice.Name == nil || *adminDevice.Name != *gear.Device.Name {
		t.Fatalf("toAdminDeviceInfo = %+v", adminDevice)
	}

	rxBytes := uint64(123)
	txBytes := uint64(456)
	adminRuntime := toAdminRuntime(apitypes.Runtime{Online: true, LastSeenAt: now, RxBytes: &rxBytes, TxBytes: &txBytes})
	if !adminRuntime.Online || !adminRuntime.LastSeenAt.Equal(now) || adminRuntime.RxBytes == nil || *adminRuntime.RxBytes != rxBytes || adminRuntime.TxBytes == nil || *adminRuntime.TxBytes != txBytes {
		t.Fatalf("toAdminRuntime = %+v", adminRuntime)
	}
}
