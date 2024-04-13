// psengine Pub Sub
resource "google_pubsub_topic" "psengine_topic" {
  name = "psengine"
  labels = {
    engine-subscriber = google_app_engine_application.engine_app.name
  }
}

resource "google_pubsub_subscription" "psengine_subscription" {
  name  = "psengine-subscription"
  topic = google_pubsub_topic.psengine_topic.name

  push_config {
    push_endpoint = google_cloudfunctions_function.func_function.https_trigger_url
  }

}
