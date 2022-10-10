/*
   Copyright 2020 Docker Compose CLI authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package compose

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/utils"
	"github.com/spf13/cobra"
)

func deployCommand(p *projectOptions, dockerCli command.Cli, backend api.Service) *cobra.Command {
	up := upOptions{}
	create := createOptions{}
	pull := pullOptions{projectOptions: p}
	deployCmd := &cobra.Command{
		Use:   "deploy [OPTIONS] [SERVICE...]",
		Short: "Deploy containers(combine pull, up, and hooks. hook supported command, scripts of shell, golang+)",
		PreRunE: AdaptCmd(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			create.timeChanged = cmd.Flags().Changed("timeout")

			pull.quiet = create.quietPull
			return validateFlags(&up, &create)
		}),
		ValidArgsFunction: completeServiceNames(p),
	}
	deployCmd.RunE = p.WithServices(func(ctx context.Context, project *types.Project, services []string) error {
		create.ignoreOrphans = utils.StringToBool(project.Environment["COMPOSE_IGNORE_ORPHANS"])
		if create.ignoreOrphans && create.removeOrphans {
			return fmt.Errorf("COMPOSE_IGNORE_ORPHANS and --remove-orphans cannot be combined")
		}
		return runDeploy(ctx, deployCmd, backend, create, up, pull, project, services)
	})

	flags := deployCmd.Flags()
	flags.BoolVarP(&up.Detach, "detach", "d", false, "Detached mode: Run containers in the background")
	flags.BoolVar(&create.Build, "build", false, "Build images before starting containers.")
	flags.BoolVar(&create.noBuild, "no-build", false, "Don't build an image, even if it's missing.")
	flags.StringVar(&create.Pull, "pull", "missing", `Pull image before running ("always"|"missing"|"never")`)
	flags.BoolVar(&create.removeOrphans, "remove-orphans", false, "Remove containers for services not defined in the Compose file.")
	flags.StringArrayVar(&up.scale, "scale", []string{}, "Scale SERVICE to NUM instances. Overrides the `scale` setting in the Compose file if present.")
	flags.BoolVar(&up.noColor, "no-color", false, "Produce monochrome output.")
	flags.BoolVar(&up.noPrefix, "no-log-prefix", false, "Don't print prefix in logs.")
	flags.BoolVar(&create.forceRecreate, "force-recreate", false, "Recreate containers even if their configuration and image haven't changed.")
	flags.BoolVar(&create.noRecreate, "no-recreate", false, "If containers already exist, don't recreate them. Incompatible with --force-recreate.")
	flags.BoolVar(&up.noStart, "no-start", false, "Don't start the services after creating them.")
	flags.BoolVar(&up.cascadeStop, "abort-on-container-exit", false, "Stops all containers if any container was stopped. Incompatible with -d")
	flags.StringVar(&up.exitCodeFrom, "exit-code-from", "", "Return the exit code of the selected service container. Implies --abort-on-container-exit")
	flags.IntVarP(&create.timeout, "timeout", "t", 10, "Use this timeout in seconds for container shutdown when attached or when containers are already running.")
	flags.BoolVar(&up.noDeps, "no-deps", false, "Don't start linked services.")
	flags.BoolVar(&create.recreateDeps, "always-recreate-deps", false, "Recreate dependent containers. Incompatible with --no-recreate.")
	flags.BoolVarP(&create.noInherit, "renew-anon-volumes", "V", false, "Recreate anonymous volumes instead of retrieving data from the previous containers.")
	flags.BoolVar(&up.attachDependencies, "attach-dependencies", false, "Attach to dependent containers.")
	flags.BoolVar(&create.quietPull, "quiet-pull", false, "Pull without printing progress information.")
	flags.StringArrayVar(&up.attach, "attach", []string{}, "Attach to service output.")
	flags.BoolVar(&up.wait, "wait", false, "Wait for services to be running|healthy. Implies detached mode.")

	flags.Bool("hook", false, "enable x-hooks, and will execute pre-deploy post-deploy")

	return deployCmd
}

func runDeploy(ctx context.Context, cmd *cobra.Command, backend api.Service, createOptions createOptions, upOptions upOptions, pullOptions pullOptions, project *types.Project, services []string) error {
	if len(project.Services) == 0 {
		return fmt.Errorf("no service selected")
	}

	hookEnable, _ := cmd.Flags().GetBool("hook")
	if createOptions.Pull == "always" {
		if err := runPull(ctx, backend, pullOptions, services); err != nil {
			return err
		}
	}

	// 啟動hook
	if hookEnable {
		h := newHook(ctx, cmd, backend, project)
		// 解析x-hooks
		if err := h.parse(); err != nil {
			return err
		}

		// 全局 pre-deploy
		if err := h.PreDeploy(createOptions, upOptions, pullOptions, nil); err != nil {
			return err
		}

		// service pre-deploy
		for k := range project.Services {
			if err := h.PreDeploy(createOptions, upOptions, pullOptions, &project.Services[k]); err != nil {
				return err
			}
		}

		// up
		if err := runUp(h.ctx, h.backend, createOptions, upOptions, h.project, services); err != nil {
			return err
		}

		// services post-deploy
		for k := range project.Services {
			if err := h.PostDeploy(createOptions, upOptions, pullOptions, &project.Services[k]); err != nil {
				return err
			}
		}

		// 全局 post-deploy
		return h.PostDeploy(createOptions, upOptions, pullOptions, nil)
	}

	return runUp(ctx, backend, createOptions, upOptions, project, services)
}
