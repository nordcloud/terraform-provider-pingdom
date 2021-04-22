package pingdom

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestAccDataSourcePingdomActiveUsers_basic(t *testing.T) {
	resourceName := "data.pingdom_active_users.test"

	rp := Provider()
	diag := rp.Configure(context.Background(), terraform.NewResourceConfigRaw(map[string]interface{}{}))
	assert.Nil(t, diag)

	client := rp.Meta().(*Clients).Solarwinds
	resp, err := client.ActiveUserService.List()
	assert.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePingdomActiveUsersConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPingdomResourceID(resourceName),
					resource.TestCheckResourceAttr(resourceName, "names.#", strconv.Itoa(len(resp.Organization.Members))),
					resource.TestCheckResourceAttr(resourceName, "ids.#", strconv.Itoa(len(resp.Organization.Members))),
				),
			},
		},
	})
}

func testAccDataSourcePingdomActiveUsersConfig() string {
	return `
data "pingdom_active_users" "test" {
}
`
}
