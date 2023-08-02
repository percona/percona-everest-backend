// Package api contains the API server implementation.
package api

//go:generate ../bin/oapi-codegen --config=server.cfg.yml  ../docs/spec/openapi.yml

import (
	"go.uber.org/zap"

	"github.com/percona/percona-everest-backend/cmd/config"
	"github.com/percona/percona-everest-backend/model"
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
	everestK8s     everestK8s
}

// List represents a general object with the list of items.
type List struct {
	Items string `json:"items"`
}

// NewEverestServer creates and configures everest API.
func NewEverestServer(c *config.EverestConfig, l *zap.SugaredLogger) (*EverestServer, error) {
	e := &EverestServer{
		config: c,
		l:      l,
	}
	err := e.initEverest()

	return e, err
}

func (e *EverestServer) initEverest() error {
	db, err := model.NewDatabase(pgStorageName, e.config.DSN, pgMigrationsDir)
	if err != nil {
		return err
	}
	e.storage = db
	e.secretsStorage = db // so far the db implements both interfaces - the regular storage and the secrets storage
	_, err = db.Migrate()
	e.everestK8s = newEverestK8s(e.storage, e.secretsStorage, e.l)
	return err
}
//func (e *EverestServer) proxyKubernetes(ctx echo.Context, kubernetesID, resourceName string) error {
//	cluster, err := e.storage.GetKubernetesCluster(ctx.Request().Context(), kubernetesID)
//	if err != nil {
//		e.l.Error(err)
//		return ctx.JSON(http.StatusInternalServerError, Error{
//			Message: pointer.ToString("Could not get a Kubernetes cluster"),
//		})
//	}
//	encodedSecret, err := e.secretsStorage.GetSecret(ctx.Request().Context(), kubernetesID)
//	if err != nil {
//		e.l.Error(err)
//		return ctx.JSON(http.StatusInternalServerError, Error{
//			Message: pointer.ToString("Could not retrieve kubeconfig"),
//		})
//	}
//
//	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", newConfigGetter(encodedSecret).loadFromString)
//	if err != nil {
//		e.l.Error(err)
//		return ctx.JSON(http.StatusBadRequest, Error{
//			Message: pointer.ToString("Could not build kubeconfig"),
//		})
//	}
//	reverseProxy := httputil.NewSingleHostReverseProxy(
//		&url.URL{
//			Host:   strings.TrimPrefix(config.Host, "https://"),
//			Scheme: "https",
//		})
//	transport, err := rest.TransportFor(config)
//	if err != nil {
//		e.l.Error(err)
//		return ctx.JSON(http.StatusBadRequest, Error{
//			Message: pointer.ToString("Could not create REST transport"),
//		})
//	}
//	reverseProxy.Transport = transport
//	req := ctx.Request()
//	req.URL.Path = buildProxiedURL(ctx.Request().URL.Path, kubernetesID, resourceName, cluster.Namespace)
//	reverseProxy.ServeHTTP(ctx.Response(), req)
//	return nil
//}
//
//func buildProxiedURL(uri, kubernetesID, resourceName, namespace string) string {
//	// cut the /kubernetes part
//	uri = strings.TrimPrefix(uri, "/v1/kubernetes/"+kubernetesID)
//
//	// cut the resource name if present
//	uri = strings.TrimSuffix(uri, resourceName)
//
//	// remove kebab-case
//	uri = strings.ReplaceAll(uri, "-", "")
//	return fmt.Sprintf("/apis/everest.percona.com/v1alpha1/namespaces/%s%s%s", namespace, uri, resourceName)
//}
