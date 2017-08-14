package topo

import (
	//"encoding/json"

	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	api "k8s.io/client-go/pkg/api/v1"
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
	ips, ipToSVC, err := getIPs(kubeClient, namespace, topoSelector)
	if err != nil {
		glog.V(3).Infof("Fetch pod IPs of topoSelector %s in namespace %s failed: %v", topoSelector, namespace, err)
		return nil, fmt.Errorf("Fetch pod IPs of topoSelector %s in namespace %s failed: %v", topoSelector, namespace, err)
	}
	glog.V(3).Infof("Fetched pod IPs %v of topoSelector %s in namespace %s", ips, topoSelector, namespace)

	// test data
	//ips := []string{"10.168.14.71", "10.168.14.99"}
	// Bool Search with podIPs
	topo, err := esClient.GetLinks(ips)
	if err != nil {
		glog.V(3).Infof("Get links of topoSelector %s in namespace %s failed: %v", topoSelector, namespace, err)
		return nil, fmt.Errorf("Get links of topoSelector %s in namespace %s failed: %v", topoSelector, namespace, err)
	}

	linkToSVC := make(map[string]int64)
	for _, link := range topo {
		ip := strings.Split(link.Key, "_")
		linkName := ipToSVC[ip[0]] + "_" + ipToSVC[ip[1]]
		if _, ok := linkToSVC[linkName]; ok {
			linkToSVC[linkName]++
		} else {
			linkToSVC[linkName] = 1
		}
	}

	links := make([]types.LinkSVC, len(linkToSVC))
	i := 0
	for k, c := range linkToSVC {
		svcName := strings.Split(k, "_")
		links[i] = types.LinkSVC{
			Key:    k,
			Count:  c,
			SrcSVC: svcName[0],
			DstSVC: svcName[1],
		}
		i++
	}

	return &types.Topology{
		Name:  topoSelector,
		Links: links,
	}, nil
}

func getIPs(kubeClient *kubernetes.Clientset, namespace, topoSelector string) ([]string, map[string]string, error) {
	//return nil, nil
	opts := metav1.ListOptions{
		LabelSelector: topoSelector,
	}
	endpointList, err := kubeClient.CoreV1().Endpoints(api.NamespaceAll).List(opts)
	if err != nil {
		return nil, nil, err
	}
	var IPs []string
	ipToSVC := make(map[string]string)
	for _, endpoint := range endpointList.Items {
		for _, subSets := range endpoint.Subsets {
			for _, address := range subSets.Addresses {
				IPs = append(IPs, address.IP)
				ipToSVC[address.IP] = endpoint.Name
			}
			for _, notReadyAddress := range subSets.NotReadyAddresses {
				IPs = append(IPs, notReadyAddress.IP)
				ipToSVC[notReadyAddress.IP] = endpoint.Name
			}
		}
	}
	return IPs, ipToSVC, nil
}
