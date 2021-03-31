package pingdom

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccUser(t *testing.T) {
	email := acctest.RandString(10) + "@foo.com"
	resourceName := "pingdom_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(email, "MEMBER", "APPOPTICS", "MEMBER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExist(resourceName),
					resource.TestCheckResourceAttr(resourceName, "role", "MEMBER"),
					resource.TestCheckResourceAttr(resourceName, "products.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "products.0.name", "APPOPTICS"),
					resource.TestCheckResourceAttr(resourceName, "products.0.role", "MEMBER"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckUserDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Clients).Solarwinds

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pingdom_user" {
			continue
		}

		email := rs.Primary.ID
		user, err := client.UserService.Retrieve(email)
		if err != nil {
			return err
		}
		if user != nil {
			return fmt.Errorf("user for resource (%s) still exists", email)
		}
	}
	return nil
}

func testAccCheckUserExist(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no id is set")
		}

		email := rs.Primary.ID
		client := testAccProvider.Meta().(*Clients).Solarwinds
		user, err := client.UserService.Retrieve(email)
		if err != nil {
			return err
		}
		if user == nil {
			return fmt.Errorf("user for resource (%s) not found", email)
		}
		return nil
	}
}

func testAccUserConfig(email string, role string, productName string, productRole string) string {
	return fmt.Sprintf(`
resource "pingdom_user" "test" {
	email = %[1]q
	role = %[2]q
	products {
		name = %[3]q
		role = %[4]q
	}
}
`, email, role, productName, productRole)
}
