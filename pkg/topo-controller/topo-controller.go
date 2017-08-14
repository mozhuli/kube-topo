package topo

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/mozhuli/kube-topo/pkg/config"
	"github.com/mozhuli/kube-topo/pkg/sets"
	"github.com/mozhuli/kube-topo/pkg/types"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
	utilruntime "k8s.io/client-go/pkg/util/runtime"
	"k8s.io/client-go/pkg/util/wait"
	"k8s.io/client-go/pkg/util/workqueue"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/tools/cache"
	// Only required to authenticate against GKE clusters
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type TopoController struct {
	kubeClient *kubernetes.Clientset
	// A cache of endpoints
	endpointsStore cache.Store
	// Watches changes to all endpoints
	endpointController *cache.Controller
	endpointsQueue     workqueue.RateLimitingInterface
}

func (topo *TopoController) enqueueEndpoint(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		fmt.Printf("Couldn't get key for object %+v: %v", obj, err)
		return
	}
	topo.endpointsQueue.Add(key)
}

func (topo *TopoController) updateEndpoint(oldObj interface{}, newObj interface{}) {
	oldEndpoints := oldObj.(*v1.Endpoints)
	newEndpoints := newObj.(*v1.Endpoints)
	if reflect.DeepEqual(oldEndpoints.Subsets, newEndpoints.Subsets) {
		return
	}
	topo.enqueueEndpoint(newObj)
}

func (topo *TopoController) endpointWorker() {
	workFunc := func() bool {
		key, quit := topo.endpointsQueue.Get()
		if quit {
			return true
		}
		defer topo.endpointsQueue.Done(key)

		obj, exists, err := topo.endpointsStore.GetByKey(key.(string))
		if !exists {
			fmt.Printf("endpoint has been deleted %v\n", key)
			return false
		}
		if err != nil {
			fmt.Printf("cannot get endpoint: %v\n", key)
			return false
		}

		endpoints := obj.(*v1.Endpoints)

		// there is no pod backing service "kubernetes"
		if endpoints.ObjectMeta.Name == "kubernetes" {
			return false
		}

		for k, v := range endpoints.Labels {
			if strings.HasPrefix(k, "topo-") {
				lable := k + "=" + v
				for _, subset := range endpoints.Subsets {
					for _, address := range subset.Addresses {
						ipSet, ok := config.Topomap.Read(lable)
						if !ok {
							fmt.Printf("Topomap can't find %s\n", lable)
							config.Topomap.Write(lable, "")
							topo.endpointsQueue.AddRateLimited(key)
							return false
						}
						config.Topomap.Write(lable, address.IP)
					}
				}
			}
		}

		return false
	}
	for {
		if quit := workFunc(); quit {
			fmt.Printf("topo worker shutting down")
			return
		}
	}
}

// NewTopoController create a new topo controller
func NewTopoController(kubeClient *kubernetes.Clientset) (*TopoController, error) {
	topo := &TopoController{
		kubeClient: kubeClient,
		endpointsQueue: workqueue.NewNamedRateLimitingQueue(
			workqueue.NewMaxOfRateLimiter(workqueue.NewItemExponentialFailureRateLimiter(100*time.Millisecond, 5*time.Second)), "endpoints"),
	}
	config.Topomap = types.NewTopoToIPs()

	topo.endpointsStore, topo.endpointController = cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return topo.kubeClient.Core().Endpoints(api.NamespaceAll).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return topo.kubeClient.Core().Endpoints(api.NamespaceAll).Watch(options)
			},
		},
		&v1.Endpoints{},
		// resync is not needed
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    topo.enqueueEndpoint,
			UpdateFunc: topo.updateEndpoint,
			DeleteFunc: topo.enqueueEndpoint,
		},
	)
	return topo, nil
}

// Run begins watching and syncing.
func (topo *TopoController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	fmt.Println("Starting topoController Manager")
	go topo.endpointController.Run(stopCh)
	// wait for the controller to List. This help avoid churns during start up.
	if !cache.WaitForCacheSync(stopCh, topo.endpointController.HasSynced) {
		return
	}
	go wait.Until(topo.endpointWorker, time.Second, stopCh)

	<-stopCh
	fmt.Printf("Shutting down topo Controller")
	topo.endpointsQueue.ShutDown()
}

func sliceToSets(slice []string) sets.String {
	ss := sets.String{}
	for _, s := range slice {
		ss.Insert(s)
	}
	return ss
}
