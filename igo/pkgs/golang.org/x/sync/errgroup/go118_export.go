// export by github.com/goplus/igop/cmd/qexp

//go:build go1.18
// +build go1.18

package errgroup

import (
	q "golang.org/x/sync/errgroup"

	"reflect"

	"github.com/goplus/igop"
)

func init() {
	igop.RegisterPackage(&igop.Package{
		Name: "errgroup",
		Path: "golang.org/x/sync/errgroup",
		Deps: map[string]string{
			"context": "context",
			"fmt":     "fmt",
			"sync":    "sync",
		},
		Interfaces: map[string]reflect.Type{},
		NamedTypes: map[string]reflect.Type{
			"Group": reflect.TypeOf((*q.Group)(nil)).Elem(),
		},
		AliasTypes: map[string]reflect.Type{},
		Vars:       map[string]reflect.Value{},
		Funcs: map[string]reflect.Value{
			"WithContext": reflect.ValueOf(q.WithContext),
		},
		TypedConsts:   map[string]igop.TypedConst{},
		UntypedConsts: map[string]igop.UntypedConst{},
	})
}
