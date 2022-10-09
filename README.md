# Public IP Terraform Provider

This Terraform provider connects to [ifconfig.co](https://ifconfig.co) to fetch information about the public IP in use.
It was mainly built to allow creating or adjusting firewall rules in cloud setups on the fly,
so that other providers that connect directly to the resources can operate.

For example and at least until version 3,
the `azurerm` provider needed to connect to the Azure Storage account directly to manage blob containers.
Likewise, a `postgresql` provider needs to directly connect to the database to do its job.

## Usage

```terraform
terraform {
  required_providers {
    publicip = {
      source = "nxt-engineering/publicip"
      version = "~> 0.0.3"
    }
  }
}

provider "publicip" {}

data "publicip_address" "main" {}

output "ip" {
  value = data.publicip_address.main.ip
}

output "all" {
  value = data.publicip_address.main
}
```


## Development

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.17

### Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

### Testrun

Add this to your `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    # Edit the path …                         … here:  |________|
    "registry.terraform.io/nxt-engineering/publicip" = "HOME/SRC/terraform-provider-publicip"
  }

  direct {}
}
```

Set up a dummy project like this:

```terraform
terraform {
  required_providers {
    publicip = {
      source = "nxt-engineering/publicip"
    }
  }
}

provider "publicip" {}

data "publicip_address" "default" {}

data "publicip_address" "default_v4" {
  source_ip = "0.0.0.0"
}

data "publicip_address" "default_v6" {
  source_ip = "::"
}

output "out" {
  value = {
    default    = data.publicip_address.default,
    default_v4 = data.publicip_address.default_v4,
    default_v6 = data.publicip_address.default_v6,
  }
}
```

Build the provider:

```bash
go build -o terraform-provider-publicip
```

Run the provider like this:

```bash
TF_LOG_PROVIDER=DEBUG terraform apply -auto-approve
```
