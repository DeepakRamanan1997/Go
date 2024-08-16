package main

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/openzipkin/zipkin-go"
	zipkinhttpreporter "github.com/openzipkin/zipkin-go/reporter/http"
	zipkinhttp "github.com/openzipkin/zipkin-go/middleware/http"
)

func main() {
	tracer, err := newTracer()
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/foo", FooHandler)
	r.Use(zipkinhttp.NewServerMiddleware(tracer, zipkinhttp.SpanName("request")))
	log.Fatal(http.ListenAndServe(":8080", r))
}

func FooHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func newTracer() (*zipkin.Tracer, error) {
	localEndpoint, err := zipkin.NewEndpoint("my_service", "localhost:8080")
	if err != nil {
		return nil, err
	}
	reporter := zipkinhttpreporter.NewReporter("http://localhost:9411/api/v2/spans")
	sampler, err := zipkin.NewCountingSampler(1)
	if err != nil {
		return nil, err
	}
	tracer, err := zipkin.NewTracer(reporter, zipkin.WithSampler(sampler), zipkin.WithLocalEndpoint(localEndpoint))
	if err != nil {
		return nil, err
	}
	return tracer, nil
}
