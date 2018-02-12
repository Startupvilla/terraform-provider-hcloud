package hcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"context"
	"strconv"
)

func dataSourceHcloudServertype() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudServertypeRead,

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

func dataSourceHcloudServertypeRead(d *schema.ResourceData, m interface{}) error {
	var st *hcloud.ServerType
	var err error

	if cid, ok := d.GetOk("id"); ok {
		id, err := strconv.Atoi(cid.(string))
		if err != nil {
			return err
		}

		st, _, err = m.(*hcloud.Client).ServerType.GetByID(context.Background(), id)
	} else {
		st, _, err = m.(*hcloud.Client).ServerType.Get(context.Background(), d.Get("name").(string))
	}

	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(st.ID))
	d.Set("name", st.Name)
	d.Set("description", st.Description)

	return nil
}