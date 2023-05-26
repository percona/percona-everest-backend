//go:build tools
// +build tools

package tools

import (
	_ "github.com/daixiang0/gci"
	_ "github.com/deepmap/oapi-codegen/cmd/oapi-codegen"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/quasilyte/go-consistent"
	_ "github.com/reviewdog/reviewdog/cmd/reviewdog"
	_ "golang.org/x/tools/cmd/goimports"
	_ "mvdan.cc/gofumpt"
)

//go:generate go build -o ../bin/gci github.com/daixiang0/gci
//go:generate go build -o ../bin/go-consistent github.com/quasilyte/go-consistent
//go:generate go build -o ../bin/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint
//go:generate go build -o ../bin/oapi-codegen github.com/deepmap/oapi-codegen/cmd/oapi-codegen
//go:generate go build -o ../bin/reviewdog github.com/reviewdog/reviewdog/cmd/reviewdog
//go:generate go build -o ../bin/gofumpt mvdan.cc/gofumpt
//go:generate go build -o ../bin/goimports golang.org/x/tools/cmd/goimports
