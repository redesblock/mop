package main

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	filekeystore "github.com/redesblock/hop/core/keystore/file"
	"github.com/spf13/cobra"
	"path/filepath"
)

func (c *command) initExportPrivateCmd() error {
	cmd := &cobra.Command{
		Use:   "export-private [password]",
		Short: "export private key.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			keystore := filekeystore.New(filepath.Join(c.config.GetString(optionNameDataDir), "keys"))
			swarmPrivateKey, _, err := keystore.Key("swarm", args[0])
			if err != nil {
				return fmt.Errorf("swarm key: %w", err)
			}
			keyBytes := math.PaddedBigBytes(swarmPrivateKey.D, 32)
			fmt.Println("private key:", hex.EncodeToString(keyBytes))
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return c.config.BindPFlags(cmd.Flags())
		},
	}
	cmd.Flags().String(optionNameDataDir, filepath.Join(c.homeDir, ".hop"), "data directory")

	c.root.AddCommand(cmd)

	return nil
}
