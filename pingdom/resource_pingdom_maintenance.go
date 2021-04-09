package pingdom

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nordcloud/go-pingdom/pingdom"
)

const timeFormat = time.RFC3339

func resourcePingdomMaintenance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePingdomMaintenanceCreate,
		ReadContext:   resourcePingdomMaintenanceRead,
		UpdateContext: resourcePingdomMaintenanceUpdate,
		DeleteContext: resourcePingdomMaintenanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"from": {
				Type:     schema.TypeString,
				Required: true,
			},
			"to": {
				Type:     schema.TypeString,
				Required: true,
			},
			"effectiveto": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"recurrencetype": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "none",
			},
			"repeatevery": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"tmsids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"uptimeids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func maintenanceForResource(d *schema.ResourceData) (*pingdom.MaintenanceWindow, error) {
	maintenance := pingdom.MaintenanceWindow{}

	// required
	if v, ok := d.GetOk("description"); ok {
		maintenance.Description = v.(string)
	}

	if v, ok := d.GetOk("from"); ok {
		t, err := time.Parse(timeFormat, v.(string))
		if err != nil {
			return nil, err
		}
		maintenance.From = t.Unix()
	}

	if v, ok := d.GetOk("to"); ok {
		t, err := time.Parse(timeFormat, v.(string))
		if err != nil {
			return nil, err
		}
		maintenance.To = t.Unix()
	}

	if v, ok := d.GetOk("effectiveto"); ok {
		t, err := time.Parse(timeFormat, v.(string))
		if err != nil {
			return nil, err
		}
		maintenance.EffectiveTo = t.Unix()
	}

	if v, ok := d.GetOk("recurrencetype"); ok {
		maintenance.RecurrenceType = v.(string)
	}

	if v, ok := d.GetOk("repeatevery"); ok {
		maintenance.RepeatEvery = v.(int)
	}

	if v, ok := d.GetOk("tmsids"); ok {
		maintenance.TmsIDs = convertIntInterfaceSliceToString(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("uptimeids"); ok {
		maintenance.UptimeIDs = convertIntInterfaceSliceToString(v.(*schema.Set).List())
	}

	return &maintenance, nil
}

func updateResourceFromMaintenanceResponse(d *schema.ResourceData, m *pingdom.MaintenanceResponse) error {
	if err := d.Set("description", m.Description); err != nil {
		return err
	}

	if err := d.Set("from", time.Unix(m.From, 0).Format(timeFormat)); err != nil {
		return err
	}

	if err := d.Set("to", time.Unix(m.To, 0).Format(timeFormat)); err != nil {
		return err
	}

	if err := d.Set("effectiveto", time.Unix(m.EffectiveTo, 0).Format(timeFormat)); err != nil {
		return err
	}

	if err := d.Set("recurrencetype", m.RecurrenceType); err != nil {
		return err
	}

	if err := d.Set("repeatevery", m.RepeatEvery); err != nil {
		return err
	}

	tmsids := schema.NewSet(
		func(tmsId interface{}) int { return tmsId.(int) },
		[]interface{}{},
	)
	for _, tms := range m.Checks.Tms {
		tmsids.Add(tms)
	}
	if err := d.Set("tmsids", tmsids); err != nil {
		return err
	}

	uptimeids := schema.NewSet(
		func(uptimeId interface{}) int { return uptimeId.(int) },
		[]interface{}{},
	)
	for _, uptime := range m.Checks.Uptime {
		uptimeids.Add(uptime)
	}
	if err := d.Set("uptimeids", uptimeids); err != nil {
		return err
	}

	return nil
}

func resourcePingdomMaintenanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*pingdom.Client)

	maintenance, err := maintenanceForResource(d)
	if err != nil {
		return diag.FromErr(err)
	}

	result, err := client.Maintenances.Create(maintenance)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(result.ID))

	return nil
}

func resourcePingdomMaintenanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*pingdom.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Error retrieving id for resource: %s", err)
	}
	maintenance, err := client.Maintenances.Read(id)
	if err != nil {
		return diag.Errorf("Error retrieving maintenance: %s", err)
	}

	if err := updateResourceFromMaintenanceResponse(d, maintenance); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourcePingdomMaintenanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*pingdom.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Error retrieving id for resource: %s", err)
	}
	maintenance, err := maintenanceForResource(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err = client.Maintenances.Update(id, maintenance); err != nil {
		return diag.Errorf("Error updating maintenance: %s", err)
	}

	return nil
}

func resourcePingdomMaintenanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*pingdom.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Error retrieving id for resource: %s", err)
	}
	if _, err := client.Maintenances.Delete(id); err != nil {
		return diag.Errorf("Error deleting maintenance: %s", err)
	}
	return nil
}

func convertIntInterfaceSliceToString(slice []interface{}) string {
	stringSlice := make([]string, len(slice))
	for i := range slice {
		stringSlice[i] = strconv.Itoa(slice[i].(int))
	}
	return strings.Join(stringSlice, ",")
}
