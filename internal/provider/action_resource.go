// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/abustany/terraform-provider-posthog/internal/posthog"
)

var _ resource.Resource = &actionResource{}
var _ resource.ResourceWithImportState = &actionResource{}

func newActionResource() resource.Resource {
	return &actionResource{}
}

type actionResource struct {
	client *posthog.Client
}

type actionResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ProjectID            types.String `tfsdk:"project_id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	Tags                 types.List   `tfsdk:"tags"`
	PostToWebhook        types.Bool   `tfsdk:"post_to_webhook"`
	WebhookMessageFormat types.String `tfsdk:"webhook_message_format"`
	MatchCustomEvents    types.List   `tfsdk:"match_custom_events"`
	MatchPageViews       types.List   `tfsdk:"match_page_views"`
	MatchAutocaptures    types.List   `tfsdk:"match_autocaptures"`
}

type matchCustomEvent struct {
	ID    string `tfsdk:"id"`
	Event string `tfsdk:"event"`
}

type matchPageViewEvent struct {
	ID  string         `tfsdk:"id"`
	URL matchableValue `tfsdk:"url"`
}

type matchAutocaptureEvent struct {
	ID          string         `tfsdk:"id"`
	URL         matchableValue `tfsdk:"url"`
	ElementText matchableValue `tfsdk:"element_text"`
	LinkHref    matchableValue `tfsdk:"link_href"`
	Selector    string         `tfsdk:"selector"`
}

type matchableValue struct {
	Value    string               `tfsdk:"value"`
	Matching posthog.TextMatching `tfsdk:"matching"`
}

func (r *actionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_action"
}

func matchableValueAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"value": schema.StringAttribute{
			MarkdownDescription: "Value to match",
			Required:            true,
		},
		"matching": schema.StringAttribute{
			MarkdownDescription: "Matching strategy, must be `exact`, `contains` or `regex`",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("contains"),
			Validators: []validator.String{
				stringvalidator.OneOf("exact", "contains", "regex"),
			},
		},
	}
}

func matchCustomEventsSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "List of custom events that trigger this action.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: "ID of the match group",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"event": schema.StringAttribute{
					MarkdownDescription: "Name of the custom event to match",
					Required:            true,
				},
			},
		},
		Optional: true,
	}
}

func matchPageViewsSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "List of page view events that trigger this action.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: "ID of the match group",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"url": schema.SingleNestedAttribute{
					MarkdownDescription: "URL of the page view event",
					Required:            true,
					Attributes:          matchableValueAttributes(),
				},
			},
		},
		Optional: true,
	}
}

func matchAutocapturesSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "List of autocapture events that trigger this action.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: "ID of the match group",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"url": schema.SingleNestedAttribute{
					MarkdownDescription: "URL where the event was captured",
					Optional:            true,
					Attributes:          matchableValueAttributes(),
				},
				"element_text": schema.SingleNestedAttribute{
					MarkdownDescription: "Text of the element that triggered the event",
					Optional:            true,
					Attributes:          matchableValueAttributes(),
				},
				"link_href": schema.SingleNestedAttribute{
					MarkdownDescription: "Href of the link that triggered the event",
					Optional:            true,
					Attributes:          matchableValueAttributes(),
				},
				"selector": schema.StringAttribute{
					MarkdownDescription: "CSS selector of the element that triggered the event",
					Optional:            true,
				},
			},
		},
		Optional: true,
	}
}

func (r *actionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Posthog Action",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the action",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "ID of the project of the action",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the action",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the action",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Action tags",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"post_to_webhook": schema.BoolAttribute{
				MarkdownDescription: "Whether to post to a webhook when this action is triggered",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"webhook_message_format": schema.StringAttribute{
				MarkdownDescription: "Format of the message sent to the webhook",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"match_custom_events": matchCustomEventsSchema(),
			"match_page_views":    matchPageViewsSchema(),
			"match_autocaptures":  matchAutocapturesSchema(),
		},
	}
}

func (r *actionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func sortedStrings(strs []string) []string {
	res := append([]string(nil), strs...)
	sort.Strings(res)
	return res
}

func updateActionModel(ctx context.Context, model *actionResourceModel, apiAction *posthog.Action) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(apiAction.ID.String())
	model.Name = types.StringValue(apiAction.Name)
	model.Description = types.StringValue(apiAction.Description)
	model.Tags, diags = types.ListValueFrom(ctx, types.StringType, sortedStrings(apiAction.Tags))
	if diags.HasError() {
		return diags
	}

	var (
		matchCustomEventSteps      []matchCustomEvent
		matchPageViewEventSteps    []matchPageViewEvent
		matchAutocaptureEventSteps []matchAutocaptureEvent
	)
	for _, ev := range apiAction.Steps {
		switch ev.Event {
		case "$autocapture":
			matchAutocaptureEventSteps = append(matchAutocaptureEventSteps, matchAutocaptureEvent{
				ID:          ev.ID,
				URL:         matchableValue{Value: ev.URL, Matching: ev.URLMatching},
				ElementText: matchableValue{Value: ev.Text, Matching: ev.TextMatching},
				LinkHref:    matchableValue{Value: ev.Href, Matching: ev.HrefMatching},
				Selector:    ev.Selector,
			})
		case "$pageview":
			matchPageViewEventSteps = append(matchPageViewEventSteps, matchPageViewEvent{
				ID:  ev.ID,
				URL: matchableValue{Value: ev.URL, Matching: ev.URLMatching},
			})
		default:
			matchCustomEventSteps = append(matchCustomEventSteps, matchCustomEvent{ID: ev.ID, Event: ev.Event})
		}
	}

	model.MatchCustomEvents, diags = types.ListValueFrom(ctx, matchCustomEventsSchema().NestedObject.Type(), matchCustomEventSteps)
	if diags.HasError() {
		return diags
	}

	model.MatchPageViews, diags = types.ListValueFrom(ctx, matchPageViewsSchema().NestedObject.Type(), matchPageViewEventSteps)
	if diags.HasError() {
		return diags
	}

	model.MatchAutocaptures, diags = types.ListValueFrom(ctx, matchAutocapturesSchema().NestedObject.Type(), matchAutocaptureEventSteps)
	if diags.HasError() {
		return diags
	}

	model.PostToWebhook = types.BoolValue(apiAction.PostToSlack)
	model.WebhookMessageFormat = types.StringValue(apiAction.SlackMessageFormat)

	return nil
}

func (r *actionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data actionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createActionRequest := posthog.CreateActionRequest{
		Name:               data.Name.ValueString(),
		Description:        data.Description.ValueString(),
		PostToSlack:        data.PostToWebhook.ValueBool(),
		SlackMessageFormat: data.WebhookMessageFormat.ValueString(),
	}

	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &createActionRequest.Tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Decode custom event steps

	var matchCustomEvents []matchCustomEvent

	resp.Diagnostics.Append(data.MatchCustomEvents.ElementsAs(ctx, &matchCustomEvents, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, ev := range matchCustomEvents {
		createActionRequest.Steps = append(createActionRequest.Steps, posthog.CreateActionStepRequest{Event: ev.Event})
	}

	// Decode page view steps

	var matchPageViewEvents []matchPageViewEvent

	resp.Diagnostics.Append(data.MatchPageViews.ElementsAs(ctx, &matchPageViewEvents, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, ev := range matchPageViewEvents {
		createActionRequest.Steps = append(createActionRequest.Steps, posthog.CreateActionStepRequest{
			Event:       "$pageview",
			URL:         ev.URL.Value,
			URLMatching: ev.URL.Matching,
		})
	}

	// Decode autocapture steps

	var matchAutocaptureEvents []matchAutocaptureEvent

	resp.Diagnostics.Append(data.MatchAutocaptures.ElementsAs(ctx, &matchAutocaptureEvents, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, ev := range matchAutocaptureEvents {
		createActionRequest.Steps = append(createActionRequest.Steps, posthog.CreateActionStepRequest{
			Event:        "$autocapture",
			URL:          ev.URL.Value,
			URLMatching:  ev.URL.Matching,
			Text:         ev.ElementText.Value,
			TextMatching: ev.ElementText.Matching,
			Href:         ev.LinkHref.Value,
			HrefMatching: ev.LinkHref.Matching,
			Selector:     ev.Selector,
		})
	}

	// Create the action

	projectID, err := posthog.ProjectIDFromString(data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid project ID", err.Error())
		return
	}

	res, err := r.client.CreateAction(ctx, projectID, createActionRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error creating action: %s", err))
		return
	}

	resp.Diagnostics.Append(updateActionModel(ctx, &data, res)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created action", map[string]interface{}{"action_id": res.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *actionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data actionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, err := posthog.ProjectIDFromString(data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid project ID", err.Error())
		return
	}

	actionID, err := posthog.ActionIDFromString(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid action ID", err.Error())
		return
	}

	res, err := r.client.GetAction(ctx, projectID, actionID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error getting action %s: %s", data.ID, err))
		return
	}

	if res == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(updateActionModel(ctx, &data, res)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read action", map[string]interface{}{"action_id": res.ID})

	if res.Deleted {
		resp.State.RemoveResource(ctx)
	} else {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func actionFromModel(ctx context.Context, data actionResourceModel) (posthog.ProjectID, posthog.Action, diag.Diagnostics) {
	var diags diag.Diagnostics

	projectID, err := posthog.ProjectIDFromString(data.ProjectID.ValueString())
	if err != nil {
		diags.AddError("Invalid project ID", err.Error())
		return posthog.ProjectID(0), posthog.Action{}, diags
	}

	actionID, err := posthog.ActionIDFromString(data.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid action ID", err.Error())
		return posthog.ProjectID(0), posthog.Action{}, diags
	}

	action := posthog.Action{
		ID:          actionID,
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}

	diags.Append(data.Tags.ElementsAs(ctx, &action.Tags, false)...)
	if diags.HasError() {
		return posthog.ProjectID(0), posthog.Action{}, diags
	}

	// Decode custom event steps

	var matchCustomEvents []matchCustomEvent

	diags.Append(data.MatchCustomEvents.ElementsAs(ctx, &matchCustomEvents, false)...)
	if diags.HasError() {
		return posthog.ProjectID(0), posthog.Action{}, diags
	}

	for _, ev := range matchCustomEvents {
		action.Steps = append(action.Steps, posthog.ActionStep{ID: ev.ID, Event: ev.Event})
	}

	// Decode page view steps

	var matchPageViewEvents []matchPageViewEvent

	diags.Append(data.MatchPageViews.ElementsAs(ctx, &matchPageViewEvents, false)...)
	if diags.HasError() {
		return posthog.ProjectID(0), posthog.Action{}, diags
	}

	for _, ev := range matchPageViewEvents {
		action.Steps = append(action.Steps, posthog.ActionStep{
			ID:          ev.ID,
			Event:       "$pageview",
			URL:         ev.URL.Value,
			URLMatching: ev.URL.Matching,
		})
	}

	// Decode autocapture steps

	var matchAutocaptureEvents []matchAutocaptureEvent

	diags.Append(data.MatchAutocaptures.ElementsAs(ctx, &matchAutocaptureEvents, false)...)
	if diags.HasError() {
		return posthog.ProjectID(0), posthog.Action{}, diags
	}

	for _, ev := range matchAutocaptureEvents {
		action.Steps = append(action.Steps, posthog.ActionStep{
			ID:           ev.ID,
			Event:        "$autocapture",
			URL:          ev.URL.Value,
			URLMatching:  ev.URL.Matching,
			Text:         ev.ElementText.Value,
			TextMatching: ev.ElementText.Matching,
			Href:         ev.LinkHref.Value,
			HrefMatching: ev.LinkHref.Matching,
			Selector:     ev.Selector,
		})
	}

	return projectID, action, diags
}

func (r *actionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data actionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, action, diags := actionFromModel(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.UpdateAction(ctx, projectID, action)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error updating action %s: %s", action.ID, err))
		return
	}

	resp.Diagnostics.Append(updateActionModel(ctx, &data, res)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated action", map[string]interface{}{"action_id": res.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *actionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data actionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, action, diags := actionFromModel(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	action.Deleted = true

	_, err := r.client.UpdateAction(ctx, projectID, action)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error deleting action %s: %s", action.ID, err))
		return
	}
}

func parseImportID(s string) (posthog.ProjectID, posthog.ActionID, error) {
	tokens := strings.SplitN(s, "/", 2)
	if len(tokens) != 2 {
		return posthog.ProjectID(0), posthog.ActionID(0), fmt.Errorf("ID not of the form PROJECT_ID/ACTION_ID")
	}

	projectID, err := posthog.ProjectIDFromString(tokens[0])
	if err != nil {
		return posthog.ProjectID(0), posthog.ActionID(0), fmt.Errorf("invalid project ID: %w", err)
	}

	actionID, err := posthog.ActionIDFromString(tokens[1])
	if err != nil {
		return posthog.ProjectID(0), posthog.ActionID(0), fmt.Errorf("invalid action ID: %w", err)
	}

	return projectID, actionID, nil
}

func (r *actionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID, actionID, err := parseImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), actionID.String())...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectID.String())...)
}
