package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cobra.EnableCommandSorting = false
}

type command struct {
	root           *cobra.Command
	config         *viper.Viper
	passwordReader passwordReader
	cfgFile        string
	homeDir        string
}

type option func(*command)

func newCommand(opts ...option) (c *command, err error) {
	c = &command{
		root: &cobra.Command{
			Use:           "hop",
			SilenceErrors: true,
			SilenceUsage:  true,
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				return c.initConfig()
			},
		},
	}

	for _, o := range opts {
		o(c)
	}
	if c.passwordReader == nil {
		c.passwordReader = new(stdInPasswordReader)
	}

	// Find home directory.
	if err := c.setHomeDir(); err != nil {
		return nil, err
	}

	c.initGlobalFlags()

	if err := c.initStartCmd(); err != nil {
		return nil, err
	}

	c.initVersionCmd()
	return c, nil
}

func (c *command) Execute() (err error) {
	return c.root.Execute()
}

// Execute parses command line arguments and runs appropriate functions.
func Execute() (err error) {
	c, err := newCommand()
	if err != nil {
		return err
	}
	return c.Execute()
}

func (c *command) initGlobalFlags() {
	globalFlags := c.root.PersistentFlags()
	globalFlags.StringVar(&c.cfgFile, "config", "", "config file (default is $HOME/.hop.yaml)")
}

func (c *command) initConfig() (err error) {
	config := viper.New()
	configName := ".hop"
	if c.cfgFile != "" {
		// Use config file from the flag.
		config.SetConfigFile(c.cfgFile)
	} else {
		// Search config in home directory with name ".hop" (without extension).
		config.AddConfigPath(c.homeDir)
		config.SetConfigName(configName)
	}

	// Environment
	config.SetEnvPrefix("hop")
	config.AutomaticEnv() // read in environment variables that match
	config.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	if c.homeDir != "" && c.cfgFile == "" {
		c.cfgFile = filepath.Join(c.homeDir, configName+".yaml")
	}

	// If a config file is found, read it in.
	if err := config.ReadInConfig(); err != nil {
		var e viper.ConfigFileNotFoundError
		if !errors.As(err, &e) {
			return err
		}
	}
	c.config = config
	return nil
}

func (c *command) setHomeDir() (err error) {
	if c.homeDir != "" {
		return
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	c.homeDir = dir
	return nil
}