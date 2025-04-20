package env

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/joho/godotenv"
)

const (
	// TagEnv specifies the environment variable name.
	TagEnv = "env"
	// TagSecret specifies the Google Secret Manager secret name (e.g., "projects/PROJECT_ID/secrets/SECRET_NAME/versions/latest").
	TagSecret = "secret"
	// TagFlag specifies the command-line flag name.
	TagFlag = "flag"
	// TagDefault specifies the default value if no other source provides one.
	TagDefault = "default"
	// TagRequired specifies if the field must have a value after processing all sources.
	TagRequired = "required"
	// TagIgnored specifies that the field should be ignored by the loader.
	TagIgnored = "ignored"
)

var (
	// smClient is a cached Secret Manager client.
	smClient *secretmanager.Client
	// errInvalidSpecification indicates that the spec argument was not a pointer to a struct.
	errInvalidSpecification = errors.New("specification must be a pointer to a struct")
)

// initSecretManagerClient initializes the Secret Manager client if needed.
// It assumes Application Default Credentials (ADC) are configured correctly.
func initSecretManagerClient(ctx context.Context) error {
	if smClient == nil {
		var err error
		smClient, err = secretmanager.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create secret manager client: %w", err)
		}
	}
	return nil
}

// CloseSecretManagerClient closes the cached Secret Manager client if it was initialized.
// It's good practice to call this when the application shuts down.
func CloseSecretManagerClient() error {
	if smClient != nil {
		err := smClient.Close()
		smClient = nil // Reset client after closing
		return err
	}
	return nil
}

// Load processes configuration from various sources into the provided struct specification.
//
// The spec argument must be a pointer to a struct. Fields in the struct can use
// tags (env, secret, flag, default, required, ignored) to control loading.
//
// Sources are processed in the following order (later sources override earlier ones):
// 1. Default values (`default` tag)
// 2. .env file (if present)
// 3. Environment variables (`env` tag)
// 4. Google Secret Manager (`secret` tag) - Requires ADC or explicit credentials.
// 5. Command-line flags (`flag` tag) - Flags must be defined and parsed *before* calling Load.
//
// A prefix can be provided to namespace environment variables (e.g., "APP_").
// Required fields (`required:"true"`) must have a value after processing all sources.
//
// Example struct field:
//
//	APIKey string `env:"API_KEY" secret:"projects/p/secrets/s/versions/1" required:"true"`
func Load(ctx context.Context, prefix string, spec interface{}) error {
	// --- Validation ---
	specValue := reflect.ValueOf(spec)
	if specValue.Kind() != reflect.Ptr || specValue.IsNil() {
		return errInvalidSpecification
	}
	specElem := specValue.Elem()
	if specElem.Kind() != reflect.Struct {
		return errInvalidSpecification
	}
	specType := specElem.Type()

	// --- Load Sources ---
	// 1. Load .env file (ignore if not found)
	_ = godotenv.Load() // Best effort

	// 2. Prepare for Secret Manager (initialize client later if needed)
	var smErr error
	needsSecretManager := false
	for i := 0; i < specElem.NumField(); i++ {
		if _, ok := specType.Field(i).Tag.Lookup(TagSecret); ok {
			needsSecretManager = true
			break
		}
	}
	if needsSecretManager {
		smErr = initSecretManagerClient(ctx)
		// We don't return immediately on error here, as other sources might still work.
		// The error will be checked when fetching a specific secret.
	}

	var processingErrors []string

	// --- Process Fields ---
	for i := 0; i < specElem.NumField(); i++ {
		field := specElem.Field(i)
		fieldType := specType.Field(i)

		// Skip unexported fields or ignored fields
		if !field.CanSet() || fieldType.Tag.Get(TagIgnored) == "true" {
			continue
		}

		var valueStr string
		var found bool

		// --- Apply Default Value ---
		defaultValue := fieldType.Tag.Get(TagDefault)
		if defaultValue != "" {
			valueStr = defaultValue
			found = true // Mark as found initially, can be overridden
		}

		// --- Load from Environment Variable ---
		envKey := fieldType.Tag.Get(TagEnv)
		if envKey != "" {
			envFullName := strings.ToUpper(prefix + envKey)
			if val, ok := os.LookupEnv(envFullName); ok {
				valueStr = val
				found = true
			}
		}

		// --- Load from Secret Manager ---
		secretName := fieldType.Tag.Get(TagSecret)
		if secretName != "" {
			if smErr != nil {
				// Record error from client initialization if we needed it
				processingErrors = append(processingErrors, fmt.Sprintf("field %q: failed to initialize secret manager client: %v", fieldType.Name, smErr))
			} else if smClient == nil {
				// Should not happen if needsSecretManager was true and init succeeded, but check anyway
				processingErrors = append(processingErrors, fmt.Sprintf("field %q: secret manager client not initialized", fieldType.Name))
			} else {
				secretValue, err := accessSecretVersion(ctx, secretName)
				if err != nil {
					// Don't fail immediately, maybe another source worked or it's not required
					processingErrors = append(processingErrors, fmt.Sprintf("field %q: failed to access secret %q: %v", fieldType.Name, secretName, err))
				} else {
					valueStr = secretValue
					found = true
				}
			}
		}

		// --- Load from Command-line Flag ---
		flagName := fieldType.Tag.Get(TagFlag)
		if flagName != "" {
			// flag.Visit checks if a flag was explicitly set on the command line
			var flagValue *string // Use pointer to check if visited
			flag.Visit(func(f *flag.Flag) {
				if f.Name == flagName {
					// The flag package stores the value in the variable it was bound to.
					// We need to retrieve it indirectly. This is a bit hacky.
					// A better approach might be to require the user to pass the parsed flagset.
					// For simplicity here, we assume the flag was registered and parsed correctly.
					fl := flag.Lookup(flagName)
					if fl != nil {
						// Get the value as a string representation
						gv := fl.Value.(flag.Getter)
						valStr := gv.String()
						flagValue = &valStr // Mark as found via flag
					}
				}
			})
			if flagValue != nil {
				valueStr = *flagValue
				found = true
			}
		}

		// --- Set Field Value ---
		if found {
			if err := setFieldValue(field, valueStr); err != nil {
				processingErrors = append(processingErrors, fmt.Sprintf("field %q: error setting value: %v", fieldType.Name, err))
				continue // Skip required check if setting failed
			}
		}

		// --- Check Required ---
		required := fieldType.Tag.Get(TagRequired)
		if required == "true" && (!found || field.IsZero()) { // Check IsZero in case default/loaded value was the zero value
			processingErrors = append(processingErrors, fmt.Sprintf("field %q is required but was not found or is zero", fieldType.Name))
		}
	} // End field loop

	if len(processingErrors) > 0 {
		return fmt.Errorf("config loading errors:\n - %s", strings.Join(processingErrors, "\n - "))
	}

	return nil
}

// accessSecretVersion fetches the payload of the given secret version.
func accessSecretVersion(ctx context.Context, name string) (string, error) {
	if smClient == nil {
		return "", errors.New("secret manager client not initialized")
	}

	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	// Call the API.
	result, err := smClient.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version %q: %w", name, err)
	}

	// Secret Manager payloads are limited to 64KiB.
	// Ensure the payload isn't excessively large. Consider adding checks if needed.

	return string(result.Payload.Data), nil
}

// setFieldValue converts the string value and sets it on the reflect.Value field.
// Supports basic types: string, int, int64, uint, uint64, bool, float64, time.Duration.
// Extend this function for more complex types (slices, maps, custom types).
func setFieldValue(field reflect.Value, value string) error {
	if !field.CanSet() {
		return errors.New("field cannot be set")
	}

	fieldType := field.Type()

	switch fieldType.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Handle time.Duration specifically if it's an int64
		if fieldType == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration format %q: %w", value, err)
			}
			field.SetInt(int64(duration))
		} else {
			intValue, err := strconv.ParseInt(value, 0, fieldType.Bits())
			if err != nil {
				return fmt.Errorf("invalid integer format %q: %w", value, err)
			}
			field.SetInt(intValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 0, fieldType.Bits())
		if err != nil {
			return fmt.Errorf("invalid unsigned integer format %q: %w", value, err)
		}
		field.SetUint(uintValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean format %q: %w", value, err)
		}
		field.SetBool(boolValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, fieldType.Bits())
		if err != nil {
			return fmt.Errorf("invalid float format %q: %w", value, err)
		}
		field.SetFloat(floatValue)
	case reflect.Ptr:
		// Handle pointer fields by creating a new value of the underlying type
		if field.IsNil() {
			field.Set(reflect.New(fieldType.Elem()))
		}
		// Recursively set the value of the element pointed to
		return setFieldValue(field.Elem(), value)
	// Add cases for other types like slices, maps, etc. as needed
	// case reflect.Slice:
	// case reflect.Map:
	default:
		return fmt.Errorf("unsupported field type %s", fieldType.Kind())
	}
	return nil
}

// --- Example Usage ---

/*
// Define your configuration struct
type Config struct {
	ServerHost    string        `env:"SERVER_HOST" flag:"host" default:"localhost"`
	ServerPort    int           `env:"SERVER_PORT" flag:"port" default:"8080" required:"true"`
	APIKey        string        `env:"API_KEY" secret:"projects/your-gcp-project/secrets/api-key/versions/latest" required:"true"`
	Timeout       time.Duration `env:"TIMEOUT" default:"5s"`
	DebugMode     bool          `env:"DEBUG_MODE" flag:"debug" default:"false"`
	OptionalVal   *string       `env:"OPTIONAL_VAL"` // Example of optional pointer
	IgnoredField  string        `ignored:"true"`
}

func main() {
	// --- IMPORTANT: Define Flags BEFORE Parsing ---
	// Define flags that correspond to the 'flag' tags in the Config struct.
	// The configloader itself does *not* define flags.
	host := flag.String("host", "localhost", "Server host") // Default here should match struct for consistency
	port := flag.Int("port", 8080, "Server port")
	debug := flag.Bool("debug", false, "Enable debug mode")

	// --- Parse Flags ---
	// This must happen before calling configloader.Load
	flag.Parse()


	// --- Load Configuration ---
	var cfg Config
	ctx := context.Background()

	// Load configuration using the utility
	err := Load(ctx, "MYAPP_", &cfg) // Using "MYAPP_" as prefix for env vars
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Close the Secret Manager client when done (optional but good practice)
	if err := CloseSecretManagerClient(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing Secret Manager client: %v\n", err)
	}


	// --- Use Configuration ---
	fmt.Printf("Configuration loaded successfully:\n")
	fmt.Printf("  Host: %s\n", cfg.ServerHost)
	fmt.Printf("  Port: %d\n", cfg.ServerPort)
	fmt.Printf("  API Key: %s\n", "***REDACTED***") // Don't print secrets!
	fmt.Printf("  Timeout: %v\n", cfg.Timeout)
	fmt.Printf("  Debug Mode: %t\n", cfg.DebugMode)
	if cfg.OptionalVal != nil {
		fmt.Printf("  Optional Value: %s\n", *cfg.OptionalVal)
	} else {
		fmt.Printf("  Optional Value: <not set>\n")
	}

	// Example: Use the flags directly if needed (though Load already incorporated them)
	fmt.Printf("\nDirect flag values (for comparison):\n")
	fmt.Printf("  Flag Host: %s\n", *host)
	fmt.Printf("  Flag Port: %d\n", *port)
	fmt.Printf("  Flag Debug: %t\n", *debug)
}

*/
