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

package api

import (
	"context"
	"errors"
	"fmt"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"github.com/zitadel/oidc/v2/pkg/client/profile"
	"github.com/zitadel/zitadel-go/v2/pkg/client/admin"
	"github.com/zitadel/zitadel-go/v2/pkg/client/management"
	zitadelMiddleware "github.com/zitadel/zitadel-go/v2/pkg/client/middleware"
	"github.com/zitadel/zitadel-go/v2/pkg/client/zitadel"
	adminPb "github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/admin"
	zitadelApp "github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/app"
	"github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/authn"
	managementPb "github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/management"
	"github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/object"
	zitadelOrg "github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/org"
	zitadelProject "github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/project"
	"github.com/zitadel/zitadel-go/v2/pkg/client/zitadel/user"
	"golang.org/x/oauth2"
)

const (
	zitadelOrgName     = "Percona"
	zitadelProjectName = "Everest"
	// TODO: rename to Everest
	zitadelSaName = "Everest1"
	// TODO: rename to Everest
	zitadelSaUsername        = "Everest1"
	zitadelSaSecretName      = "zitadel/proxy-service-account-json"
	zitadelWebAppName        = "Frontend"
	zitadelBackendAppName    = "Backend token introspect"
	zitadelBackendSecretName = "zitadel/introspect-key-json"
)

func (e *EverestServer) initZitadelOrganization(
	ctx context.Context, mngClient *management.Client, adminClient *admin.Client,
	orgName string,
) (string, error) {
	e.l.Debug("Creating Zitadel organization")
	org, err := mngClient.AddOrg(ctx, &managementPb.AddOrgRequest{Name: orgName})
	if err != nil && !isGrpcAlreadyExistsErr(err) {
		return "", errors.Join(err, errors.New("could not create a Zitadel new organization"))
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
			return "", errors.Join(err, errors.New("could not list Zitadel organizations"))
		}
		orgs := orgsRes.GetResult()
		if len(orgs) != 1 {
			return "", errors.Join(err, errors.New("could not find Zitadel organization in the list"))
		}

		orgID = orgs[0].Id
	}
	e.l.Debugf("orgID %s", orgID)

	return orgID, nil
}

func (e *EverestServer) initZitadelProject(
	ctx context.Context, mngClient *management.Client,
	orgID, projectName string,
) (string, error) {
	e.l.Debug("Creating Zitadel project")
	proj, err := mngClient.AddProject(
		zitadelMiddleware.SetOrgID(ctx, orgID),
		&managementPb.AddProjectRequest{
			Name:            projectName,
			HasProjectCheck: true,
		},
	)
	if err != nil && !isGrpcAlreadyExistsErr(err) {
		return "", errors.Join(err, errors.New("could not create a Zitadel new project"))
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
			return "", errors.Join(err, errors.New("could not list Zitadel projects"))
		}
		projs := projsRes.GetResult()
		if len(projs) != 1 {
			return "", errors.Join(err, errors.New("could not find Zitadel project in the list"))
		}

		projID = projs[0].Id
	}
	e.l.Debugf("projID %s", projID)

	return projID, nil
}

func (e *EverestServer) initZitadelServiceAccount(
	ctx context.Context, mngClient *management.Client,
	orgID, saUsername, saName, saSecretName string,
) error {
	e.l.Debug("Creating Zitadel service account")

	serviceAccount, err := mngClient.AddMachineUser(
		zitadelMiddleware.SetOrgID(ctx, orgID),
		&managementPb.AddMachineUserRequest{
			UserName:        saUsername,
			Name:            saName,
			AccessTokenType: user.AccessTokenType_ACCESS_TOKEN_TYPE_BEARER,
		},
	)
	if err != nil && !isGrpcAlreadyExistsErr(err) {
		return errors.Join(err, errors.New("could not create a new Zitadel service account"))
	}

	var serviceAccountID string
	if serviceAccount != nil {
		serviceAccountID = serviceAccount.UserId
	} else {
		e.l.Debug("Looking up Zitadel service account")
		saRes, err := mngClient.ListUsers(
			zitadelMiddleware.SetOrgID(ctx, orgID),
			&managementPb.ListUsersRequest{
				Query: &object.ListQuery{
					Offset: 0,
					Limit:  2,
				},
				Queries: []*user.SearchQuery{
					{
						Query: &user.SearchQuery_UserNameQuery{
							UserNameQuery: &user.UserNameQuery{UserName: saUsername},
						},
					},
					{
						Query: &user.SearchQuery_TypeQuery{
							TypeQuery: &user.TypeQuery{Type: user.Type_TYPE_MACHINE},
						},
					},
				},
			},
		)
		if err != nil {
			return errors.Join(err, errors.New("could not list Zitadel service accounts"))
		}
		sas := saRes.GetResult()
		if len(sas) != 1 {
			return errors.Join(err, errors.New("could not find Zitadel service account in the list"))
		}

		serviceAccountID = sas[0].Id
	}
	e.l.Debugf("serviceAccountID %s", serviceAccountID)

	mkList, err := mngClient.ListMachineKeys(
		zitadelMiddleware.SetOrgID(ctx, orgID),
		&managementPb.ListMachineKeysRequest{
			Query: &object.ListQuery{
				Offset: 0,
				Limit:  1,
			},
			UserId: serviceAccountID,
		},
	)
	if err != nil {
		return errors.Join(err, errors.New("could not list Zitadel machine keys for a service account"))
	}

	mk := mkList.GetResult()
	var serviceAccountJsonSecret string
	if len(mk) == 1 {
		e.l.Debugf("machineKeyID %s", mk[0].Id)
		e.l.Debug("Using service account secret from secrets storage")
		serviceAccountJsonSecret, err = e.secretsStorage.GetSecret(ctx, saSecretName)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Join(err, errors.New("could not retrieve service account json secret from secrets storage"))
		}
	}

	if len(mk) == 0 || serviceAccountJsonSecret == "" {
		e.l.Debug("Creating a new machine key for service account")
		mkRes, err := mngClient.AddMachineKey(
			zitadelMiddleware.SetOrgID(ctx, orgID),
			&managementPb.AddMachineKeyRequest{
				UserId: serviceAccountID,
				Type:   authn.KeyType_KEY_TYPE_JSON,
			},
		)
		if err != nil {
			return errors.Join(err, errors.New("could not create a machine key for service account"))
		}

		err = e.secretsStorage.SetSecret(ctx, saSecretName, string(mkRes.KeyDetails))
		if err != nil {
			return errors.Join(err, errors.New("could not store service account json key in secrets storage"))
		}

		serviceAccountJsonSecret = string(mkRes.KeyDetails)
	}

	e.serviceAccountProxyJsonSecret = []byte(serviceAccountJsonSecret)

	_, err = mngClient.AddOrgMember(
		zitadelMiddleware.SetOrgID(ctx, orgID),
		&managementPb.AddOrgMemberRequest{
			UserId: serviceAccountID,
			Roles:  []string{"ORG_USER_MANAGER"},
		},
	)
	if err != nil && !isGrpcAlreadyExistsErr(err) {
		return errors.Join(err, errors.New("could not add service account to organization"))
	}

	return nil
}

func (e *EverestServer) initZitadelBackendApp(
	ctx context.Context, mngClient *management.Client,
	orgID, projID, backendAppName, backendSecretName string,
) error {
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

	var (
		beAppID   string
		getNewKey bool
	)
	if be != nil {
		beAppID = be.AppId
		getNewKey = true
		if err := e.secretsStorage.SetSecret(ctx, backendSecretName, be.ClientSecret); err != nil {
			return errors.Join(err, errors.New("could not store Zitadel's backend application secret in secrets storage"))
		}
	} else {
		e.l.Debug("Looking up Zitadel backend application")
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
			return errors.Join(err, errors.New("could not find Zitadel backend application in the list"))
		}

		beAppID = apps[0].Id
	}
	e.l.Debugf("beAppID %s", beAppID)

	var beAppClientSecret string
	if !getNewKey {
		e.l.Debug("Retrieving backend app secret from secrets storage")
		beAppClientSecret, err = e.secretsStorage.GetSecret(ctx, backendSecretName)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Join(err, errors.New("could not retrieve Zitadel backend application secret from secrets storage"))
		}
	}

	if beAppClientSecret == "" || getNewKey {
		e.l.Debug("Creating a new backend app key")
		secretRes, err := mngClient.AddAppKey(
			zitadelMiddleware.SetOrgID(ctx, orgID),
			&managementPb.AddAppKeyRequest{
				ProjectId: projID,
				AppId:     beAppID,
				Type:      authn.KeyType_KEY_TYPE_JSON,
			},
		)
		if err != nil {
			return errors.Join(err, errors.New("could not create a new Zitadel backend application secret"))
		}

		if err := e.secretsStorage.SetSecret(ctx, backendSecretName, string(secretRes.KeyDetails)); err != nil {
			return errors.Join(err, errors.New("could not store a Zitadel backend application secret in secrets storage"))
		}

		beAppClientSecret = string(secretRes.KeyDetails)
	}

	e.serviceAccountIntrospectJsonSecret = []byte(beAppClientSecret)

	return nil
}

func (e *EverestServer) initZitadelReverseProxy() error {
	scopes := []string{
		"openid",
		zitadel.ScopeZitadelAPI(),
	}

	ts, err := profile.NewJWTProfileTokenSourceFromKeyFileData(
		e.config.Auth.Issuer,
		e.serviceAccountProxyJsonSecret,
		scopes,
	)
	if err != nil {
		return errors.Join(err, errors.New("could not initialize token source for Zitadel proxy"))
	}

	scheme := "https"
	if e.config.Auth.Insecure {
		scheme = "http"
	}
	rp := httputil.NewSingleHostReverseProxy(
		&url.URL{
			Host:   e.config.Auth.Hostname,
			Scheme: scheme,
		})

	client := oauth2.NewClient(context.Background(), ts)
	rp.Transport = client.Transport

	e.zitadelReverseProxy = rp

	return nil
}

func (e *EverestServer) proxyZitadel(ctx echo.Context) error {
	req := ctx.Request().Clone(ctx.Request().Context())
	req.URL.Path = strings.TrimPrefix(req.URL.Path, "/v1/zitadel/")
	e.zitadelReverseProxy.ServeHTTP(ctx.Response(), req)

	return nil
}

func (e *EverestServer) zitadelRedirectURIs() []string {
	webAppRedirectURIs := []string{fmt.Sprintf("%s/callback", e.config.URL)}
	return webAppRedirectURIs
}

func (e *EverestServer) zitadelLogoutRedirectURIs() []string {
	webAppLogoutRedirectURIs := []string{e.config.URL}
	return webAppLogoutRedirectURIs
}
