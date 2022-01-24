package provider

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// provider satisfies the tfsdk.Provider interface and usually is included
// with all Resource and DataSource implementations.
type provider struct {
	timeout time.Duration
	ipURL   *url.URL

	// configured is set to true at the end of the Configure method.
	// This can be used in Resource and DataSource implementations to verify
	// that the provider was previously configured.
	configured bool

	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string

	// toolName is the name of this provider.
	toolName string
}

// providerData can be used to store data from the Terraform configuration.
type providerData struct {
	ProviderURL types.String `tfsdk:"provider_url"`
	Timeout     types.String `tfsdk:"timeout"`
}

const DefaultTimeout = "5s"
const DefaultProviderURL = "https://ifconfig.co/"

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var data providerData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var providerURL string
	if data.ProviderURL.Null {
		providerURL = DefaultProviderURL
	} else {
		providerURL = data.ProviderURL.Value
	}

	var err error
	p.ipURL, err = url.Parse(providerURL)
	if err != nil {
		resp.Diagnostics.AddError("Unable to parse the provider_url", fmt.Sprintf("The provider_url '%s' can't be parsed: %s", providerURL, err))
		return
	}

	var timeout string
	if data.Timeout.Null {
		timeout = DefaultTimeout
	} else {
		timeout = data.Timeout.Value
	}

	p.timeout, err = time.ParseDuration(timeout)
	if err != nil {
		resp.Diagnostics.AddError("Unable to parse the timeout", fmt.Sprintf("The timeout '%s' can't be parsed: %s", timeout, err))
		return
	}

	p.configured = true
}

func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		// "scaffolding_example": exampleResourceType{},
	}, nil
}

func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"publicip_address": ipDataSourceType{},
	}, nil
}

func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"timeout": {
				MarkdownDescription: fmt.Sprintf("Timeout of the request to the IP information provider, defaults to `%s`.", DefaultTimeout),
				Optional:            true,
				Type:                types.StringType,
			},
			"provider_url": {
				MarkdownDescription: fmt.Sprintf("URL to a ifconfig.co-compatible IP information provider, defaults to `%s`.", DefaultProviderURL),
				Optional:            true,
				Type:                types.StringType,
			},
		},
	}, nil
}

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in tfsdk.Provider) (provider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*provider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return provider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return provider{}, diags
	}

	return *p, diags
}
