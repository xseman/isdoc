// Package types provides custom types for ISDOC with validation and XML marshaling.
//
// Types in this package:
//   - Decimal: String-backed decimal to prevent floating-point drift
//   - Date: YYYY-MM-DD date format
//   - Bool: Strict true/false only (rejects 0/1)
//   - UUID: 36-character UUID with pattern validation
package types
