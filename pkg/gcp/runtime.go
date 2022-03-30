package gcp

import (
	"net/http"

	"cloud.google.com/go/compute/metadata"
	"github.com/finiteloopme/goutils/pkg/log"
)

// Check if the runtime platform is GCP
// Return true if GCP
func IsRuntimeGCP() bool {
	p := GetPRojectID()
	if p != "" {
		return true
	} else {
		return false
	}
}

// Get the GCP project ID
func GetProjectID() string {
	c := metadata.NewClient(&http.Client{Transport: userAgentTransport{
		userAgent: "kl-gcp-user-agent",
		base:      http.DefaultTransport,
	}})
	p, err := c.ProjectID()
	if err != nil {
		log.Fatal(err)
	}
	return p
}

// userAgentTransport sets the User-Agent header before calling base.
type userAgentTransport struct {
	userAgent string
	base      http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface.
func (t userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", t.userAgent)
	return t.base.RoundTrip(req)
}
