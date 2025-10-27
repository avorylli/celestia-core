package commands

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/store"
)

var PruneTxInfoCmd = &cobra.Command{
	Use:   "prune-txinfo",
	Short: "prune old transaction info records to free up disk space",
	Long: `
This command removes old TxInfo records from the blockstore database.
TxInfo records are used for transaction status queries but can accumulate
over time and consume significant disk space.

The retain_height parameter specifies the height below which TxInfo records
will be deleted. Records for heights >= retain_height will be preserved.

Example:
  cometbft prune-txinfo --retain-height=1000

This will delete TxInfo records for all transactions in blocks with height < 1000.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		retainHeight, err := cmd.Flags().GetInt64("retain-height")
		if err != nil {
			return err
		}

		if retainHeight <= 0 {
			return fmt.Errorf("retain-height must be greater than 0")
		}

		// Open the blockstore database
		dbPath := filepath.Join(config.RootDir, "data", "blockstore.db")
		db, err := dbm.NewDB("blockstore", dbm.BackendType(config.DBBackend), dbPath)
		if err != nil {
			return fmt.Errorf("failed to open blockstore database: %w", err)
		}
		defer db.Close()

		// Create blockstore instance
		blockStore := store.NewBlockStore(db)

		// Get current height for validation
		currentHeight := blockStore.Height()
		if retainHeight >= currentHeight {
			return fmt.Errorf("retain-height %d must be less than current height %d", retainHeight, currentHeight)
		}

		logger.Info("starting TxInfo pruning",
			"retain_height", retainHeight,
			"current_height", currentHeight)

		// Perform pruning
		err = blockStore.PruneTxInfo(retainHeight)
		if err != nil {
			return fmt.Errorf("failed to prune TxInfo records: %w", err)
		}

		logger.Info("TxInfo pruning completed successfully",
			"retain_height", retainHeight)

		return nil
	},
}

func init() {
	PruneTxInfoCmd.Flags().Int64("retain-height", 0, "height below which TxInfo records will be deleted")
	PruneTxInfoCmd.MarkFlagRequired("retain-height")
}
