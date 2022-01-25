data "publicip_address" "default" {
}

data "publicip_address" "v4" {
  ip_version = "v4"
}

data "publicip_address" "v6" {
  ip_version = "v6"
}
