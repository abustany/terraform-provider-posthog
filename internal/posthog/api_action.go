package posthog

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type CreateActionRequest struct {
	Name               string                    `json:"name"`
	Description        string                    `json:"description,omitempty"`
	Tags               []string                  `json:"tags"`
	PostToSlack        bool                      `json:"post_to_slack"`
	SlackMessageFormat string                    `json:"slack_message_format,omitempty"`
	Steps              []CreateActionStepRequest `json:"steps"`
}

type TextMatching string

const (
	TextMatchingExact    TextMatching = "exact"
	TextMatchingContains TextMatching = "contains"
	TextMatchingRegex    TextMatching = "regex"
)

type CreateActionStepRequest struct {
	Event string `json:"event"` // $autocapture or $pageview or custom event name

	// $autocapture and $pageview

	URL         string       `json:"url,omitempty"` // only match on given URL
	URLMatching TextMatching `json:"url_matching,omitempty"`

	// $autocapture

	Text         string       `json:"text,omitempty"` // only match elements with the given text
	TextMatching TextMatching `json:"text_matching,omitempty"`

	Href         string       `json:"href,omitempty"` // only match <a> elements with the given href
	HrefMatching TextMatching `json:"href_matching,omitempty"`

	Selector string `json:"selector,omitempty"` // only match elements that satisfy this CSS selector

	// TODO: properties
}

type ActionID uint64

func (i ActionID) String() string {
	return strconv.FormatUint(uint64(i), 10)
}

func ActionIDFromString(s string) (ActionID, error) {
	res, err := strconv.ParseUint(s, 10, 64)
	return ActionID(res), err
}

type Action struct {
	ID                 ActionID     `json:"id"`
	Name               string       `json:"name"`
	Description        string       `json:"description"`
	Tags               []string     `json:"tags"`
	PostToSlack        bool         `json:"post_to_slack"`
	SlackMessageFormat string       `json:"slack_message_format"`
	Steps              []ActionStep `json:"steps"`
	Deleted            bool         `json:"deleted"`
	IsCalculating      bool         `json:"is_calculating"`
	CreatedAt          time.Time    `json:"created_at"`
	LastCalculatedAt   time.Time    `json:"last_calculated_at"`
}

type ActionStep struct {
	ID    string `json:"id"`
	Event string `json:"event"` // $autocapture or $pageview or custom event name

	// $autocapture and $pageview

	URL         string       `json:"url,omitempty"` // only match on given URL
	URLMatching TextMatching `json:"url_matching,omitempty"`

	// $autocapture

	Text         string       `json:"text,omitempty"` // only match elements with the given text
	TextMatching TextMatching `json:"text_matching,omitempty"`

	Href         string       `json:"href,omitempty"` // only match <a> elements with the given href
	HrefMatching TextMatching `json:"href_matching,omitempty"`

	Selector string `json:"selector,omitempty"` // only match elements that satisfy this CSS selector

	// TODO: properties
}

func nilSliceToEmpty[T interface{}](v *[]T) {
	if *v == nil {
		*v = make([]T, 0)
	}
}

func (c *Client) CreateAction(ctx context.Context, projectID ProjectID, a CreateActionRequest) (*Action, error) {
	nilSliceToEmpty(&a.Tags)
	nilSliceToEmpty(&a.Steps)

	var res *Action
	err := c.do(ctx, apiRequest{
		Method:       "POST",
		Path:         "/projects/" + url.PathEscape(projectID.String()) + "/actions",
		Input:        a,
		ExpectedCode: http.StatusCreated,
		Output:       &res,
	})
	return res, err
}

func (c *Client) UpdateAction(ctx context.Context, projectID ProjectID, a Action) (*Action, error) {
	nilSliceToEmpty(&a.Tags)
	nilSliceToEmpty(&a.Steps)

	var res *Action
	err := c.do(ctx, apiRequest{
		Method:       "PATCH",
		Path:         "/projects/" + url.PathEscape(projectID.String()) + "/actions/" + url.PathEscape(a.ID.String()),
		Input:        a,
		ExpectedCode: http.StatusOK,
		Output:       &res,
	})
	return res, err
}

func (c *Client) GetAction(ctx context.Context, projectID ProjectID, actionID ActionID) (*Action, error) {
	var res *Action
	err := c.do(ctx, apiRequest{
		Method:         "GET",
		Path:           "/projects/" + url.PathEscape(projectID.String()) + "/actions/" + url.PathEscape(actionID.String()),
		ExpectedCode:   http.StatusOK,
		Output:         &res,
		OutputNilIf404: true,
	})
	return res, err
}
