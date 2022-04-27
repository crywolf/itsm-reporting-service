package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-openapi/runtime/middleware"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

func main() {
	l, _ := zap.NewProduction()
	defer func(l *zap.Logger) {
		_ = l.Sync()
	}(l)

	logger := l.Sugar()

	r := httprouter.New()

	// API documentation
	opts := middleware.RedocOpts{Path: "/", SpecURL: "/swagger.yaml", Title: "Reporting service API documentation"}
	docsHandler := middleware.Redoc(opts, nil)
	// handlers for API documentation
	r.Handler(http.MethodGet, "/", docsHandler)
	r.Handler(http.MethodGet, "/swagger.yaml", http.FileServer(http.Dir("./internal/http/rest/api")))

	// Port for the documentation server can be specified by Env var HTTP_DOCS_PORT.
	// If HTTP_DOCS_PORT is not defined, then --port flag can be used
	// Otherwise the default port is 3001
	var HTTPDocsPort string
	var ok bool

	if HTTPDocsPort, ok = os.LookupEnv("HTTP_DOCS_PORT"); !ok {
		port := flag.String("port", "3001", "http server port")
		flag.Parse()
		HTTPDocsPort = *port
	}

	addr := fmt.Sprintf("localhost:%s", HTTPDocsPort)
	logger.Infof("Starting API documentation server at %s", addr)
	logger.Fatal(http.ListenAndServe(addr, r))
}
