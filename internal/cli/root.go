// Package cli provides command-line interface for archlint.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mshogin/archlint/pkg/tracer"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "archlint",
	Short: "Tool for building architecture graphs",
	Long: `archlint - tool for building structural and behavioral graphs
from Go source code.`,
	Version: version,
}

// Execute runs the root command.
func Execute() error {
	tracer.Enter("cli.Execute")

	if err := rootCmd.Execute(); err != nil {
		tracer.ExitError("cli.Execute", err)
		return fmt.Errorf("command execution failed: %w", err)
	}

	tracer.ExitSuccess("cli.Execute")
	return nil
}
