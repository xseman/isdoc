package types

import (
	"testing"
	"time"
)

func TestDecimal(t *testing.T) {
	tests := []struct {
		input   string
		valid   bool
		wantStr string
	}{
		{"123", true, "123"},
		{"123.45", true, "123.45"},
		{"-123.45", true, "-123.45"},
		{"0", true, "0"},
		{"0.00", true, "0.00"},
		{"-0.5", true, "-0.5"},
		{"1234567890.123456", true, "1234567890.123456"},
		{"", false, ""},
		{"abc", false, ""},
		{"12.34.56", false, ""},
		{"12,34", false, ""},
		{" 123", false, ""},
		{"123 ", false, ""},
		{"--123", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := NewDecimal(tt.input)
			if tt.valid {
				if err != nil {
					t.Errorf("NewDecimal(%q) unexpected error: %v", tt.input, err)
				}
				if d.String() != tt.wantStr {
					t.Errorf("NewDecimal(%q).String() = %q, want %q", tt.input, d.String(), tt.wantStr)
				}
			} else {
				if err == nil {
					t.Errorf("NewDecimal(%q) expected error, got nil", tt.input)
				}
			}
		})
	}
}

func TestDate(t *testing.T) {
	tests := []struct {
		input   string
		valid   bool
		wantStr string
	}{
		{"2024-01-15", true, "2024-01-15"},
		{"2024-12-31", true, "2024-12-31"},
		{"2000-01-01", true, "2000-01-01"},
		{"", true, ""},
		{"2024-1-15", false, ""},
		{"2024/01/15", false, ""},
		{"15-01-2024", false, ""},
		{"2024-01-15T10:00:00", false, ""},
		{"invalid", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := ParseDate(tt.input)
			if tt.valid {
				if err != nil {
					t.Errorf("ParseDate(%q) unexpected error: %v", tt.input, err)
				}
				if d.String() != tt.wantStr {
					t.Errorf("ParseDate(%q).String() = %q, want %q", tt.input, d.String(), tt.wantStr)
				}
			} else {
				if err == nil {
					t.Errorf("ParseDate(%q) expected error, got nil", tt.input)
				}
			}
		})
	}
}

func TestDateFromTime(t *testing.T) {
	tm := time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC)
	d := NewDate(tm)
	if d.String() != "2024-06-15" {
		t.Errorf("NewDate() = %q, want %q", d.String(), "2024-06-15")
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		input   string
		valid   bool
		wantVal bool
	}{
		{"true", true, true},
		{"false", true, false},
		{"", true, false},
		{"True", false, false},
		{"FALSE", false, false},
		{"1", false, false},
		{"0", false, false},
		{"yes", false, false},
		{"no", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			b, err := ParseBool(tt.input)
			if tt.valid {
				if err != nil {
					t.Errorf("ParseBool(%q) unexpected error: %v", tt.input, err)
				}
				if bool(b) != tt.wantVal {
					t.Errorf("ParseBool(%q) = %v, want %v", tt.input, bool(b), tt.wantVal)
				}
			} else {
				if err == nil {
					t.Errorf("ParseBool(%q) expected error, got nil", tt.input)
				}
			}
		})
	}
}

func TestBoolString(t *testing.T) {
	if Bool(true).String() != "true" {
		t.Error("Bool(true).String() should be 'true'")
	}
	if Bool(false).String() != "false" {
		t.Error("Bool(false).String() should be 'false'")
	}
}

func TestUUID(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12345678-1234-1234-1234-123456789012", true},
		{"ABCDEF00-1234-5678-9ABC-DEF012345678", true},
		{"abcdef00-1234-5678-9abc-def012345678", true},
		{"", true}, // empty is valid (optional field)
		{"12345678-1234-1234-1234-12345678901", false},   // too short
		{"12345678-1234-1234-1234-1234567890123", false}, // too long
		{"12345678123412341234123456789012", false},      // no dashes
		{"12345678-1234-1234-1234-12345678901G", false},  // invalid char
		{"not-a-uuid-at-all", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			u, err := NewUUID(tt.input)
			if tt.valid {
				if err != nil {
					t.Errorf("NewUUID(%q) unexpected error: %v", tt.input, err)
				}
				if u.String() != tt.input {
					t.Errorf("NewUUID(%q).String() = %q", tt.input, u.String())
				}
			} else {
				if err == nil {
					t.Errorf("NewUUID(%q) expected error, got nil", tt.input)
				}
			}
		})
	}
}
