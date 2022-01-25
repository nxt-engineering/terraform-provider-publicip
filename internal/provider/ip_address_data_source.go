package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ipDataSourceType struct{}

func (t ipDataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The current (public) IP as reported by the IP information provider.",

		Attributes: map[string]tfsdk.Attribute{
			"ip_version": {
				Type:                types.StringType,
				MarkdownDescription: "Whether to use IPv4 or IPv6 only. Valid values: 'V4', 'V6'",
				Optional:            true,
				Validators:          []tfsdk.AttributeValidator{ipVersionValidator{}},
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
		},
	}, nil
}

func (t ipDataSourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return ipDataSource{
		provider: provider,
	}, diags
}

type ipDataSourceData struct {
	IPVersion types.String `tfsdk:"ip_version"`
	IP        types.String `tfsdk:"ip"`
	ASNID     types.String `tfsdk:"asn_id"`
	ASNOrg    types.String `tfsdk:"asn_org"`
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

	if !data.IPVersion.Null {
		switch data.IPVersion.Value {
		case IPVersion6:
			forceV6(client)
		case IPVersion4:
			forceV4(client)
		}
	}

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

	httpResp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("HTTP Client Error 🚨: %s", err)
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

	log.Printf("got to apply ✅: %+v", respData)

	data.IP = types.String{Value: respData.IP}
	data.ASNID = types.String{Value: respData.ASN}
	data.ASNOrg = types.String{Value: respData.ASNOrg}
	data.IPVersion = types.String{Value: ipVersion(data.IPVersion, respData.IP)}

	log.Printf("got to state update ✅: %+v", data)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

	log.Printf("done ✅")
}

func ipVersion(version types.String, ip string) string {
	if !version.Null {
		return version.Value
	}

	if strings.Contains(ip, ":") {
		return IPVersion6
	}
	return IPVersion4
}
