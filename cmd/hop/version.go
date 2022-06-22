package main

import (
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

var (
	version    = "0.5.0" // manually set semantic version number
	commitHash string    // automatically set git commit hash
	commitTime string    // automatically set git commit time

	Version = func() string {
		if commitHash != "" {
			return version + "-" + commitHash
		}
		return version + "-dev"
	}()

	// CommitTime returns the time of the commit from which this code was derived.
	// If it's not set (in the case of running the code directly without compilation)
	// then the current time will be returned.
	CommitTime = func() string {
		if commitTime == "" {
			commitTime = strconv.Itoa(int(time.Now().Unix()))
		}
		return commitTime
	}
)

func (c *command) initVersionCmd() {
	c.root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version number",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(Version)
		},
	})
}
