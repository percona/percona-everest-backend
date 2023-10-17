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
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/AlekSi/pointer"
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
		beAppID           string
		beAppClientSecret string
	)
	if be != nil {
		beAppID = be.AppId
		beAppClientSecret = be.ClientSecret
		if err := e.secretsStorage.SetSecret(ctx, backendSecretName, be.ClientSecret); err != nil {
			return errors.Join(err, errors.New("could not store Zitadel's backend application secret in secrets storage"))
		}
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

	if beAppClientSecret == "" {
		e.l.Debug("Retrieving backend app secret from secrets storage")
		beAppClientSecret, err = e.secretsStorage.GetSecret(ctx, backendSecretName)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Join(err, errors.New("could not retrieve Zitadel BE application secret from secrets storage"))
		}
	}

	if beAppClientSecret == "" {
		e.l.Debug("Creating a new backend app client secret")
		secretRes, err := mngClient.RegenerateAPIClientSecret(
			zitadelMiddleware.SetOrgID(ctx, orgID),
			&managementPb.RegenerateAPIClientSecretRequest{
				ProjectId: projID,
				AppId:     beAppID,
			},
		)
		if err != nil {
			return errors.Join(err, errors.New("could not create a new Zitadel BE application secret"))
		}

		if err := e.secretsStorage.SetSecret(ctx, backendSecretName, secretRes.ClientSecret); err != nil {
			return errors.Join(err, errors.New("could not store a Zitadel BE application secret in secrets storage"))
		}

		beAppClientSecret = secretRes.ClientSecret
	}

	e.serviceAccountIntrospectJsonSecret = []byte(beAppClientSecret)

	return nil
}

func (e *EverestServer) proxyZitadel(ctx echo.Context) error {
	// TODO: where to get issuer?
	issuer := "localhost:8080"
	scopes := []string{
		"openid",
		zitadel.ScopeZitadelAPI(),
		// zitadel.ScopeProjectID("236525724985458690"),
	}

	// everest-sa
	// e.serviceAccountProxyJsonSecret = []byte(`{"type":"serviceaccount","keyId":"234896340004372483","key":"-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA1UC7RbdjN4MhtMYu4AihehhRbYIUb4hpqULay81UzcUKG8B9\n9ftHzyB+WZRFEPLso2bab45vlErupFiDsVERnVhHzgl4uyVhMf1R2nSZcqwzuhq4\nEOJQ4fFKNMcpn56gkvwTHN7o63AQMOA84bx7YBzlYPsmG8qzUmPDfASelNXy8vpz\nzbympeo2CHt4Ark/P2JEX4yW+iAXZUW1dR6DcWB12RYPxO8W+RNkjAUYN627wj0d\ndDS//1YfnFeDu11CWMvnWZaoyhMxfDLGNnB1hA1WLVb000K5r03r6mngIr/ItW60\nuBZ9NHBrHeSUC4TERSrWsycx0h9fuBLYTZq4ZQIDAQABAoIBABrOTjwPN0uNEjmV\nB+NlclbUo7euOD9k2FNMchBYOSa8c+7VHYBEG9yvavJ7rsrYnmJT1XVcZC4x1RmX\nfsZVOG+c3znI+wIbSsJr41QggAFoIABux2Bn8l7UY82Kk3LbD7gqM4TXiFO//GkI\nZt7BQIjuWO794uZvbmcW30XBluWCXIXPd46jZntFMd4tVxa/WIccjcdgkytZKv+k\nBIJbnN6agpFZGt3rK6tBVyUbH3d4YDiVI00TQdJm7RzpjUQOgsMnT2RAib+bJC94\nR5VHRXZPtgGbg0D3t58rzUSP/2XlhUx6RzrFgoguPmQTot6ZU4EITCGr0MMgJ2Is\nVvd3DcECgYEA4TB/gBLJE3ECoWECvcO494Tbp+CyD2c5kVnIUwpSzhBukcrrvlI8\nYeWJrqe0fQcwXBws6lKidM2sA/iizvJ2mcnpyKpmukEwRBLwcSCHVA8HBo3Mgaug\nNVjWqFgrz/ueNqMVQdwDDwsAQzpev5cGVRfKdPerjjn4upweLZUGNi0CgYEA8m4k\nRWS34BENtL5eJlLwJLkv4gttn8DB5+hXx0bOd5gHq2DGL6sRYEoDU92oCO3+OOYZ\nRwNJXNsAswtaZS52c04eo1QDjlIKpfsNCUZsjpt1rP7FZ4SFObqb614e3VEN9LST\noZCKGSYeuGFA4AIBcgZp+p/zo6CTNPqG71d/5hkCgYAsfz7SeePNuakBZn/6K3Cj\nSFd3JslIjecsN4eEESgnm8udd3F53BoeZhL8thrOEduWd+LQMp/zYKi66CiTqAmT\nffh6NGG0MClvaiak0/6pt4Z13xMoFFfF8tYH0dRmdpvew/7xUp4wHMZigmgyh48y\nxU62KjJ2GjJx/WNhMm9VyQKBgQCWNfaVZK2l0Qs8BYRSnKsdJf1sQwZ+qLG83rKc\nz9uYMIP4BTNnT8ipb9KWAU5fkg8l9DSPUpL/TNcnGQ6+iMZt9WZ4btLxORZN97sB\nFzimN972/LkVxf/CYETB1oSrPtC14bljrypSINOCDQhkg/mfTCgYWUleBl2Pwvce\nj4m46QKBgQCG+peVdw8AbfWHLsX3XItr+1SxHpXqTAVVWii1ghqDoVwXyxQCqOy2\nPkjLX4MjryyVWtIadU6RZGrcI8O+A9cjmt/kOUSxDX+nBfapbpp9yGlqoZ/7wm9r\nwZY0VR870ea+Ndf4iBN0SXub/4Mjx5LIzMidS7NxE30rskYaIpIDQw==\n-----END RSA PRIVATE KEY-----\n","userId":"234888754085494787"}`)
	// evtest
	// e.serviceAccountProxyJsonSecret = []byte(`{"type":"serviceaccount","keyId":"236523842263056386","key":"-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAqkYMAgBLpROBAoifSCJ3CVMnuBkZVgP2SEmYjS6fcUkXqUOn\nUtsK337Dam/IHhSfXidw7zYjlCdmqPzqn5bz0YUSm7pH7XtC+nxAAQsCGc1NlPN/\nlaPsWtOPqerzxH3DGFQvcNqVBXF5cde9fmMj6e6RnVgbXI3KF2ePbtO8qKqIaRSt\nvRnN8aJuNjmEkkPobvMFvamWt0GbRf78Qvm6ePdRWu5QEeXGK9Lt2nRxRMo/UWkz\naVusCUFKbhjJcxmEd9ZtNvc2IZCCSO6F+CRao2vTXXl6j/oW/HSXiyh1z+1UTXsa\nFt8m9GvAme7rgIqMu7QE3Bb9dI5W6yLWJczM/QIDAQABAoIBAEYwP6zngEcYxhpM\nRRRQGK+AVqQdvILneTMNG1Q/PrxM+/LrD2MpJc9BCr6qO1yi9ZqzOWtx7rKYl0nb\nj7+fUvwwFZ6Z6CJtqAtnAl8rsX7/URawVQxTGQ/Lm7HYRwndKXmy4idsAvfOcdhK\nrTMXHOvGSsIIWqcJT5/cMZTmtSL1FMslLzZjmzBy+oqfB1UxVNJfwpl/tWFSDH6j\nCOBqPLXRCbulLYaGerMwR2lcbQJVU4P5+tRfJub3F/0mq2SoS+6aGwBoPoLgnqrQ\n7+5TCB6zKpf87CQmhzureOAfy6DRPtsMO2ZQsPQr5z/dLo45R5fO0xQnZyybf1bT\nOy8WLWECgYEAz/sFzd9i/8NXsavF1qEp7aFdijJKWWrGZZvowxj2k9AtnxtgHkcH\nnDa1lPGlQaSFjSTZ5DxVuRX9XjLZGQVu19uo45SsdMIg2MWfx1xSDKs6G9heJn/R\nBtEF3m2dKa3P8Ag84IYFph5dulAeLOW+AAc8LqnfEiCo1rrHmdzE05UCgYEA0ZZN\nxjL0Ucuklw6EObHoY/kQlDL3Kc5PlTdQuX8KlEUtIAE7MNG4Fhk1sgvOeMjjw+RH\nri5cDB4zvOLND0g/gIvvbyvnPxLbjZ7K2PWlnEsBEC2PgWzSbY5PZieq+CFZMv3K\no6Ms/uPc7J6V5xopM0lavuw7ZNdVsK9jnzHsuckCgYAGGkSCVPKvtIinMvYcJSB4\n04pOGsmptANcSeXbi6j4j1w3VfNNECJ+B/DuDOUfdvdgO9uU4dxWEPodQHq0TD+D\nX/OlseAZkPSrx6i3jdLugjuzQ3cHxCpa+9kjPK4m4e2/Ck7W+7fAtxVi+STZhmg7\n0fqHF/7upjyuCE8BCcRQvQKBgQCnsK7Bqfs5hspF4mOBFgtuEdVl/fEsDdo29W8t\nO6xnPYIBXXrScLntVHZV4oRst68lCP0hLA6R04hp1L1lQNUuMMh+Fo6LNLdd9HMw\nbDr5djl/jDSJxVwINBjrD0oIBgasecssal6SAha9a5Vctt3IHyTwJWrQIEp7d5kp\nwnQ5oQKBgF2w886c8+WITaPoXGQeiv3We5jZu6Ra0MJyFxMcPPEuxZt0S+O3iHgv\nn/rLVWUdAt4h+flqerB58DBas/ZABbl6UGlqSWxwKZpmNWeUkZ3y52HDXUeR6x/m\n3oIsfD5nVJ0lYppQ8bWtt2frhlm/6Gk/raTS+wB8ffTGNrPkiFVS\n-----END RSA PRIVATE KEY-----\n","userId":"236523829780807682"}`)
	// TODO: cache this so it's not requested on every request
	ts, err := profile.NewJWTProfileTokenSourceFromKeyFileData(
		"http://"+issuer,
		e.serviceAccountProxyJsonSecret,
		scopes,
	)
	// ts, err := profile.NewJWTProfileTokenSourceFromKeyFile("http://"+issuer, keyPath, scopes)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	client := oauth2.NewClient(context.Background(), ts)
	rp := httputil.NewSingleHostReverseProxy(
		&url.URL{
			Host:   issuer,
			Scheme: "http",
		})

	rp.Transport = client.Transport

	req := ctx.Request()
	req.URL.Path = strings.TrimPrefix(req.URL.Path, "/v1/zitadel/")
	rp.ServeHTTP(ctx.Response(), req)
	return nil
}
