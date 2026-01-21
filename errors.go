// Package errors defines error types for ISDOC parsing, validation, and encoding.
package isdoc

import (
	"fmt"
	"strings"
)

// Severity indicates the severity of a validation error.
type Severity int

const (
	// SeverityError indicates a fatal error that must be fixed.
	SeverityError Severity = iota
	// SeverityWarning indicates a non-fatal issue (used in non-strict mode).
	SeverityWarning
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "ERROR"
	case SeverityWarning:
		return "WARNING"
	default:
		return "UNKNOWN"
	}
}

// Error codes for validation errors.
const (
	ErrCodeRequiredField     = "REQUIRED_FIELD"
	ErrCodeInvalidEnum       = "INVALID_ENUM"
	ErrCodeInvalidDecimal    = "INVALID_DECIMAL"
	ErrCodeInvalidDate       = "INVALID_DATE"
	ErrCodeInvalidUUID       = "INVALID_UUID"
	ErrCodeInvalidPattern    = "INVALID_PATTERN"
	ErrCodeInvalidLength     = "INVALID_LENGTH"
	ErrCodeTotalMismatch     = "TOTAL_MISMATCH"
	ErrCodeVATMismatch       = "VAT_MISMATCH"
	ErrCodeReferenceNotFound = "REFERENCE_NOT_FOUND"
	ErrCodeDuplicateID       = "DUPLICATE_ID"
	ErrCodeInvalidXML        = "INVALID_XML"
	ErrCodeSchemaViolation   = "SCHEMA_VIOLATION"
)

// DecodeError represents an error during XML decoding.
type DecodeError struct {
	// Path is the JSON-style path to the element (e.g., "Invoice.InvoiceLines[0].ID").
	Path string
	// Err is the underlying error.
	Err error
}

func (e *DecodeError) Error() string {
	if e.Path == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Path, e.Err)
}

func (e *DecodeError) Unwrap() error {
	return e.Err
}

// NewDecodeError creates a new DecodeError.
func NewDecodeError(path string, err error) *DecodeError {
	return &DecodeError{Path: path, Err: err}
}

// DecodeErrors is a collection of decode errors.
type DecodeErrors []*DecodeError

func (e DecodeErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d decode errors:\n", len(e))
	for _, err := range e {
		fmt.Fprintf(&b, "  - %s\n", err.Error())
	}
	return b.String()
}

// HasErrors returns true if there are any decode errors.
func (e DecodeErrors) HasErrors() bool {
	return len(e) > 0
}

// ValidationError represents a validation error.
type ValidationError struct {
	// Field is the path to the field (e.g., "Invoice.ID").
	Field string
	// Code is a machine-readable error code.
	Code string
	// Severity indicates error vs warning.
	Severity Severity
	// Msg is a human-readable error message.
	Msg string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s (%s)", e.Severity, e.Field, e.Msg, e.Code)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []*ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d validation issues:\n", len(e))
	for _, err := range e {
		fmt.Fprintf(&b, "  - %s\n", err.Error())
	}
	return b.String()
}

// HasErrors returns true if there are any errors (not just warnings).
func (e ValidationErrors) HasErrors() bool {
	for _, err := range e {
		if err.Severity == SeverityError {
			return true
		}
	}
	return false
}

// HasWarnings returns true if there are any warnings.
func (e ValidationErrors) HasWarnings() bool {
	for _, err := range e {
		if err.Severity == SeverityWarning {
			return true
		}
	}
	return false
}

// Errors returns only errors (excluding warnings).
func (e ValidationErrors) Errors() ValidationErrors {
	var result ValidationErrors
	for _, err := range e {
		if err.Severity == SeverityError {
			result = append(result, err)
		}
	}
	return result
}

// Warnings returns only warnings.
func (e ValidationErrors) Warnings() ValidationErrors {
	var result ValidationErrors
	for _, err := range e {
		if err.Severity == SeverityWarning {
			result = append(result, err)
		}
	}
	return result
}
