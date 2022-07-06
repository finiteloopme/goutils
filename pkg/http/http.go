// Simple utility to start HTTP Server
// The server will start serving files from the current working directory
// Call: StartHTTPServer()
package http

import (
	"net/http"

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
	http.Handle("/", http.FileServer(http.Dir("./")))
	listenAt := config.Host + ":" + config.Port
	log.Info("Server listening at: " + listenAt)
	http.ListenAndServe(listenAt, nil)
}
