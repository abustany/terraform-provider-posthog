resource "posthog_project" "test" {
  name = "test"
}

# Action triggered by a custom event and posting to a webhook
resource "posthog_action" "custom_event_webhook" {
  name       = "Custom Event"
  project_id = posthog_project.test.id

  match_custom_events = [
    { event = "Custom Event name" }
  ]

  post_to_webhook        = true
  webhook_message_format = "This just happened: [event.link]"
}

# Action triggered by an autocapture event
resource "posthog_action" "name_autocapture" {
  name       = "User signed up"
  project_id = posthog_project.test.id

  match_autocaptures = [
    {
      url          = { value = "/signup", matching = "contains" }
      element_text = { value = "Sign up", matching = "exact" }
      selector     = ".button"
    }
  ]
}

# Action triggered by a page view
resource "posthog_action" "page_view" {
  name       = "User opened contact details"
  project_id = posthog_project.test.id

  match_page_views = [
    {
      url = { value = "/contacts/\\d+", matching = "regex" }
    }
  ]
}
