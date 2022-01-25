package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccExampleDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: defaultConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.publicip_address.default", "ip_version"),
				),
			},
			{
				Config: v6Config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.publicip_address.v6", "ip_version", "v6"),
				),
			},
			{
				Config: v4Config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.publicip_address.v4", "ip_version", "v4"),
				),
			},
		},
	})
}

const defaultConfig = `
data "publicip_address" "default" {
}
`

const v4Config = `
data "publicip_address" "v4" {
  ip_version = "v4"
}
`

const v6Config = `
data "publicip_address" "v6" {
  ip_version = "v6"
}
`
