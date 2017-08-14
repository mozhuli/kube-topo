package config

import (
	"io/ioutil"
	"time"

	"github.com/mozhuli/kube-topo/pkg/elasticsearch"
	"github.com/mozhuli/kube-topo/pkg/types"

	"k8s.io/client-go/kubernetes"

	"gopkg.in/yaml.v2"
)

var (
	// KubeConfig is the path to kubernetes admin config file (default "/etc/kubernetes/admin.conf")
	KubeConfig string
	// ElasticsearchEndpoint is elasticsearch endpoint
	ElasticsearchEndpoint string
	// Address is to listen and serve
	Address string
	// Timeout of sniffer report (default: -1s)
	Timeout time.Duration
	// Debug enable debug output
	Debug bool
	// KubeClient the kubernetes client
	KubeClient *kubernetes.Clientset
	// EsClient the es client
	EsClient *elasticsearch.Client
	// Topomap save the topo's ip informations.
	Topomap *types.TopoToIPs
)

// Config is the internal representation of the yaml that
// determines how the app start
type Config struct {
	KubeConfig            string `yaml:"kube-config"`
	ElasticsearchEndpoint string `yaml:"endpoint"`
	Address               string `yaml:"address"`
	Timeout               int    `yaml:"timeout"`
	Debug                 bool   `yaml:"debug"`
}

// ReadConfig reads from a file with the given name and returns
// a config or an error if the file was unable to be parsed.
func ReadConfig(filepath string) (*Config, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, err
}
