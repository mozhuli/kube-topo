package elasticsearch

import (
	//"encoding/json"
	"errors"
	//"fmt"
	//"os"
	//"reflect"

	"golang.org/x/net/context"
	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/golang/glog"
	"github.com/mozhuli/kube-topo/pkg/elasticsearch/types"
	//"github.com/mozhuli/kube-topo/pkg/util"
	//gcfg "gopkg.in/gcfg.v1"
)

const (
	StatusCodeAlreadyExists int = 409

	podNamePrefix     = "kube"
	securitygroupName = "kube-securitygroup-default"
	HostnameMaxLen    = 63
)

var (
	ctx                = context.Background()
	ErrNotFound        = errors.New("NotFound")
	ErrMultipleResults = errors.New("MultipleResults")
)

type Client struct {
	ES *elastic.Client
}

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

func (c *Client) CreateIndex(index string) error {
	_, err := c.ES.CreateIndex(index).Do(ctx)
	if err != nil {
		glog.Errorf("Create index %s error: %v", index, err)
	}
	return err
}

func (c *Client) AddDocument(link types.Link) error {
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

/*func (c *Client) DeleteTenant(tenantName string) error {
	return tenants.List(c.Identity, nil).EachPage(func(page pagination.Page) (bool, error) {
		tenantList, err := tenants.ExtractTenants(page)
		if err != nil {
			return false, err
		}
		for _, t := range tenantList {
			if t.Name == tenantName {
				err := tenants.Delete(c.Identity, t.ID).ExtractErr()
				if err != nil {
					glog.Errorf("Delete openstack tenant %s error: %v", tenantName, err)
					return false, err
				}
				glog.V(4).Infof("Tenant %s deleted", tenantName)
				break
			}
		}
		return true, nil
	})
}

func (c *Client) CreateUser(username, password, tenantID string) error {
	opts := users.CreateOpts{
		Name:     username,
		TenantID: tenantID,
		Enabled:  gophercloud.Enabled,
		Password: password,
	}
	_, err := users.Create(c.Identity, opts).Extract()
	if err != nil && !IsAlreadyExists(err) {
		glog.Errorf("Failed to create user %s: %v", username, err)
		return err
	}
	glog.V(4).Infof("User %s created", username)
	return nil
}*/
