// percona-everest-cli
// Copyright (C) 2023 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Package install holds the main logic for installation commands.

// Package version implements version reporting command to the end user.
package version

import (
	"fmt"
	"strings"
)

const (
	devCatalogImage     = "docker.io/percona/everest-catalog:latest"
	releaseCatalogImage = "docker.io/percona/everest-catalog:%s"
)

var (
	// ProjectName is a component name, e.g. everestctl.
	ProjectName string //nolint:gochecknoglobals
	// Version is a component version e.g. v0.3.0-1-a93bef.
	Version string //nolint:gochecknoglobals
	// FullCommit is a git commit hash.
	FullCommit string //nolint:gochecknoglobals
	// CatalogImage is a image path for OLM catalog.
	catalogImage string //nolint:gochecknoglobals
)

// CatalogImage returns a catalog image needed for the build of everestctl
//
// for dev builds it returns everest-catalog:latest
// for the release it returns everest-catalog:X.Y.Z.
func CatalogImage() string {
	catalogImage = fmt.Sprintf(releaseCatalogImage, Version)
	if strings.Contains(Version, "dirty") {
		catalogImage = devCatalogImage
	}
	return catalogImage
}

// FullVersionInfo returns full version report.
func FullVersionInfo() string {
	out := []string{
		"ProjectName: " + ProjectName,
		"Version: " + Version,
		"FullCommit: " + FullCommit,
	}
	return strings.Join(out, "\n")
}
