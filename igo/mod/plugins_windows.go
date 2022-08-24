package mod

import "fmt"

func (m *Modules) LoadPlugins(plugins []string) error {
	fmt.Println("WARNING: Loading plugin in Windows is not supported.")
	return nil
}
