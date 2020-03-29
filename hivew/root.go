package main

import (
	"github.com/spf13/cobra"
	"github.com/suborbital/hive-wasm/hivew/command"
	"github.com/suborbital/hive-wasm/hivew/release"
)

func rootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "hivew",
		Short:   "Hive WASM Runnable toolchain",
		Long:    `hivew includes a full toolchain for managing and running WASM Runnables with Hive.`,
		Version: release.HiveWDotVersion,
	}

	cmd.SetVersionTemplate("{{.Version}}\n")

	cmd.AddCommand(command.BuildCmd())

	return cmd
}
