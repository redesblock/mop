package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/node"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	optionNameDataDir                    = "data-dir"
	optionNameCacheCapacity              = "cache-capacity"
	optionNameMemCacheCapacity           = "mem-cache-capacity"
	optionNameDBOpenFilesLimit           = "db-open-files-limit"
	optionNameDBBlockCacheCapacity       = "db-block-cache-capacity"
	optionNameDBWriteBufferSize          = "db-write-buffer-size"
	optionNameDBDisableSeeksCompaction   = "db-disable-seeks-compaction"
	optionNamePassword                   = "password"
	optionNamePasswordFile               = "password-file"
	optionNameAPIAddr                    = "api-addr"
	optionNameP2PAddr                    = "p2p-addr"
	optionNameNATAddr                    = "nat-addr"
	optionNameP2PWSEnable                = "p2p-ws-enable"
	optionNameDebugAPIEnable             = "debug-api-enable"
	optionNameDebugAPIAddr               = "debug-api-addr"
	optionNameBootnodes                  = "bootnode"
	optionNameNetworkID                  = "network-id"
	optionWelcomeMessage                 = "welcome-message"
	optionCORSAllowedOrigins             = "cors-allowed-origins"
	optionNameTracingEnabled             = "tracer-enable"
	optionNameTracingEndpoint            = "tracer-endpoint"
	optionNameTracingHost                = "tracer-host"
	optionNameTracingPort                = "tracer-port"
	optionNameTracingServiceName         = "tracer-service-name"
	optionNameVerbosity                  = "verbosity"
	optionNamePaymentThreshold           = "payment-threshold"
	optionNamePaymentTolerance           = "payment-tolerance-percent"
	optionNamePaymentEarly               = "payment-early-percent"
	optionNameResolverEndpoints          = "resolver-options"
	optionNameBootnodeMode               = "bootnode-mode"
	optionNameClefSignerEnable           = "clef-signer-enable"
	optionNameClefSignerEndpoint         = "clef-signer-endpoint"
	optionNameClefSignerBSCAddress       = "clef-signer-bsc-address"
	optionNameBSCEndpoint                = "bsc-rpc-endpoint"
	optionNameSwapFactoryAddress         = "swap-factory-address"
	optionNameSwapLegacyFactoryAddresses = "swap-legacy-factory-addresses"
	optionNameSwapInitialDeposit         = "swap-initial-deposit"
	optionNameSwapEnable                 = "swap-enable"
	optionNameChequebookEnable           = "chequebook-enable"
	optionNameTransactionHash            = "transaction"
	optionNameBlockHash                  = "block-hash"
	optionNameSwapDeploymentGasPrice     = "swap-deployment-gas-price"
	optionNameFullNode                   = "full-node"
	optionNameVoucherContractAddress     = "voucher-stamp-address"
	optionNamePriceOracleAddress         = "price-oracle-address"
	optionNamePledgeAddress              = "pledge-address"
	optionNameRewardAddress              = "reward-address"
	optionNameBlockTime                  = "block-time"
	optionWarmUpTime                     = "warmup-time"
	optionNameMainNet                    = "mainnet"
	optionNameRetrievalCaching           = "cache-retrieval"
	optionNameDevReserveCapacity         = "dev-reserve-capacity"
	optionNameResync                     = "resync"
	optionNamePProfBlock                 = "pprof-profile"
	optionNamePProfMutex                 = "pprof-mutex"
	optionNameStaticNodes                = "static-nodes"
	optionNameAllowPrivateCIDRs          = "allow-private-cidrs"
	optionNameSleepAfter                 = "sleep-after"
	optionNameRestrictedAPI              = "restricted"
	optionNameTokenEncryptionKey         = "token-encryption-key"
	optionNameAdminPasswordHash          = "admin-password"
	optionNameUseVoucherSnapshot         = "use-voucher-snapshot"
	optionNameRemoteEndpoint             = "remote-endpoint"
	optionNameMaxWorker                  = "max-worker"
	optionTrustNode                      = "trust-node"
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
			Use:           "mop",
			Short:         "BNB Smart Chain Cluster MOP",
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

	if err := c.initStartDevCmd(); err != nil {
		return nil, err
	}

	if err := c.initHasherCmd(); err != nil {
		return nil, err
	}

	if err := c.initInitCmd(); err != nil {
		return nil, err
	}

	if err := c.initDeployCmd(); err != nil {
		return nil, err
	}

	if err := c.initBuyStampCmd(); err != nil {
		return nil, err
	}

	if err := c.initListStampCmd(); err != nil {
		return nil, err
	}

	if err := c.initShowStampCmd(); err != nil {
		return nil, err
	}

	if err := c.initUploadCmd(); err != nil {
		return nil, err
	}

	if err := c.initDownloadCmd(); err != nil {
		return nil, err
	}

	if err := c.initExportPrivateCmd(); err != nil {
		return nil, err
	}

	c.initVersionCmd()
	c.initDBCmd()

	if err := c.initConfigurateOptionsCmd(); err != nil {
		return nil, err
	}

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
	globalFlags.StringVar(&c.cfgFile, "config", "", "config file (default is $HOME/.mop.yaml)")
}

func (c *command) initConfig() (err error) {
	config := viper.New()
	configName := ".mop"
	if c.cfgFile != "" {
		// Use config file from the flag.
		config.SetConfigFile(c.cfgFile)
	} else {
		// Search config in home directory with name ".mop" (without extension).
		config.AddConfigPath(c.homeDir)
		config.SetConfigName(configName)
	}

	// Environment
	config.SetEnvPrefix("mop")
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

func (c *command) setAllFlags(cmd *cobra.Command) {
	cmd.Flags().String(optionNameDataDir, filepath.Join(c.homeDir, ".mop"), "data directory")
	cmd.Flags().Uint64(optionNameCacheCapacity, 1000000, fmt.Sprintf("cache capacity in chunks, multiply by %d to get approximate capacity in bytes", cluster.ChunkSize))
	cmd.Flags().Uint64(optionNameMemCacheCapacity, 1000, fmt.Sprintf("memory cache capacity in chunks, multiply by %d to get approximate capacity in bytes", cluster.ChunkSize))
	cmd.Flags().Uint64(optionNameDBOpenFilesLimit, 200, "number of open files allowed by database")
	cmd.Flags().Uint64(optionNameDBBlockCacheCapacity, 32*1024*1024, "size of block cache of the database in bytes")
	cmd.Flags().Uint64(optionNameDBWriteBufferSize, 32*1024*1024, "size of the database write buffer in bytes")
	cmd.Flags().Bool(optionNameDBDisableSeeksCompaction, false, "disables db compactions triggered by seeks")
	cmd.Flags().String(optionNamePassword, "", "password for decrypting keys")
	cmd.Flags().String(optionNamePasswordFile, "", "path to a file that contains password for decrypting keys")
	cmd.Flags().String(optionNameAPIAddr, ":1683", "HTTP API listen address")
	cmd.Flags().String(optionNameP2PAddr, ":1684", "P2P listen address")
	cmd.Flags().String(optionNameNATAddr, "", "NAT exposed address")
	cmd.Flags().Bool(optionNameP2PWSEnable, false, "enable P2P WebSocket transport")
	cmd.Flags().StringSlice(optionNameBootnodes, []string{"/ip4/202.83.246.155/tcp/1684/p2p/16Uiu2HAmPqr2vmnwZi6HhTmWoCEVx2pD37m3p9G5dfNYCMrormLf"}, "initial nodes to connect to")
	cmd.Flags().Bool(optionNameDebugAPIEnable, false, "enable debug HTTP API")
	cmd.Flags().String(optionNameDebugAPIAddr, ":1685", "debug HTTP API listen address")
	cmd.Flags().Uint64(optionNameNetworkID, 97, "ID of the Cluster network")
	cmd.Flags().StringSlice(optionCORSAllowedOrigins, []string{}, "origins with CORS headers enabled")
	cmd.Flags().Bool(optionNameTracingEnabled, false, "enable tracer")
	cmd.Flags().String(optionNameTracingEndpoint, "127.0.0.1:1680", "endpoint to send tracer data")
	cmd.Flags().String(optionNameTracingHost, "", "host to send tracer data")
	cmd.Flags().String(optionNameTracingPort, "", "port to send tracer data")
	cmd.Flags().String(optionNameTracingServiceName, "mop", "service name identifier for tracer")
	cmd.Flags().String(optionNameVerbosity, "info", "log verbosity level 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=trace")
	cmd.Flags().String(optionWelcomeMessage, "", "send a welcome message string during handshakes")
	cmd.Flags().String(optionNamePaymentThreshold, "13500000", "threshold in MOP where you expect to get paid from your peers")
	cmd.Flags().Int64(optionNamePaymentTolerance, 25, "excess debt above payment threshold in percentages where you disconnect from your peer")
	cmd.Flags().Int64(optionNamePaymentEarly, 50, "percentage below the peers payment threshold when we initiate settlement")
	cmd.Flags().StringSlice(optionNameResolverEndpoints, []string{}, "ENS compatible API endpoint for a TLD and with contract address, can be repeated, format [tld:][contract-addr@]url")
	cmd.Flags().Bool(optionNameBootnodeMode, false, "cause the node to always accept incoming connections")
	cmd.Flags().Bool(optionNameClefSignerEnable, false, "enable clef signer")
	cmd.Flags().String(optionNameClefSignerEndpoint, "", "clef signer endpoint")
	cmd.Flags().String(optionNameClefSignerBSCAddress, "", "BNB Smart Chain to use from clef signer")
	cmd.Flags().StringSlice(optionNameBSCEndpoint, []string{"http://202.83.246.155:8575", "https://data-seed-prebsc-1-s1.binance.org:8545"}, "swap BNB Smart Chain endpoint")
	cmd.Flags().String(optionNameSwapFactoryAddress, "", "swap factory addresses")
	cmd.Flags().StringSlice(optionNameSwapLegacyFactoryAddresses, nil, "legacy swap factory addresses")
	cmd.Flags().String(optionNameSwapInitialDeposit, "10000000000000000", "initial deposit if deploying a new chequebook")
	cmd.Flags().Bool(optionNameSwapEnable, true, "enable swap")
	cmd.Flags().Bool(optionNameChequebookEnable, true, "enable chequebook")
	cmd.Flags().Bool(optionNameFullNode, false, "cause the node to start in full mode")
	cmd.Flags().String(optionNameVoucherContractAddress, "", "voucher stamp contract address")
	cmd.Flags().String(optionNamePriceOracleAddress, "", "price oracle contract address")
	cmd.Flags().String(optionNamePledgeAddress, "", "pledge contract address")
	cmd.Flags().String(optionNameRewardAddress, "", "reward contract address")
	cmd.Flags().String(optionNameTransactionHash, "", "proof-of-identity transaction hash")
	cmd.Flags().String(optionNameBlockHash, "", "block hash of the block whose parent is the block that contains the transaction hash")
	cmd.Flags().Uint64(optionNameBlockTime, 3, "chain block time")
	cmd.Flags().String(optionNameSwapDeploymentGasPrice, "", "gas price in wei to use for deployment and funding")
	cmd.Flags().Duration(optionWarmUpTime, time.Minute*5, "time to warmup the node before some major protocols can be kicked off.")
	cmd.Flags().Bool(optionNameMainNet, false, "triggers connect to main net bootnodes.")
	cmd.Flags().Bool(optionNameRetrievalCaching, true, "enable forwarded content caching")
	cmd.Flags().Bool(optionNameResync, false, "forces the node to resync voucher contract data")
	cmd.Flags().Bool(optionNamePProfBlock, false, "enable pprof block profile")
	cmd.Flags().Bool(optionNamePProfMutex, false, "enable pprof mutex profile")
	cmd.Flags().StringSlice(optionNameStaticNodes, []string{}, "protect nodes from getting kicked out on bootnode")
	cmd.Flags().Bool(optionNameAllowPrivateCIDRs, false, "allow to advertise private CIDRs to the public network")
	cmd.Flags().Bool(optionNameRestrictedAPI, false, "enable permission check on the http APIs")
	cmd.Flags().String(optionNameTokenEncryptionKey, "", "admin username to get the security token")
	cmd.Flags().String(optionNameAdminPasswordHash, "", "bcrypt hash of the admin password to get the security token")
	cmd.Flags().Bool(optionNameUseVoucherSnapshot, false, "bootstrap node using voucher snapshot from the network")
	cmd.Flags().String(optionNameRemoteEndpoint, "", "push remote server")
	cmd.Flags().Int(optionNameMaxWorker, 0, "number of workers")
	cmd.Flags().Bool(optionTrustNode, false, "ensure the locally chunk is valid")
}

func newLogger(cmd *cobra.Command, verbosity string) (log.Logger, error) {
	var (
		sink   = cmd.OutOrStdout()
		vLevel = log.VerbosityNone
	)

	switch verbosity {
	case "0", "silent":
		sink = io.Discard
	case "1", "error":
		vLevel = log.VerbosityError
	case "2", "warn":
		vLevel = log.VerbosityWarning
	case "3", "info":
		vLevel = log.VerbosityInfo
	case "4", "debug":
		vLevel = log.VerbosityDebug
	case "5", "trace":
		vLevel = log.VerbosityDebug + 1 // For backwards compatibility, just enable v1 debugging as trace.
	default:
		return nil, fmt.Errorf("unknown verbosity level %q", verbosity)
	}

	log.ModifyDefaults(
		log.WithTimestamp(),
		log.WithLogMetrics(),
	)

	return log.NewLogger(
		node.LoggerName,
		log.WithSink(sink),
		log.WithVerbosity(vLevel),
	).Register(), nil
}

func newFileLogger(cmd *cobra.Command, verbosity string, dataDir string, remote bool) (log.Logger, error) {
	var (
		sink   = cmd.OutOrStdout()
		vLevel = log.VerbosityNone
	)

	if dataDir != "" {
		if v, err := rotatelogs.New(
			filepath.Join(dataDir, "logs", "mop.log.%Y%m%d"),
			rotatelogs.WithLinkName(filepath.Join(dataDir, "logs", "mop.log")),
			rotatelogs.WithMaxAge(15*24*time.Hour),
			rotatelogs.WithRotationTime(24*time.Hour),
			rotatelogs.WithHandler(rotatelogs.HandlerFunc(func(e rotatelogs.Event) {
				if e.Type() != rotatelogs.FileRotatedEventType {
					return
				}

				if remote {
					if err := storeTrafficAnalysis(e.(*rotatelogs.FileRotatedEvent).PreviousFile()); err != nil {
						fmt.Println("storeTrafficAnalysis", "error", err)
					}
				}
			})),
		); err != nil {
			return nil, fmt.Errorf("unknown log file %s", err)
		} else {
			sink = v
		}
	}

	switch verbosity {
	case "0", "silent":
		sink = io.Discard
	case "1", "error":
		vLevel = log.VerbosityError
	case "2", "warn":
		vLevel = log.VerbosityWarning
	case "3", "info":
		vLevel = log.VerbosityInfo
	case "4", "debug":
		vLevel = log.VerbosityDebug
	case "5", "trace":
		vLevel = log.VerbosityDebug + 1 // For backwards compatibility, just enable v1 debugging as trace.
	default:
		return nil, fmt.Errorf("unknown verbosity level %q", verbosity)
	}

	log.ModifyDefaults(
		log.WithTimestamp(),
		log.WithLogMetrics(),
	)

	return log.NewLogger(
		node.LoggerName,
		log.WithSink(sink),
		log.WithVerbosity(vLevel),
	).Register(), nil
}

func storeTrafficAnalysis(file string) error {
	w, err := os.Create(file + ".traffic")
	if err != nil {
		return err
	}
	defer w.Close()

	readLine := func(fileName string, handler func(string) error) error {
		f, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer f.Close()

		br := bufio.NewReader(f)
		for {
			line, _, err := br.ReadLine()
			if err != nil {
				// file read complete
				if err == io.EOF {
					return nil
				}
				return err
			}
			if err := handler(string(line)); err != nil {
				return err
			}
		}
	}

	handler := func(line string) error {
		if strings.Contains(line, "/bytes") || strings.Contains(line, "/chunks") || strings.Contains(line, "/mop") {
			if _, err := w.WriteString(line + "\n"); err != nil {
				return err
			}
		}
		return nil
	}

	return readLine(file, handler)
}
