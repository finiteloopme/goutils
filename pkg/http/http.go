// Simple utility to start HTTP Server
// The server will start serving files from the current working directory
// If index.html doesn't exist, it will create a simple handler at /
// Call: StartHTTPServer()
package http

import (
	"fmt"
	"net/http"
	"os"

	log "github.com/finiteloopme/goutils/pkg/log"
	"github.com/kelseyhightower/envconfig"
)

// Config for HTTP Server
type HTTPConfig struct {
	// Set env variable GCP_HOST. Default value is 0.0.0.0
	Host string `default:"0.0.0.0"`
	// Set env variable GCP_PORT. Default value is 8080
	Port string `default:"8080"`
}

// Start the HTTP Server
func StartHTTPServer() {
	var config HTTPConfig
	envconfig.Process("gcp", &config)
	// Check if ./index.html exists
	if _, err := os.Stat("./index.html"); os.IsNotExist(err) {
		// ./index.html doesn't exist
		// Creating a simple handler
		http.Handle("/", DefaultHandler{})
	} else {
		// ./index.html exists. So serve the current directory
		http.Handle("/", http.FileServer(http.Dir("./")))
	}
	listenAt := config.Host + ":" + config.Port
	log.Info("Server listening at: " + listenAt)
	http.ListenAndServe(listenAt, nil)
}

type DefaultHandler struct{}

func (DefaultHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "Hello from the default handler")
}

// Data structure to keep a map of URL to the function handler
type URLMap map[string]http.Handler

// Register Handlers for URL
func StartServer(opts ...URLMap) {
	urlMap := URLMap{"/": DefaultHandler{}}
	if len(opts) > 0 {
		urlMap = opts[0]
	}

	for url, funcHandler := range urlMap {
		http.Handle(url, funcHandler)
	}
}
