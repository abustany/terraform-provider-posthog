// Package typeutil provides additional helpers to convert between Go and Terraform types.
package typeutil

import "github.com/hashicorp/terraform-plugin-framework/types"

func IsStringValueUnset(v types.String) bool {
	return v.IsNull() || v.IsUnknown() || v.ValueString() == ""
}

func NullableStringValue(s string) types.String {
	if s == "" {
		return types.StringNull()
	}

	return types.StringValue(s)
}
