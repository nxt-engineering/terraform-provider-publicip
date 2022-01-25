package provider

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/time/rate"
)

// provider satisfies the tfsdk.Provider interface and usually is included
// with all Resource and DataSource implementations.
type provider struct {
	timeout     time.Duration
	ipURL       *url.URL
	rateLimiter *rate.Limiter

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
	ProviderURL    types.String `tfsdk:"provider_url"`
	Timeout        types.String `tfsdk:"timeout"`
	RateLimitRate  types.String `tfsdk:"rate_limit_rate"`
	RateLimitBurst types.Int64  `tfsdk:"rate_limit_burst"`
}

const DefaultTimeout = "5s"
const DefaultProviderURL = "https://ifconfig.co/"
const DefaultRateLimitRate = "500ms"
const DefaultRateLimitBurst = 1

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

	if !p.configureProviderURL(providerURL, resp) ||
		!p.configureTimeout(data, resp) ||
		!p.configureRateLimiter(data, resp) {
		return
	}

	p.configured = true
}

func (p *provider) configureProviderURL(providerURL string, resp *tfsdk.ConfigureProviderResponse) bool {
	var err error
	p.ipURL, err = url.Parse(providerURL)

	if err != nil {
		resp.Diagnostics.AddError("Unable to parse the provider_url", fmt.Sprintf("The provider_url value '%s' can't be parsed: %s", providerURL, err))
		return false
	}
	return true
}

func (p *provider) configureTimeout(data providerData, resp *tfsdk.ConfigureProviderResponse) bool {
	var timeout string
	if data.Timeout.Null {
		timeout = DefaultTimeout
	} else {
		timeout = data.Timeout.Value
	}

	var err error
	p.timeout, err = time.ParseDuration(timeout)
	if err != nil {
		resp.Diagnostics.AddError("Unable to parse the timeout", fmt.Sprintf("The timeout value '%s' can't be parsed: %s", timeout, err))
		return false
	}
	return true
}

func (p *provider) configureRateLimiter(data providerData, resp *tfsdk.ConfigureProviderResponse) bool {
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

	p.rateLimiter = rate.NewLimiter(rate.Every(rateLimitRateDuration), rateLimitBurst)

	return true
}

func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{}, nil
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
