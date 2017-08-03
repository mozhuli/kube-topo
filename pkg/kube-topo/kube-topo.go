package topo

import (
	//"encoding/json"

	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	//"golang.org/x/net/context"
	"github.com/mozhuli/kube-topo/pkg/config"
	"github.com/mozhuli/kube-topo/pkg/elasticsearch"
	"github.com/mozhuli/kube-topo/pkg/elasticsearch/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Index is the es url index
var Index = "search"

func parseParams(r *http.Request) (string, string, error) {
	var namespace, topoSelector string
	r.ParseForm()
	for k, v := range r.Form {
		if len(v) != 1 {
			return "", "", fmt.Errorf("Wrong request params")
		}
		if k == "namespace" {
			namespace = v[0]
		}
		if k == "topoSelector" {
			topoSelector = v[0]
		}
	}
	if namespace != "" && topoSelector != "" {
		return namespace, topoSelector, nil
	}
	return "", "", fmt.Errorf("Wrong request params")
}

// TopoHandler hand the get topological graph request.
func TopoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	// Parse namespace topoSelector params
	namespace, topoSelector, err := parseParams(r)
	if err != nil {
		glog.Errorf("Failed parse params: %v", err)
		w.Write([]byte("Wrong params!\n"))
		return
	}
	glog.V(3).Infof("Parsed params namespace: %s,topoSelector: %s", namespace, topoSelector)

	//Generate topological graph
	topoData, err := generateTopo(config.KubeClient, config.EsClient, namespace, topoSelector)
	if err != nil {
		glog.Errorf("Failed generate topological graph data: %v", err)
		w.Write([]byte("Failed generate topological graph data!\n"))
		//panic(err)
	}
	glog.V(3).Infof("Generated topological graph data: %#v", topoData)

	b, _ := json.Marshal(topoData)
	//fmt.Fprint(w, *b)
	w.Write([]byte(b))

	//w.Write([]byte("Gorilla Map!\n"))
	//fmt.Println(client.ClusterState())
	//fmt.Fprint(w, searchResult.TotalHits())

}

func generateTopo(kubeClient *kubernetes.Clientset, esClient *elasticsearch.Client, namespace, topoSelector string) (*types.Topology, error) {
	podIPs, err := getPodIPs(kubeClient, namespace, topoSelector)
	if err != nil {
		glog.V(3).Infof("Fetch pod IPs of topoSelector %s in namespace %s failed: %v", topoSelector, namespace, err)
		return nil, fmt.Errorf("Fetch pod IPs of topoSelector %s in namespace %s failed: %v", topoSelector, namespace, err)
	}
	glog.V(3).Infof("Fetch pod IPs of topoSelector %s in namespace %s: %v", topoSelector, namespace, podIPs)

	// test data
	ips := []string{"10.168.14.71", "10.168.14.99"}
	// Bool Search with podIPs
	topo, err := esClient.GetLinks(ips)
	if err != nil {
		glog.V(3).Infof("Get links of topoSelector %s in namespace %s failed: %v", topoSelector, namespace, err)
		return nil, fmt.Errorf("Get links of topoSelector %s in namespace %s failed: %v", topoSelector, namespace, err)
	}

	return &types.Topology{
		Name:  topoSelector,
		Links: topo,
	}, nil
}

func getPodIPs(kubeClient *kubernetes.Clientset, namespace, topoSelector string) ([]string, error) {
	return nil, nil
	opts := metav1.ListOptions{
		LabelSelector: topoSelector,
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
