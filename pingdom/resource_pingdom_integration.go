package pingdom

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nordcloud/go-pingdom/pingdomext"
)

func resourcePingdomIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePingdomIntegrationCreate,
		ReadContext:   resourcePingdomIntegrationRead,
		UpdateContext: resourcePingdomIntegrationUpdate,
		DeleteContext: resourcePingdomIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"provider_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"data": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func integrationForResource(d *schema.ResourceData, client *pingdomext.Client) (pingdomext.Integration, error) {
	integration := &pingdomext.WebHookIntegration{}

	// required
	if v, ok := d.GetOk("provider_name"); ok {
		integrationProvider, err := getIntegrationProvider(v.(string), client)
		if err != nil {
			return nil, err
		}
		integration.ProviderID = integrationProvider.ID
	}

	if v, ok := d.GetOk("active"); ok {
		integration.Active = v.(bool)
	}

	if v, ok := d.GetOk("data"); ok {
		data := v.(map[string]interface{})
		userData := &pingdomext.WebHookData{
			Name: data["name"].(string),
			URL:  data["url"].(string),
		}
		integration.UserData = userData
	}

	return integration, nil
}

func resourcePingdomIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Clients).PingdomExt

	integration, err := integrationForResource(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	//log.Printf("[DEBUG] Integration create configuration: %#v", d.Get("data").(map[string]interface{})["Name"].(string))
	result, err := client.Integrations.Create(integration)
	if err != nil {
		return diag.FromErr(err)
	}

	if !result.Status {
		return diag.Errorf("Integration create failed.")
	}

	d.SetId(strconv.Itoa(result.ID))
	return nil
}

func resourcePingdomIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Clients).PingdomExt

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Error retrieving id for resource: %s", err)
	}

	integration, err := integrationForResource(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	//log.Printf("[DEBUG] Integration update configuration: %#v", d.Get("data").(map[string]interface{})["Name"].(string))
	result, err := client.Integrations.Update(id, integration)
	if err != nil {
		return diag.FromErr(err)
	}
	if !result.Status {
		return diag.Errorf("Integration update failed.")
	}

	return nil
}

func resourcePingdomIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Clients).PingdomExt

	integrations, err := client.Integrations.List()
	if err != nil {
		return diag.Errorf("Error retrieving list of integrations: %s", err)
	}
	exists := false
	for _, integration := range integrations {
		if strconv.Itoa(integration.ID) == d.Id() {
			exists = true
			break
		}
	}
	if !exists {
		d.SetId("")
		return nil
	}
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Error retrieving id for resource: %s", err)
	}
	integration, err := client.Integrations.Read(id)
	if err != nil {
		return diag.Errorf("Error retrieving integration: %s", err)
	}

	if err := d.Set("provider_name", integration.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("active", integration.ActivatedAt != 0); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("data", integration.UserData); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePingdomIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Clients).PingdomExt
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Error retrieving id for resource: %s", err)
	}

	result, err := client.Integrations.Delete(id)
	if err != nil {
		return diag.Errorf("Error deleting integration: %s", err)
	}

	if !result.Status {
		return diag.Errorf("Integration delete failed.")
	}
	return nil
}

func getIntegrationProvider(providerName string, client *pingdomext.Client) (*pingdomext.IntegrationProvider, error) {

	providers, err := client.Integrations.ListProviders()
	if err != nil {
		return nil, err
	}

	for _, provider := range providers {
		if provider.Name == providerName {
			return &provider, nil
		}
	}
	return nil, fmt.Errorf("Unable find the integration provider %s", providerName)
}
