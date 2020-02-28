apiVersion: v1
kind: ConfigMap
metadata:
  name: processor
data:
  processorImage: {{ echo -n $PROCESSOR_IMAGE_REPO }}-native:{{ echo -n $VERSION_SLUG }}
