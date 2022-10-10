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
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/sync/errgroup"

	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/utils"
)

func (s *composeService) Logs(
	ctx context.Context,
	projectName string,
	consumer api.LogConsumer,
	options api.LogOptions,
) error {
	projectName = strings.ToLower(projectName)

	containers, err := s.getContainers(ctx, projectName, oneOffExclude, true, options.Services...)
	if err != nil {
		return err
	}

	project := options.Project
	if project == nil {
		project, err = s.getProjectWithResources(ctx, containers, projectName)
		if err != nil {
			return err
		}
	}

	if len(options.Services) == 0 {
		options.Services = project.ServiceNames()
	}

	containers = containers.filter(isService(options.Services...))
	eg, ctx := errgroup.WithContext(ctx)
	for _, c := range containers {
		c := c
		eg.Go(func() error {
			return s.logContainers(ctx, consumer, c, options)
		})
	}

	if options.Follow {
		printer := newLogPrinter(consumer)
		eg.Go(func() error {
			for _, c := range containers {
				printer.HandleEvent(api.ContainerEvent{
					Type:      api.ContainerEventAttach,
					Container: getContainerNameWithoutProject(c),
					Service:   c.Labels[api.ServiceLabel],
				})
			}
			return nil
		})

		eg.Go(func() error {
			return s.watchContainers(ctx, projectName, options.Services, printer.HandleEvent, containers, func(c types.Container) error {
				printer.HandleEvent(api.ContainerEvent{
					Type:      api.ContainerEventAttach,
					Container: getContainerNameWithoutProject(c),
					Service:   c.Labels[api.ServiceLabel],
				})
				return s.logContainers(ctx, consumer, c, options)
			})
		})

		eg.Go(func() error {
			_, err := printer.Run(ctx, false, "", nil)
			return err
		})
	}

	return eg.Wait()
}

func (s *composeService) logContainers(ctx context.Context, consumer api.LogConsumer, c types.Container, options api.LogOptions) error {
	cnt, err := s.apiClient().ContainerInspect(ctx, c.ID)
	if err != nil {
		return err
	}

	service := c.Labels[api.ServiceLabel]
	r, err := s.apiClient().ContainerLogs(ctx, cnt.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     options.Follow,
		Since:      options.Since,
		Until:      options.Until,
		Tail:       options.Tail,
		Timestamps: options.Timestamps,
	})
	if err != nil {
		return err
	}
	defer r.Close() //nolint:errcheck

	name := getContainerNameWithoutProject(c)
	w := utils.GetWriter(func(line string) {
		consumer.Log(name, service, line)
	})
	if cnt.Config.Tty {
		_, err = io.Copy(w, r)
	} else {
		_, err = stdcopy.StdCopy(w, w, r)
	}
	return err
}
