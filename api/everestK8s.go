package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	perconak8s "github.com/percona/percona-everest-backend/pkg/kubernetes"
)

type everestK8sImpl struct {
	storage        storage
	secretsStorage secretsStorage
	l              *zap.SugaredLogger
}

// NewEverestK8s creates Everest+k8s communication interface.
func newEverestK8s(storage storage, secretsStorage secretsStorage, l *zap.SugaredLogger) *everestK8sImpl {
	return &everestK8sImpl{
		storage:        storage,
		secretsStorage: secretsStorage,
		l:              l,
	}
}

// ApplyObjectStorage creates k8s objects in the given k8s cluster.
func (e *everestK8sImpl) ApplyObjectStorage(ctx echo.Context, kubernetesID string, bs BackupStorage, secretFields map[string]string) error { //nolint:funlen
	k, err := e.storage.GetKubernetesCluster(ctx.Request().Context(), kubernetesID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get a Kubernetes cluster"),
		})
	}

	everestClient, err := perconak8s.NewFromSecretsStorage(
		ctx.Request().Context(), e.secretsStorage, k.ID, k.Namespace, e.l)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not create kubernetes client"),
		})
	}

	secretName := buildSecretName(bs.Name)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: k.Namespace,
		},
		StringData: secretFields,
		Type:       corev1.SecretTypeOpaque,
	}
	_, err = everestClient.CreateSecret(ctx.Request().Context(), secret)
	// if such Secret is already present in k8s - consider it as created and do nothing (fixme)
	if err != nil && !k8serr.IsAlreadyExists(err) {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to create k8s Secret"),
		})
	}

	var url string
	if bs.Url != nil {
		url = *bs.Url
	}
	backupStorage := &everestv1alpha1.ObjectStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bs.Name,
			Namespace: k.Namespace,
		},
		Spec: everestv1alpha1.ObjectStorageSpec{
			Type:                  everestv1alpha1.ObjectStorageType(bs.Type),
			Bucket:                bs.BucketName,
			Region:                bs.Region,
			EndpointURL:           url,
			CredentialsSecretName: secretName,
		},
	}

	err = everestClient.CreateObjectStorage(ctx.Request().Context(), backupStorage)
	// if such ObjectStorage is already present in k8s - consider it as created and do nothing (fixme)
	if err != nil && !k8serr.IsAlreadyExists(err) {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to create ObjectStorage resource"),
		})
	}

	return nil
}

// RemoveObjectStorage removes k8s objects from the given k8s cluster.
func (e *everestK8sImpl) RemoveObjectStorage(ctx echo.Context, kubernetesID, storageName string) error {
	k, err := e.storage.GetKubernetesCluster(ctx.Request().Context(), kubernetesID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get a Kubernetes cluster"),
		})
	}

	everestClient, err := perconak8s.NewFromSecretsStorage(
		ctx.Request().Context(), e.secretsStorage, k.ID, k.Namespace, e.l)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not create kubernetes client"),
		})
	}

	dbClusters, err := e.getDBClustersByObjectStorage(ctx.Request().Context(), everestClient, storageName)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to check if ObjectStorage is in use"),
		})
	}

	if err = buildObjectStorageInUseError(dbClusters, storageName); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	err = everestClient.DeleteObjectStorage(ctx.Request().Context(), storageName, k.Namespace)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to delete ObjectStorage"),
		})
	}

	secretName := buildSecretName(storageName)
	err = everestClient.DeleteSecret(ctx.Request().Context(), secretName, k.Namespace)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to delete Secret"),
		})
	}

	return nil
}

func (e *everestK8sImpl) ProxyKubernetes(ctx echo.Context, kubernetesID, resourceName string) error {
	cluster, err := e.storage.GetKubernetesCluster(ctx.Request().Context(), kubernetesID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get a Kubernetes cluster"),
		})
	}
	encodedSecret, err := e.secretsStorage.GetSecret(ctx.Request().Context(), kubernetesID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not retrieve kubeconfig"),
		})
	}

	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", newConfigGetter(encodedSecret).loadFromString)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not build kubeconfig"),
		})
	}
	reverseProxy := httputil.NewSingleHostReverseProxy(
		&url.URL{
			Host:   strings.TrimPrefix(config.Host, "https://"),
			Scheme: "https",
		})
	transport, err := rest.TransportFor(config)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not create REST transport"),
		})
	}
	reverseProxy.Transport = transport
	req := ctx.Request()
	req.URL.Path = buildProxiedURL(ctx.Request().URL.Path, kubernetesID, resourceName, cluster.Namespace)
	reverseProxy.ServeHTTP(ctx.Response(), req)
	return nil
}

func (e *everestK8sImpl) getDBClustersByObjectStorage(ctx context.Context, everestClient *perconak8s.Kubernetes, storageName string) ([]everestv1alpha1.DatabaseCluster, error) {
	list, err := everestClient.ListDatabaseClusters(ctx)
	if err != nil {
		return nil, err
	}

	dbClusters := make([]everestv1alpha1.DatabaseCluster, 0, len(list.Items))
	for _, dbCluster := range list.Items {
		for _, schedule := range dbCluster.Spec.Backup.Schedules {
			if schedule.ObjectStorageName == storageName {
				dbClusters = append(dbClusters, dbCluster)
				break
			}
		}
	}

	return dbClusters, nil
}

func buildObjectStorageInUseError(dbClusters []everestv1alpha1.DatabaseCluster, storageName string) error {
	if len(dbClusters) == 0 {
		return nil
	}
	names := make([]string, 0, len(dbClusters))
	for _, cluster := range dbClusters {
		names = append(names, cluster.Name)
	}

	return errors.Errorf("the ObjectStorage '%s' is used in following DatabaseClusters: %s. Please update the DatabaseClusters configuration first", storageName, strings.Join(names, ","))
}

func buildProxiedURL(uri, kubernetesID, resourceName, namespace string) string {
	// cut the /kubernetes part
	uri = strings.TrimPrefix(uri, "/v1/kubernetes/"+kubernetesID)

	// cut the resource name if present
	uri = strings.TrimSuffix(uri, resourceName)

	// remove kebab-case
	uri = strings.ReplaceAll(uri, "-", "")
	return fmt.Sprintf("/apis/everest.percona.com/v1alpha1/namespaces/%s%s%s", namespace, uri, resourceName)
}

func buildSecretName(storageName string) string {
	return storageName + "-secret"
}
