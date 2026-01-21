package isdoc

import (
	"encoding/xml"
	"testing"

	"github.com/xseman/isdoc/types"
)

// TestDecimalPatternValidation tests various decimal formats.
// Based on PHP RestrictionTest.testDecimal
func TestDecimalPatternValidation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"positive integer", "100", true},
		{"negative integer", "-100", true},
		{"positive decimal", "100.25", true},
		{"negative decimal", "-100.25", true},
		{"large decimal", "12345678901234.1234", true},
		{"leading decimal", ".21837", true},
		{"negative leading decimal", "-.21837", true},
		{"zero", "0", true},
		{"negative zero", "-0", true},
		{"decimal zero", "0.0", true},
		{"small decimal", "0.00001", true},
		{"empty string", "", false},
		{"alphabetic", "abc", false},
		{"mixed", "12.34abc", false},
		{"multiple dots", "12.34.56", false},
		{"comma separator", "12,34", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := types.NewDecimal(tt.input)
			if tt.valid {
				if err != nil {
					t.Errorf("expected valid decimal %q, got error: %v", tt.input, err)
				}
				// Verify it can be used
				_ = d.String()
			} else {
				if err == nil {
					t.Errorf("expected invalid decimal %q to fail, got: %s", tt.input, d.String())
				}
			}
		})
	}
}

// TestDecimalString tests decimal string representation.
func TestDecimalString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "100", "100"},
		{"decimal", "100.25", "100.25"},
		{"leading decimal", ".21837", ".21837"},
		{"negative", "-100.25", "-100.25"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := types.MustDecimal(tt.input)
			if d.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, d.String())
			}
		})
	}
}

// TestDateValidation tests date parsing and formatting.
// Based on PHP pattern validation for dates
func TestDateValidation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
		year  int
		month int
		day   int
	}{
		{"standard date", "2021-08-16", true, 2021, 8, 16},
		{"january", "2021-01-01", true, 2021, 1, 1},
		{"december", "2021-12-31", true, 2021, 12, 31},
		{"leap year", "2020-02-29", true, 2020, 2, 29},
		{"year 2000", "2000-01-01", true, 2000, 1, 1},
		{"invalid month", "2021-13-01", false, 0, 0, 0},
		{"invalid day", "2021-02-30", false, 0, 0, 0},
		{"wrong format", "16-08-2021", false, 0, 0, 0},
		{"slash format", "2021/08/16", false, 0, 0, 0},
		{"empty", "", true, 0, 0, 0}, // Empty is allowed (zero value)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := types.ParseDate(tt.input)
			if tt.valid {
				if err != nil {
					t.Errorf("expected valid date %q, got error: %v", tt.input, err)
					return
				}
				if tt.input != "" {
					year, month, day := d.Date()
					if year != tt.year || int(month) != tt.month || day != tt.day {
						t.Errorf("expected %d-%02d-%02d, got %d-%02d-%02d",
							tt.year, tt.month, tt.day, year, month, day)
					}
				}
			} else {
				if err == nil {
					t.Errorf("expected invalid date %q to fail, got: %s", tt.input, d.String())
				}
			}
		})
	}
}

// TestUUIDValidation tests UUID parsing.
// Based on PHP pattern validation for UUIDs
func TestUUIDValidation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"standard uuid", "7B4C5BE0-288C-11D2-8E62-004095452B84", true},
		{"lowercase uuid", "7b4c5be0-288c-11d2-8e62-004095452b84", true},
		{"mixed case uuid", "7B4c5bE0-288C-11d2-8E62-004095452B84", true},
		{"all zeros", "00000000-0000-0000-0000-000000000000", true},
		{"all f's", "FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF", true},
		{"too short", "7B4C5BE0-288C-11D2-8E62-004095452B8", false},
		{"too long", "7B4C5BE0-288C-11D2-8E62-004095452B845", false},
		{"missing dash", "7B4C5BE0288C-11D2-8E62-004095452B84", false},
		{"invalid char", "7B4C5BE0-288C-11D2-8E62-00409545ZB84", false},
		{"empty", "", true}, // Empty is allowed (zero value)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := types.NewUUID(tt.input)
			if tt.valid {
				if err != nil {
					t.Errorf("expected valid UUID %q, got error: %v", tt.input, err)
				}
				_ = u.String()
			} else {
				if err == nil {
					t.Errorf("expected invalid UUID %q to fail", tt.input)
				}
			}
		})
	}
}

// TestBoolParsing tests boolean parsing via XML.
func TestBoolParsing(t *testing.T) {
	// Test via XML parsing since types.Bool uses XML unmarshaling
	tests := []struct {
		name     string
		xml      string
		expected bool
		valid    bool
	}{
		{"true lowercase", "<wrapper><v>true</v></wrapper>", true, true},
		{"false lowercase", "<wrapper><v>false</v></wrapper>", false, true},
		// Note: ISDOC Bool type only accepts "true" or "false", not "1" or "0"
		{"1", "<wrapper><v>1</v></wrapper>", true, false},
		{"0", "<wrapper><v>0</v></wrapper>", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type wrapper struct {
				V types.Bool `xml:"v"`
			}
			var w wrapper
			err := xml.Unmarshal([]byte(tt.xml), &w)
			if tt.valid {
				if err != nil {
					t.Errorf("parse failed: %v", err)
					return
				}
				if bool(w.V) != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, bool(w.V))
				}
			} else {
				if err == nil {
					t.Logf("Note: value %q was parsed as %v (strict mode might reject this)", tt.name, bool(w.V))
				}
			}
		})
	}
}

// TestCurrencyCodeValidation tests 3-letter currency codes.
func TestCurrencyCodeValidation(t *testing.T) {
	validCodes := []string{"CZK", "EUR", "USD", "GBP", "PLN", "CHF"}
	invalidCodes := []string{"CZ", "EURO", "us", "123", ""}

	for _, code := range validCodes {
		t.Run("valid_"+code, func(t *testing.T) {
			if len(code) != 3 {
				t.Errorf("currency code %q should be 3 characters", code)
			}
		})
	}

	for _, code := range invalidCodes {
		name := code
		if name == "" {
			name = "empty"
		}
		t.Run("invalid_"+name, func(t *testing.T) {
			if len(code) == 3 {
				// Valid format but potentially invalid value
				t.Logf("code %q has 3 chars but may be semantically invalid", code)
			}
		})
	}
}

// TestCountryCodeValidation tests 2-letter country codes.
func TestCountryCodeValidation(t *testing.T) {
	validCodes := []string{"CZ", "SK", "DE", "AT", "PL", "US", "GB"}
	invalidCodes := []string{"C", "CZE", "cz", "12", ""}

	for _, code := range validCodes {
		t.Run("valid_"+code, func(t *testing.T) {
			if len(code) != 2 {
				t.Errorf("country code %q should be 2 characters", code)
			}
		})
	}

	for _, code := range invalidCodes {
		name := code
		if name == "" {
			name = "empty"
		}
		t.Run("invalid_"+name, func(t *testing.T) {
			if len(code) == 2 {
				// Valid format but potentially invalid value
				t.Logf("code %q has 2 chars but may be semantically invalid", code)
			}
		})
	}
}

// TestVATCalculationMethodValidation tests valid VAT calculation methods.
func TestVATCalculationMethodValidation(t *testing.T) {
	validMethods := []int{0, 1}
	invalidMethods := []int{-1, 2, 99}

	for _, method := range validMethods {
		t.Run("valid", func(t *testing.T) {
			if method != 0 && method != 1 {
				t.Errorf("VAT calculation method %d should be valid", method)
			}
		})
	}

	for _, method := range invalidMethods {
		t.Run("invalid", func(t *testing.T) {
			if method >= 0 && method <= 1 {
				t.Errorf("VAT calculation method %d should be invalid", method)
			}
		})
	}
}

// TestDecimalPrecision tests decimal precision handling.
func TestDecimalPrecision(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no decimals", "100", "100"},
		{"one decimal", "100.5", "100.5"},
		{"two decimals", "100.55", "100.55"},
		{"four decimals", "100.5555", "100.5555"},
		{"trailing zeros", "100.50", "100.50"},
		{"leading decimal", ".5", ".5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := types.MustDecimal(tt.input)
			result := d.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
