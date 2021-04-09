package pingdom

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nordcloud/go-pingdom/pingdom"
)

func TestAccResourcePingdomMaintenance_basic(t *testing.T) {
	resourceName := "pingdom_maintenance.test"
	checkResourceName := "pingdom_check.test"

	description := acctest.RandomWithPrefix("tf-acc-test")
	updatedDescription := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPingdomMaintenanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePingdomMaintenanceConfig(description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPingdomResourceID(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "from", "2717878696"),
					resource.TestCheckResourceAttr(resourceName, "to", "2718878696"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourcePingdomMaintenanceConfigUpdate(updatedDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPingdomResourceID(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
					resource.TestCheckResourceAttr(resourceName, "from", "2717878693"),
					resource.TestCheckResourceAttr(resourceName, "to", "2718878693"),
					resource.TestCheckResourceAttr(resourceName, "effectiveto", "2718978693"),
					resource.TestCheckResourceAttr(resourceName, "recurrencetype", "week"),
					resource.TestCheckResourceAttr(resourceName, "repeatevery", "4"),
					resource.TestCheckResourceAttr(resourceName, "uptimeids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "uptimeids.*", checkResourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckPingdomMaintenanceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pingdom.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pingdom_maintenance" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Maintenance ID is not valid: %s", rs.Primary.ID)
		}

		resp, err := client.Maintenances.Read(id)
		if err == nil {
			if strconv.Itoa(resp.ID) == rs.Primary.ID {
				return fmt.Errorf("Maintenance (%s) still exists.", rs.Primary.ID)
			}
		}

		if !strings.Contains(err.Error(), "404") {
			return err
		}
	}

	return nil
}

func testAccResourcePingdomMaintenanceConfig(name string) string {
	return fmt.Sprintf(`
resource "pingdom_maintenance" "test" {
	description = "%s"
	from        = 2717878696
	to          = 2718878696
}
`, name)
}

func testAccResourcePingdomMaintenanceConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "pingdom_check" "test" {
	name = "%s"
	host = "www.example.com"
	type = "http"
}

resource "pingdom_maintenance" "test" {
	description    = "%s"
	from           = 2717878693
	to             = 2718878693
	effectiveto    = 2718978693
	recurrencetype = "week"
	repeatevery    = 4
	uptimeids      = [pingdom_check.test.id]
}
`, name, name)
}
