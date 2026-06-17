package gizclaw

import (
	"context"
	"errors"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/acl"
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/credential"
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/peer"
)

type Migrator struct {
	ACL         *acl.Server
	Credentials *credential.Server
	Peers       *peer.Server
}

func (m *Migrator) Migrate(ctx context.Context) error {
	if m == nil {
		return errors.New("gizclaw: nil migrator")
	}
	if m.ACL != nil {
		if err := m.ACL.Migration(ctx); err != nil {
			return err
		}
	}
	if m.Peers != nil {
		if err := m.Peers.Migration(ctx); err != nil {
			return err
		}
	}
	if m.Credentials != nil {
		if err := m.Credentials.Migration(ctx); err != nil {
			return err
		}
	}
	return nil
}
