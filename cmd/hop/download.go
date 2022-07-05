package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
)

func (c *command) initDownloadCmd() error {
	cmd := &cobra.Command{
		Use:   "download reference",
		Short: "get file or index document from a collection of files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			client := &http.Client{}
			fmt.Println(fmt.Sprintf("http://localhost:1633/hop/%s", args[0]))
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:1633/hop/%s", args[0]), nil)
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
