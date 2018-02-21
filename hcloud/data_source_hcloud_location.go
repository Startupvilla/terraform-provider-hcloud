package hcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"context"
	"strconv"
)

func dataSourceHcloudLocation() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudLocationRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"description": {
				Type: schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceHcloudLocationRead(d *schema.ResourceData, m interface{}) error {
	var loc *hcloud.Location
	var err error

	if cid, ok := d.GetOk("id"); ok {
		id, err := strconv.Atoi(cid.(string))
		if err != nil {
			return err
		}

		loc, _, err = m.(*hcloud.Client).Location.GetByID(context.Background(), id)
	} else {
		loc, _, err = m.(*hcloud.Client).Location.Get(context.Background(), d.Get("name").(string))
	}

	if err != nil {
		return err
	}

	if loc != nil {
		d.SetId(strconv.Itoa(loc.ID))
		d.Set("name", loc.Name)
		d.Set("description", loc.Description)
	} else {
		d.SetId("")
	}

	return nil
}