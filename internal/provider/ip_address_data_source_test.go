package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestIpAddressDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: defaultConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.publicip_address.default", "ip_version"),
					resource.TestCheckResourceAttrSet("data.publicip_address.default", "ip"),
					resource.TestCheckResourceAttrSet("data.publicip_address.default", "id"),
					resource.TestCheckResourceAttr("data.publicip_address.default", "source_ip", ""),
				),
			},
			{
				Config: v6Config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.publicip_address.v6", "ip_version", "v6"),
					resource.TestCheckResourceAttrSet("data.publicip_address.v6", "ip"),
					resource.TestCheckResourceAttrSet("data.publicip_address.v6", "id"),
					resource.TestCheckResourceAttr("data.publicip_address.v6", "source_ip", ""),
				),
			},
			{
				Config: v4Config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.publicip_address.v4", "ip_version", "v4"),
					resource.TestCheckResourceAttrSet("data.publicip_address.v4", "ip"),
					resource.TestCheckResourceAttrSet("data.publicip_address.v4", "id"),
					resource.TestCheckResourceAttr("data.publicip_address.v4", "source_ip", ""),
				),
			},
			{
				Config: v6SrcConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.publicip_address.v6src", "ip_version", "v6"),
					resource.TestCheckResourceAttrSet("data.publicip_address.v6src", "ip"),
					resource.TestCheckResourceAttrSet("data.publicip_address.v6src", "id"),
					resource.TestCheckResourceAttr("data.publicip_address.v6src", "source_ip", "::"),
				),
			},
			{
				Config: v4SrcConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.publicip_address.v4src", "ip_version", "v4"),
					resource.TestCheckResourceAttrSet("data.publicip_address.v4src", "ip"),
					resource.TestCheckResourceAttrSet("data.publicip_address.v4src", "id"),
					resource.TestCheckResourceAttr("data.publicip_address.v4src", "source_ip", "0.0.0.0"),
				),
			},
		},
	})
}

const defaultConfig = `
data "publicip_address" "default" {
}
`

const v6Config = `
data "publicip_address" "v6" {
  ip_version = "v6"
}
`

const v4Config = `
data "publicip_address" "v4" {
  ip_version = "v4"
}
`

const v6SrcConfig = `
data "publicip_address" "v6src" {
  source_ip = "::"
}
`

const v4SrcConfig = `
data "publicip_address" "v4src" {
  source_ip = "0.0.0.0"
}
`
