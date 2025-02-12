package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func nodeDataSource() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches nodes from Trustgrid",

		ReadContext: nodeRead,

		Schema: map[string]*schema.Schema{
			"tags": {
				Description: "Tag Filters",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"exclude_tags": {
				Description: "Exclude Tag Filters",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"node_ids": {
				Type:        schema.TypeSet,
				Description: "List of matching nodes",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

type Node struct {
	UID     string            `json:"uid"`
	Name    string            `json:"name"`
	FQDN    string            `json:"fqdn"`
	Cluster string            `json:"cluster"`
	Tags    map[string]string `json:"tags" `
	Config  struct {
		SNMP `json:"snmp"`
	} `json:"config"`
}

type filter struct {
	Tags        map[string]interface{} `tf:"tags"`
	ExcludeTags map[string]interface{} `tf:"exclude_tags"`
}

func (f *filter) match(n Node) bool {
	for tag, value := range f.Tags {
		nv := n.Tags[tag]
		if nv != value {
			return false
		}
	}

	for tag, value := range f.ExcludeTags {
		nv := n.Tags[tag]
		if nv == value {
			return false
		}
	}

	return true
}

func nodeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	//nodes := make([]map[string]interface{}, 0)
	//nodes = append(nodes, map[string]interface{}{"uid": "59838ae6-a2b2-4c45-b7be-9378f0b265f5"})

	// TODO set this to hash the filters
	d.SetId(fmt.Sprintf("%d", time.Now().Unix()))

	tg := meta.(*tgClient)

	f := filter{}
	err := marshalResourceData(d, &f)
	if err != nil {
		return diag.FromErr(err)
	}

	nodes := make([]Node, 0)
	err = tg.get(ctx, "/node", &nodes)
	if err != nil {
		return diag.FromErr(err)
	}

	nodeIDs := make([]string, 0)
	for _, node := range nodes {
		if f.match(node) {
			nodeIDs = append(nodeIDs, node.UID)
		}
	}

	err = d.Set("node_ids", nodeIDs)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
