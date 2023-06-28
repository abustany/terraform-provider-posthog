package posthog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type ProjectID uint64

func (i ProjectID) String() string {
	return strconv.FormatUint(uint64(i), 10)
}

func ProjectIDFromString(s string) (ProjectID, error) {
	res, err := strconv.ParseUint(s, 10, 64)
	return ProjectID(res), err
}

type ProjectToolbarMode string

const (
	ProjectToolbarModeDisabled ProjectToolbarMode = "disabled"
	ProjectToolbarModeToolbar  ProjectToolbarMode = "toolbar"
)

type ProjectSessionRecordingVersion string

const (
	ProjectSessionRecordingVersionV1 = "v1"
	ProjectSessionRecordingVersionV2 = "v2"
)

type CreateProjectRequest struct {
	Name              string `json:"name"`
	AutocaptureOptOut bool   `json:"autocapture_opt_out"`
	Timezone          string `json:"timezone"`
	// TODO: filters
	// TODO: correlation analysis exclusions
	// TODO: path cleaning rules
	AppURLs                     []string                       `json:"app_urls"`
	DataAttributes              []string                       `json:"data_attributes"`
	PersonDisplayNameProperties []string                       `json:"person_display_name_properties"`
	SlackIncomingWebhook        string                         `json:"slack_incoming_webhook"`
	AnonymizeIPs                bool                           `json:"anonymize_ips"`
	ToolbarMode                 ProjectToolbarMode             `json:"toolbar_mode"`
	CapturePerformanceOptIn     bool                           `json:"capture_performance_opt_in"`
	CaptureConsoleLogOptIn      bool                           `json:"capture_console_log_opt_in"`
	SessionRecordingOptIn       bool                           `json:"session_recording_opt_in"`
	SessionRecordingVersion     ProjectSessionRecordingVersion `json:"session_recording_version"`
	RecordingDomains            []string                       `json:"recording_domains"`
	AccessControl               bool                           `json:"access_control"`
	CompletedSnippetOnboarding  bool                           `json:"completed_snippet_onboarding"`
}

type Project struct {
	ID                ProjectID `json:"id"`
	Name              string    `json:"name"`
	AutocaptureOptOut bool      `json:"autocapture_opt_out"`
	Timezone          string    `json:"timezone"`
	// TODO: filters
	// TODO: correlation analysis exclusions
	// TODO: path cleaning rules
	AppURLs                     []string                       `json:"app_urls"`
	DataAttributes              []string                       `json:"data_attributes"`
	PersonDisplayNameProperties []string                       `json:"person_display_name_properties"`
	SlackIncomingWebhook        string                         `json:"slack_incoming_webhook"`
	AnonymizeIPs                bool                           `json:"anonymize_ips"`
	ToolbarMode                 ProjectToolbarMode             `json:"toolbar_mode"`
	CapturePerformanceOptIn     bool                           `json:"capture_performance_opt_in"`
	CaptureConsoleLogOptIn      bool                           `json:"capture_console_log_opt_in"`
	SessionRecordingOptIn       bool                           `json:"session_recording_opt_in"`
	SessionRecordingVersion     ProjectSessionRecordingVersion `json:"session_recording_version"`
	RecordingDomains            []string                       `json:"recording_domains"`
	AccessControl               bool                           `json:"access_control"`
	APIToken                    string                         `json:"api_token"`
	CompletedSnippetOnboarding  bool                           `json:"completed_snippet_onboarding"`
	CreatedAt                   time.Time                      `json:"created_at"`
	UpdatedAt                   time.Time                      `json:"updated_at"`
}

func (p *Project) UnmarshalJSON(b []byte) error {
	type rawProject Project
	if err := json.Unmarshal(b, (*rawProject)(p)); err != nil {
		return err
	}

	if p.ToolbarMode == "" {
		// Somehow this is not returned in GET requests
		p.ToolbarMode = ProjectToolbarModeToolbar
	}

	return nil
}

func (c *Client) CreateProject(ctx context.Context, p CreateProjectRequest) (*Project, error) {
	nilSliceToEmpty(&p.AppURLs)
	nilSliceToEmpty(&p.DataAttributes)
	nilSliceToEmpty(&p.PersonDisplayNameProperties)
	nilSliceToEmpty(&p.RecordingDomains)

	var res *Project
	err := c.do(ctx, apiRequest{
		Method:       "POST",
		Path:         "/projects/",
		Input:        p,
		ExpectedCode: http.StatusCreated,
		Output:       &res,
	})
	return res, err
}

func (c *Client) UpdateProject(ctx context.Context, p Project) (*Project, error) {
	nilSliceToEmpty(&p.AppURLs)
	nilSliceToEmpty(&p.DataAttributes)
	nilSliceToEmpty(&p.PersonDisplayNameProperties)
	nilSliceToEmpty(&p.RecordingDomains)

	var res *Project
	err := c.do(ctx, apiRequest{
		Method:       "PATCH",
		Path:         "/projects/" + url.PathEscape(p.ID.String()),
		Input:        p,
		ExpectedCode: http.StatusOK,
		Output:       &res,
	})
	return res, err
}

func (c *Client) GetProject(ctx context.Context, projectID ProjectID) (*Project, error) {
	var res *Project
	err := c.do(ctx, apiRequest{
		Method:         "GET",
		Path:           "/projects/" + url.PathEscape(projectID.String()),
		ExpectedCode:   http.StatusOK,
		Output:         &res,
		OutputNilIf404: true,
	})
	return res, err
}

func (c *Client) DeleteProject(ctx context.Context, projectID ProjectID) error {
	err := c.do(ctx, apiRequest{
		Method:       "DELETE",
		Path:         "/projects/" + url.PathEscape(projectID.String()),
		ExpectedCode: http.StatusNoContent,
	})
	return err
}
