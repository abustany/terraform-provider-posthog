// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/abustany/terraform-provider-posthog/internal/posthog"
	"github.com/abustany/terraform-provider-posthog/internal/typeutil"
)

var _ provider.Provider = &postHogProvider{}

type postHogProvider struct {
	version string
}

type postHogProviderModel struct {
	Host   types.String `tfsdk:"host"`
	APIKey types.String `tfsdk:"api_key"`
}

func (p *postHogProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "posthog"
	resp.Version = p.version
}

func (p *postHogProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "API host: https://app.posthog.com for US customers, https://eu.posthog.com for EU customers, or the address of the server for self hosted instances.",
				Required:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Personal API key used to authenticate against PostHog.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *postHogProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data postHogProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if typeutil.IsStringValueUnset(data.Host) {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing PostHog host",
			"The host parameter is required for this provider to manage resources.",
		)
	}

	if typeutil.IsStringValueUnset(data.APIKey) {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing PostHog API key",
			"The api_key parameter is required for this provider to manage resources.",
		)
	}

	client := &posthog.Client{
		HTTPClient: http.DefaultClient,
		Host:       data.Host.ValueString(),
		APIKey:     data.APIKey.ValueString(),
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *postHogProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newActionResource,
		newProjectResource,
	}
}

func (p *postHogProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		//NewExampleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &postHogProvider{
			version: version,
		}
	}
}
