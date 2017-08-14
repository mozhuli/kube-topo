package elasticsearch

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"
	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/golang/glog"
	"github.com/mozhuli/kube-topo/pkg/types"
	//"github.com/mozhuli/kube-topo/pkg/util"
	//gcfg "gopkg.in/gcfg.v1"
)

const (
	StatusCodeAlreadyExists int = 409

	podNamePrefix = "kube"
)

var (
	ctx         = context.Background()
	ErrNotFound = errors.New("NotFound")
)

type Client struct {
	ES *elastic.Client
}

// NewClient return a es client
func NewClient(endpoint string) (*Client, error) {
	// Create a new elasticsearch client
	esClient, err := elastic.NewClient(elastic.SetURL(endpoint))
	if err != nil {
		return nil, err
	}
	client := &Client{
		ES: esClient,
	}
	return client, nil
}

// CreateIndex create es index.
func (c *Client) CreateIndex(index string) error {
	_, err := c.ES.CreateIndex(index).Do(ctx)
	if err != nil {
		glog.Errorf("Create index %s error: %v", index, err)
	}
	return err
}

// AddDocument add document to es.
func (c *Client) AddDocument(link types.IPLink) error {
	// Add a document to the index
	_, err := c.ES.Index().
		Index("topo").
		Type("aggregation").
		BodyJson(link).
		Do(ctx)
	if err != nil {
		glog.Errorf("Add a document %v error: %v", link, err)
	}
	return err
}

// GetLinks get links of specific ips.
func (c *Client) GetLinks(ips []string) ([]types.IPLink, error) {
	//ips := []string{"10.168.14.71", "10.168.14.99"}
	res, err := FindTopo(c.ES, ips)
	if err != nil {
		glog.Errorf("Get links of %v error: %v", ips, err)
		return nil, err
	}

	// Output results
	// fmt.Println(res)
	return res, nil
}

// FindTopo executes the search and returns a response.
func FindTopo(client *elastic.Client, ips []string) ([]types.IPLink, error) {
	// Create service and use query, aggregations, filter, pagination funcs
	search := client.Search().Index("topo").Type("log")
	search = query(search, ips)
	search = aggs(search)
	search = paginate(search)

	// TODO Add other properties here, e.g. timeouts, explain or pretty printing

	// Execute query
	sr, err := search.Do(ctx)
	if err != nil {
		glog.Errorf("Search error: %v", err)
		return nil, err
	}

	// Decode response
	/*links, err := f.decodeLinks(sr)
	if err != nil {
		return resp, err
	}
	fmt.Println(links)*/
	//resp.Links = links
	//resp.Total = sr.Hits.TotalHits

	// Deserialize aggregations
	var links []types.IPLink
	if agg, found := sr.Aggregations.Terms("links"); found {
		links = make([]types.IPLink, len(agg.Buckets))
		for i, bucket := range agg.Buckets {
			fmt.Println(bucket.DocCount)
			fmt.Println(bucket.Key.(string))
			links[i] = types.IPLink{
				Key:   bucket.Key.(string),
				Count: bucket.DocCount,
			}
		}
	}

	return links, nil
}

// query sets up the query in the search service.
func query(service *elastic.SearchService, ips []string) *elastic.SearchService {
	q := elastic.NewBoolQuery()
	for i := 0; i < len(ips); i++ {
		q = q.Should(elastic.NewTermQuery("Data.dstIP", ips[i]))
	}
	// TODO Add other queries and filters here, maybe differentiating between AND/OR etc.

	service = service.Query(q)
	return service
}

// aggs sets up the aggregations in the service.
func aggs(service *elastic.SearchService) *elastic.SearchService {
	// Terms aggregation by genre
	agg := elastic.NewTermsAggregation().Field("Data.link")
	service = service.Aggregation("links", agg)

	return service
}

// paginate sets up pagination in the service.
func paginate(service *elastic.SearchService) *elastic.SearchService {
	service = service.From(0)
	service = service.Size(0)
	return service
}
