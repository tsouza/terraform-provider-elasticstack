package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccElasticstackAuthUserBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckElasticstackAuthUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckElasticstackAuthUserConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticstackAuthUserExists("elasticstack_auth_user.test"),
				),
			},
		},
	})
}

func testAccCheckElasticstackAuthUserDestroy(s *terraform.State) error {
	es := testAccProvider.Meta().(*apiClient).es

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_auth_user" {
			continue
		}

		username := rs.Primary.ID

		req := esapi.SecurityGetUserRequest{
			Username: []string{username},
		}

		res, err := req.Do(context.Background(), es)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != 404 {
			return fmt.Errorf("User '%s' still exists: %s", username, res)
		}
	}

	return nil
}

func testAccCheckElasticstackAuthUserConfigBasic() string {
	username := "test1"
	fullName := "test tester"
	email := "test@example.com"
	password := "foobar"
	return fmt.Sprintf(`
	resource "elasticstack_auth_user" "test" {
		username = "%s"
		full_name = "%s"
		email = "%s"
		password = "%s"
		roles = []
	}
	`, username, fullName, email, password)
}

func testAccCheckElasticstackAuthUserExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No user ID set")
		}

		return nil
	}
}
