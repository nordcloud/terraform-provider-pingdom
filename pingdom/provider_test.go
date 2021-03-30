package pingdom

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ProviderNamePingdom = "pingdom"

	envSolarwindsUser     = "SOLARWINDS_USER"
	envSolarwindsPassword = "SOLARWINDS_PASSWD"
)

var testAccProviderFactories map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider

var testAccProviderConfigure sync.Once

func init() {
	testAccProvider = Provider()
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		ProviderNamePingdom: func() (*schema.Provider, error) { return Provider(), nil },
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
		t.Fatalf("err: %v", err)
	}

	config := rp.Meta().(*Clients).Pingdom

	if config.APIToken != expectedToken {
		t.Fatalf("bad: %#v", config)
	}
}

func testAccPreCheck(t *testing.T) {
	testAccProviderConfigure.Do(func() {
		for _, envvar := range []string{envSolarwindsUser, envSolarwindsPassword} {
			if os.Getenv(envvar) == "" {
				t.Fatalf("environment variable %s is required for solarwinds resource tests to run", envvar)
			}
		}

		err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
		if err != nil {
			t.Fatal(err)
		}
	})
}
