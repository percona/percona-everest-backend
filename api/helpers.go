package api

import "github.com/percona/percona-everest-backend/model"

func toKubernetesClusterAPIItem(k model.KubernetesCluster) *KubernetesCluster {
	return &KubernetesCluster{Name: &k.Name, Id: &k.ID}
}
func toKubernetesClusterAPIList(list []model.KubernetesCluster) []*KubernetesCluster {
	items := make([]*KubernetesCluster, len(list))
	for i := range list {
		items[i] = toKubernetesClusterAPIItem(list[i])
	}
	return items
}
