package hcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"context"
	"strconv"
)

func dataSourceHcloudImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudImageRead,

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
			"status": {
				Type: schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceHcloudImageRead(d *schema.ResourceData, m interface{}) error {
	var img *hcloud.Image
	var err error

	if cid, ok := d.GetOk("id"); ok {
		id, err := strconv.Atoi(cid.(string))
		if err != nil {
			return err
		}

		img, _, err = m.(*hcloud.Client).Image.GetByID(context.Background(), id)
	} else {
		img, _, err = m.(*hcloud.Client).Image.Get(context.Background(), d.Get("name").(string))
	}

	if err != nil {
		return err
	}

	if img != nil {
		d.SetId(strconv.Itoa(img.ID))
		d.Set("name", img.Name)
		d.Set("description", img.Description)
		d.Set("status", img.Status)
	} else {
		d.SetId("")
	}

	return nil
}