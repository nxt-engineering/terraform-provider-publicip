package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"inet.af/netaddr"
)

const IPVersion4 = "v4"
const IPVersion6 = "v6"
const IPUnknown = "unknown"

type ipDataSourceType struct{}

func (t ipDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The current (public) IP as reported by the IP information provider.",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "An ID, which is only used internally. *Do not use this field in your terraform definitions.*",
				Computed:            true,
				Type:                types.StringType,
			},
			"ip_version": {
				MarkdownDescription: fmt.Sprintf("Whether the returned IP is an IPv6 or IPv4. Expected values: '%s', '%s', '%s'", IPVersion6, IPVersion4, IPUnknown),
				Computed:            true,
				Type:                types.StringType,
			},
			"is_ipv4": {
				MarkdownDescription: "`true` if the returned IP is an IPv6.",
				Computed:            true,
				Type:                types.BoolType,
			},
			"is_ipv6": {
				MarkdownDescription: "`true` if the returned IP is an IPv4.",
				Computed:            true,
				Type:                types.BoolType,
			},
			"ip": {
				MarkdownDescription: "The IP as returned by the IP information provider.",
				Computed:            true,
				Type:                types.StringType,
			},
			"asn_id": {
				MarkdownDescription: "The ASN as returned by the IP information provider.",
				Computed:            true,
				Type:                types.StringType,
			},
			"asn_org": {
				MarkdownDescription: "The organisation to which the ASN is registered to as returned by the IP information provider.",
				Computed:            true,
				Type:                types.StringType,
			},
			"source_ip": {
				MarkdownDescription: `Set the source IP address that is used to make the request to the IP information provider.
The address must be configured on a local network interface and that interface will be used.
Leave empty or ` + "`null`" + ` for default interface and IP stack.
` + "Set to `::` to get your public IPv6 address and `0.0.0.0` to get your IPv4 address.",
				Optional: true,
				Type:     types.StringType,
			},
		},
	}, nil
}

func (t ipDataSourceType) NewDataSource(_ context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return ipDataSource{
		provider: provider,
	}, diags
}

type ipDataSourceData struct {
	ID        types.String `tfsdk:"id"`
	IPVersion types.String `tfsdk:"ip_version"`
	IsIPv6    types.Bool   `tfsdk:"is_ipv6"`
	IsIPv4    types.Bool   `tfsdk:"is_ipv4"`
	IP        types.String `tfsdk:"ip"`
	ASNID     types.String `tfsdk:"asn_id"`
	ASNOrg    types.String `tfsdk:"asn_org"`
	SourceIP  types.String `tfsdk:"source_ip"`
}

type ipDataSource struct {
	provider provider
}

func (d ipDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data ipDataSourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	log.Printf("got to configuration ✅")

	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("got to client ✅")

	client := &http.Client{
		Timeout: d.provider.timeout,
	}

	if data.SourceIP.Null {
		data.SourceIP = types.String{Value: ""}
	}

	sourceIP := netaddr.IP{}
	if data.SourceIP.Value != "" {
		sourceIPStr := data.SourceIP.Value

		var err error
		sourceIP, err = netaddr.ParseIP(sourceIPStr)
		if err != nil || !sourceIP.IsValid() {
			log.Printf("Could not parse IP '%s' 🚨: %s", sourceIPStr, err)
			resp.Diagnostics.AddError("Invalid IP", fmt.Sprintf("The IP '%s' could not be parsed as valid IP: %s", sourceIPStr, err))
			return
		}
	}

	network := "tcp"
	if data.SourceIP.Value != "" {
		if sourceIP.Is6() {
			network = "tcp6"
		} else if sourceIP.Is4() {
			network = "tcp4"
		}
	}

	forceNetwork(client, network, sourceIP)

	baseURL := d.provider.ipURL
	requestURL := url.URL{
		Scheme:     baseURL.Scheme,
		Opaque:     baseURL.Opaque,
		User:       baseURL.User,
		Host:       baseURL.Host,
		Path:       path.Join(baseURL.Path, "json"),
		ForceQuery: baseURL.ForceQuery,
		RawQuery:   baseURL.RawQuery,
		Fragment:   baseURL.Fragment,
	}
	requestURLstr := requestURL.String()

	log.Printf("got to prepare request ✅: %s", requestURLstr)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", requestURLstr, nil)
	if err != nil {
		log.Printf("HTTP Client Creation Error 🚨: %s", err)
		resp.Diagnostics.AddError("Error preparing the HTTP request", fmt.Sprintf("There was an error when preparing the HTTP client with the url '%s': %s", requestURLstr, err))
		return
	}

	userAgent := fmt.Sprintf("%s (%s)", d.provider.toolName, d.provider.version)
	httpReq.Header.Set("User-Agent", userAgent)

	log.Printf("got to send request ✅: %s", userAgent)

	if !d.provider.rateLimiter.Allow() {
		log.Printf("the rate limit may be triggered ⏳")
	}

	timeoutCtx, cancelFunc := context.WithTimeout(ctx, d.provider.timeout)
	defer cancelFunc()
	err = d.provider.rateLimiter.Wait(timeoutCtx)
	if err != nil {
		log.Printf("Rate limiter error 🚨: %s", err)
		resp.Diagnostics.AddError("Error waiting for rate limit", fmt.Sprintf("There was an error while awaiting a slot from the rate limiter: %s", err))
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("HTTP client error 🚨: %s", err)
		resp.Diagnostics.AddError("Error fetching information from the IP information provider", fmt.Sprintf("There was an error when contacting '%s': %s", requestURLstr, err))
		return
	}
	defer httpResp.Body.Close()

	log.Printf("got to response ✅")

	if httpResp.StatusCode != http.StatusOK {
		log.Printf("HTTP Request Error 🚨: %d %s", httpResp.StatusCode, httpResp.Status)
		resp.Diagnostics.AddError("Error in response from the IP information provider", fmt.Sprintf("The IP information provider responded with the status code %d '%s'", httpResp.StatusCode, httpResp.Status))
		return
	}

	log.Printf("got to reading ✅")

	reader := httpResp.Body

	respData := new(IPResponse)
	err = json.NewDecoder(reader).Decode(respData)
	if err != nil {
		log.Printf("JSON decode error 🚨: %s", err)
		resp.Diagnostics.AddError("Error parsing the response from the IP information provider", fmt.Sprintf("There was an error when parsing the response from the IP information provider: %s", err))
		return
	}

	log.Printf("got to parse ip response ✅: %+v", respData)

	ip, err := netaddr.ParseIP(respData.IP)
	if err != nil {
		log.Printf("IP '%s' decode error 🚨: %s", respData.IP, err)
		resp.Diagnostics.AddError("Error parsing the IP from the IP information provider", fmt.Sprintf("There was an error when parsing the IP '%s' of the response from the IP information provider: %s", respData.IP, err))
		return
	}

	log.Printf("got to apply ✅: %+v", respData)

	data.ID = types.String{Value: fmt.Sprintf("%s$%s", data.SourceIP.Value, respData.IP)}
	data.IP = types.String{Value: ip.String()}
	data.IPVersion = types.String{Value: ipVersion(ip)}
	data.IsIPv6 = types.Bool{Value: ip.Is6()}
	data.IsIPv4 = types.Bool{Value: ip.Is4()}
	data.IP = types.String{Value: ip.String()}
	data.ASNID = types.String{Value: respData.ASN}
	data.ASNOrg = types.String{Value: respData.ASNOrg}

	log.Printf("got to state update ✅: %+v", data)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

	log.Printf("done ✅")
}

func ipVersion(netIP netaddr.IP) string {
	if netIP.Is6() {
		return IPVersion6
	}
	if netIP.Is4() {
		return IPVersion4
	}

	return IPUnknown
}
