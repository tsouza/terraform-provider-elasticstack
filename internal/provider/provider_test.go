package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = New("dev")()
	testAccProviders = map[string]*schema.Provider{
		"elasticstack": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := New("dev")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if err := os.Getenv("ELASTICSEARCH_URL"); err == "" {
		t.Fatal("ELASTICSEARCH_URL must be set for acceptance tests")
	}
	if err := os.Getenv("ELASTICSEARCH_USER"); err == "" {
		t.Fatal("ELASTICSEARCH_USER must be set for acceptance tests")
	}
	if err := os.Getenv("ELASTICSEARCH_PASSWORD"); err == "" {
		t.Fatal("ELASTICSEARCH_PASSWORD must be set for acceptance tests")
	}
}
