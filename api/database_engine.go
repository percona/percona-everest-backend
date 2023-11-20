// percona-everest-backend
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

// Package api ...
package api

import "github.com/labstack/echo/v4"

// ListDatabaseEngines List of the available database engines on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseEngines(ctx echo.Context, _ string) error {
	return e.proxyKubernetes(ctx, "")
}

// GetDatabaseEngine Get the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseEngine(ctx echo.Context, _ string, name string) error {
	return e.proxyKubernetes(ctx, name)
}

// UpdateDatabaseEngine Get the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseEngine(ctx echo.Context, _ string, name string) error {
	return e.proxyKubernetes(ctx, name)
}
