package hcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Hetzner Cloud Token",
				Sensitive:   true,
			},
		},
		DataSourcesMap: map[string]*schema.Resource {
			"hcloud_datacenter": dataSourceHcloudDatacenter(),
			"hcloud_image": dataSourceHcloudImage(),
			"hcloud_servertype": dataSourceHcloudServertype(),
			"hcloud_location": dataSourceHcloudLocation(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"hcloud_server" : resourceHcloudServer(),
			"hcloud_sshkey" : resourceHcloudSSHKey(),
			"hcloud_rescue" : resourceHcloudRescue(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	client := hcloud.NewClient(hcloud.WithToken(d.Get("token").(string)))
	return client, nil
}