a) build the docker image using dockerfile for that main.go file and push it to dockerhub, this app sends traces to the endpoint for every 5 seconds, nothing have to be done from outside.

b) deploy both values.yaml files of jaeger and otel for required trace collection depending on required receivers and exporters.

C) deploying this files as it is will sends traces to jaeger and awsxray backends

Note:
  i) To receive traces from opentelemetry through OTLP the otlp port (4317) of jaeger collector should be exposed, otherwise traces wont be sent to jaeger from otel.Make Change accordingly on values.yaml file before deploying jaeger.
 ii) Similarly atleast "xray read only policy" should be attached to eks nodegroup role to send the traces to awsxray, otherwise traces wont be sent to awsxray from otel.
