package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/suborbital/hivew/hivew/util"
)

var dockerImageForLang = map[string]string{
	"rust": "suborbital/hive-rs:1.42",
}

type runnableDir struct {
	Name string
	Lang string
}

// BuildCmd returns the build command
func BuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "build a WASM runnable",
		Long:  `build a WASM runnable from local source files`,
		RunE:  runBuild,
	}

	cmd.Flags().Bool("bundle", false, "if true, bundle all resulting runnables into a deployable .wasm.zip bundle")

	return cmd
}

func runBuild(cmd *cobra.Command, args []string) error {
	dirs, err := getRunnableDirs()
	if err != nil {
		return errors.Wrap(err, "failed to getRunnableDirs")
	}

	if len(dirs) == 0 {
		return errors.New("🚫 no runnables found in current directory (no .hive.yaml files found)")
	}

	results := make([]os.File, len(dirs))

	for i, r := range dirs {
		fmt.Println(fmt.Sprintf("✨ START: building runnable: %s (%s)", r.Name, r.Lang))

		file, err := doBuildForRunnable(r)
		if err != nil {
			fmt.Println("🚫 FAIL:", errors.Wrapf(err, "failed to doBuild for %s", r.Name))
		} else {
			results[i] = *file
			fmt.Println(fmt.Sprintf("✨ DONE: %s was built -> %s/%s.wasm", r.Name, r.Name, r.Name))
		}

	}

	shouldBundle, err := cmd.Flags().GetBool("bundle")
	if err != nil {
		return errors.Wrap(err, "🚫 failed to get bundle flag")
	} else if shouldBundle {
		bundlePath, err := bundleTargetPath()
		if err != nil {
			return errors.Wrap(err, "🚫 failed to determine bundle path")
		}

		if err := util.WriteBundle(results, bundlePath); err != nil {
			return errors.Wrap(err, "🚫 failed to WriteBundle")
		}

		fmt.Println(fmt.Sprintf("✨ DONE: bundle was created -> %s", bundlePath))
	}

	return nil
}

func doBuildForRunnable(r runnableDir) (*os.File, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get CWD")
	}

	img := imageForLang(r.Lang)
	if img == "" {
		return nil, fmt.Errorf("%s is not a supported language", r.Lang)
	}

	_, _, err = util.Run(fmt.Sprintf("docker run --rm --mount type=bind,source=%s/%s,target=/root/rs-wasm %s", cwd, r.Name, img))
	if err != nil {
		return nil, errors.Wrap(err, "failed to Run docker command")
	}

	targetPath := filepath.Join(cwd, r.Name, fmt.Sprintf("%s.wasm", r.Name))
	os.Rename(filepath.Join(cwd, r.Name, "wasm_runner_bg.wasm"), targetPath)

	file, err := os.Open(targetPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open resulting built file %s", targetPath)
	}

	return file, nil
}

func imageForLang(lang string) string {
	img, ok := dockerImageForLang[lang]
	if !ok {
		return ""
	}

	return img
}

func getRunnableDirs() ([]runnableDir, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get CWD")
	}

	runnables := []runnableDir{}

	// go through all of the dirs in the current dir
	topLvlFiles, err := ioutil.ReadDir(cwd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list directory")
	}

	for _, tf := range topLvlFiles {
		if !tf.IsDir() {
			continue
		}

		// determine if a .hive.yaml exists in that dir
		innerFiles, err := ioutil.ReadDir(filepath.Join(cwd, tf.Name()))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list files in %s", tf.Name())
		}

		if containsDotHiveYaml(innerFiles) {
			runnable := runnableDir{
				Name: tf.Name(),
				Lang: "rust",
			}

			runnables = append(runnables, runnable)
		}
	}

	return runnables, nil
}

func containsDotHiveYaml(files []os.FileInfo) bool {
	for _, f := range files {
		if f.Name() == ".hive.yaml" || f.Name() == ".hive.yml" {
			return true
		}
	}

	return false
}

func bundleTargetPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "failed to get CWD")
	}

	return filepath.Join(cwd, "runnables.wasm.zip"), nil
}
