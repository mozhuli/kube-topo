package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mozhuli/kube-topo/pkg/config"
	"github.com/mozhuli/kube-topo/pkg/elasticsearch"
	topo "github.com/mozhuli/kube-topo/pkg/kube-topo"
	"github.com/mozhuli/kube-topo/pkg/util"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
	//"k8s.io/client-go/kubernetes"
)

var (
	kubeconfig = pflag.String("kubeconfig", "/etc/kubernetes/admin.conf", "path to kubernetes admin config file")
	endpoint   = pflag.String("endpoint", "", "elasticsearch endpoint")
	address    = pflag.String("address", "localhost:8000", "the address to listen and serve")
)

func initHttpHandle() {
	r := mux.NewRouter()
	r.HandleFunc("/topo", topo.TopoHandler)
	http.Handle("/", r)
	http.ListenAndServe(*address, r)
}

func verifyClientSetting() error {
	/*conf, err := util.NewClusterConfig(*kubeconfig)
	if err != nil {
		return fmt.Errorf("Init kubernetes cluster failed: %v", err)
	}

	_, err = kubernetes.NewForConfig(conf)
	if err != nil {
		return fmt.Errorf("Init kubernetes clientset failed: %v", err)
	}*/

	_, err := elasticsearch.NewClient(*endpoint)
	if err != nil {
		return fmt.Errorf("Init elasticsearch client failed: %v", err)
	}

	return nil
}

func main() {
	util.InitFlags()
	util.InitLogs()
	defer util.FlushLogs()

	// Verify client setting at the beginning and fail early if there are errors.
	err := verifyClientSetting()
	if err != nil {
		glog.Fatal(err)
	}

	// Iint config
	config.KubeConfig = *kubeconfig
	config.ElasticsearchEndpoint = *endpoint

	// Start http handle
	initHttpHandle()
}
