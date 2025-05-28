package err

import (
	"errors"
	"testing"
)

// TestIsError tests the IsError function.
func TestIsError(t *testing.T) {
	// Test case 1: Input is nil
	if IsError(nil) {
		t.Errorf("IsError(nil) = true; want false")
	}

	// Test case 2: Input is a non-nil error
	testErr := errors.New("this is a test error")
	if !IsError(testErr) {
		t.Errorf("IsError(testErr) = false; want true")
	}
}

// TestExitIfError tests the ExitIfError function.
func TestExitIfError(t *testing.T) {
	// Test case 1: Input is nil (should not panic)
	func() {
		// Use a deferred function to check if a panic occurred unexpectedly
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ExitIfError(nil) panicked unexpectedly: %v", r)
			}
		}()
		PanicIfError("Test", nil) // Call the function with nil
	}() // Execute the anonymous function immediately

	// Test case 2: Input is a non-nil error (should panic)
	testErr := errors.New("this should cause a panic")
	func() {
		// Use a deferred function to recover from the expected panic
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("ExitIfError(testErr) did not panic as expected")
				return // Exit defer early if no panic occurred
			}
			// Check if the recovered value is the error we passed in
			if recoveredErr, ok := r.(error); !ok || recoveredErr != testErr {
				t.Errorf("ExitIfError(testErr) panicked with unexpected value: got %v, want %v", r, testErr)
			}
		}()
		PanicIfError("Test", testErr) // Call the function with the error
	}() // Execute the anonymous function immediately
}
