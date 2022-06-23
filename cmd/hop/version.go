package main

import (
	"github.com/redesblock/hop/cmd/version"
	"github.com/spf13/cobra"
)

func (c *command) initVersionCmd() {
	c.root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version number",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(version.Version)
		},
	})
}
