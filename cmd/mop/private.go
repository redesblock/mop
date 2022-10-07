package main

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	filekeystore "github.com/redesblock/mop/core/keystore/file"
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
			flockPrivateKey, _, err := keystore.Key("flock", args[0])
			if err != nil {
				return fmt.Errorf("flock key: %w", err)
			}
			keyBytes := math.PaddedBigBytes(flockPrivateKey.D, 32)
			fmt.Println("private key:", hex.EncodeToString(keyBytes))
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return c.config.BindPFlags(cmd.Flags())
		},
	}
	cmd.Flags().String(optionNameDataDir, filepath.Join(c.homeDir, ".mop"), "data directory")

	c.root.AddCommand(cmd)

	return nil
}
