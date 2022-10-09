package provider

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/time/rate"
)

const TypeName = "publicip"
const UserAgent = "terraform-provider-publicip"

type IpProvider struct {
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

// ProviderModel can be used to store data from the Terraform configuration.
type ProviderModel struct {
	ProviderURL    types.String `tfsdk:"provider_url"`
	Timeout        types.String `tfsdk:"timeout"`
	RateLimitRate  types.String `tfsdk:"rate_limit_rate"`
	RateLimitBurst types.Int64  `tfsdk:"rate_limit_burst"`

	version       string
	ipProviderURL *url.URL
	timeout       time.Duration
	rateLimiter   *rate.Limiter
}

const DefaultTimeout = "5s"
const DefaultProviderURL = "https://ifconfig.co/"
const DefaultRateLimitRate = "500ms"
const DefaultRateLimitBurst = 1

func (p *IpProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProviderModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.version = p.version
	if !p.configureProviderURL(&data, resp) ||
		!p.configureTimeout(&data, resp) ||
		!p.configureRateLimiter(&data, resp) {
		return
	}

	resp.DataSourceData = &data
	p.configured = true
}

func (p *IpProvider) configureProviderURL(data *ProviderModel, resp *provider.ConfigureResponse) bool {
	var providerURL string
	if data.ProviderURL.Null {
		providerURL = DefaultProviderURL
	} else {
		providerURL = data.ProviderURL.Value
	}

	var err error
	data.ipProviderURL, err = url.Parse(providerURL)

	if err != nil {
		resp.Diagnostics.AddError("Unable to parse the provider_url", fmt.Sprintf("The provider_url value '%s' can't be parsed: %s", providerURL, err))
		return false
	}
	return true
}

func (p *IpProvider) configureTimeout(data *ProviderModel, resp *provider.ConfigureResponse) bool {
	var timeout string
	if data.Timeout.Null {
		timeout = DefaultTimeout
	} else {
		timeout = data.Timeout.Value
	}

	var err error
	data.timeout, err = time.ParseDuration(timeout)
	if err != nil {
		resp.Diagnostics.AddError("Unable to parse the timeout", fmt.Sprintf("The timeout value '%s' can't be parsed: %s", timeout, err))
		return false
	}
	return true
}

func (p *IpProvider) configureRateLimiter(data *ProviderModel, resp *provider.ConfigureResponse) bool {
	var rateLimitRate string
	if data.RateLimitRate.Null {
		rateLimitRate = DefaultRateLimitRate
	} else {
		rateLimitRate = data.RateLimitRate.Value
	}

	rateLimitRateDuration, err := time.ParseDuration(rateLimitRate)
	if err != nil {
		resp.Diagnostics.AddError("Unable to parse the rate_limit_rate", fmt.Sprintf("The rate_limit_rate value '%s' can't be parsed: %s", rateLimitRate, err))
		return false
	}

	var rateLimitBurst int
	if data.RateLimitBurst.Null {
		rateLimitBurst = DefaultRateLimitBurst
	} else if data.RateLimitBurst.Value > math.MaxInt {
		resp.Diagnostics.AddError("Unable to use the rate_limit_burst", fmt.Sprintf("The rate_limit_burst value '%d' is too big. Maximum allowed is %d", data.RateLimitBurst.Value, math.MaxInt))
		return false
	} else if data.RateLimitBurst.Value <= 0 {
		resp.Diagnostics.AddError("Unable to use the rate_limit_burst", fmt.Sprintf("The rate_limit_burst value '%d' must be bigger than 0", data.RateLimitBurst.Value))
		return false
	} else {
		rateLimitBurst = int(data.RateLimitBurst.Value)
	}

	data.rateLimiter = rate.NewLimiter(rate.Every(rateLimitRateDuration), rateLimitBurst)

	return true
}

func (p *IpProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = TypeName
}

func (p *IpProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}

func (p *IpProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewIpDataSource,
	}
}

func (p *IpProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"timeout": {
				MarkdownDescription: fmt.Sprintf("Timeout of the request to the IP information provider. Defaults to `%s`.", DefaultTimeout),
				Optional:            true,
				Type:                types.StringType,
			},
			"rate_limit_rate": {
				MarkdownDescription: fmt.Sprintf("Limit the number of the request to the IP information provider. Defines the time until the limit is reset. Defaults to `%s`.", DefaultRateLimitRate),
				Optional:            true,
				Type:                types.StringType,
			},
			"rate_limit_burst": {
				MarkdownDescription: fmt.Sprintf("Limit the number of the request to the IP information provider. Defines the number of events per rate until the limit is reached. Defaults to `%d`.", DefaultRateLimitBurst),
				Optional:            true,
				Type:                types.Int64Type,
			},
			"provider_url": {
				MarkdownDescription: fmt.Sprintf("URL to an ifconfig.co-compatible IP information provider, defaults to `%s`.", DefaultProviderURL),
				Optional:            true,
				Type:                types.StringType,
			},
		},
	}, nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &IpProvider{
			version: version,
		}
	}
}
