package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/incentives/settlement/swap/erc20"
	"github.com/redesblock/mop/core/node"
	"github.com/spf13/cobra"
)

const blocktime = 15

func (c *command) initDeployCmd() error {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy and fund the chequebook contract",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if (len(args)) > 0 {
				return cmd.Help()
			}

			v := strings.ToLower(c.config.GetString(optionNameVerbosity))
			logger, err := newLogger(cmd, v)
			if err != nil {
				return fmt.Errorf("new logger: %w", err)
			}

			dataDir := c.config.GetString(optionNameDataDir)
			factoryAddress := c.config.GetString(optionNameSwapFactoryAddress)
			swapInitialDeposit := c.config.GetString(optionNameSwapInitialDeposit)
			bscEndpoints := c.config.GetStringSlice(optionNameBSCEndpoint)
			deployGasPrice := c.config.GetString(optionNameSwapDeploymentGasPrice)
			networkID := c.config.GetUint64(optionNameNetworkID)

			stateStore, err := node.InitStateStore(logger, dataDir)
			if err != nil {
				return err
			}

			defer stateStore.Close()

			signerConfig, err := c.configureSigner(cmd, logger)
			if err != nil {
				return err
			}
			signer := signerConfig.signer

			ctx := cmd.Context()

			swapBackend, overlayEthAddress, chainID, transactionMonitor, transactionService, err := node.InitChain(
				ctx,
				logger,
				stateStore,
				bscEndpoints,
				0,
				signer,
				blocktime,
				true,
			)
			if err != nil {
				return err
			}
			defer swapBackend.Close()
			defer transactionMonitor.Close()

			chequebookFactory, err := node.InitChequebookFactory(
				logger,
				swapBackend,
				chainID,
				transactionService,
				factoryAddress,
				nil,
			)
			if err != nil {
				return err
			}

			erc20Address, err := chequebookFactory.ERC20Address(ctx)
			if err != nil {
				return err
			}

			erc20Service := erc20.New(transactionService, erc20Address)

			_, err = node.InitChequebookService(
				ctx,
				logger,
				stateStore,
				signer,
				chainID,
				swapBackend,
				overlayEthAddress,
				transactionService,
				chequebookFactory,
				swapInitialDeposit,
				deployGasPrice,
				erc20Service,
			)
			if err != nil {
				return err
			}

			optionTrxHash := c.config.GetString(optionNameTransactionHash)
			optionBlockHash := c.config.GetString(optionNameBlockHash)

			txHash, err := node.GetTxHash(stateStore, logger, optionTrxHash)
			if err != nil {
				return fmt.Errorf("invalid transaction hash: %w", err)
			}

			blockTime := time.Duration(c.config.GetUint64(optionNameBlockTime)) * time.Second

			blockHash, err := node.GetTxNextBlock(ctx, logger, swapBackend, transactionMonitor, blockTime, txHash, optionBlockHash)
			if err != nil {
				return err
			}

			pubKey, err := signer.PublicKey()
			if err != nil {
				return err
			}

			clusterAddress, err := crypto.NewOverlayAddress(*pubKey, networkID, blockHash)

			err = node.CheckOverlayWithStore(clusterAddress, stateStore)

			return err
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return c.config.BindPFlags(cmd.Flags())
		},
	}

	c.setAllFlags(cmd)
	c.root.AddCommand(cmd)

	return nil
}
