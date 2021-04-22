package pingdom

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func dataSourceSolarwindsActiveUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSolarwindsActiveUsersRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSolarwindsActiveUsersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Clients).Solarwinds

	log.Printf("[INFO] Reading all active users on Solarwinds")

	resp, err := client.ActiveUserService.List()
	if err != nil {
		return diag.FromErr(err)
	}

	var ids []string
	var names []string

	for _, member := range resp.Organization.Members {
		ids = append(ids, member.User.Id)
		names = append(names, fmt.Sprintf("%s %s", member.User.FirstName, member.User.LastName))
	}
	d.SetId(fmt.Sprintf("%s-%d", resp.Organization.Id, len(resp.Organization.Members)))
	if err := d.Set("ids", ids); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("names", names); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
