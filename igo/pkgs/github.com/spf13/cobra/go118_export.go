// export by github.com/goplus/igop/cmd/qexp

//go:build go1.18
// +build go1.18

package cobra

import (
	q "github.com/spf13/cobra"

	"go/constant"
	"reflect"

	"github.com/goplus/igop"
)

func init() {
	igop.RegisterPackage(&igop.Package{
		Name: "cobra",
		Path: "github.com/spf13/cobra",
		Deps: map[string]string{
			"bytes":                                "bytes",
			"context":                              "context",
			"errors":                               "errors",
			"fmt":                                  "fmt",
			"github.com/inconshreveable/mousetrap": "mousetrap",
			"github.com/spf13/pflag":               "pflag",
			"io":                                   "io",
			"os":                                   "os",
			"path/filepath":                        "filepath",
			"reflect":                              "reflect",
			"sort":                                 "sort",
			"strconv":                              "strconv",
			"strings":                              "strings",
			"sync":                                 "sync",
			"text/template":                        "template",
			"time":                                 "time",
			"unicode":                              "unicode",
		},
		Interfaces: map[string]reflect.Type{},
		NamedTypes: map[string]reflect.Type{
			"Command":            reflect.TypeOf((*q.Command)(nil)).Elem(),
			"CompletionOptions":  reflect.TypeOf((*q.CompletionOptions)(nil)).Elem(),
			"FParseErrWhitelist": reflect.TypeOf((*q.FParseErrWhitelist)(nil)).Elem(),
			"PositionalArgs":     reflect.TypeOf((*q.PositionalArgs)(nil)).Elem(),
			"ShellCompDirective": reflect.TypeOf((*q.ShellCompDirective)(nil)).Elem(),
		},
		AliasTypes: map[string]reflect.Type{},
		Vars: map[string]reflect.Value{
			"EnableCommandSorting":     reflect.ValueOf(&q.EnableCommandSorting),
			"EnablePrefixMatching":     reflect.ValueOf(&q.EnablePrefixMatching),
			"MousetrapDisplayDuration": reflect.ValueOf(&q.MousetrapDisplayDuration),
			"MousetrapHelpText":        reflect.ValueOf(&q.MousetrapHelpText),
		},
		Funcs: map[string]reflect.Value{
			"AddTemplateFunc":     reflect.ValueOf(q.AddTemplateFunc),
			"AddTemplateFuncs":    reflect.ValueOf(q.AddTemplateFuncs),
			"AppendActiveHelp":    reflect.ValueOf(q.AppendActiveHelp),
			"ArbitraryArgs":       reflect.ValueOf(q.ArbitraryArgs),
			"CheckErr":            reflect.ValueOf(q.CheckErr),
			"CompDebug":           reflect.ValueOf(q.CompDebug),
			"CompDebugln":         reflect.ValueOf(q.CompDebugln),
			"CompError":           reflect.ValueOf(q.CompError),
			"CompErrorln":         reflect.ValueOf(q.CompErrorln),
			"Eq":                  reflect.ValueOf(q.Eq),
			"ExactArgs":           reflect.ValueOf(q.ExactArgs),
			"ExactValidArgs":      reflect.ValueOf(q.ExactValidArgs),
			"FixedCompletions":    reflect.ValueOf(q.FixedCompletions),
			"GetActiveHelpConfig": reflect.ValueOf(q.GetActiveHelpConfig),
			"Gt":                  reflect.ValueOf(q.Gt),
			"MarkFlagCustom":      reflect.ValueOf(q.MarkFlagCustom),
			"MarkFlagDirname":     reflect.ValueOf(q.MarkFlagDirname),
			"MarkFlagFilename":    reflect.ValueOf(q.MarkFlagFilename),
			"MarkFlagRequired":    reflect.ValueOf(q.MarkFlagRequired),
			"MatchAll":            reflect.ValueOf(q.MatchAll),
			"MaximumNArgs":        reflect.ValueOf(q.MaximumNArgs),
			"MinimumNArgs":        reflect.ValueOf(q.MinimumNArgs),
			"NoArgs":              reflect.ValueOf(q.NoArgs),
			"NoFileCompletions":   reflect.ValueOf(q.NoFileCompletions),
			"OnInitialize":        reflect.ValueOf(q.OnInitialize),
			"OnlyValidArgs":       reflect.ValueOf(q.OnlyValidArgs),
			"RangeArgs":           reflect.ValueOf(q.RangeArgs),
			"WriteStringAndCheck": reflect.ValueOf(q.WriteStringAndCheck),
		},
		TypedConsts: map[string]igop.TypedConst{
			"ShellCompDirectiveDefault":       {reflect.TypeOf(q.ShellCompDirectiveDefault), constant.MakeInt64(int64(q.ShellCompDirectiveDefault))},
			"ShellCompDirectiveError":         {reflect.TypeOf(q.ShellCompDirectiveError), constant.MakeInt64(int64(q.ShellCompDirectiveError))},
			"ShellCompDirectiveFilterDirs":    {reflect.TypeOf(q.ShellCompDirectiveFilterDirs), constant.MakeInt64(int64(q.ShellCompDirectiveFilterDirs))},
			"ShellCompDirectiveFilterFileExt": {reflect.TypeOf(q.ShellCompDirectiveFilterFileExt), constant.MakeInt64(int64(q.ShellCompDirectiveFilterFileExt))},
			"ShellCompDirectiveNoFileComp":    {reflect.TypeOf(q.ShellCompDirectiveNoFileComp), constant.MakeInt64(int64(q.ShellCompDirectiveNoFileComp))},
			"ShellCompDirectiveNoSpace":       {reflect.TypeOf(q.ShellCompDirectiveNoSpace), constant.MakeInt64(int64(q.ShellCompDirectiveNoSpace))},
		},
		UntypedConsts: map[string]igop.UntypedConst{
			"BashCompCustom":            {"untyped string", constant.MakeString(string(q.BashCompCustom))},
			"BashCompFilenameExt":       {"untyped string", constant.MakeString(string(q.BashCompFilenameExt))},
			"BashCompOneRequiredFlag":   {"untyped string", constant.MakeString(string(q.BashCompOneRequiredFlag))},
			"BashCompSubdirsInDir":      {"untyped string", constant.MakeString(string(q.BashCompSubdirsInDir))},
			"ShellCompNoDescRequestCmd": {"untyped string", constant.MakeString(string(q.ShellCompNoDescRequestCmd))},
			"ShellCompRequestCmd":       {"untyped string", constant.MakeString(string(q.ShellCompRequestCmd))},
		},
	})
}
