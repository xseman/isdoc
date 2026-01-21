// Package types provides custom types for ISDOC with validation and XML marshaling.
package types

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
)

// decimalPattern validates decimal format: optional minus, digits, optional decimal part
// Allows formats like: 123, 123.45, -123.45, .45, -.45
var decimalPattern = regexp.MustCompile(`^-?(\d+\.?\d*|\d*\.?\d+)$`)

// Decimal represents a decimal number as a string to preserve precision.
// This avoids floating-point drift issues common with float64.
type Decimal string

// NewDecimal creates a Decimal from a string, validating the format.
func NewDecimal(s string) (Decimal, error) {
	if !decimalPattern.MatchString(s) {
		return "", fmt.Errorf("invalid decimal format: %q", s)
	}
	return Decimal(s), nil
}

// MustDecimal creates a Decimal from a string, panicking on invalid format.
func MustDecimal(s string) Decimal {
	d, err := NewDecimal(s)
	if err != nil {
		panic(err)
	}
	return d
}

// String returns the decimal as a string.
func (d Decimal) String() string {
	return string(d)
}

// IsZero returns true if the decimal is empty (zero value).
func (d Decimal) IsZero() bool {
	return d == ""
}

// Float64 returns the decimal as a float64.
// Returns 0 if the decimal is empty or cannot be parsed.
// Note: This may lose precision for very large or precise values.
func (d Decimal) Float64() float64 {
	if d == "" {
		return 0
	}
	f, err := strconv.ParseFloat(string(d), 64)
	if err != nil {
		return 0
	}
	return f
}

// Equal returns true if two decimals represent the same numeric value.
// Compares as float64, allowing for string format differences (e.g., "1.0" == "1.00").
func (d Decimal) Equal(other Decimal) bool {
	return d.Float64() == other.Float64()
}

// MarshalXML implements xml.Marshaler for Decimal.
func (d Decimal) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(string(d), start)
}

// UnmarshalXML implements xml.Unmarshaler for Decimal.
func (d *Decimal) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := dec.DecodeElement(&s, &start); err != nil {
		return err
	}
	if s == "" {
		*d = ""
		return nil
	}
	parsed, err := NewDecimal(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// MarshalXMLAttr implements xml.MarshalerAttr for Decimal.
func (d Decimal) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: string(d)}, nil
}

// UnmarshalXMLAttr implements xml.UnmarshalerAttr for Decimal.
func (d *Decimal) UnmarshalXMLAttr(attr xml.Attr) error {
	if attr.Value == "" {
		*d = ""
		return nil
	}
	parsed, err := NewDecimal(attr.Value)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
