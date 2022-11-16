//go:build !windows
// +build !windows

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

package e2e

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/docker/compose/v2/pkg/utils"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/icmd"
)

func TestComposeCancel(t *testing.T) {
	c := NewParallelCLI(t)

	t.Run("metrics on cancel Compose build", func(t *testing.T) {
		c.RunDockerComposeCmd(t, "ls")
		buildProjectPath := "fixtures/build-infinite/compose.yaml"

		// require a separate groupID from the process running tests, in order to simulate ctrl+C from a terminal.
		// sending kill signal
		stdout := &utils.SafeBuffer{}
		stderr := &utils.SafeBuffer{}
		cmd, err := StartWithNewGroupID(context.Background(),
			c.NewDockerComposeCmd(t, "-f", buildProjectPath, "build", "--progress", "plain"),
			stdout,
			stderr)
		assert.NilError(t, err)

		c.WaitForCondition(t, func() (bool, string) {
			out := stdout.String()
			errors := stderr.String()
			return strings.Contains(out,
					"RUN sleep infinity"), fmt.Sprintf("'RUN sleep infinity' not found in : \n%s\nStderr: \n%s\n", out,
					errors)
		}, 30*time.Second, 1*time.Second)

		err = syscall.Kill(-cmd.Process.Pid, syscall.SIGINT) // simulate Ctrl-C : send signal to processGroup, children will have same groupId by default

		assert.NilError(t, err)
		c.WaitForCondition(t, func() (bool, string) {
			out := stdout.String()
			errors := stderr.String()
			return strings.Contains(out, "CANCELED"), fmt.Sprintf("'CANCELED' not found in : \n%s\nStderr: \n%s\n", out,
				errors)
		}, 10*time.Second, 1*time.Second)
	})
}

func StartWithNewGroupID(ctx context.Context, command icmd.Cmd, stdout *utils.SafeBuffer, stderr *utils.SafeBuffer) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, command.Command[0], command.Command[1:]...)
	cmd.Env = command.Env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}
	err := cmd.Start()
	return cmd, err
}
