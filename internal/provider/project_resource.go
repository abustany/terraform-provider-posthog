// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/abustany/terraform-provider-posthog/internal/posthog"
)

var _ resource.Resource = &projectResource{}
var _ resource.ResourceWithImportState = &projectResource{}

func newProjectResource() resource.Resource {
	return &projectResource{}
}

type projectResource struct {
	client *posthog.Client
}

type projectResourceModel struct {
	ID                             types.String `tfsdk:"id"`
	Name                           types.String `tfsdk:"name"`
	DisableAutocapture             types.Bool   `tfsdk:"disable_autocapture"`
	Timezone                       types.String `tfsdk:"timezone"`
	AuthorizedURLs                 types.List   `tfsdk:"authorized_urls"`
	DataAttributes                 types.List   `tfsdk:"data_attributes"`
	PersonDisplayNameProperties    types.List   `tfsdk:"person_display_name_properties"`
	WebhookURL                     types.String `tfsdk:"webhook_url"`
	AnonymizeIPs                   types.Bool   `tfsdk:"anonymize_ips"`
	EnableToolbar                  types.Bool   `tfsdk:"enable_toolbar"`
	RecordUserSessions             types.Bool   `tfsdk:"record_user_sessions"`
	CaptureConsoleLogs             types.Bool   `tfsdk:"capture_console_logs"`
	CaptureNetworkPerformance      types.Bool   `tfsdk:"capture_network_performance"`
	UseSessionRecorderV2           types.Bool   `tfsdk:"use_session_recorder_v2"`
	AuthorizedSessionRecordingURLs types.List   `tfsdk:"authorized_session_recording_urls"`
	EnableAccessControl            types.Bool   `tfsdk:"enable_access_control"`
	APIToken                       types.String `tfsdk:"api_token"`
}

func (r *projectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Posthog Project",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the project",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the action",
				Required:            true,
			},
			"disable_autocapture": schema.BoolAttribute{
				MarkdownDescription: "Whether to disable capturing frontend interactions like pageviews, clicks, and more when using the JavaScript or React Native libraries.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"timezone": schema.StringAttribute{
				MarkdownDescription: "Timezone for the project. All charts will be based on this timezone, including how PostHog buckets data in day/week/month intervals.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("UTC"),
			},
			"authorized_urls": schema.ListAttribute{
				MarkdownDescription: "URLs where the Toolbar will automatically launch when logged in.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"data_attributes": schema.ListAttribute{
				MarkdownDescription: "Attributes used when using the toolbar and defining actions to match unique elements on your pages.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"person_display_name_properties": schema.ListAttribute{
				MarkdownDescription: "Properties of an identified person used for their Display Name.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "URL where notifications are sent when selected actions are performed by users.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"anonymize_ips": schema.BoolAttribute{
				MarkdownDescription: "Whether to discard client IP data.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"enable_toolbar": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable the PostHog Toolbar which gives access to heatmaps, stats and allows to create actions directly in the website.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"record_user_sessions": schema.BoolAttribute{
				MarkdownDescription: "Whether to record user interactions.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"capture_console_logs": schema.BoolAttribute{
				MarkdownDescription: "Whether to include console logs in session recordings.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"capture_network_performance": schema.BoolAttribute{
				MarkdownDescription: "Whether to capture network information in session recordings.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"use_session_recorder_v2": schema.BoolAttribute{
				MarkdownDescription: "Whether to use rrweb 2 to record user sessions.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"authorized_session_recording_urls": schema.ListAttribute{
				MarkdownDescription: "Restricts where sessions are recorded. An empty list means no restriction.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"enable_access_control": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable granular access control for this project.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "API token used to send events to this project",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *projectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*posthog.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *posthog.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func updateProjectModel(ctx context.Context, model *projectResourceModel, apiProject *posthog.Project) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(apiProject.ID.String())
	model.Name = types.StringValue(apiProject.Name)
	model.DisableAutocapture = types.BoolValue(apiProject.AutocaptureOptOut)
	model.Timezone = types.StringValue(apiProject.Timezone)

	model.AuthorizedURLs, diags = types.ListValueFrom(ctx, types.StringType, sortedStrings(apiProject.AppURLs))
	if diags.HasError() {
		return diags
	}

	model.DataAttributes, diags = types.ListValueFrom(ctx, types.StringType, sortedStrings(apiProject.DataAttributes))
	if diags.HasError() {
		return diags
	}

	model.PersonDisplayNameProperties, diags = types.ListValueFrom(ctx, types.StringType, sortedStrings(apiProject.PersonDisplayNameProperties))
	if diags.HasError() {
		return diags
	}

	model.WebhookURL = types.StringValue(apiProject.SlackIncomingWebhook)
	model.AnonymizeIPs = types.BoolValue(apiProject.AnonymizeIPs)
	model.EnableToolbar = types.BoolValue(apiProject.ToolbarMode == posthog.ProjectToolbarModeToolbar)
	model.RecordUserSessions = types.BoolValue(apiProject.SessionRecordingOptIn)
	model.CaptureConsoleLogs = types.BoolValue(apiProject.CaptureConsoleLogOptIn)
	model.CaptureNetworkPerformance = types.BoolValue(apiProject.CapturePerformanceOptIn)
	model.UseSessionRecorderV2 = types.BoolValue(apiProject.SessionRecordingVersion == posthog.ProjectSessionRecordingVersionV2)

	model.AuthorizedSessionRecordingURLs, diags = types.ListValueFrom(ctx, types.StringType, sortedStrings(apiProject.RecordingDomains))
	if diags.HasError() {
		return diags
	}

	model.EnableAccessControl = types.BoolValue(apiProject.AccessControl)
	model.APIToken = types.StringValue(apiProject.APIToken)

	return diags
}

func projectFromModel(ctx context.Context, data projectResourceModel) (posthog.Project, diag.Diagnostics) {
	var diags diag.Diagnostics

	projectID, err := posthog.ProjectIDFromString(data.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid project ID", err.Error())
		return posthog.Project{}, diags
	}

	p := posthog.Project{
		ID:                      projectID,
		Name:                    data.Name.ValueString(),
		AutocaptureOptOut:       data.DisableAutocapture.ValueBool(),
		Timezone:                data.Timezone.ValueString(),
		SlackIncomingWebhook:    data.WebhookURL.ValueString(),
		AnonymizeIPs:            data.AnonymizeIPs.ValueBool(),
		CapturePerformanceOptIn: data.CaptureNetworkPerformance.ValueBool(),
		SessionRecordingOptIn:   data.RecordUserSessions.ValueBool(),
		AccessControl:           data.EnableAccessControl.ValueBool(),
		APIToken:                data.APIToken.ValueString(),
	}

	diags.Append(data.AuthorizedURLs.ElementsAs(ctx, &p.AppURLs, false)...)
	if diags.HasError() {
		return posthog.Project{}, diags
	}

	diags.Append(data.DataAttributes.ElementsAs(ctx, &p.DataAttributes, false)...)
	if diags.HasError() {
		return posthog.Project{}, diags
	}

	diags.Append(data.PersonDisplayNameProperties.ElementsAs(ctx, &p.PersonDisplayNameProperties, false)...)
	if diags.HasError() {
		return posthog.Project{}, diags
	}

	if data.EnableToolbar.ValueBool() {
		p.ToolbarMode = posthog.ProjectToolbarModeToolbar
	} else {
		p.ToolbarMode = posthog.ProjectToolbarModeDisabled
	}

	if data.UseSessionRecorderV2.ValueBool() {
		p.SessionRecordingVersion = posthog.ProjectSessionRecordingVersionV2
	} else {
		p.SessionRecordingVersion = posthog.ProjectSessionRecordingVersionV1
	}

	diags.Append(data.AuthorizedSessionRecordingURLs.ElementsAs(ctx, &p.RecordingDomains, false)...)
	if diags.HasError() {
		return posthog.Project{}, diags
	}

	return p, diags
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data projectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createProjectRequest := posthog.CreateProjectRequest{
		Name:                       data.Name.ValueString(),
		AutocaptureOptOut:          data.DisableAutocapture.ValueBool(),
		Timezone:                   data.Timezone.ValueString(),
		SlackIncomingWebhook:       data.WebhookURL.ValueString(),
		AnonymizeIPs:               data.AnonymizeIPs.ValueBool(),
		CapturePerformanceOptIn:    data.CaptureNetworkPerformance.ValueBool(),
		CaptureConsoleLogOptIn:     data.CaptureConsoleLogs.ValueBool(),
		SessionRecordingOptIn:      data.RecordUserSessions.ValueBool(),
		AccessControl:              data.EnableAccessControl.ValueBool(),
		CompletedSnippetOnboarding: true,
	}

	resp.Diagnostics.Append(data.AuthorizedURLs.ElementsAs(ctx, &createProjectRequest.AppURLs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.DataAttributes.ElementsAs(ctx, &createProjectRequest.DataAttributes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.PersonDisplayNameProperties.ElementsAs(ctx, &createProjectRequest.PersonDisplayNameProperties, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.EnableToolbar.ValueBool() {
		createProjectRequest.ToolbarMode = posthog.ProjectToolbarModeToolbar
	} else {
		createProjectRequest.ToolbarMode = posthog.ProjectToolbarModeDisabled
	}

	if data.UseSessionRecorderV2.ValueBool() {
		createProjectRequest.SessionRecordingVersion = posthog.ProjectSessionRecordingVersionV2
	} else {
		createProjectRequest.SessionRecordingVersion = posthog.ProjectSessionRecordingVersionV1
	}

	resp.Diagnostics.Append(data.AuthorizedSessionRecordingURLs.ElementsAs(ctx, &createProjectRequest.RecordingDomains, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the project

	res, err := r.client.CreateProject(ctx, createProjectRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error creating action: %s", err))
		return
	}

	resp.Diagnostics.Append(updateProjectModel(ctx, &data, res)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created project", map[string]interface{}{"project_id": res.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data projectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, err := posthog.ProjectIDFromString(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid project ID", err.Error())
		return
	}

	res, err := r.client.GetProject(ctx, projectID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error getting project %s: %s", data.ID, err))
		return
	}

	if res == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(updateProjectModel(ctx, &data, res)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read project", map[string]interface{}{"action_id": res.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data projectResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, diags := projectFromModel(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.UpdateProject(ctx, project)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error updating project %s: %s", project.ID, err))
		return
	}

	resp.Diagnostics.Append(updateProjectModel(ctx, &data, res)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated project", map[string]interface{}{"project_id": res.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data projectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, diags := projectFromModel(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteProject(ctx, project.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error deleting project %s: %s", project.ID, err))
		return
	}
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID, err := posthog.ProjectIDFromString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid project ID", err.Error())
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), projectID.String())...)
}
