package api

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"github.com/zitadel/oidc/v2/pkg/client/profile"
	"golang.org/x/oauth2"
)

func (e *EverestServer) proxyZitadel(ctx echo.Context) error {
	// TODO: where to get the file from?
	keyPath := "~/Desktop/everest-sa.json"
	// TODO: where to get issuer?
	issuer := "localhost:8080"
	scopes := []string{"openid urn:zitadel:iam:org:project:id:zitadel:aud"}

	// TODO: cache this so it's not requested on every request
	ts, err := profile.NewJWTProfileTokenSourceFromKeyFile("http://"+issuer, keyPath, scopes)
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
	req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/zitadel/")
	rp.ServeHTTP(ctx.Response(), req)
	return nil
}
