package main

import (
	"github.com/spf13/cobra"
	"github.com/suborbital/hivew/hivew/command"
	"github.com/suborbital/hivew/hivew/context"
	"github.com/suborbital/hivew/hivew/release"
)

func rootCommand(bctx *context.BuildContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "hivew",
		Short:   "Hive WASM Runnable toolchain",
		Long:    `hivew includes a full toolchain for managing and running WASM Runnables with Hive.`,
		Version: release.HiveWDotVersion,
	}

	cmd.SetVersionTemplate("{{.Version}}\n")

	cmd.AddCommand(command.BuildCmd(bctx))

	return cmd
}
