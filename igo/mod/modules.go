package mod

import (
	"fmt"
	"github.com/goplus/igop"
	"golang.org/x/mod/modfile"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

import (
	_ "github.com/goplus/igop/pkg"
	_ "github.com/goplus/ipkg/github.com/modern-go/reflect2"
	_ "github.com/goplus/reflectx/icall/icall8192"
	// _ "igop/src/pkgs"
)

type Module struct {
	Name      string
	Path      string
	Version   string
	GoVersion string
}

type Modules struct {
	projectDir string
	modules    map[string]*Module
	rkeys      []string
}

func canonicalize(path string) string {
	if path == "" {
		return path
	}
	nPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	nPath = filepath.Clean(nPath)
	return nPath
}

func NewModules(projectDir string) *Modules {
	var m = &Modules{modules: map[string]*Module{}}

	m.projectDir = canonicalize(projectDir)

	return m
}

func (m *Modules) resortKeys() {
	for k := range m.modules {
		m.rkeys = append(m.rkeys, k)
	}
	// rkeys 倒序排序
	sort.Slice(m.rkeys, func(i, j int) bool {
		return m.rkeys[i] > m.rkeys[j]
	})
}

func (m *Modules) LoadGoMod(goModPath string) error {
	goModPath = canonicalize(goModPath)
	// go.mod存在
	if stat, err := os.Stat(goModPath); err == nil && !stat.IsDir() {
		if err = m.parseGoMod(goModPath); err != nil {
			return err
		}
	}
	return nil
}

func (m *Modules) LoadVendor(vendorPath string) error {
	if vendorPath == "" { // vendor 目录没有传递，尝试使用项目下的
		vendorPath = filepath.Join(m.projectDir, "vendor")
	}

	vendorPath = canonicalize(vendorPath)

	// vendor/modules.txt文件存在
	if stat, err := os.Stat(filepath.Join(vendorPath, "modules.txt")); err == nil && !stat.IsDir() {
		if err = m.parseVendor(vendorPath); err != nil {
			return err
		}
	}

	return nil
}

func (m *Modules) parseGoMod(goModPath string) error {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}
	f, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return err
	}
	if f.Module == nil {
		// No module declaration. Must add module path.
		return fmt.Errorf("no module declaration in go.mod. To specify the module path:\n\tgo mod edit -module=example.com/mod")
	}

	goVersion, err := modFileGoVersion(f)
	if err != nil {
		return err
	}

	m.modules[f.Module.Mod.Path] = &Module{
		Name:      f.Module.Mod.Path,
		Path:      m.projectDir,
		Version:   f.Module.Mod.Version,
		GoVersion: goVersion,
	}

	m.resortKeys()

	return nil
}

func (m *Modules) parseVendor(vendorPath string) error {
	vendorList, err := readVendorList(vendorPath)
	if err != nil {
		return fmt.Errorf("[Vendor]%w", err)
	}

	for k, v := range vendorList.vendorMeta {
		m.modules[k.Path] = &Module{
			Name:      k.Path,
			Path:      filepath.Join(vendorPath, k.Path),
			Version:   k.Version,
			GoVersion: v.GoVersion,
		}
	}

	m.resortKeys()

	return nil
}

func (m *Modules) Lookup(root, pkg string) (dir string, found bool) {

	module, ok := m.modules[pkg]
	if ok {
		return module.Path, ok
	}

	// 因为是倒序排列，故第一个匹配项是最长匹配
	for _, v := range m.rkeys {
		if strings.HasPrefix(pkg, v+"/") {
			module = m.modules[v]
			break
		}
	}

	if module != nil && module.Path != "" {
		return filepath.Join(module.Path, pkg[len(module.Name+"/"):]), true
	}

	return "", false
}

func (m *Modules) SetLookup(ctx *igop.Context) {
	ctx.Lookup = m.Lookup
}
