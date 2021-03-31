package pingdom

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"pingdom": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderConfigure(t *testing.T) {
	var expectedToken string

	if v := os.Getenv("PINGDOM_API_TOKEN"); v != "" {
		expectedToken = v
	} else {
		expectedToken = "foo"
	}

	raw := map[string]interface{}{
		"api_token": expectedToken,
	}

	rp := Provider()
	err := rp.Configure(context.Background(), terraform.NewResourceConfigRaw(raw))
	if err != nil {
		t.Fatal(err)
	}

	config := rp.Meta().(*Clients).Pingdom

	if config.APIToken != expectedToken {
		t.Fatalf("bad: %#v", config)
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("PINGDOM_API_TOKEN"); v == "" {
		t.Fatal("PINGDOM_API_TOKEN environment variable must be set for acceptance tests")
	}
}
