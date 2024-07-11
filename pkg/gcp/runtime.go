package gcp

import (
	"os"
	"runtime"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"github.com/finiteloopme/goutils/pkg/log"
)

// Check if the runtime platform is GCP
// Return true if GCP
func IsRuntimeGCP() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	_product_name, _ := os.ReadFile("/sys/class/dmi/id/product_name")
	product_name := strings.TrimSpace(string(_product_name))
	if strings.Contains(product_name, "Google") {
		return true
	} else {
		return false
	}
}

// Get the GCP project ID
// If the runtime is not GCP, then specify an optional
// environment variable to read the project ID from
// 'GCP_PROJECT' is the default environment variable used
func GetProjectID(env ...string) string {
	projectID := ""
	if IsRuntimeGCP() {
		// Runtime is GCP
		var err error
		projectID, err = metadata.ProjectID()
		if err != nil {
			log.Warn("Unexpected error reading ProjectID from metadata server. ", err)
		}
	} else {
		env_variable := "GCP_PROJECT"
		if len(env) > 0 {
			env_variable = env[0]
		}
		projectID = os.Getenv(env_variable)
	}

	return projectID
}
