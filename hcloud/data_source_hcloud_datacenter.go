package hcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"context"
	"strconv"
)

func dataSourceHcloudDatacenter() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudDatacenterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type: schema.TypeString,
				Computed: true,
			},
			"location": {
				Type: schema.TypeInt,
				Computed: true,
			},
			"server_types_supported": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:     schema.TypeInt,
				},
				Computed: true,
			},
			"server_types_available": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:     schema.TypeInt,
				},
				Computed: true,
			},
		},
	}
}

func dataSourceHcloudDatacenterRead(d *schema.ResourceData, m interface{}) error {
	var dc *hcloud.Datacenter
	var err error

	if cid, ok := d.GetOk("id"); ok {
		id, err := strconv.Atoi(cid.(string))
		if err != nil {
			return err
		}

		dc, _, err = m.(*hcloud.Client).Datacenter.GetByID(context.Background(), id)
	} else {
		dc, _, err = m.(*hcloud.Client).Datacenter.Get(context.Background(), d.Get("name").(string))
	}

	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(dc.ID))
	d.Set("name", dc.Name)
	d.Set("description", dc.Description)
	d.Set("location", dc.Location.ID)

	supp := make([]int, len(dc.ServerTypes.Supported))
	for i, v := range dc.ServerTypes.Supported {
		supp[i] = v.ID
	}
	d.Set("server_types_supported", supp)

	avail := make([]int, len(dc.ServerTypes.Available))
	for i, v := range dc.ServerTypes.Available {
		avail[i] = v.ID
	}
	d.Set("server_types_supported", avail)

	return nil
}