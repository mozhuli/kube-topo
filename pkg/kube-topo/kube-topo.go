package topo

import (
	//"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	//"golang.org/x/net/context"
	"github.com/mozhuli/kube-topo/pkg/config"
	"github.com/mozhuli/kube-topo/pkg/elasticsearch"
	"github.com/mozhuli/kube-topo/pkg/elasticsearch/types"
	"github.com/mozhuli/kube-topo/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var Index = "search"

func parseParams(r *http.Request) (string, string, error) {
	var namespace, topoID string
	r.ParseForm()
	for k, v := range r.Form {
		if len(v) != 1 {
			return "", "", fmt.Errorf("Wrong request params")
		}
		if k == "namespace" {
			namespace = v[0]
		}
		if k == "topoID" {
			topoID = v[0]
		}
	}
	if namespace != "" && topoID != "" {
		return namespace, topoID, nil
	}
	return "", "", fmt.Errorf("Wrong request params")
}

func TopoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	// Parse namespace topoID params
	namespace, topoID, err := parseParams(r)
	if err != nil {
		glog.Errorf("Failed parse params: %v", err)
		w.Write([]byte("Wrong params!\n"))
		//fmt.Fprint(w, "Wrong params")
		//panic(err)

	}
	glog.V(3).Infof("Parsed params namespace: %s,topoID: %s", namespace, topoID)

	//Generate topological graph
	topoData, err := generateTopo(namespace, topoID)
	if err != nil {
		glog.Errorf("Failed generate topological graph data: %v", err)
		w.Write([]byte("Failed generate topological graph data!\n"))
		//panic(err)
	}
	glog.V(3).Infof("Generated topological graph data: %v", topoData)

	//w.Write([]byte(*topoData))

	//w.Write([]byte("Gorilla Map!\n"))
	//fmt.Println(client.ClusterState())
	//fmt.Fprint(w, searchResult.TotalHits())

}

func generateTopo(namespace, topoID string) (*types.Topology, error) {
	_, err := elasticsearch.NewClient(config.ElasticsearchEndpoint)
	if err != nil {
		glog.V(3).Infof("Init elasticsearch client failed: %v", err)
		return nil, fmt.Errorf("Init elasticsearch client failed: %v", err)
	}

	podIPs, err := getPodIPs(namespace, topoID)
	if err != nil {
		glog.V(3).Infof("Fetch pod IPs of topoID %s in namespace %s failed: %v", namespace, topoID, err)
		return nil, fmt.Errorf("Fetch pod IPs of topoID %s in namespace %s failed: %v", namespace, topoID, err)
	}
	glog.V(3).Infof("Fetch pod IPs of topoID %s in namespace %s: %v", namespace, topoID, podIPs)

	// Search with a term query
	/*termQuery := elastic.NewTermQuery("user", "olivere")
	searchResult, err := client.Search().
		Index("twitter").        // search in index "twitter"
		Query(termQuery).        // specify the query
		Sort("user", true).      // sort by "user" field, ascending
		From(0).Size(10).        // take documents 0-9
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		// Handle error
		fmt.Println(err)
		//panic(err)
	}*/
	return nil, nil
}

func getPodIPs(namespace, topoID string) ([]string, error) {
	// Create kubernetes client config. Use kubeconfig if given, otherwise assume in-cluster.
	config, err := util.NewClusterConfig(config.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %v", err)
	}

	opts := metav1.ListOptions{
		LabelSelector: "topoID=" + topoID,
	}
	endpointList, err := kubeClient.CoreV1().Endpoints(namespace).List(opts)
	if err != nil {
		return nil, err
	}

	var podIPs []string
	for i := 0; i < len(endpointList.Items); i++ {
		for j := 0; j < len(endpointList.Items[i].Subsets[0].Addresses); j++ {
			ip := endpointList.Items[i].Subsets[0].Addresses[j].IP
			podIPs = append(podIPs, ip)
		}

	}
	return podIPs, nil
}
