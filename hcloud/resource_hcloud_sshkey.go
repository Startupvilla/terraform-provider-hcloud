package hcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"context"
	"strconv"
)

func resourceHcloudSSHKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceHcloudSSHKeyCreate,
		Read:   resourceHcloudSSHKeyRead,
		Update: resourceHcloudSSHKeyUpdate,
		Delete: resourceHcloudSSHKeyDelete,

		Importer: &schema.ResourceImporter{
			State: resourceHcloudSSHKeyImport,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"public_key": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"fingerprint": &schema.Schema{
				Type: schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHcloudSSHKeyImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := resourceHcloudSSHKeyRead(d, meta)

	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHcloudSSHKeyCreate(d *schema.ResourceData, m interface{}) error {
	key, _, err := m.(*hcloud.Client).SSHKey.Create(context.Background(), hcloud.SSHKeyCreateOpts{
		Name: d.Get("name").(string),
		PublicKey: d.Get("public_key").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(key.ID))
	d.Set("fingerprint", key.Fingerprint)

	return nil
}

func resourceHcloudSSHKeyRead(d *schema.ResourceData, m interface{}) error {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	key, _, err := m.(*hcloud.Client).SSHKey.GetByID(context.Background(), id)
	if err != nil {
		return err
	}

	d.Set("name", key.Name)
	d.Set("public_key", key.PublicKey)
	d.Set("fingerprint", key.Fingerprint)

	return nil
}

func resourceHcloudSSHKeyUpdate(d *schema.ResourceData, m interface{}) error {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	key, _, err := m.(*hcloud.Client).SSHKey.GetByID(context.Background(), id)
	if err != nil {
		return err
	}

	if key.Name != d.Get("name").(string) {
		key, _, err = m.(*hcloud.Client).SSHKey.Update(context.Background(), key, hcloud.SSHKeyUpdateOpts{
			Name: d.Get("name").(string),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceHcloudSSHKeyDelete(d *schema.ResourceData, m interface{}) error {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	key, _, err := m.(*hcloud.Client).SSHKey.GetByID(context.Background(), id)
	if err != nil {
		return err
	}
	_, err = m.(*hcloud.Client).SSHKey.Delete(context.Background(), key)

	return err
}