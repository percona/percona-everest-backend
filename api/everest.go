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

// Package api contains the API server implementation.
package api

//go:generate ../bin/oapi-codegen --config=server.cfg.yml  ../docs/spec/openapi.yml

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"sync"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/zitadel/oidc/pkg/oidc"
	"github.com/zitadel/zitadel-go/v2/pkg/client/admin"
	"github.com/zitadel/zitadel-go/v2/pkg/client/management"
	zitadelMiddleware "github.com/zitadel/zitadel-go/v2/pkg/client/middleware"
	"github.com/zitadel/zitadel-go/v2/pkg/client/zitadel"
	adminPb "github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/admin"
	zitadelApp "github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/app"
	managementPb "github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/management"
	"github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/object"
	zitadelOrg "github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/org"
	zitadelProject "github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/project"
	"github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/percona/percona-everest-backend/cmd/config"
	"github.com/percona/percona-everest-backend/model"
	"github.com/percona/percona-everest-backend/pkg/kubernetes"
	"github.com/percona/percona-everest-backend/public"
)

const (
	pgStorageName   = "postgres"
	pgMigrationsDir = "migrations"
)

// EverestServer represents the server struct.
type EverestServer struct {
	config         *config.EverestConfig
	l              *zap.SugaredLogger
	storage        storage
	secretsStorage secretsStorage
	waitGroup      *sync.WaitGroup
	echo           *echo.Echo

	publicConfiguration *Configuration
}

// NewEverestServer creates and configures everest API.
func NewEverestServer(c *config.EverestConfig, l *zap.SugaredLogger) (*EverestServer, error) {
	e := &EverestServer{
		config:    c,
		l:         l,
		echo:      echo.New(),
		waitGroup: &sync.WaitGroup{},
	}
	if err := e.initHTTPServer(); err != nil {
		return e, err
	}

	if err := e.initEverest(); err != nil {
		return e, err
	}

	if err := e.initZitadel(context.TODO()); err != nil {
		return e, err
	}

	return e, nil
}

func (e *EverestServer) initEverest() error {
	db, err := model.NewDatabase(pgStorageName, e.config.DSN, pgMigrationsDir)
	if err != nil {
		return err
	}
	e.storage = db
	e.secretsStorage = db // so far the db implements both interfaces - the regular storage and the secrets storage
	_, err = db.Migrate()
	return err
}

func (e *EverestServer) initKubeClient(ctx context.Context, kubernetesID string) (*model.KubernetesCluster, *kubernetes.Kubernetes, int, error) {
	k, err := e.storage.GetKubernetesCluster(ctx, kubernetesID)
	if err != nil {
		e.l.Error(err)
		return nil, nil, http.StatusBadRequest, errors.New("could not find Kubernetes cluster")
	}

	kubeClient, err := kubernetes.NewFromSecretsStorage(
		ctx, e.secretsStorage, k.ID,
		k.Namespace, e.l,
	)
	if err != nil {
		e.l.Error(err)
		return k, nil, http.StatusInternalServerError, errors.New("could not create Kubernetes client from kubeconfig")
	}

	return k, kubeClient, 0, nil
}

// initHTTPServer configures http server for the current EverestServer instance.
func (e *EverestServer) initHTTPServer() error {
	swagger, err := GetSwagger()
	if err != nil {
		return err
	}
	fsys, err := fs.Sub(public.Static, "dist")
	if err != nil {
		return errors.Join(err, errors.New("error reading filesystem"))
	}
	staticFilesHandler := http.FileServer(http.FS(fsys))
	indexFS := echo.MustSubFS(public.Index, "dist")
	// FIXME: Ideally it should be redirected to /everest/ and FE app should be served using this endpoint.
	//
	// We tried to do this with Fabio and FE app requires the following changes to be implemented:
	// 1. Add basePath configuration for react router
	// 2. Add apiUrl configuration for FE app
	//
	// Once it'll be implemented we can serve FE app on /everest/ location
	e.echo.FileFS("/*", "index.html", indexFS)
	e.echo.GET("/favicon.ico", echo.WrapHandler(staticFilesHandler))
	e.echo.GET("/assets-manifest.json", echo.WrapHandler(staticFilesHandler))
	e.echo.GET("/static/*", echo.WrapHandler(staticFilesHandler))
	// Log all requests
	e.echo.Use(echomiddleware.Logger())
	e.echo.Pre(echomiddleware.RemoveTrailingSlash())

	basePath, err := swagger.Servers.BasePath()
	if err != nil {
		return errors.Join(err, errors.New("could not get base path"))
	}

	// Use our validation middleware to check all requests against the OpenAPI schema.
	apiGroup := e.echo.Group(basePath)
	apiGroup.Use(middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{
		SilenceServersWarning: true,
	}))
	RegisterHandlers(apiGroup, e)

	return nil
}

func (e *EverestServer) initZitadel(ctx context.Context) error {
	issuer := "http://localhost:8080"
	api := "localhost:8080"
	keyFile := "/home/ceecko/Desktop/everest-sa.json"
	orgName := "Percona"
	projectName := "Everest"
	saName := "Everest"
	saUsername := "Everest"
	webAppName := "Frontend"
	webAppRedirectURIs := []string{"http://localhost:8081"}
	backendAppName := "Backend token introspect"

	e.l.Info("Initializing Zitadel instance")

	mngClient, err := management.NewClient(
		issuer,
		api,
		[]string{oidc.ScopeOpenID, zitadel.ScopeZitadelAPI()},
		zitadel.WithJWTProfileTokenSource(zitadelMiddleware.JWTProfileFromPath(keyFile)),
		// TODO: make this secure
		zitadel.WithInsecure(),
	)
	if err != nil {
		return errors.Join(err, errors.New("could not create Zitadel management client"))
	}
	defer mngClient.Connection.Close()

	adminClient, err := admin.NewClient(
		issuer,
		api,
		[]string{oidc.ScopeOpenID, zitadel.ScopeZitadelAPI()},
		zitadel.WithJWTProfileTokenSource(zitadelMiddleware.JWTProfileFromPath(keyFile)),
		// TODO: make this secure
		zitadel.WithInsecure(),
	)
	if err != nil {
		return errors.Join(err, errors.New("could not create Zitadel admin client"))
	}
	defer adminClient.Connection.Close()

	e.l.Debug("Creating Zitadel organization")
	org, err := mngClient.AddOrg(ctx, &managementPb.AddOrgRequest{Name: orgName})
	if err != nil && !isGrpcAlreadyExistsErr(err) {
		return errors.Join(err, errors.New("could not create a Zitadel new organization"))
	}
	var orgID string
	if org != nil {
		orgID = org.Id
	} else {
		e.l.Debug("Looking up Zitadel organization")
		orgsRes, err := adminClient.ListOrgs(
			ctx,
			&adminPb.ListOrgsRequest{
				Query: &object.ListQuery{
					Limit: 2,
				},
				Queries: []*zitadelOrg.OrgQuery{
					{
						Query: &zitadelOrg.OrgQuery_NameQuery{
							NameQuery: &zitadelOrg.OrgNameQuery{Name: orgName},
						},
					},
				},
			},
		)
		if err != nil {
			return errors.Join(err, errors.New("could not list Zitadel organizations"))
		}
		orgs := orgsRes.GetResult()
		if len(orgs) != 1 {
			return errors.Join(err, errors.New("could not find Zitadel organization in the list"))
		}

		orgID = orgs[0].Id
	}
	e.l.Debugf("orgID %s", orgID)

	e.l.Debug("Creating Zitadel project")
	proj, err := mngClient.AddProject(
		zitadelMiddleware.SetOrgID(ctx, orgID),
		&managementPb.AddProjectRequest{
			Name:            projectName,
			HasProjectCheck: true,
		},
	)
	if err != nil && !isGrpcAlreadyExistsErr(err) {
		return errors.Join(err, errors.New("could not create a Zitadel new project"))
	}
	var projID string
	if proj != nil {
		projID = proj.Id
	} else {
		e.l.Debug("Looking up Zitadel project")
		projsRes, err := mngClient.ListProjects(
			zitadelMiddleware.SetOrgID(ctx, orgID),
			&managementPb.ListProjectsRequest{
				Query: &object.ListQuery{
					Offset: 0,
					Limit:  2,
				},
				Queries: []*zitadelProject.ProjectQuery{
					{
						Query: &zitadelProject.ProjectQuery_NameQuery{
							NameQuery: &zitadelProject.ProjectNameQuery{Name: projectName},
						},
					},
				},
			},
		)
		if err != nil {
			return errors.Join(err, errors.New("could not list Zitadel projects"))
		}
		projs := projsRes.GetResult()
		if len(projs) != 1 {
			return errors.Join(err, errors.New("could not find Zitadel project in the list"))
		}

		projID = projs[0].Id
	}
	e.l.Debugf("projID %s", projID)

	e.l.Debug("Creating Zitadel service account")
	_, err = mngClient.AddMachineUser(
		zitadelMiddleware.SetOrgID(ctx, orgID),
		&managementPb.AddMachineUserRequest{
			UserName:        saUsername,
			Name:            saName,
			AccessTokenType: user.AccessTokenType_ACCESS_TOKEN_TYPE_JWT,
		},
	)
	if err != nil && !isGrpcAlreadyExistsErr(err) {
		return errors.Join(err, errors.New("could not create a new Zitadel service account"))
	}

	e.l.Debug("Creating Zitadel web application")
	fe, err := mngClient.AddOIDCApp(
		zitadelMiddleware.SetOrgID(ctx, orgID),
		&managementPb.AddOIDCAppRequest{
			ProjectId:     projID,
			Name:          webAppName,
			AppType:       zitadelApp.OIDCAppType_OIDC_APP_TYPE_WEB,
			ResponseTypes: []zitadelApp.OIDCResponseType{zitadelApp.OIDCResponseType_OIDC_RESPONSE_TYPE_CODE},
			GrantTypes: []zitadelApp.OIDCGrantType{
				zitadelApp.OIDCGrantType_OIDC_GRANT_TYPE_AUTHORIZATION_CODE,
				zitadelApp.OIDCGrantType_OIDC_GRANT_TYPE_REFRESH_TOKEN,
			},
			RedirectUris:    webAppRedirectURIs,
			AuthMethodType:  zitadelApp.OIDCAuthMethodType_OIDC_AUTH_METHOD_TYPE_NONE,
			AccessTokenType: zitadelApp.OIDCTokenType_OIDC_TOKEN_TYPE_BEARER,
		},
	)
	if err != nil && !isGrpcAlreadyExistsErr(err) {
		return errors.Join(err, errors.New("could not create a new Zitadel web application"))
	}

	var (
		feAppID       string
		feAppClientID string
	)
	if fe != nil {
		feAppID = fe.AppId
		feAppClientID = fe.ClientId
	} else {
		e.l.Debug("Looking up Zitadel FE application")
		appsRes, err := mngClient.ListApps(
			zitadelMiddleware.SetOrgID(ctx, orgID),
			&managementPb.ListAppsRequest{
				ProjectId: projID,
				Query: &object.ListQuery{
					Offset: 0,
					Limit:  2,
				},
				Queries: []*zitadelApp.AppQuery{
					{
						Query: &zitadelApp.AppQuery_NameQuery{
							NameQuery: &zitadelApp.AppNameQuery{Name: webAppName},
						},
					},
				},
			},
		)
		if err != nil {
			return errors.Join(err, errors.New("could not list Zitadel applications"))
		}
		apps := appsRes.GetResult()
		if len(apps) != 1 {
			return errors.Join(err, errors.New("could not find Zitadel FE application in the list"))
		}

		feAppID = apps[0].Id
		feAppClientID = apps[0].GetOidcConfig().ClientId
	}
	e.l.Debugf("feAppID %s", feAppID)

	e.l.Debug("Creating Zitadel backend application")
	be, err := mngClient.AddAPIApp(
		zitadelMiddleware.SetOrgID(ctx, orgID),
		&managementPb.AddAPIAppRequest{
			ProjectId:      projID,
			Name:           backendAppName,
			AuthMethodType: zitadelApp.APIAuthMethodType_API_AUTH_METHOD_TYPE_PRIVATE_KEY_JWT,
		},
	)
	if err != nil && !isGrpcAlreadyExistsErr(err) {
		return errors.Join(err, errors.New("could not create a new Zitadel backend application"))
	}

	var beAppID string
	if be != nil {
		beAppID = fe.AppId
	} else {
		e.l.Debug("Looking up Zitadel BE application")
		appsRes, err := mngClient.ListApps(
			zitadelMiddleware.SetOrgID(ctx, orgID),
			&managementPb.ListAppsRequest{
				ProjectId: projID,
				Query: &object.ListQuery{
					Offset: 0,
					Limit:  2,
				},
				Queries: []*zitadelApp.AppQuery{
					{
						Query: &zitadelApp.AppQuery_NameQuery{
							NameQuery: &zitadelApp.AppNameQuery{Name: backendAppName},
						},
					},
				},
			},
		)
		if err != nil {
			return errors.Join(err, errors.New("could not list Zitadel applications"))
		}
		apps := appsRes.GetResult()
		if len(apps) != 1 {
			return errors.Join(err, errors.New("could not find Zitadel BE application in the list"))
		}

		beAppID = apps[0].Id
	}
	e.l.Debugf("beAppID %s", beAppID)

	e.publicConfiguration = &Configuration{
		Auth: AuthConfiguration{
			Web: &WebAuthConfiguration{
				ClientID: feAppClientID,
				Url:      issuer,
			},
		},
	}
	e.l.Info("Zitadel initialization finished")

	return nil
}

func isGrpcAlreadyExistsErr(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}

	if s.Code() == codes.AlreadyExists {
		return true
	}

	return false
}

// Start starts everest server.
func (e *EverestServer) Start() error {
	return e.echo.Start(fmt.Sprintf("0.0.0.0:%d", e.config.HTTPPort))
}

// Shutdown gracefully stops the Everest server.
func (e *EverestServer) Shutdown(ctx context.Context) error {
	e.l.Info("Shutting down http server")
	if err := e.echo.Shutdown(ctx); err != nil {
		e.l.Error(errors.Join(err, errors.New("could not shut down http server")))
	} else {
		e.l.Info("http server shut down")
	}

	e.l.Info("Shutting down Everest")
	e.waitGroup.Wait()

	e.waitGroup.Add(1)
	go func() {
		defer e.waitGroup.Done()
		e.l.Info("Shutting down database storage")
		if err := e.storage.Close(); err != nil {
			e.l.Error(errors.Join(err, errors.New("could not shut down database storage")))
		} else {
			e.l.Info("Database storage shut down")
		}
	}()

	e.waitGroup.Add(1)
	go func() {
		defer e.waitGroup.Done()
		e.l.Info("Shutting down secrets storage")
		if err := e.secretsStorage.Close(); err != nil {
			e.l.Error(errors.Join(err, errors.New("could not shut down secret storage")))
		} else {
			e.l.Info("Secret storage shut down")
		}
	}()

	done := make(chan struct{}, 1)
	go func() {
		e.waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *EverestServer) getBodyFromContext(ctx echo.Context, into any) error {
	// GetBody creates a copy of the body to avoid "spoiling" the request before proxing
	reader, err := ctx.Request().GetBody()
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(into); err != nil {
		return errors.Join(err, errors.New("could not decode body"))
	}
	return nil
}
