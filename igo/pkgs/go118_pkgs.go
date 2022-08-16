//go:build go1.18
// +build go1.18

package pkgs

/*
qexp -outdir . -addtags "//+build go1.18" -filename go118_export github.com/compose-spec/compose-go/types
qexp -outdir . -addtags "//+build go1.18" -filename go118_export github.com/compose-spec/compose-go/loader
qexp -outdir . -addtags "//+build go1.18" -filename go118_export golang.org/x/sync/errgroup
qexp -outdir . -addtags "//+build go1.18" -filename go118_export github.com/opencontainers/go-digest
qexp -outdir . -addtags "//+build go1.18" -filename go118_export github.com/mitchellh/mapstructure
qexp -outdir . -addtags "//+build go1.18" -filename go118_export github.com/docker/go-connections/nat
qexp -outdir . -addtags "//+build go1.18" -filename go118_export github.com/distribution/distribution/v3/reference
qexp -outdir . -addtags "//+build go1.18" -filename go118_export github.com/distribution/distribution/v3/digestset
*/

import (
	_ "github.com/docker/compose/v2/igo/pkgs/github.com/compose-spec/compose-go/types"
	_ "github.com/docker/compose/v2/igo/pkgs/github.com/distribution/distribution/v3/digestset"
	_ "github.com/docker/compose/v2/igo/pkgs/github.com/distribution/distribution/v3/reference"
	_ "github.com/docker/compose/v2/igo/pkgs/github.com/docker/go-connections/nat"
	_ "github.com/docker/compose/v2/igo/pkgs/github.com/mitchellh/mapstructure"
	_ "github.com/docker/compose/v2/igo/pkgs/github.com/opencontainers/go-digest"
	_ "github.com/docker/compose/v2/igo/pkgs/golang.org/x/sync/errgroup"
)
