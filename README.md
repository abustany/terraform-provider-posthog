# Terraform Provider for PostHog

⚠️ This code is in very early state! Lots of features and tests are missing. Contributions welcome.

The PostHog Terraform provider allows managing resources within PostHog.

## Usage example

```terraform
terraform {
  required_providers {
    posthog = {
      source = "hashicorp.com/abustany/posthog"
    }
  }
  required_version = ">= 0.0.1"
}

provider "posthog" {
  api_key = "API key created in https://eu.posthog.com/me/settings"
  host    = "https://eu.posthog.com" # or https://app.posthog.com in the US
}

resource "posthog_project" "test" {
  name = "test project"
}

resource "posthog_action" "test" {
  name        = "Action name"
  description = "Action created from the TF provider"
  project_id  = posthog_project.test.id

  # The set of filters below don't really make sense but are here to demonstrate
  # the various features of this provider.

  match_custom_events = [
    { event = "Custom Event name" }
  ]

  match_page_views = [
    {
      url = { value = "/mypage/mysubpage", matching = "contains" }
    }
  ]

  match_autocaptures = [
    {
      url          = { value = "/url/with/autocapture/enabled" }
      element_text = { value = "bar", matching = "regex" }
      link_href    = { value = "link", matching = "exact" }
      selector     = ".link_class"
    }
  ]

  post_to_webhook        = true
  webhook_message_format = "This just happened: [event.link]"
}
```

## Supported features

The set of resources that can/could be managed by this provider are listed in
[the PostHog API docs](https://posthog.com/docs/api).

| Resource type | Supported | Notes |
|---------------|-----------|-------|
| [Projects](docs/resources/project.md) | ✅ | Missing: event filters, correlation analysis exclusions, path cleaning rules |
| [Actions](docs/resources/action.md)   | ✅ | Missing: filters |
