package gcp

import (
	"io/ioutil"
	"runtime"
	"strings"

	"cloud.google.com/go/compute/metadata"
)

// Check if the runtime platform is GCP
// Return true if GCP
func IsRuntimeGCP() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	_product_name, _ := ioutil.ReadFile("/sys/class/dmi/id/product_name")
	product_name := strings.TrimSpace(string(_product_name))
	if strings.Contains(product_name, "Google") {
		return true
	} else {
		return false
	}
}

// Get the GCP project ID
func GetProjectID() string {
	// c := metadata.NewClient(&http.Client{
	// 	Transport: userAgentTransport{
	// 		userAgent: "kl-gcp-user-agent",
	// 		base:      http.DefaultTransport,
	// 	},
	// 	Timeout: 1000000000, // 1 sec timeout
	// })
	// p, err := c.ProjectID()

	if IsRuntimeGCP() {
		p, _ := metadata.ProjectID()
		return p
	} else {
		return ""
	}
}

// // userAgentTransport sets the User-Agent header before calling base.
// type userAgentTransport struct {
// 	userAgent string
// 	base      http.RoundTripper
// }

// // RoundTrip implements the http.RoundTripper interface.
// func (t userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
// 	req.Header.Set("User-Agent", t.userAgent)
// 	return t.base.RoundTrip(req)
// }
