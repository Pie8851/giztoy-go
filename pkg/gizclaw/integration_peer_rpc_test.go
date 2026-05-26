package gizclaw_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/apitypes"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw"
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/adminservice"
)

func TestIntegrationPeerRPCRefresh(t *testing.T) {
	ts := startTestServer(t)

	admin := newTestClient(t, ts)
	ensureAdminPeer(t, ts, admin, apitypes.DeviceInfo{Name: strPtr("admin")})

	device := newTestClient(t, ts)
	devicePublicKey := ensureGearInfo(t, device, apitypes.DeviceInfo{Name: strPtr("gear")})

	device.Device = apitypes.DeviceInfo{
		Hardware: &apitypes.HardwareInfo{
			Manufacturer: strPtr("Acme"),
			Model:        strPtr("M1"),
		},
		Sn: strPtr("sn-r1"),
	}

	result, err := waitForRefreshPeerSuccess(admin, devicePublicKey)
	if err != nil {
		t.Fatalf("RefreshPeer error: %v", err)
	}
	if result.Gear.Device.Hardware == nil || result.Gear.Device.Hardware.Manufacturer == nil || *result.Gear.Device.Hardware.Manufacturer != "Acme" {
		t.Fatalf("manufacturer = %+v", result.Gear.Device.Hardware)
	}
}

func TestIntegrationPeerRPCRefreshReportsOfflineWhenDeviceDisconnected(t *testing.T) {
	ts := startTestServer(t)

	admin := newTestClient(t, ts)
	ensureAdminPeer(t, ts, admin, apitypes.DeviceInfo{Name: strPtr("admin")})

	device := newTestClient(t, ts)
	devicePublicKey := ensureGearInfo(t, device, apitypes.DeviceInfo{Name: strPtr("gear")})
	if err := device.Close(); err != nil {
		t.Fatalf("device close error: %v", err)
	}

	err := waitForRefreshPeerFailure(admin, devicePublicKey)
	if err == nil {
		t.Fatal("RefreshPeer should fail when peer disconnects")
	}
	if !isOfflineRefreshError(err) {
		t.Fatalf("RefreshPeer error = %v, want offline-equivalent error", err)
	}
}

func waitForRefreshPeerSuccess(admin *gizclaw.Client, publicKey string) (adminservice.RefreshResult, error) {
	var lastResult adminservice.RefreshResult
	err := waitUntil(testReadyTimeout, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		result, err := refreshPeer(ctx, admin, publicKey)
		cancel()
		lastResult = result
		if err == nil &&
			result.Gear.Device.Hardware != nil &&
			result.Gear.Device.Hardware.Manufacturer != nil &&
			*result.Gear.Device.Hardware.Manufacturer == "Acme" {
			return nil
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("refresh gear did not return expected manufacturer, got %+v", lastResult.Gear.Device.Hardware)
	})
	if err != nil {
		return lastResult, err
	}
	return lastResult, nil
}

func waitForRefreshPeerFailure(admin *gizclaw.Client, publicKey string) error {
	var offlineErr error
	err := waitUntil(testReadyTimeout, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := refreshPeer(ctx, admin, publicKey)
		cancel()
		if isOfflineRefreshError(err) {
			offlineErr = err
			return nil
		}
		if err != nil {
			return err
		}
		return errors.New("refresh gear did not return expected failure")
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = refreshPeer(ctx, admin, publicKey)
	if isOfflineRefreshError(err) {
		return err
	}
	if offlineErr != nil {
		return offlineErr
	}
	return err
}

func isOfflineRefreshError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "DEVICE_OFFLINE") || strings.Contains(msg, "conn closed")
}
