package main

import (
	"fmt"
	"github.com/redesblock/hop/core/postage/postagecontract"
	"github.com/spf13/cobra"
	"io"
	"math/big"
	"net/http"
	"os"
)

func (c *command) initListStampCmd() error {
	cmd := &cobra.Command{
		Use:   "list-stamp",
		Short: "get all available stamps for this node",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			client := &http.Client{}

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:1635/stamps"), nil)
			if err != nil {
				return err
			}
			response, err := client.Do(req)
			if err != nil {
				return err
			}

			stdout := os.Stdout
			_, err = io.Copy(stdout, response.Body)
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return c.config.BindPFlags(cmd.Flags())
		},
	}

	c.setAllFlags(cmd)
	c.root.AddCommand(cmd)

	return nil
}

func (c *command) initShowStampCmd() error {
	cmd := &cobra.Command{
		Use:   "show-stamp id",
		Short: "get an individual postage batch status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			client := &http.Client{}
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:1635/stamps/%s", args[0]), nil)
			if err != nil {
				return err
			}
			response, err := client.Do(req)
			if err != nil {
				return err
			}

			stdout := os.Stdout
			_, err = io.Copy(stdout, response.Body)
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return c.config.BindPFlags(cmd.Flags())
		},
	}

	c.setAllFlags(cmd)
	c.root.AddCommand(cmd)

	return nil
}

func (c *command) initBuyStampCmd() error {
	cmd := &cobra.Command{
		Use:   "buy-stamp amount depth",
		Short: "buy a new postage batch.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if _, ok := new(big.Int).SetString(args[0], 10); !ok {
				return fmt.Errorf("invalid amount")
			}

			if depth, ok := new(big.Int).SetString(args[1], 10); !ok {
				return fmt.Errorf("invalid depth")
			} else if depth.Int64() <= int64(postagecontract.BucketDepth) {
				return fmt.Errorf("invalid depth")
			}

			client := &http.Client{}
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:1635/stamps/%s/%s", args[0], args[1]), nil)
			if err != nil {
				return err
			}
			response, err := client.Do(req)
			if err != nil {
				return err
			}

			stdout := os.Stdout
			_, err = io.Copy(stdout, response.Body)
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return c.config.BindPFlags(cmd.Flags())
		},
	}

	c.setAllFlags(cmd)
	c.root.AddCommand(cmd)

	return nil
}
