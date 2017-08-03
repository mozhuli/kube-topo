package main

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/mozhuli/kube-topo/pkg/config"
	"github.com/mozhuli/kube-topo/pkg/elasticsearch"
	topo "github.com/mozhuli/kube-topo/pkg/kube-topo"
	"github.com/mozhuli/kube-topo/pkg/util"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
)

var (
	kubeconfig = pflag.String("kubeconfig", "/etc/kubernetes/admin.conf", "path to kubernetes admin config file")
	endpoint   = pflag.String("endpoint", "http://10.10.101.145:9200", "elasticsearch endpoint")
	address    = pflag.String("address", "localhost:8000", "the address to listen and serve")
)

func initHTTPHandle() {
	r := mux.NewRouter()
	r.HandleFunc("/topo", topo.TopoHandler)
	http.Handle("/", r)
	http.ListenAndServe(*address, r)
}

func initClients() (*kubernetes.Clientset, *elasticsearch.Client, error) {
	// Create kubernetes client config. Use kubeconfig if given, otherwise assume in-cluster.
	/*config, err := util.NewClusterConfig(*kubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build kubeconfig: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kubernetes clientset: %v", err)
	}*/

	// Create elasticsearch client.
	esClient, err := elasticsearch.NewClient(*endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("could't initialize openstack client: %v", err)
	}

	return nil, esClient, nil
}

func main() {
	util.InitFlags()
	util.InitLogs()
	defer util.FlushLogs()

	// Initilize kubernetes and elasticsearch clients.
	kubeClient, esClient, err := initClients()
	if err != nil {
		glog.Fatal(err)
	}

	config.KubeClient = kubeClient
	config.EsClient = esClient

	// Start http handle
	initHTTPHandle()
}
