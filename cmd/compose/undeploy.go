package compose

import (
	"context"
	"fmt"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func undeployCommand(p *projectOptions, backend api.Service) *cobra.Command {
	opts := downOptions{
		projectOptions: p,
	}
	undeployCmd := &cobra.Command{
		Use:   "undeploy [OPTIONS] [--hook]",
		Short: "Stop and remove containers, networks with hooks",
		PreRunE: AdaptCmd(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			opts.timeChanged = cmd.Flags().Changed("timeout")
			if opts.images != "" {
				if opts.images != "all" && opts.images != "local" {
					return fmt.Errorf("invalid value for --rmi: %q", opts.images)
				}
			}
			return nil
		}),
		Args:              cobra.NoArgs,
		ValidArgsFunction: noCompletion(),
	}
	undeployCmd.RunE = Adapt(func(ctx context.Context, args []string) error {
		return runUndeploy(ctx, undeployCmd, backend, opts)
	})

	flags := undeployCmd.Flags()
	removeOrphans := utils.StringToBool(os.Getenv("COMPOSE_REMOVE_ORPHANS"))
	flags.BoolVar(&opts.removeOrphans, "remove-orphans", removeOrphans, "Remove containers for services not defined in the Compose file.")
	flags.IntVarP(&opts.timeout, "timeout", "t", 10, "Specify a shutdown timeout in seconds")
	flags.BoolVarP(&opts.volumes, "volumes", "v", false, "Remove named volumes declared in the `volumes` section of the Compose file and anonymous volumes attached to containers.")
	flags.StringVar(&opts.images, "rmi", "", `Remove images used by services. "local" remove only images that don't have a custom tag ("local"|"all")`)
	flags.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		if name == "volume" {
			name = "volumes"
			logrus.Warn("--volume is deprecated, please use --volumes")
		}
		return pflag.NormalizedName(name)
	})

	flags.Bool("hook", false, "enable x-hooks, and will execute pre-undeploy post-undeploy")

	return undeployCmd
}

func runUndeploy(ctx context.Context, cmd *cobra.Command, backend api.Service, downOpts downOptions) error {
	project, _, err := downOpts.projectOrName()
	if err != nil {
		return err
	}

	if len(project.Services) == 0 {
		return fmt.Errorf("no service selected")
	}

	hookEnable, _ := cmd.Flags().GetBool("hook")
	// 啟動hook
	if hookEnable {
		yamlBuf, _ := yaml.Marshal(project)
		_ = os.WriteFile(filepath.Join(filepath.Dir(project.ComposeFiles[0]), ".current-docker-compose.yml"), yamlBuf, 0644)

		h := newHook(ctx, cmd, backend, project)
		// 解析x-hooks
		if err := h.parse(); err != nil {
			return err
		}

		// 全局 pre-undeploy
		if err := h.PreUndeploy(downOpts, nil); err != nil {
			return err
		}

		// service pre-undeploy
		for k := range project.Services {
			if err := h.PreUndeploy(downOpts, &project.Services[k]); err != nil {
				return err
			}
		}

		if err = runDown(ctx, backend, downOpts); err != nil {
			return err
		}

		// service post-undeploy
		for k := range project.Services {
			if err := h.PostUndeploy(downOpts, &project.Services[k]); err != nil {
				return err
			}
		}

		// 全局 post-undeploy
		if err := h.PostUndeploy(downOpts, nil); err != nil {
			return err
		}
	}

	return runDown(ctx, backend, downOpts)
}
