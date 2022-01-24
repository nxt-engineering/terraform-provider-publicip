package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

const IPVersion4 = "v4"
const IPVersion6 = "v6"

type ipVersionValidator struct{}

func (i ipVersionValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Ensures the correct values for the 'ip_version' field, which is either '%s' or '%s'.", IPVersion6, IPVersion4)
}

func (i ipVersionValidator) MarkdownDescription(ctx context.Context) string {
	return i.Description(ctx)
}

func (i ipVersionValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var data ipDataSourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if data.IPVersion.Null || data.IPVersion.Unknown {
		return
	}

	value := data.IPVersion.Value
	if !(value == IPVersion4 || value == IPVersion6) {
		resp.Diagnostics.AddError("Unrecognized 'ip_version'", fmt.Sprintf("The field 'ip_version' has the value '%s'. The allowed values are either '%s' or '%s'.", value, IPVersion6, IPVersion4))
	}
}
