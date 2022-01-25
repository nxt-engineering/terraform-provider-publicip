package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/nxt-engineering/terraform-provider-publicip/internal/provider"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version = "dev"
	commit  = ""
	date    = ""
)

const toolName = "terraform-provider-publicip"

func main() {
	log.Printf("%s Version: %s Commit: %s Date: %s", toolName, version, commit, date)

	opts := tfsdk.ServeOpts{
		Name: "registry.terraform.io/nxt-engineering/publicip",
	}

	err := tfsdk.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
