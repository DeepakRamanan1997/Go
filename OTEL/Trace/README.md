a) build the docker image using dockerfile for that main.go file and push it to dockerhub, this app sends traces to the endpoint for every 5 seconds, nothing have to be done from outside.
   use this commands i) docker buildx create --use  
                    ii) docker buildx inspect --bootstrap
                   iii) docker buildx build --platform linux/amd64,linux/arm64 -t deepakramanan/goapp:multiarchtraceEKS --push .
   to build the dockerimage which supports both amd and arm architectures .

   To run docker buildx command install docker buildx following the steps given in this link https://github.com/docker/buildx/blob/master/README.md
   
b) deploy both values.yaml files of jaeger and otel for required trace collection depending on required receivers and exporters.

C) deploying this files as it is will sends traces to jaeger and awsxray backends

Note:
  i) To receive traces from opentelemetry through OTLP the otlp port (4317) of jaeger collector should be exposed, otherwise traces wont be sent to jaeger from otel.Make Change accordingly on values.yaml file before deploying jaeger.
 ii) Similarly atleast "xray read only policy" should be attached to eks nodegroup role to send the traces to awsxray, otherwise traces wont be sent to awsxray from otel.
