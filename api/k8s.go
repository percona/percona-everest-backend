package api

import (
	"encoding/base64"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// configGetter stores kubeconfig string to convert it to the final object.
type configGetter struct {
	kubeconfig string
}

// newConfigGetter creates a new configGetter struct.
func newConfigGetter(kubeconfig string) *configGetter {
	return &configGetter{kubeconfig: kubeconfig}
}

// loadFromString takes a kubeconfig and deserializes the contents into Config object.
func (g *configGetter) loadFromString() (*clientcmdapi.Config, error) {
	decoded, err := base64.StdEncoding.DecodeString(g.kubeconfig)
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.Load(decoded)
	if err != nil {
		return nil, err
	}

	if config.AuthInfos == nil {
		config.AuthInfos = make(map[string]*clientcmdapi.AuthInfo)
	}
	if config.Clusters == nil {
		config.Clusters = make(map[string]*clientcmdapi.Cluster)
	}
	if config.Contexts == nil {
		config.Contexts = make(map[string]*clientcmdapi.Context)
	}

	return config, nil
}
