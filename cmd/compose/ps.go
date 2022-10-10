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
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/docker/compose/v2/cmd/formatter"
	"github.com/docker/compose/v2/pkg/utils"
	"github.com/docker/docker/api/types"

	formatter2 "github.com/docker/cli/cli/command/formatter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/docker/compose/v2/pkg/api"
)

type psOptions struct {
	*projectOptions
	Format   string
	All      bool
	Quiet    bool
	Services bool
	Filter   string
	Status   []string
}

func (p *psOptions) parseFilter() error {
	if p.Filter == "" {
		return nil
	}
	parts := strings.SplitN(p.Filter, "=", 2)
	if len(parts) != 2 {
		return errors.New("arguments to --filter should be in form KEY=VAL")
	}
	switch parts[0] {
	case "status":
		p.Status = append(p.Status, parts[1])
	case "source":
		return api.ErrNotImplemented
	default:
		return fmt.Errorf("unknown filter %s", parts[0])
	}
	return nil
}

func psCommand(p *projectOptions, backend api.Service) *cobra.Command {
	opts := psOptions{
		projectOptions: p,
	}
	psCmd := &cobra.Command{
		Use:   "ps [OPTIONS] [SERVICE...]",
		Short: "List containers",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.parseFilter()
		},
		RunE: Adapt(func(ctx context.Context, args []string) error {
			return runPs(ctx, backend, args, opts)
		}),
		ValidArgsFunction: completeServiceNames(p),
	}
	flags := psCmd.Flags()
	flags.StringVar(&opts.Format, "format", "pretty", "Format the output. Values: [pretty | json]")
	flags.StringVar(&opts.Filter, "filter", "", "Filter services by a property (supported filters: status).")
	flags.StringArrayVar(&opts.Status, "status", []string{}, "Filter services by status. Values: [paused | restarting | removing | running | dead | created | exited]")
	flags.BoolVarP(&opts.Quiet, "quiet", "q", false, "Only display IDs")
	flags.BoolVar(&opts.Services, "services", false, "Display services")
	flags.BoolVarP(&opts.All, "all", "a", false, "Show all stopped containers (including those created by the run command)")
	return psCmd
}

func runPs(ctx context.Context, backend api.Service, services []string, opts psOptions) error {
	project, name, err := opts.projectOrName()
	if err != nil {
		return err
	}
	containers, err := backend.Ps(ctx, name, api.PsOptions{
		Project:  project,
		All:      opts.All,
		Services: services,
	})
	if err != nil {
		return err
	}

SERVICES:
	for _, s := range services {
		for _, c := range containers {
			if c.Service == s {
				continue SERVICES
			}
		}
		return fmt.Errorf("no such service: %s", s)
	}

	if len(opts.Status) != 0 {
		containers = filterByStatus(containers, opts.Status)
	}

	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Name < containers[j].Name
	})

	if opts.Quiet {
		for _, c := range containers {
			fmt.Println(c.ID)
		}
		return nil
	}

	if opts.Services {
		services := []string{}
		for _, s := range containers {
			if !utils.StringContains(services, s.Service) {
				services = append(services, s.Service)
			}
		}
		fmt.Println(strings.Join(services, "\n"))
		return nil
	}

	return formatter.Print(containers, opts.Format, os.Stdout,
		writer(containers),
		"NAME", "COMMAND", "SERVICE", "STATUS", "PORTS")
}

func writer(containers []api.ContainerSummary) func(w io.Writer) {
	return func(w io.Writer) {
		for _, container := range containers {
			ports := displayablePorts(container)
			status := container.State
			if status == "running" && container.Health != "" {
				status = fmt.Sprintf("%s (%s)", container.State, container.Health)
			} else if status == "exited" || status == "dead" {
				status = fmt.Sprintf("%s (%d)", container.State, container.ExitCode)
			}
			command := formatter2.Ellipsis(container.Command, 20)
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", container.Name, strconv.Quote(command), container.Service, status, ports)
		}
	}
}

func filterByStatus(containers []api.ContainerSummary, statuses []string) []api.ContainerSummary {
	var filtered []api.ContainerSummary
	for _, c := range containers {
		if hasStatus(c, statuses) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func hasStatus(c api.ContainerSummary, statuses []string) bool {
	for _, status := range statuses {
		if c.State == status {
			return true
		}
	}
	return false
}

func displayablePorts(c api.ContainerSummary) string {
	if c.Publishers == nil {
		return ""
	}

	ports := make([]types.Port, len(c.Publishers))
	for i, pub := range c.Publishers {
		ports[i] = types.Port{
			IP:          pub.URL,
			PrivatePort: uint16(pub.TargetPort),
			PublicPort:  uint16(pub.PublishedPort),
			Type:        pub.Protocol,
		}
	}

	return formatter2.DisplayablePorts(ports)
}
