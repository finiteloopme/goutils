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
	// flagsDefined ensures we only define flags once per application run.
	flagsDefined = false
	// flagValues holds the pointers to the variables bound to the flags.
	// We need this because flag.Xxx returns the pointer, and we define them before parsing.
	flagValues = make(map[string]interface{})
	// flagWasSet tracks which flags were explicitly set on the command line.
	flagWasSet = make(map[string]bool)
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

// defineFlags defines command-line flags based on struct tags.
// It uses the global flag set (flag.CommandLine).
func defineFlags(spec interface{}) error {
	if flagsDefined {
		return nil // Avoid defining flags multiple times
	}

	specValue := reflect.ValueOf(spec)
	specElem := specValue.Elem()
	specType := specElem.Type()

	var flagDefinitionErrors []string

	for i := 0; i < specElem.NumField(); i++ {
		field := specElem.Field(i)
		fieldType := specType.Field(i)

		// Skip unexported fields or ignored fields
		if !field.CanSet() || fieldType.Tag.Get(TagIgnored) == "true" {
			continue
		}

		flagName := fieldType.Tag.Get(TagFlag)
		if flagName == "" {
			continue // No flag defined for this field
		}

		if flg := flag.Lookup(flagName); flg != nil {
			// Flag already defined on command line
			flagWasSet[flagName] = true      // Mark as set
			flagValues[flagName] = flg.Value // Set the value
			continue                         // Skip this field
		}

		// Prevent duplicate flag definition
		if _, exists := flagValues[flagName]; exists {
			// This could happen if Load is called multiple times with overlapping flag names
			// or if the user manually defined a flag with the same name.
			// We'll allow it but maybe log a warning in a real app.
			// return fmt.Errorf("flag %q is already defined", flagName)
			continue
		}

		defaultValueStr := fieldType.Tag.Get(TagDefault)
		usage := fmt.Sprintf("Set value for %s", fieldType.Name) // Basic usage message
		envVar := fieldType.Tag.Get(TagEnv)
		if envVar != "" {
			usage = fmt.Sprintf("%s (env: %s)", usage, envVar)
		}

		// Define the flag based on the field type
		kind := field.Kind()
		// If field is a pointer, get the underlying type
		if kind == reflect.Ptr {
			kind = field.Type().Elem().Kind()
		}

		switch kind {
		case reflect.String:
			flagValues[flagName] = flag.String(flagName, defaultValueStr, usage)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// Handle time.Duration specifically
			if field.Type() == reflect.TypeOf(time.Duration(0)) || (field.Kind() == reflect.Ptr && field.Type().Elem() == reflect.TypeOf(time.Duration(0))) {
				defaultDuration := time.Duration(0)
				if defaultValueStr != "" {
					var err error
					defaultDuration, err = time.ParseDuration(defaultValueStr)
					if err != nil {
						flagDefinitionErrors = append(flagDefinitionErrors, fmt.Sprintf("field %q (flag %q): invalid default duration format %q: %v", fieldType.Name, flagName, defaultValueStr, err))
						continue // Skip defining this flag
					}
				}
				flagValues[flagName] = flag.Duration(flagName, defaultDuration, usage)
			} else { // Handle regular integers
				defaultInt := int64(0)
				if defaultValueStr != "" {
					var err error
					defaultInt, err = strconv.ParseInt(defaultValueStr, 0, 64) // Parse as int64 for flag
					if err != nil {
						flagDefinitionErrors = append(flagDefinitionErrors, fmt.Sprintf("field %q (flag %q): invalid default integer format %q: %v", fieldType.Name, flagName, defaultValueStr, err))
						continue // Skip defining this flag
					}
				}
				// Use flag.Int for simplicity if the target is int, otherwise keep int64
				if kind == reflect.Int {
					flagValues[flagName] = flag.Int(flagName, int(defaultInt), usage)
				} else {
					// For int8, int16, int32, int64, the flag package uses int64 internally
					flagValues[flagName] = flag.Int64(flagName, defaultInt, usage)
				}

			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			defaultUint := uint64(0)
			if defaultValueStr != "" {
				var err error
				defaultUint, err = strconv.ParseUint(defaultValueStr, 0, 64) // Parse as uint64 for flag
				if err != nil {
					flagDefinitionErrors = append(flagDefinitionErrors, fmt.Sprintf("field %q (flag %q): invalid default unsigned integer format %q: %v", fieldType.Name, flagName, defaultValueStr, err))
					continue // Skip defining this flag
				}
			}
			if kind == reflect.Uint {
				flagValues[flagName] = flag.Uint(flagName, uint(defaultUint), usage)
			} else {
				flagValues[flagName] = flag.Uint64(flagName, defaultUint, usage)
			}

		case reflect.Bool:
			defaultBool := false
			if defaultValueStr != "" {
				var err error
				defaultBool, err = strconv.ParseBool(defaultValueStr)
				if err != nil {
					flagDefinitionErrors = append(flagDefinitionErrors, fmt.Sprintf("field %q (flag %q): invalid default boolean format %q: %v", fieldType.Name, flagName, defaultValueStr, err))
					continue // Skip defining this flag
				}
			}
			flagValues[flagName] = flag.Bool(flagName, defaultBool, usage)
		case reflect.Float32, reflect.Float64:
			defaultFloat := float64(0)
			if defaultValueStr != "" {
				var err error
				defaultFloat, err = strconv.ParseFloat(defaultValueStr, 64) // Parse as float64 for flag
				if err != nil {
					flagDefinitionErrors = append(flagDefinitionErrors, fmt.Sprintf("field %q (flag %q): invalid default float format %q: %v", fieldType.Name, flagName, defaultValueStr, err))
					continue // Skip defining this flag
				}
			}
			flagValues[flagName] = flag.Float64(flagName, defaultFloat, usage)
		// Add cases for other flag types if needed (e.g., slices using flag.Func)
		default:
			flagDefinitionErrors = append(flagDefinitionErrors, fmt.Sprintf("field %q (flag %q): unsupported type for flag: %s", fieldType.Name, flagName, kind))
		}
	} // End field loop

	if len(flagDefinitionErrors) > 0 {
		return fmt.Errorf("flag definition errors:\n - %s", strings.Join(flagDefinitionErrors, "\n - "))
	}

	flagsDefined = true
	return nil
}

// ProcessConfig processes configuration from various sources into the provided struct specification.
//
// The spec argument must be a pointer to a struct. Fields in the struct can use
// tags (env, secret, flag, default, required, ignored) to control loading.
//
// Load will **define and parse** command-line flags based on `flag` tags.
// Call this function *instead* of manually defining/parsing flags related to the config struct.
//
// Sources are processed in the following order (later sources override earlier ones):
// 1. Default values (`default` tag) - Also used as defaults for flags.
// 2. .env file (if present)
// 3. Environment variables (`env` tag)
// 4. Google Secret Manager (`secret` tag) - Requires ADC or explicit credentials.
// 5. Command-line flags (`flag` tag)
//
// A prefix can be provided to namespace environment variables (e.g., "APP_").
// Required fields (`required:"true"`) must have a value after processing all sources.
//
// Example struct field:
//
//		APIKey string `env:"API_KEY" secret:"projects/p/secrets/s/versions/1" required:"true"`
//	 Host   string `flag:"host" default:"localhost"`
func ProcessConfig(ctx context.Context, prefix string, spec interface{}) error {
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

	// --- Define Flags (if not already done) ---
	if err := defineFlags(spec); err != nil {
		return fmt.Errorf("error defining flags: %w", err)
	}

	// --- Parse Flags (if not already parsed) ---
	// This ensures flags are parsed only once, even if Load is called multiple times.
	// It uses the default CommandLine flag set.
	if !flag.Parsed() {
		flag.Parse()
		// Record which flags were actually set by the user
		flag.Visit(func(f *flag.Flag) {
			flagWasSet[f.Name] = true
		})
	}

	// --- Load Other Sources ---
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
		var source string // Keep track of where the value came from (for debugging/info)

		// --- 1. Apply Default Value ---
		defaultValue := fieldType.Tag.Get(TagDefault)
		if defaultValue != "" {
			valueStr = defaultValue
			found = true
			source = "default"
		}

		// --- 2. Load from Environment Variable (from .env or actual env) ---
		envKey := fieldType.Tag.Get(TagEnv)
		if envKey != "" {
			envFullName := strings.ToUpper(prefix + envKey)
			if val, ok := os.LookupEnv(envFullName); ok {
				valueStr = val
				found = true
				source = "environment"
			}
		}

		// --- 3. Load from Secret Manager ---
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
					source = "secret"
				}
			}
		}

		// --- 4. Load from Command-line Flag ---
		flagName := fieldType.Tag.Get(TagFlag)
		// Check if the flag was *defined* for this field AND *set* on the command line
		if flagName != "" {
			if pointer, defined := flagValues[flagName]; defined && flagWasSet[flagName] {
				// Get the value from the pointer stored during flag definition
				// Need to use reflection to get the underlying value from the interface{} pointer
				ptrValue := reflect.ValueOf(pointer) // e.g., ValueOf(**string)
				if ptrValue.IsValid() && ptrValue.Kind() == reflect.Ptr && !ptrValue.IsNil() {
					actualValue := ptrValue.Elem().Interface() // e.g., value of type string, int, bool etc.
					// Convert the actual value back to string for setFieldValue
					valueStr = fmt.Sprintf("%v", actualValue)
					found = true
					source = "flag"
				} else {
					// This shouldn't happen if defineFlags worked correctly
					processingErrors = append(processingErrors, fmt.Sprintf("field %q: internal error retrieving value for flag %q", fieldType.Name, flagName))
				}
			} else if !defined && flagWasSet[flagName] {
				// Flag was set but somehow not defined by our logic (e.g. user defined it manually)
				// We could potentially try flag.Lookup here as a fallback, but it might indicate an issue.
				// For now, we rely on flags being defined via `defineFlags`.
				processingErrors = append(processingErrors, fmt.Sprintf("field %q: flag %q was set but not defined by config loader", fieldType.Name, flagName))
			}
		}

		// --- Set Field Value ---
		if found {
			// fmt.Printf("Debug: Setting field %s from %s with value: %q\n", fieldType.Name, source, valueStr) // Optional debug line
			if err := setFieldValue(field, valueStr); err != nil {
				processingErrors = append(processingErrors, fmt.Sprintf("field %q (source: %s): error setting value '%s': %v", fieldType.Name, source, valueStr, err))
				continue // Skip required check if setting failed
			}
		}

		// --- Check Required ---
		required := fieldType.Tag.Get(TagRequired)
		// Check IsZero AFTER attempting to set. Handles cases where the loaded value IS the zero value (e.g., port 0, empty string).
		// If 'found' is false, it definitely wasn't set. If 'found' is true, check if the resulting field value is zero.
		if required == "true" && (!found || field.IsZero()) {
			// Construct a more informative error message
			errMsg := fmt.Sprintf("field %q is required but was not provided", fieldType.Name)
			if found && field.IsZero() { // It was found, but the value resulted in zero
				errMsg = fmt.Sprintf("field %q is required but received zero value (source: %s, raw value: '%s')", fieldType.Name, source, valueStr)
			}
			envDetail := ""
			if envKey != "" {
				envDetail = fmt.Sprintf(" (env: %s%s)", prefix, envKey)
			}
			flagDetail := ""
			if flagName != "" {
				flagDetail = fmt.Sprintf(" (flag: --%s)", flagName)
			}
			secretDetail := ""
			if secretName != "" {
				secretDetail = fmt.Sprintf(" (secret: %s)", secretName)
			}
			processingErrors = append(processingErrors, fmt.Sprintf("%s%s%s%s", errMsg, envDetail, flagDetail, secretDetail))
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
// Supports basic types: string, int, int64, uint, uint64, bool, float64, time.Duration, and pointers to these.
func setFieldValue(field reflect.Value, value string) error {
	if !field.CanSet() {
		return errors.New("field cannot be set")
	}

	fieldType := field.Type()

	// If the field is a pointer, allocate memory if nil and set the pointed-to value
	if fieldType.Kind() == reflect.Ptr {
		// If the pointer is nil, create a new instance of the element type
		if field.IsNil() {
			field.Set(reflect.New(fieldType.Elem()))
		}
		// Dereference the pointer and call setFieldValue recursively on the element
		return setFieldValue(field.Elem(), value)
	}

	// Handle non-pointer types
	switch fieldType.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Handle time.Duration specifically
		if fieldType == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				// Allow plain integers for duration (treat as nanoseconds)
				intVal, intErr := strconv.ParseInt(value, 10, 64)
				if intErr != nil {
					// Neither valid duration string nor plain int
					return fmt.Errorf("invalid duration format %q: %w (also not parsable as nanoseconds)", value, err)
				}
				duration = time.Duration(intVal)
			}
			field.SetInt(int64(duration))
		} else { // Handle regular integers
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
			// Allow 1/0 for bools as well? Often used in env vars.
			if value == "1" {
				boolValue = true
			} else if value == "0" {
				boolValue = false
			} else {
				return fmt.Errorf("invalid boolean format %q: %w", value, err)
			}
		}
		field.SetBool(boolValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, fieldType.Bits())
		if err != nil {
			return fmt.Errorf("invalid float format %q: %w", value, err)
		}
		field.SetFloat(floatValue)
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
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"your_module_path/env" // Replace with your actual module path
)


// Define your configuration struct
type Config struct {
	ServerHost   string        `env:"SERVER_HOST" flag:"host" default:"localhost"`
	ServerPort   int           `env:"SERVER_PORT" flag:"port" default:"8080" required:"true"`
	APIKey       string        `env:"API_KEY" secret:"projects/your-gcp-project/secrets/api-key/versions/latest" required:"true"`
	Timeout      time.Duration `env:"TIMEOUT" flag:"timeout" default:"5s"`
	Retries      int           `env:"RETRIES" flag:"retries" default:"3"`
	DebugMode    bool          `env:"DEBUG_MODE" flag:"debug" default:"false"`
	OptionalVal  *string       `env:"OPTIONAL_VAL" flag:"optional"` // Example of optional pointer
	IgnoredField string        `ignored:"true"`
}

func main() {
	// --- Load Configuration ---
	var cfg Config
	ctx := context.Background()

	// Load configuration using the utility.
	// It now defines and parses flags internally.
	// Do NOT define/parse flags manually for these fields anymore.
	err := env.Load(ctx, "MYAPP_", &cfg) // Using "MYAPP_" as prefix for env vars
	if err != nil {
		// flag package prints usage on error automatically if parsing fails
		// But Load can fail for other reasons (required fields, secrets)
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		// Optionally print usage again if needed, although flag.Parse() might have done it
		// flag.Usage()
		os.Exit(1)
	}

	// Close the Secret Manager client when done (optional but good practice)
	if err := env.CloseSecretManagerClient(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing Secret Manager client: %v\n", err)
	}


	// --- Use Configuration ---
	fmt.Printf("Configuration loaded successfully:\n")
	fmt.Printf("  Host: %s\n", cfg.ServerHost)
	fmt.Printf("  Port: %d\n", cfg.ServerPort)
	fmt.Printf("  API Key: %s\n", "***REDACTED***") // Don't print secrets!
	fmt.Printf("  Timeout: %v\n", cfg.Timeout)
	fmt.Printf("  Retries: %d\n", cfg.Retries)
	fmt.Printf("  Debug Mode: %t\n", cfg.DebugMode)
	if cfg.OptionalVal != nil {
		fmt.Printf("  Optional Value: %s\n", *cfg.OptionalVal)
	} else {
		fmt.Printf("  Optional Value: <not set>\n")
	}

	// You can still define OTHER flags manually if needed, just don't
	// redefine the ones handled by the Config struct.
	// For example:
	// var showVersion = flag.Bool("version", false, "Show application version")
	// Need to re-parse if you define flags *after* calling Load, which is not recommended.
	// It's better to define all flags before Load, or ensure Load handles all config flags.

	// Example: Accessing a flag value directly *after* Load has parsed them
	// Note: You need to lookup the flag value, you don't have the original pointer anymore.
	// timeoutFlag := flag.Lookup("timeout")
	// if timeoutFlag != nil {
	//     fmt.Printf("\nDirect lookup of timeout flag value: %s\n", timeoutFlag.Value.String())
	// }

}

*/
