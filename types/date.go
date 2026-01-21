package types

import (
	"encoding/xml"
	"fmt"
	"time"
)

// DateFormat is the standard ISDOC date format (ISO 8601 date only).
const DateFormat = "2006-01-02"

// Date represents an ISDOC date (YYYY-MM-DD format).
type Date struct {
	time.Time
}

// NewDate creates a Date from a time.Time value.
func NewDate(t time.Time) Date {
	// Normalize to midnight UTC
	return Date{time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)}
}

// ParseDate parses a date string in YYYY-MM-DD format.
func ParseDate(s string) (Date, error) {
	if s == "" {
		return Date{}, nil
	}
	t, err := time.Parse(DateFormat, s)
	if err != nil {
		return Date{}, fmt.Errorf("invalid date format %q: expected YYYY-MM-DD", s)
	}
	return NewDate(t), nil
}

// MustParseDate parses a date string, panicking on invalid format.
func MustParseDate(s string) Date {
	d, err := ParseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}

// String returns the date in YYYY-MM-DD format.
func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return d.Time.Format(DateFormat)
}

// IsZero returns true if the date is the zero value.
func (d Date) IsZero() bool {
	return d.Time.IsZero()
}

// MarshalXML implements xml.Marshaler for Date.
func (d Date) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d.IsZero() {
		return nil
	}
	return e.EncodeElement(d.String(), start)
}

// UnmarshalXML implements xml.Unmarshaler for Date.
func (d *Date) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := dec.DecodeElement(&s, &start); err != nil {
		return err
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// MarshalXMLAttr implements xml.MarshalerAttr for Date.
func (d Date) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: d.String()}, nil
}

// UnmarshalXMLAttr implements xml.UnmarshalerAttr for Date.
func (d *Date) UnmarshalXMLAttr(attr xml.Attr) error {
	parsed, err := ParseDate(attr.Value)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
