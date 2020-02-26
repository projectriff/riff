apiVersion: v1
kind: ConfigMap
metadata:
  name: processor
data:
  processorImage: {{ gcloud container images describe gcr.io/projectriff/streaming-processor/processor-native:0.6.0-SNAPSHOT --format="value(image_summary.fully_qualified_digest)" }}
