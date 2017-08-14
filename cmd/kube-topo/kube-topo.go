package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

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

func startController(kubeClient *kubernetes.Clientset,*elasticsearch.Client) error {
	// Creates a new topo controller
	topoController, err := topo.NewTopoController(kubeClient)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error { return topoController.Run(ctx.Done()) })
	wg.Go(func() error { return initHTTPHandle(ctx.Done()) })

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	select {
	case <-term:
		glog.V(4).Info("Received SIGTERM, exiting gracefully...")
	case <-ctx.Done():
	}

	cancel()
	if err := wg.Wait(); err != nil {
		glog.Errorf("Unhandled error received: %v", err)
		return err
	}

	return nil
}

func initHTTPHandle(stopCh <-chan struct{})error{
	r := mux.NewRouter()
	r.HandleFunc("/topo", topo.TopoHandler)
	http.Handle("/", r)
	http.ListenAndServe(*address, r)
	<-stopCh
	return nil
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

	// Start topo controller.
	if err := startController(kubeClient, esClient); err != nil {
		glog.Fatal(err)
	}

	// Start http handle
	//initHTTPHandle()
}
