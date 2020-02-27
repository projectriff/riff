apiVersion: v1
kind: ConfigMap
metadata:
  name: processor
data:
  processorImage: gcr.io/projectriff/streaming-processor/processor-native:{{ echo -n $VERSION_SLUG }}
