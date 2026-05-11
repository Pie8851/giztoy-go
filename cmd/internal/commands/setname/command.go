package setnamecmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GizClaw/gizclaw-go/cmd/internal/client"
)

func NewCmd() *cobra.Command {
	var ctxName string
	cmd := &cobra.Command{
		Use:   "set-name <name>",
		Short: "Set current device name",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}
			if strings.TrimSpace(args[0]) == "" {
				return fmt.Errorf("device name must not be empty")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.ConnectFromContext(ctxName)
			if err != nil {
				return err
			}
			defer c.Close()
			info, err := client.SetName(context.Background(), c, args[0])
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(info)
		},
	}
	cmd.Flags().StringVar(&ctxName, "context", "", "context name (default: current)")
	return cmd
}
