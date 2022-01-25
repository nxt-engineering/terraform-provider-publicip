data "publicip_address" "default" {
}

data "publicip_address" "v6" {
  ip_version = "v6"
}

data "publicip_address" "v4" {
  ip_version = "v4"
}

data "publicip_address" "source_v6" {
  source_ip = "::"
}

data "publicip_address" "source_v4" {
  source_ip = "0.0.0.0"
}
