package compose

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type xHooks struct {
	PreDeploy  []types.ShellCommand `mapstructure:"pre-deploy"`
	PostDeploy []types.ShellCommand `mapstructure:"post-deploy"`

	PreUndeploy  []types.ShellCommand `mapstructure:"pre-undeploy"`
	PostUndeploy []types.ShellCommand `mapstructure:"post-undeploy"`
}

type hook struct {
	ctx     context.Context
	cmd     *cobra.Command
	project *types.Project
	backend api.Service

	globalHooks   xHooks
	servicesHooks map[string]xHooks
}

func newHook(ctx context.Context, cmd *cobra.Command, backend api.Service, project *types.Project) *hook {
	return &hook{
		ctx:           ctx,
		cmd:           cmd,
		project:       project,
		backend:       backend,
		globalHooks:   xHooks{},
		servicesHooks: map[string]xHooks{},
	}
}

func (h *hook) parse() error {
	if err := loader.Transform(h.project.Extensions["x-hooks"], &h.globalHooks); err != nil {
		return err
	}

	for _, service := range h.project.Services {
		if _, ok := service.Extensions["x-hooks"]; ok {
			var xHooks xHooks
			if err := loader.Transform(service.Extensions["x-hooks"], &xHooks); err != nil {
				return err
			}
			h.servicesHooks[service.Name] = xHooks
		}
	}

	return nil
}

func (h *hook) PreDeploy(createOptions createOptions, upOptions upOptions, pullOptions pullOptions, service *types.ServiceConfig) error {
	xHook := h.globalHooks
	if service != nil {
		var ok bool
		if xHook, ok = h.servicesHooks[service.Name]; !ok { // service中不存在 x-hook则退出
			return nil
		}
		fmt.Printf("[Hook]pre-deploy of service: \"%s\"\n", service.Name)
	} else {
		fmt.Printf("[Hook]pre-deploy of global\n")
	}

	return h.handle(xHook.PreDeploy, service)
}

func (h *hook) PostDeploy(createOptions createOptions, upOptions upOptions, pullOptions pullOptions, service *types.ServiceConfig) error {
	xHook := h.globalHooks
	if service != nil {
		var ok bool
		if xHook, ok = h.servicesHooks[service.Name]; !ok { // service中不存在 x-hook则退出
			return nil
		}
		fmt.Printf("[Hook]post-deploy of service: \"%s\"\n", service.Name)
	} else {
		fmt.Printf("[Hook]post-deploy of global\n")
	}

	return h.handle(xHook.PostDeploy, service)
}

func (h *hook) PreUndeploy(downOptions downOptions, service *types.ServiceConfig) error {
	xHook := h.globalHooks
	if service != nil {
		var ok bool
		if xHook, ok = h.servicesHooks[service.Name]; !ok { // service中不存在 x-hook则退出
			return nil
		}
		fmt.Printf("[Hook]pre-undeploy of service: \"%s\"\n", service.Name)
	} else {
		fmt.Printf("[Hook]pre-undeploy of global\n")
	}

	return h.handle(xHook.PreUndeploy, service)
}

func (h *hook) PostUndeploy(downOptions downOptions, service *types.ServiceConfig) error {
	xHook := h.globalHooks
	if service != nil {
		var ok bool
		if xHook, ok = h.servicesHooks[service.Name]; !ok { // service中不存在 x-hook则退出
			return nil
		}
		fmt.Printf("[Hook]post-undeploy of service: \"%s\"\n", service.Name)
	} else {
		fmt.Printf("[Hook]post-undeploy of global\n")
	}

	return h.handle(xHook.PostUndeploy, service)
}

func (h *hook) handle(commands []types.ShellCommand, service *types.ServiceConfig) error {
	for _, command := range commands {
		if exe := h.parseCommand(command, service); exe != nil {
			if err := exe.run(h, service); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *hook) parseCommand(command types.ShellCommand, service *types.ServiceConfig) *execute {
	if len(command) <= 0 {
		return nil
	}

	workDir := filepath.Dir(h.project.ComposeFiles[0]) // 相對docker-compose.yml文件的工作目錄

	if len(command) >= 2 && command[0] == "shell-key" {
		path := filepath.Join(workDir, "."+strings.TrimSpace(command[1])+".sh")
		if err := os.WriteFile(path, []byte(h.getXKey(command[1], service)), 0o644); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "write x-key to %s error: %s", path, err.Error())
			return nil
		}
		return &execute{
			env:         h.project.Environment,
			workDir:     workDir,
			command:     append(types.ShellCommand{"bash", path}, command[2:]...),
			deleteAfter: path,
		}
	}

	return &execute{
		env:     h.project.Environment,
		workDir: workDir,
		command: command,
	}
}

func (h *hook) getXKey(name string, service *types.ServiceConfig) string {
	name = strings.TrimSpace(name)
	if service != nil {
		if c, ok := service.Extensions[name]; ok {
			return c.(string)
		}
	}

	return h.project.Extensions[name].(string)
}

type execute struct {
	env         map[string]string
	workDir     string
	command     types.ShellCommand
	deleteAfter string
}

func (e *execute) run(h *hook, service *types.ServiceConfig) error {
	workDir := filepath.Dir(h.project.ComposeFiles[0]) // 相對docker-compose.yml文件的工作目錄
	if e.deleteAfter != "" {
		defer os.Remove(e.deleteAfter)
	}

	fmt.Printf(" - executing: %+q\n", e.command)

	var env = os.Environ()
	for k, v := range e.env {
		env = append(env, k+"="+v)
	}

	cmd := exec.CommandContext(h.ctx, e.command[0], e.command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = workDir
	cmd.Env = env
	return cmd.Run()
}
