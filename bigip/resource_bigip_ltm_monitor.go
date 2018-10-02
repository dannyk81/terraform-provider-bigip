package bigip

import (
	"fmt"
	"log"
	"strings"

	"github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceBigipLtmMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceBigipLtmMonitorCreate,
		Read:   resourceBigipLtmMonitorRead,
		Update: resourceBigipLtmMonitorUpdate,
		Delete: resourceBigipLtmMonitorDelete,
		Exists: resourceBigipLtmMonitorExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Name of the monitor",
				ForceNew:     true,
				ValidateFunc: validateF5Name,
			},

			"parent": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateParent,
				ForceNew:     true,
				Description:  "Existing monitor to inherit from. Must be one of /Common/http, /Common/https, /Common/icmp, /Common/gateway-icmp, /Common/tcp-half-open or /Common/tcp",
			},
			"defaults_from": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Specifies the existing monitor from which the system imports settings for the new monitor",
			},

			"interval": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Check interval in seconds",
				Default:     3,
			},

			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Timeout in seconds",
				Default:     16,
			},

			"send": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Request string to send.",
				StateFunc: func(s interface{}) string {
					return strings.Replace(s.(string), "\r\n", "\\r\\n", -1)
				},
			},

			"receive": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Expected response string.",
			},

			"receive_disable": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Expected response string.",
			},

			"reverse": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"transparent": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "disabled",
			},

			"manual_resume": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "disabled",
			},

			"ip_dscp": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},

			"time_until_up": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Time in seconds",
			},

			"destination": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "*:*",
				Description: "Alias for the destination",
			},
		},
	}
}

func resourceBigipLtmMonitorCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)

	name := d.Get("name").(string)
	parent := monitorParent(d.Get("parent").(string))
	log.Printf("[DEBUG] Creating monitor %s::%s", name, parent)

	err := client.CreateMonitor(
		name,
		parent,
		d.Get("defaults_from").(string),
		d.Get("interval").(int),
		d.Get("timeout").(int),
		d.Get("send").(string),
		d.Get("receive").(string),
		d.Get("receive_disable").(string),
	)
	if err != nil {
		return fmt.Errorf("Error creating Monitor %s: %v", name, err)
	}

	d.SetId(name)

	return resourceBigipLtmMonitorUpdate(d, meta)
}

func resourceBigipLtmMonitorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	parent := monitorParent(d.Get("parent").(string))
	log.Printf("[DEBUG] Reading Monitor %s::%s", name, parent)

	monitors, err := client.Monitors()
	if err != nil {
		return fmt.Errorf("Unable to retrieve Monitors: %v", err)
	}
	if monitors == nil {
		log.Printf("[DEBUG] Monitors not found, removing Monitor %s::%s from state", name, parent)
		d.SetId("")
		return nil
	}

	for _, m := range monitors {
		if m.FullPath == name {
			d.Set("name", m.FullPath)
			d.Set("parent", m.ParentMonitor)
			d.Set("defaults_from", m.DefaultsFrom)
			d.Set("interval", m.Interval)
			d.Set("timeout", m.Timeout)
			d.Set("send", m.SendString)
			d.Set("receive", m.ReceiveString)
			d.Set("receive_disable", m.ReceiveDisable)
			d.Set("reverse", m.Reverse)
			d.Set("transparent", m.Transparent)
			d.Set("ip_dscp", m.IPDSCP)
			d.Set("time_until_up", m.TimeUntilUp)
			d.Set("manual_resume", m.ManualResume)
			d.Set("destination", m.Destination)
			return nil
		}
	}

	log.Printf("[DEBUG] Monitor %s::%s not found, removing it from state", name, parent)
	d.SetId("")
	return nil
}

func resourceBigipLtmMonitorExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	parent := monitorParent(d.Get("parent").(string))
	log.Printf("[DEBUG] Checking if Monitor %s::%s exists", name, parent)

	monitors, err := client.Monitors()
	if err != nil {
		return false, fmt.Errorf("Unable to retrieve Monitors: %v", err)
	}
	if monitors == nil {
		log.Println("[DEBUG] Monitors not found")
		return false, nil
	}
	for _, m := range monitors {
		if m.FullPath == name {
			return true, nil
		}
	}

	return false, nil
}

func resourceBigipLtmMonitorUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	parent := monitorParent(d.Get("parent").(string))
	log.Printf("[DEBUG] Updating Monitor %s::%s", name, parent)

	m := &bigip.Monitor{
		Interval:       d.Get("interval").(int),
		Timeout:        d.Get("timeout").(int),
		SendString:     d.Get("send").(string),
		ReceiveString:  d.Get("receive").(string),
		ReceiveDisable: d.Get("receive_disable").(string),
		Reverse:        d.Get("reverse").(string),
		Transparent:    d.Get("transparent").(string),
		IPDSCP:         d.Get("ip_dscp").(int),
		TimeUntilUp:    d.Get("time_until_up").(int),
		ManualResume:   d.Get("manual_resume").(string),
		Destination:    d.Get("destination").(string),
	}

	err := client.ModifyMonitor(name, parent, m)
	if err != nil {
		return fmt.Errorf("Error updating Monitor %s::%s: %v", name, parent, err)
	}

	return resourceBigipLtmMonitorRead(d, meta)
}

func resourceBigipLtmMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	parent := monitorParent(d.Get("parent").(string))

	log.Printf("[DEBUG] Deleting Monitor %s::%s", name, parent)

	err := client.DeleteMonitor(name, parent)
	if err != nil {
		return fmt.Errorf("Error deleting Monitor %s::%s: %v", name, parent, err)
	}

	d.SetId("")
	return nil
}

func validateParent(v interface{}, k string) ([]string, []error) {
	p := v.(string)
	if p == "/Common/http" || p == "/Common/https" || p == "/Common/icmp" || p == "/Common/gateway-icmp" || p == "/Common/tcp" || p == "/Common/tcp-half-open" {
		return nil, nil
	}

	return nil, []error{fmt.Errorf("parent must be one of /Common/http, /Common/https, /Common/icmp, /Common/gateway-icmp, /Common/tcp-half-open,  or /Common/tcp")}
}

func monitorParent(s string) string {
	return strings.TrimPrefix(s, "/Common/")
}
