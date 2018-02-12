package hcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"context"
	"strconv"
	"time"
)

func resourceHcloudServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceHcloudServerCreate,
		Read:   resourceHcloudServerRead,
		Update: resourceHcloudServerUpdate,
		Delete: resourceHcloudServerDelete,

		Importer: &schema.ResourceImporter{
			State: resourceHcloudServerImport,
		},

		/*
		 *	@TODO: implement ISO attach and detach
		 */
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"server_type": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"datacenter": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ConflictsWith: []string{"location"},
			},
			"location": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ConflictsWith: []string{"datacenter"},
			},
			"image": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
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
			"user_data": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv4": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv4_ptr": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ipv6": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_ptr": &schema.Schema{
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type:     schema.TypeString,
				},
				Computed: true,
				//@TODO: implement setting ipv6_ptr
			},
			"root_pw": &schema.Schema{
				Type: schema.TypeString,
				Computed: true,
				Sensitive: true,
			},
			"upgrade_disk": &schema.Schema{
				Type: schema.TypeBool,
				Optional: true,
				Default: true,
			},
			"backup": &schema.Schema{
				Type: schema.TypeBool,
				Optional: true,
				Default: false,
			},
			"backup_window": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Default: "",
			},
		},
	}
}

func resourceHcloudServerImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := resourceHcloudServerRead(d, meta)

	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHcloudServerCreate(d *schema.ResourceData, m interface{}) error {
	// get server type object
	st, _, err := m.(*hcloud.Client).ServerType.GetByID(context.Background(), d.Get("server_type").(int))
	if err != nil {
		return err
	}

	// get image object
	img, _, err := m.(*hcloud.Client).Image.GetByID(context.Background(), d.Get("image").(int))
	if err != nil {
		return err
	}

	// check if location is set and get location object
	var loc *hcloud.Location = nil
	if lid, ok := d.GetOk("location"); ok {
		loc, _, err = m.(*hcloud.Client).Location.GetByID(context.Background(), lid.(int))
		if err != nil {
			return err
		}
	}

	// check if datacenter is set and get datacenter object
	var dc *hcloud.Datacenter = nil
	if dcid, ok := d.GetOk("datacenter"); ok {
		dc, _, err = m.(*hcloud.Client).Datacenter.GetByID(context.Background(), dcid.(int))
		if err != nil {
			return err
		}
	}

	// transform ssh key ids into array of SSHKey objects
	ids := d.Get("ssh_keys").([]interface{})
	ssh := make([]*hcloud.SSHKey, len(ids))
	for i, v := range ids {
		ssh[i] = &hcloud.SSHKey{
			ID: v.(int),
		}
	}

	// create server
	server, _, err := m.(*hcloud.Client).Server.Create(context.Background(), hcloud.ServerCreateOpts{
		Name: d.Get("name").(string),
		ServerType: st,
		Image: img,
		SSHKeys: ssh,
		Location: loc,
		Datacenter: dc,
		UserData: d.Get("user_data").(string),
	})

	// wait for server_create action and check for errors
	_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), server.Action)
	err = <-errch
	if err != nil {
		return err
	}

	// set server id and save root password
	d.SetId(strconv.Itoa(server.Server.ID))
	d.Set("root_pw", server.RootPassword)

	// check if IPv4 PTR is set
	if ptr, ok := d.GetOk("ipv4_ptr"); ok {
		sptr := ptr.(string)

		// send change dns ptr request
		act, _, err := m.(*hcloud.Client).Server.ChangeDNSPtr(context.Background(), server.Server, server.Server.PublicNet.IPv4.IP.String(), &sptr)
		if err != nil {
			return nil
		}

		// wait for action to finish and check for errors
		_, errch = m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
		err = <-errch
		if err != nil {
			return err
		}
	}

	// send enable or disable backup request
	var act *hcloud.Action
	if backup := d.Get("backup").(bool); backup {
		act, _, err = m.(*hcloud.Client).Server.EnableBackup(context.Background(), server.Server, d.Get("backup_window").(string))
		if err != nil {
			return err
		}
	} else {
		act, _, err = m.(*hcloud.Client).Server.DisableBackup(context.Background(), server.Server)
		if err != nil {
			return err
		}
	}

	// wait for backup action and check for error
	_, errch = m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
	err = <-errch
	if err != nil {
		return err
	}

	// read server data
	return resourceHcloudServerRead(d, m)
}

func resourceHcloudServerRead(d *schema.ResourceData, m interface{}) error {
	// convert id from string to int
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	// get server object
	server, _, err := m.(*hcloud.Client).Server.GetByID(context.Background(), id)
	if err != nil {
		return err
	}

	// update resource data
	d.Set("name", server.Name)
	d.Set("server_type", server.ServerType.ID)
	d.Set("datacenter", server.Datacenter.ID)
	d.Set("location", server.Datacenter.Location.ID)
	d.Set("image", server.Image.ID)
	d.Set("status", server.Status)
	d.Set("created", server.Created.String())
	d.Set("ipv4", server.PublicNet.IPv4.IP.String())
	d.Set("ipv4_ptr", server.PublicNet.IPv4.DNSPtr)
	d.Set("ipv6", server.PublicNet.IPv6.IP.String())
	d.Set("ipv6_ptr", server.PublicNet.IPv6.DNSPtr)
	d.Set("backup_window", server.BackupWindow)

	// check if backup is enabled or disabled
	if len(server.BackupWindow) > 0 {
		d.Set("backup", true)
	} else {
		d.Set("backup", false)
	}

	return nil
}

func resourceHcloudServerUpdate(d *schema.ResourceData, m interface{}) error {
	// convert id from string to int
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	// get server object
	server, _, err := m.(*hcloud.Client).Server.GetByID(context.Background(), id)
	if err != nil {
		return err
	}

	// update server name, when necessary
	if server.Name != d.Get("name").(string) {
		server, _, err = m.(*hcloud.Client).Server.Update(context.Background(), server, hcloud.ServerUpdateOpts{
			Name: d.Get("name").(string),
		})
		if err != nil {
			return err
		}
	}

	// update IPv4 DNS PTR, when necessary
	if server.PublicNet.IPv4.DNSPtr != d.Get("ipv4_ptr").(string) {
		ptr := d.Get("ipv4_ptr").(string)

		// send dns ptr change request
		act, _, err := m.(*hcloud.Client).Server.ChangeDNSPtr(context.Background(), server, server.PublicNet.IPv4.IP.String(), &ptr)
		if err != nil {
			return err
		}

		// wait for action to finish and check for errors
		_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
		err = <-errch
		if err != nil {
			return err
		}
	}

	// upgrade server type, if necessary
	if server.ServerType.ID != d.Get("server_type").(int) {
		restart := false

		// check if server is running
		if server.Status != "off" {
			// restart server afterwards
			restart = true

			// try acpi shutdown
			act, _, err := m.(*hcloud.Client).Server.Shutdown(context.Background(), server)
			if err != nil {
				return err
			}

			// wait for action to finish and check for errors
			_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
			err = <-errch
			if err != nil {
				return err
			}

			// try for 5 minutes if server is powered off
			for i := 0; i < 10; i++ {
				time.Sleep(30 * time.Second)

				server, _, err := m.(*hcloud.Client).Server.GetByID(context.Background(), id)
				if err != nil {
					return err
				}

				if server.Status == "off" {
					break
				}
			}
		}

		// check if server is still running
		if server.Status != "off" {
			// power off
			act, _, err := m.(*hcloud.Client).Server.Poweroff(context.Background(), server)
			if err != nil {
				return err
			}

			// wait for action to finish and check for errors
			_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
			err = <-errch
			if err != nil {
				return err
			}
		}

		// send server change type request
		act, _, err := m.(*hcloud.Client).Server.ChangeType(context.Background(), server, hcloud.ServerChangeTypeOpts{
			ServerType: &hcloud.ServerType{
				ID: d.Get("server_type").(int),
			},
			UpgradeDisk: d.Get("upgrade_disk").(bool),
		})
		if err != nil {
			return nil
		}

		// wait for action to finish and check for errors
		_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
		err = <-errch
		if err != nil {
			return err
		}

		// start server, if it was running beforehand
		if restart {
			// power on request
			act, _, err := m.(*hcloud.Client).Server.Poweron(context.Background(), server)
			if err != nil {
				return err
			}

			// wait for action to finish and check for errors
			_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
			err = <-errch
			if err != nil {
				return err
			}
		}
	}

	// enable backup and/or update backup window, if necessary
	if d.Get("backup").(bool) && (d.Get("backup_window").(string) != server.BackupWindow || len(server.BackupWindow) < 1) {
		// send enable backup action
		act, _, err := m.(*hcloud.Client).Server.EnableBackup(context.Background(), server, d.Get("backup_window").(string))
		if err != nil {
			return err
		}

		// wait for action to finish and check for errors
		_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
		err = <-errch
		if err != nil {
			return err
		}

		// get randomly assigned backup window, if none was set
		if len(d.Get("backup_window").(string)) < 1 {
			server, _, err := m.(*hcloud.Client).Server.GetByID(context.Background(), id)
			if err != nil {
				return err
			}

			d.Set("backup_windows", server.BackupWindow)
		}
	}

	// disable backup, if neccessary
	if !d.Get("backup").(bool) {
		// send disable backup action
		act, _, err := m.(*hcloud.Client).Server.DisableBackup(context.Background(), server)
		if err != nil {
			return err
		}

		// wait for action to finish and check for errors
		_, errch := m.(*hcloud.Client).Action.WatchProgress(context.Background(), act)
		err = <-errch
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceHcloudServerDelete(d *schema.ResourceData, m interface{}) error {
	// convert if from string to int
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	// get server object
	server, _, err := m.(*hcloud.Client).Server.GetByID(context.Background(), id)
	if err != nil {
		return err
	}

	// send server delete request
	_, err = m.(*hcloud.Client).Server.Delete(context.Background(), server)
	return err

	// hcloud go library currently doesn't support checking the action status
}