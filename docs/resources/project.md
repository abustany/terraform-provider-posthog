---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "posthog_project Resource - terraform-provider-posthog"
subcategory: ""
description: |-
  Manages a Posthog Project
---

# posthog_project (Resource)

Manages a Posthog Project

## Example Usage

```terraform
resource "posthog_project" {
  name          = "test project"
  anonymize_ips = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the action

### Optional

- `anonymize_ips` (Boolean) Whether to discard client IP data.
- `authorized_session_recording_urls` (List of String) Restricts where sessions are recorded. An empty list means no restriction.
- `authorized_urls` (List of String) URLs where the Toolbar will automatically launch when logged in.
- `capture_console_logs` (Boolean) Whether to include console logs in session recordings.
- `capture_network_performance` (Boolean) Whether to capture network information in session recordings.
- `data_attributes` (List of String) Attributes used when using the toolbar and defining actions to match unique elements on your pages.
- `disable_autocapture` (Boolean) Whether to disable capturing frontend interactions like pageviews, clicks, and more when using the JavaScript or React Native libraries.
- `enable_access_control` (Boolean) Whether to enable granular access control for this project.
- `enable_toolbar` (Boolean) Whether to enable the PostHog Toolbar which gives access to heatmaps, stats and allows to create actions directly in the website.
- `person_display_name_properties` (List of String) Properties of an identified person used for their Display Name.
- `record_user_sessions` (Boolean) Whether to record user interactions.
- `timezone` (String) Timezone for the project. All charts will be based on this timezone, including how PostHog buckets data in day/week/month intervals.
- `use_session_recorder_v2` (Boolean) Whether to use rrweb 2 to record user sessions.
- `webhook_url` (String) URL where notifications are sent when selected actions are performed by users.

### Read-Only

- `api_token` (String) API token used to send events to this project
- `id` (String) ID of the project

## Import

Import is supported using the following syntax:

```shell
# Projects can be imported using their ID, found on the project settings page
# next to the API key.
terraform import posthog_project.test 1234
```
