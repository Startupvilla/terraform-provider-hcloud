package hcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"context"
	"github.com/satori/go.uuid"
)

func resourceHcloudRescue() *schema.Resource {
	return &schema.Resource{
		Create: resourceHcloudRescueCreate,
		Read:   resourceHcloudRescueRead,
		Delete: resourceHcloudRescueDelete,
		Update: resourceHcloudRescueUpdate,

		Schema: map[string]*schema.Schema{
			"server": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Default:  "linux64",
				Optional: true,
				ForceNew: true,
			},
			"ssh_keys": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:     schema.TypeInt,
				},
				Optional: true,
				ForceNew: true,
			},
			"password": &schema.Schema{
				Type: schema.TypeString,
				Computed: true,
			},
			"reset_on_activation": &schema.Schema{
				Type: schema.TypeBool,
				Optional: true,
				Default: true,
			},
			"reboot_on_activation": &schema.Schema{
				Type: schema.TypeBool,
				Optional: true,
				Default: false,
			},
			"reset_on_deactivation": &schema.Schema{
				Type: schema.TypeBool,
				Optional: true,
				Default: true,
			},
			"reboot_on_deactivation": &schema.Schema{
				Type: schema.TypeBool,
				Optional: true,
				Default: false,
			},
		},
	}
}

func resourceHcloudRescueCreate(d *schema.ResourceData, m interface{}) error {
	// get server object
	server, _, err := m.(*hcloud.Client).Server.GetByID(context.Background(), d.Get("server").(int))
	if err != nil {
		return err
	}

	// deactivate rescue first, if it's already enabled
	if server.RescueEnabled {
		// disable rescue request
		act, _, err := m.(*hcloud.Client).Server.DisableRescue(context.Background(), server)
		if err != nil {
			return err
		}

		// wait for action and check for error
		_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
		err = <-errch
		if err != nil {
			return err
		}
	}

	// transform ssh key ids into SSHKey array
	ids := d.Get("ssh_keys").([]interface{})
	ssh := make([]*hcloud.SSHKey, len(ids))
	for i, v := range ids {
		ssh[i] = &hcloud.SSHKey{
			ID: v.(int),
		}
	}

	// set rescue type
	var rt hcloud.ServerRescueType
	switch d.Get("type").(string) {
		case "linux32":
			rt = hcloud.ServerRescueTypeLinux32

		case "freebsd64":
			rt = hcloud.ServerRescueTypeFreeBSD64

		default:
			rt = hcloud.ServerRescueTypeLinux64
	}


	// enable rescue request
	rescue, _, err := m.(*hcloud.Client).Server.EnableRescue(context.Background(), server, hcloud.ServerEnableRescueOpts{
		Type: rt,
		SSHKeys: ssh,
	})
	if err != nil {
		return err
	}

	// wait for action and check for error
	_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), rescue.Action)
	err = <-errch
	if err != nil {
		return err
	}

	// save root password
	d.Set("password", rescue.RootPassword)

	if d.Get("reboot_on_activation").(bool) {	// soft reboot instead of reset, if reboot is set to true
		act, _, err := m.(*hcloud.Client).Server.Reboot(context.Background(), server)
		if err != nil {
			return err
		}

		// wait for action and check for error
		_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
		err = <-errch
		if err != nil {
			return err
		}

	} else if d.Get("reset_on_activation").(bool) {		// reset if reboot is false and reset is true
		act, _, err := m.(*hcloud.Client).Server.Reset(context.Background(), server)
		if err != nil {
			return err
		}

		// wait for action and check for error
		_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
		err = <-errch
		if err != nil {
			return err
		}
	}

	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	d.SetId(id.String())

	return nil
}

func resourceHcloudRescueRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceHcloudRescueDelete(d *schema.ResourceData, m interface{}) error {
	// get server object
	server, _, err := m.(*hcloud.Client).Server.GetByID(context.Background(), d.Get("server").(int))
	if err != nil {
		return err
	}

	// disable rescue if enabled
	if server.RescueEnabled {
		// disable rescue request
		act, _, err := m.(*hcloud.Client).Server.DisableRescue(context.Background(), server)
		if err != nil {
			return err
		}

		// wait for action and check for error
		_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
		err = <-errch
		if err != nil {
			return err
		}
	}

	// check for reboot or reset
	if d.Get("reboot_on_deactivation").(bool) {	// soft reboot instead of reset, if reboot is set to true
		act, _, err := m.(*hcloud.Client).Server.Reboot(context.Background(), server)
		if err != nil {
			return err
		}

		// wait for action and check for error
		_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
		err = <-errch
		if err != nil {
			return err
		}

	} else if d.Get("reset_on_deactivation").(bool) {		// reset if reboot is false and reset is true
		act, _, err := m.(*hcloud.Client).Server.Reset(context.Background(), server)
		if err != nil {
			return err
		}

		// wait for action and check for error
		_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
		err = <-errch
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceHcloudRescueUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}