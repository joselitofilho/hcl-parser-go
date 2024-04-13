// func Function
resource "google_cloudfunctions_function" "func_function" {
  name        = "func"
  description = "func function"
  runtime     = "go1.x"

  source_archive_bucket = google_storage_bucket.archive_funcs_bucket.name
  source_archive_object = google_storage_bucket_object.func_archive_funcs_object.name
  trigger_http          = "true"
  entry_point           = "funcFunction"

  environment_variables = {
    PSAPPENGINE_TO_TOPIC_NAME = google_pubsub_topic.psappengine_topic.name
    PSFUNC_FROM_TOPIC_NAME    = google_pubsub_topic.psfunc_topic.name

  }
}
