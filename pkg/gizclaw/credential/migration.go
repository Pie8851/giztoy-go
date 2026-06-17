package credential

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/apitypes"
	"github.com/GizClaw/gizclaw-go/pkg/store/kv"
)

// Migration upgrades persisted credential records after schema changes.
func (s *Server) Migration(ctx context.Context) error {
	store, err := s.store()
	if err != nil {
		return err
	}
	var updates []kv.Entry
	for entry, err := range store.List(ctx, credentialsRoot) {
		if err != nil {
			return err
		}
		record, changed, err := migrateCredentialRecord(entry.Value)
		if err != nil {
			return fmt.Errorf("credential: migrate %s: %w", entry.Key.String(), err)
		}
		if !changed {
			continue
		}
		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("credential: encode migrated %s: %w", entry.Key.String(), err)
		}
		updates = append(updates,
			kv.Entry{Key: credentialKey(record.Name), Value: data},
			kv.Entry{Key: credentialByProviderKey(record.Provider, record.Name), Value: []byte{}},
		)
	}
	if len(updates) == 0 {
		return nil
	}
	return store.BatchSet(ctx, updates)
}

func migrateCredentialRecord(data []byte) (credentialRecord, bool, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return credentialRecord{}, false, err
	}
	changed := false
	if _, ok := raw["method"]; ok {
		delete(raw, "method")
		changed = true
	}
	bodyChanged := false
	if bodyRaw, ok := raw["body"]; ok {
		var body map[string]json.RawMessage
		if err := json.Unmarshal(bodyRaw, &body); err == nil && body != nil {
			if _, ok := body["method"]; ok {
				delete(body, "method")
				bodyChanged = true
			}
			if bodyChanged {
				nextBody, err := json.Marshal(body)
				if err != nil {
					return credentialRecord{}, false, err
				}
				raw["body"] = nextBody
				changed = true
			}
		}
	}
	next, err := json.Marshal(raw)
	if err != nil {
		return credentialRecord{}, false, err
	}
	var record credentialRecord
	if err := json.Unmarshal(next, &record); err != nil {
		return credentialRecord{}, false, err
	}
	if !apitypes.IsZeroCredentialBody(record.Body) {
		body := apitypes.CredentialBodyMap(record.Body)
		if _, ok := body["method"]; ok {
			delete(body, "method")
			encoded, err := json.Marshal(body)
			if err != nil {
				return credentialRecord{}, false, err
			}
			if err := record.Body.UnmarshalJSON(encoded); err != nil {
				return credentialRecord{}, false, err
			}
			changed = true
		}
	}
	return record, changed, nil
}
