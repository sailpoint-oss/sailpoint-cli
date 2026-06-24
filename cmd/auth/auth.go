package auth

import (
	"github.com/spf13/cobra"
)

func NewAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication for the active environment",
		Long:  "\nManage authentication sessions for the active CLI environment.\n\nUse 'sail auth login' to explicitly authenticate,\n'sail auth status' to check your current session,\nand 'sail auth logout' to clear cached tokens.\n",
		Example: `  sail auth login
  sail auth status
  sail auth logout
  printf '%s' "$SAIL_CLIENT_SECRET" | sail auth pat set --client-id "$SAIL_CLIENT_ID" --client-secret-stdin`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newLoginCommand(),
		newLogoutCommand(),
		newPATCommand(),
		newStatusCommand(),
	)

	return cmd
}
