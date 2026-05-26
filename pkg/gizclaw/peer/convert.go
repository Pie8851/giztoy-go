package peer

import (
	"encoding/json"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/adminservice"
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/apitypes"
)

func convertViaJSON[T any](in any) (T, error) {
	var out T
	data, err := json.Marshal(in)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return out, err
	}
	return out, nil
}

func toAdminRegistrationList(items []apitypes.Gear, hasNext bool, nextCursor *string) adminservice.RegistrationList {
	out := make([]apitypes.Registration, 0, len(items))
	for _, item := range items {
		out = append(out, toAdminRegistration(item))
	}
	return adminservice.RegistrationList{
		HasNext:    hasNext,
		Items:      out,
		NextCursor: nextCursor,
	}
}

func toAdminRegistration(gear apitypes.Gear) apitypes.Registration {
	return apitypes.Registration{
		ApprovedAt:     gear.ApprovedAt,
		AutoRegistered: gear.AutoRegistered,
		CreatedAt:      gear.CreatedAt,
		Device:         &gear.Device,
		PublicKey:      gear.PublicKey,
		Role:           apitypes.GearRole(gear.Role),
		Status:         apitypes.GearStatus(gear.Status),
		UpdatedAt:      gear.UpdatedAt,
	}
}

func toAdminRuntime(in apitypes.Runtime) apitypes.Runtime {
	return apitypes.Runtime{
		LastAddr:   in.LastAddr,
		LastSeenAt: in.LastSeenAt,
		Online:     in.Online,
		RxBytes:    in.RxBytes,
		TxBytes:    in.TxBytes,
	}
}

func toGearDeviceInfo(in apitypes.DeviceInfo) (apitypes.DeviceInfo, error) {
	return convertViaJSON[apitypes.DeviceInfo](in)
}

func toAdminDeviceInfo(in apitypes.DeviceInfo) (apitypes.DeviceInfo, error) {
	return convertViaJSON[apitypes.DeviceInfo](in)
}
