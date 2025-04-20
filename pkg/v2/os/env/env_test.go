package env

// import (
// 	"context"
// 	"errors"
// 	"flag"
// 	"net"
// 	"os"
// 	"strings"
// 	"testing"
// 	"time"

// 	// Mock Secret Manager client
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"

// 	secretmanager "cloud.google.com/go/secretmanager/apiv1"
// 	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
// 	"google.golang.org/grpc"
// )

// // --- Mock Secret Manager ---

// // mockSecretManagerServer implements the SecretManagerServiceServer interface for testing.
// type mockSecretManagerServer struct {
// 	secretmanagerpb.UnimplementedSecretManagerServiceServer
// 	secrets map[string]string // Map secret name to payload
// 	err     error             // Optional error to return
// }

// func (s *mockSecretManagerServer) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
// 	if s.err != nil {
// 		return nil, s.err
// 	}
// 	if payload, ok := s.secrets[req.Name]; ok {
// 		return &secretmanagerpb.AccessSecretVersionResponse{
// 			Name: req.Name,
// 			Payload: &secretmanagerpb.SecretPayload{
// 				Data: []byte(payload),
// 			},
// 		}, nil
// 	}
// 	return nil, status.Errorf(codes.NotFound, "secret %q not found", req.Name)
// }

// // newTestSecretManagerClient creates a client connected to the mock server.
// func newTestSecretManagerClient(t *testing.T, mockServer *mockSecretManagerServer) *secretmanager.Client {
// 	// Create a mock gRPC connection (in-memory)
// 	// conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
// 	_, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
// 		// This part is a bit complex for a simple example, usually involves bufconn listener.
// 		// For simplicity here, we'll directly set the client, assuming the mock works.
// 		// In a real scenario, you'd use bufconn:
// 		// lis := bufconn.Listen(1024 * 1024)
// 		// grpcServer := grpc.NewServer()
// 		// secretmanagerpb.RegisterSecretManagerServiceServer(grpcServer, mockServer)
// 		// go func() {
// 		// 	if err := grpcServer.Serve(lis); err != nil {
// 		// 		t.Fatalf("Server exited with error: %v", err)
// 		// 	}
// 		// }()
// 		// return lis.Dial()
// 		// Since direct client setting is easier for this context:
// 		t.Skip("Skipping complex gRPC bufconn setup for mock, directly setting client.")
// 		return nil, nil // Placeholder, won't actually be used with direct client set below.
// 	}))
// 	if err != nil && !strings.Contains(err.Error(), "transport is closing") { // Ignore closing error during skip
// 		// This error check might not be hit due to the skip above.
// 		// t.Fatalf("Failed to dial bufnet: %v", err)
// 	}

// 	// Create client with the mock connection (or directly if skipping bufconn)
// 	// client, err := secretmanager.NewClient(context.Background(), option.WithGRPCConn(conn))
// 	// if err != nil {
// 	// 	t.Fatalf("Failed to create test client: %v", err)
// 	// }

// 	// --- Direct Client Setting (Simpler for Example) ---
// 	// This bypasses the gRPC connection setup for simplicity.
// 	// Create a minimal client structure sufficient for the mock interaction.
// 	// NOTE: This is NOT how you'd typically mock; it's a shortcut.
// 	// A proper mock would use interfaces or the bufconn approach above.
// 	client := &secretmanager.Client{} // Placeholder client
// 	// We will replace the global smClient directly in tests needing the mock.

// 	return client // Return the placeholder or the bufconn client
// }

// // --- Test Setup ---

// // Helper to set environment variables and clean them up
// func setenv(t *testing.T, key, value string) {
// 	t.Helper()
// 	originalValue, isset := os.LookupEnv(key)
// 	os.Setenv(key, value)
// 	t.Cleanup(func() {
// 		if isset {
// 			os.Setenv(key, originalValue)
// 		} else {
// 			os.Unsetenv(key)
// 		}
// 	})
// }

// // Helper to create a dummy .env file and clean it up
// func createDotEnv(t *testing.T, content string) string {
// 	t.Helper()
// 	tmpFile, err := os.CreateTemp("", ".env")
// 	if err != nil {
// 		t.Fatalf("Failed to create temp .env file: %v", err)
// 	}
// 	_, err = tmpFile.WriteString(content)
// 	if err != nil {
// 		tmpFile.Close()
// 		t.Fatalf("Failed to write to temp .env file: %v", err)
// 	}
// 	err = tmpFile.Close()
// 	if err != nil {
// 		t.Fatalf("Failed to close temp .env file: %v", err)
// 	}

// 	originalDir, err := os.Getwd()
// 	if err != nil {
// 		t.Fatalf("Failed to get working directory: %v", err)
// 	}
// 	err = os.Chdir(os.TempDir()) // Change to temp dir so Load finds the file
// 	if err != nil {
// 		t.Fatalf("Failed to change directory: %v", err)
// 	}

// 	t.Cleanup(func() {
// 		os.Remove(tmpFile.Name())
// 		os.Chdir(originalDir) // Change back
// 	})
// 	return tmpFile.Name()
// }

// // Helper to reset flags between test runs
// func resetFlags() {
// 	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
// }

// // --- Test Cases ---

// type TestConfig struct {
// 	Host     string        `env:"HOST" flag:"host" default:"localhost"`
// 	Port     int           `env:"PORT" flag:"port" default:"8080" required:"true"`
// 	Timeout  time.Duration `env:"TIMEOUT" default:"5s"`
// 	Debug    bool          `env:"DEBUG" flag:"debug" default:"false"`
// 	Rate     float64       `env:"RATE" default:"0.5"`
// 	Empty    string        // No tags
// 	Optional *string       `env:"OPTIONAL"`
// 	Required string        `env:"REQUIRED" required:"true"`
// 	Ignored  string        `env:"IGNORED" ignored:"true" default:"should-be-ignored"`
// 	DefOnly  string        `default:"default-val"`
// 	// TODO: Elegant secrets handling
// 	// APIKey    string        `env:"API_KEY" secret:"projects/p/secrets/k/versions/1" required:"true"`
// 	// Token     string        `secret:"projects/p/secrets/t/versions/latest"`
// 	// SecretErr string        `secret:"projects/p/secrets/nonexistent/versions/1"`
// 	APIKey    string `env:"API_KEY"`
// 	Token     string `env:"TOKEN"`
// 	SecretErr string `env:"SECRET_ERR"`
// }

// func TestLoadDefaults(t *testing.T) {
// 	resetFlags()
// 	var cfg TestConfig
// 	ctx := context.Background()
// 	err := ProcessConfig(ctx, "TEST_", &cfg) // No env, flags, or secrets set

// 	if err != nil {
// 		// Expecting required field errors
// 		if !strings.Contains(err.Error(), "Port") || !strings.Contains(err.Error(), "Required") {
// 			t.Fatalf("Expected required field errors for APIKey and Required, got: %v", err)
// 		}
// 	} else {
// 		t.Fatal("Expected required field errors, but got nil")
// 	}

// 	// Check defaults
// 	if cfg.Host != "localhost" {
// 		t.Errorf("Expected Host 'localhost', got %q", cfg.Host)
// 	}
// 	if cfg.Port != 8080 {
// 		t.Errorf("Expected Port 8080, got %d", cfg.Port)
// 	}
// 	if cfg.Timeout != 5*time.Second {
// 		t.Errorf("Expected Timeout 5s, got %v", cfg.Timeout)
// 	}
// 	if cfg.Debug != false {
// 		t.Errorf("Expected Debug false, got %t", cfg.Debug)
// 	}
// 	if cfg.Rate != 0.5 {
// 		t.Errorf("Expected Rate 0.5, got %f", cfg.Rate)
// 	}
// 	if cfg.Ignored != "" {
// 		t.Errorf("Expected Ignored field to be empty, got %q", cfg.Ignored)
// 	}
// 	if cfg.DefOnly != "default-val" {
// 		t.Errorf("Expected DefOnly 'default-val', got %q", cfg.DefOnly)
// 	}
// }

// func TestLoadEnv(t *testing.T) {
// 	resetFlags()
// 	setenv(t, "TEST_HOST", "envhost")
// 	setenv(t, "TEST_PORT", "9090")
// 	setenv(t, "TEST_API_KEY", "envkey")
// 	setenv(t, "TEST_TIMEOUT", "10s")
// 	setenv(t, "TEST_DEBUG", "true")
// 	setenv(t, "TEST_RATE", "1.23")
// 	setenv(t, "TEST_REQUIRED", "env-required")
// 	setenv(t, "TEST_OPTIONAL", "env-opt")
// 	setenv(t, "TEST_IGNORED", "env-ignored") // Should still be ignored

// 	var cfg TestConfig
// 	ctx := context.Background()
// 	err := ProcessConfig(ctx, "TEST_", &cfg) // No flags or secrets

// 	if err != nil {
// 		t.Fatalf("Load failed: %v", err)
// 	}

// 	if cfg.Host != "envhost" {
// 		t.Errorf("Expected Host 'envhost', got %q", cfg.Host)
// 	}
// 	if cfg.Port != 9090 {
// 		t.Errorf("Expected Port 9090, got %d", cfg.Port)
// 	}
// 	if cfg.APIKey != "envkey" {
// 		t.Errorf("Expected APIKey 'envkey', got %q", cfg.APIKey)
// 	}
// 	if cfg.Timeout != 10*time.Second {
// 		t.Errorf("Expected Timeout 10s, got %v", cfg.Timeout)
// 	}
// 	if cfg.Debug != true {
// 		t.Errorf("Expected Debug true, got %t", cfg.Debug)
// 	}
// 	if cfg.Rate != 1.23 {
// 		t.Errorf("Expected Rate 1.23, got %f", cfg.Rate)
// 	}
// 	if cfg.Required != "env-required" {
// 		t.Errorf("Expected Required 'env-required', got %q", cfg.Required)
// 	}
// 	if cfg.Optional == nil || *cfg.Optional != "env-opt" {
// 		t.Errorf("Expected Optional 'env-opt', got %v", cfg.Optional)
// 	}
// 	if cfg.Ignored != "" {
// 		t.Errorf("Expected Ignored field to be empty, got %q", cfg.Ignored)
// 	}
// 	if cfg.DefOnly != "default-val" { // Should retain default
// 		t.Errorf("Expected DefOnly 'default-val', got %q", cfg.DefOnly)
// 	}
// }

// // TODO: changing directory for .env file doesn't work
// // func TestLoadDotEnv(t *testing.T) {
// // 	resetFlags()
// // 	// Create .env file in temp dir
// // 	dotEnvContent := `
// // TEST_HOST=dotenvhost
// // TEST_PORT=7070
// // # Comment
// // TEST_API_KEY=dotenvkey
// // TEST_REQUIRED=dotenv-required
// // TEST_TIMEOUT=1m
// // `
// // 	createDotEnv(t, dotEnvContent) // Changes wd to temp dir

// // 	// Set an overlapping env var - should be overridden by .env
// // 	setenv(t, "TEST_HOST", "envhost-overridden")

// // 	var cfg TestConfig
// // 	ctx := context.Background()
// // 	err := Load(ctx, "TEST_", &cfg)

// // 	if err != nil {
// // 		t.Fatalf("Load failed: %v", err)
// // 	}

// // 	if cfg.Host != "dotenvhost" {
// // 		t.Errorf("Expected Host 'dotenvhost', got %q", cfg.Host)
// // 	}
// // 	if cfg.Port != 7070 {
// // 		t.Errorf("Expected Port 7070, got %d", cfg.Port)
// // 	}
// // 	if cfg.APIKey != "dotenvkey" {
// // 		t.Errorf("Expected APIKey 'dotenvkey', got %q", cfg.APIKey)
// // 	}
// // 	if cfg.Required != "dotenv-required" {
// // 		t.Errorf("Expected Required 'dotenv-required', got %q", cfg.Required)
// // 	}
// // 	if cfg.Timeout != 1*time.Minute {
// // 		t.Errorf("Expected Timeout 1m, got %v", cfg.Timeout)
// // 	}
// // 	// Check defaults for others
// // 	if cfg.Debug != false {
// // 		t.Errorf("Expected Debug false, got %t", cfg.Debug)
// // 	}
// // }

// func TestLoadFlags(t *testing.T) {
// 	resetFlags()
// 	// Define flags corresponding to the struct tags
// 	flag.String("host", "default-flag-host", "Host name")
// 	flag.Int("port", 1111, "Port number")
// 	flag.Bool("debug", false, "Enable debug")
// 	// Note: No flag defined for APIKey, Timeout, Rate, Required, Optional

// 	// Simulate command line arguments
// 	// IMPORTANT: os.Args needs to be manipulated *before* flag.Parse()
// 	originalArgs := os.Args
// 	os.Args = []string{"cmd", "--host=flaghost", "--port=2222", "--debug"}
// 	t.Cleanup(func() { os.Args = originalArgs })

// 	flag.Parse() // Parse the simulated args

// 	// Set some env vars that should be overridden by flags
// 	setenv(t, "TEST_HOST", "envhost-overridden")
// 	setenv(t, "TEST_PORT", "9999")   // Should be overridden
// 	setenv(t, "TEST_DEBUG", "false") // Should be overridden
// 	setenv(t, "TEST_API_KEY", "envkey-for-flag-test")
// 	setenv(t, "TEST_REQUIRED", "env-required-for-flag-test")

// 	var cfg TestConfig
// 	ctx := context.Background()
// 	err := ProcessConfig(ctx, "TEST_", &cfg)

// 	// Expecting error because APIKey is required but not provided by flag
// 	if err != nil {
// 		// APIKey is required and not set by flag, Token is not required
// 		if !strings.Contains(err.Error(), "APIKey") || strings.Contains(err.Error(), "Token") {
// 			t.Fatalf("Expected required field error only for APIKey, got: %v", err)
// 		}
// 	} else {
// 		t.Fatal("Expected required field error for APIKey, but got nil")
// 	}

// 	// Check values that were set
// 	if cfg.Host != "flaghost" {
// 		t.Errorf("Expected Host 'flaghost', got %q", cfg.Host)
// 	}
// 	if cfg.Port != 2222 {
// 		t.Errorf("Expected Port 2222, got %d", cfg.Port)
// 	}
// 	if cfg.Debug != true { // --debug sets it to true
// 		t.Errorf("Expected Debug true, got %t", cfg.Debug)
// 	}
// 	// Check values from env that were not overridden by flags
// 	if cfg.APIKey != "envkey-for-flag-test" {
// 		t.Errorf("Expected APIKey 'envkey-for-flag-test', got %q", cfg.APIKey)
// 	}
// 	if cfg.Required != "env-required-for-flag-test" {
// 		t.Errorf("Expected Required 'env-required-for-flag-test', got %q", cfg.Required)
// 	}
// 	// Check default
// 	if cfg.Timeout != 5*time.Second {
// 		t.Errorf("Expected Timeout 5s, got %v", cfg.Timeout)
// 	}
// }

// func TestLoadSecrets(t *testing.T) {
// 	resetFlags()
// 	ctx := context.Background()

// 	// // --- Setup Mock Secret Manager ---
// 	// mockSrv := &mockSecretManagerServer{
// 	// 	secrets: map[string]string{
// 	// 		"projects/p/secrets/k/versions/1":      "secretkey",
// 	// 		"projects/p/secrets/t/versions/latest": "secrettoken",
// 	// 		// "projects/p/secrets/nonexistent/versions/1" // This one doesn't exist
// 	// 	},
// 	// }
// 	// // Since direct client creation with mock server is complex without bufconn,
// 	// // we'll directly replace the global client for this test.
// 	// // Ensure cleanup happens.
// 	// originalSMClient := smClient
// 	// // Use a placeholder client; the mock logic is self-contained for AccessSecretVersion
// 	// // For a real test, you'd use the client from newTestSecretManagerClient connected via bufconn.
// 	// // smClient = newTestSecretManagerClient(t, mockSrv)
// 	// // --- Direct Mocking (Simpler for this example) ---
// 	// // Replace the global client with our mock logic directly.
// 	// // This requires careful handling in concurrent tests.
// 	// smClient = &secretmanager.Client{} // Minimal placeholder
// 	// // Override the function that uses the client
// 	// originalAccessFn := accessSecretVersion
// 	// accessSecretVersion = func(ctx context.Context, name string) (string, error) {
// 	// 	// Directly use the mock server logic
// 	// 	resp, err := mockSrv.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{Name: name})
// 	// 	if err != nil {
// 	// 		return "", err
// 	// 	}
// 	// 	return string(resp.Payload.Data), nil
// 	// }

// 	// t.Cleanup(func() {
// 	// 	smClient = originalSMClient            // Restore original client
// 	// 	accessSecretVersion = originalAccessFn // Restore original function
// 	// 	if smClient != nil {
// 	// 		// Attempt to close the original client if it existed
// 	// 		// smClient.Close() // Be cautious closing potentially shared clients
// 	// 	}
// 	// })
// 	// // --- End Mock Setup ---

// 	// Set env vars that should be overridden by secrets
// 	setenv(t, "TEST_API_KEY", "envkey-overridden")
// 	setenv(t, "TEST_REQUIRED", "env-required-for-secret-test") // Required, but not a secret

// 	var cfg TestConfig
// 	err := ProcessConfig(ctx, "TEST_", &cfg)

// 	// Expecting error because SecretErr tries to access a non-existent secret
// 	if err != nil {
// 		if !strings.Contains(err.Error(), "SecretErr") || !strings.Contains(err.Error(), "nonexistent") {
// 			t.Fatalf("Expected error accessing nonexistent secret for SecretErr, got: %v", err)
// 		}
// 	} else {
// 		t.Fatal("Expected error accessing nonexistent secret, but got nil")
// 	}

// 	// Check values set by secrets
// 	if cfg.APIKey != "secretkey" {
// 		t.Errorf("Expected APIKey 'secretkey', got %q", cfg.APIKey)
// 	}
// 	if cfg.Token != "secrettoken" {
// 		t.Errorf("Expected Token 'secrettoken', got %q", cfg.Token)
// 	}

// 	// Check value set by env (not overridden by secret)
// 	if cfg.Required != "env-required-for-secret-test" {
// 		t.Errorf("Expected Required 'env-required-for-secret-test', got %q", cfg.Required)
// 	}

// 	// Check defaults for others
// 	if cfg.Host != "localhost" {
// 		t.Errorf("Expected Host 'localhost', got %q", cfg.Host)
// 	}
// 	if cfg.Port != 8080 {
// 		t.Errorf("Expected Port 8080, got %d", cfg.Port)
// 	}
// 	if cfg.SecretErr != "" {
// 		t.Errorf("Expected SecretErr to be empty due to loading error, got %q", cfg.SecretErr)
// 	}
// }

// func TestLoadPrecedence(t *testing.T) {
// 	resetFlags()
// 	ctx := context.Background()

// 	// 1. Default: Port=8080
// 	// 2. .env: Port=7070
// 	dotEnvContent := "TEST_PORT=7070\nTEST_API_KEY=dotenvkey"
// 	createDotEnv(t, dotEnvContent)
// 	// 3. Env Var: Port=9090, Host=envhost
// 	setenv(t, "TEST_PORT", "9090")
// 	setenv(t, "TEST_HOST", "envhost")
// 	setenv(t, "TEST_REQUIRED", "env-required")
// 	// // 4. Secret: APIKey=secretkey
// 	// mockSrv := &mockSecretManagerServer{
// 	// 	secrets: map[string]string{"projects/p/secrets/k/versions/1": "secretkey"},
// 	// }
// 	// originalSMClient := smClient
// 	// smClient = &secretmanager.Client{} // Placeholder
// 	// originalAccessFn := accessSecretVersion
// 	// accessSecretVersion = func(ctx context.Context, name string) (string, error) {
// 	// 	resp, err := mockSrv.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{Name: name})
// 	// 	if err != nil {
// 	// 		return "", err
// 	// 	}
// 	// 	return string(resp.Payload.Data), nil
// 	// }
// 	// t.Cleanup(func() {
// 	// 	smClient = originalSMClient
// 	// 	accessSecretVersion = originalAccessFn
// 	// })
// 	// // 5. Flag: Host=flaghost, Port=2222
// 	flag.String("host", "default-flag-host", "")
// 	flag.Int("port", 1111, "")
// 	flag.Bool("debug", false, "") // Not set on command line
// 	originalArgs := os.Args
// 	os.Args = []string{"cmd", "--host=flaghost", "--port=2222"}
// 	t.Cleanup(func() { os.Args = originalArgs })
// 	flag.Parse()

// 	var cfg TestConfig
// 	err := ProcessConfig(ctx, "TEST_", &cfg) // Prefix is TEST_

// 	if err != nil {
// 		t.Fatalf("Load failed: %v", err)
// 	}

// 	// Check final values based on precedence (Flag > Secret > Env > .env > Default)
// 	if cfg.Host != "flaghost" { // Flag overrides Env
// 		t.Errorf("Expected Host 'flaghost', got %q", cfg.Host)
// 	}
// 	if cfg.Port != 2222 { // Flag overrides Env overrides .env overrides Default
// 		t.Errorf("Expected Port 2222, got %d", cfg.Port)
// 	}
// 	if cfg.APIKey != "secretkey" { // Secret overrides .env
// 		t.Errorf("Expected APIKey 'secretkey', got %q", cfg.APIKey)
// 	}
// 	if cfg.Required != "env-required" { // Only set by Env
// 		t.Errorf("Expected Required 'env-required', got %q", cfg.Required)
// 	}
// 	if cfg.Debug != false { // Default, as flag wasn't set true
// 		t.Errorf("Expected Debug false, got %t", cfg.Debug)
// 	}
// 	if cfg.Timeout != 5*time.Second { // Default
// 		t.Errorf("Expected Timeout 5s, got %v", cfg.Timeout)
// 	}
// }

// func TestLoadInvalidSpec(t *testing.T) {
// 	resetFlags()
// 	ctx := context.Background()
// 	var cfg TestConfig // Not a pointer
// 	err := ProcessConfig(ctx, "", cfg)
// 	if !errors.Is(err, errInvalidSpecification) {
// 		t.Errorf("Expected errInvalidSpecification for non-pointer spec, got %v", err)
// 	}

// 	var i int // Not a struct
// 	err = ProcessConfig(ctx, "", &i)
// 	if !errors.Is(err, errInvalidSpecification) {
// 		t.Errorf("Expected errInvalidSpecification for non-struct pointer spec, got %v", err)
// 	}

// 	var nilPtr *TestConfig // Nil pointer
// 	err = ProcessConfig(ctx, "", nilPtr)
// 	if !errors.Is(err, errInvalidSpecification) {
// 		t.Errorf("Expected errInvalidSpecification for nil pointer spec, got %v", err)
// 	}
// }

// func TestLoadPointerField(t *testing.T) {
// 	resetFlags()
// 	ctx := context.Background()

// 	// Test case 1: Env var is set
// 	setenv(t, "TEST_OPTIONAL", "hello pointer")
// 	var cfg1 TestConfig
// 	err1 := ProcessConfig(ctx, "TEST_", &cfg1)
// 	if err1 != nil && !strings.Contains(err1.Error(), "APIKey") && !strings.Contains(err1.Error(), "Required") {
// 		// Ignore required errors for this specific test focus
// 		t.Fatalf("Load failed unexpectedly: %v", err1)
// 	}
// 	if cfg1.Optional == nil {
// 		t.Errorf("Expected Optional field to be non-nil when env var is set")
// 	} else if *cfg1.Optional != "hello pointer" {
// 		t.Errorf("Expected Optional field value 'hello pointer', got %q", *cfg1.Optional)
// 	}
// 	os.Unsetenv("TEST_OPTIONAL") // Clean up for next case

// 	// Test case 2: Env var is not set
// 	var cfg2 TestConfig
// 	err2 := ProcessConfig(ctx, "TEST_", &cfg2)
// 	// Ignore required field errors
// 	if err2 != nil && !strings.Contains(err2.Error(), "APIKey") && !strings.Contains(err2.Error(), "Required") {
// 		t.Fatalf("Load failed unexpectedly: %v", err2)
// 	}
// 	if cfg2.Optional != nil {
// 		t.Errorf("Expected Optional field to be nil when env var is not set, got %v", *cfg2.Optional)
// 	}
// }

// // Add more tests for:
// // - Different data types (uint, float, bool variations)
// // - Secret Manager client initialization errors
// // - Complex secret names
// // - Empty prefix
// // - Type conversion errors (e.g., non-integer value for int field)
