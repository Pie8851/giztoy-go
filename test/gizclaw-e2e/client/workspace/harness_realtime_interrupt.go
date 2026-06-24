//go:build gizclaw_e2e

package workspace

import "context"

func (d *personaDriver) runRealtimeInterrupt(ctx context.Context) ([]interruptStats, error) {
	return d.runInterruptRounds(ctx, conversationMode{Realtime: true})
}
