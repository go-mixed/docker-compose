//go:build !windows
// +build !windows

package mod

import (
	"fmt"
	"plugin"
)

func (m *Modules) LoadPlugins(plugins []string) error {
	for _, path := range plugins {
		p, err := plugin.Open(path)
		if err != nil {
			return err
		}
		loadFunc, err := p.Lookup("Load")
		if err != nil {
			return fmt.Errorf("func \"Load\" not found in plugin \"%s\": %w", path, err)
		}
		loadFunc.(func())()
	}
	return nil
}
