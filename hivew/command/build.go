package command

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/suborbital/hive-wasm/hivew/util"
)

// BuildCmd returns the build command
func BuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "build a WASM runnable",
		Long:  `build a WASM runnable from local source files`,
		RunE:  runBuild,
	}

	return cmd
}

func runBuild(cmd *cobra.Command, args []string) error {
	_, _, err := util.Run("docker run nonsense/container")
	if err != nil {
		return errors.Wrap(err, "failed to Run docker command")
	}

	return nil
}
