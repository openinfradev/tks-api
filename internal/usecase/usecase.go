package usecase

import (
	kube "github.com/openinfradev/tks-api/internal/kubernetes"
	gcache "github.com/patrickmn/go-cache"
	"k8s.io/client-go/kubernetes"
)

func GetKubeClient(cache *gcache.Cache, clusterId string) (*kubernetes.Clientset, error) {
	const prefix = "CACHE_KEY_KUBE_CLIENT_"
	value, found := cache.Get(prefix + clusterId)
	if found {
		return value.(*kubernetes.Clientset), nil
	}
	client, err := kube.GetClientFromClusterId(clusterId)
	if err != nil {
		return nil, err
	}

	cache.Set(prefix+clusterId, client, gcache.DefaultExpiration)
	return client, nil
}
