package types

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// Bool represents an ISDOC boolean that only accepts "true" or "false" literals.
// Unlike standard Go bool, it rejects "0", "1", "True", "False", etc.
type Bool bool

// ParseBool parses an ISDOC boolean string.
// Only "true" and "false" (lowercase) are accepted per XSD BooleanType.
func ParseBool(s string) (Bool, error) {
	switch s {
	case "true":
		return Bool(true), nil
	case "false":
		return Bool(false), nil
	case "":
		return Bool(false), nil
	default:
		return Bool(false), fmt.Errorf("invalid boolean %q: must be 'true' or 'false'", s)
	}
}

// MustParseBool parses a boolean string, panicking on invalid format.
func MustParseBool(s string) Bool {
	b, err := ParseBool(s)
	if err != nil {
		panic(err)
	}
	return b
}

// String returns "true" or "false".
func (b Bool) String() string {
	if b {
		return "true"
	}
	return "false"
}

// Bool returns the underlying bool value.
func (b Bool) Bool() bool {
	return bool(b)
}

// MarshalXML implements xml.Marshaler for Bool.
func (b Bool) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(b.String(), start)
}

// UnmarshalXML implements xml.Unmarshaler for Bool.
func (b *Bool) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := dec.DecodeElement(&s, &start); err != nil {
		return err
	}
	s = strings.TrimSpace(s)
	parsed, err := ParseBool(s)
	if err != nil {
		return err
	}
	*b = parsed
	return nil
}

// MarshalXMLAttr implements xml.MarshalerAttr for Bool.
func (b Bool) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: b.String()}, nil
}

// UnmarshalXMLAttr implements xml.UnmarshalerAttr for Bool.
func (b *Bool) UnmarshalXMLAttr(attr xml.Attr) error {
	parsed, err := ParseBool(strings.TrimSpace(attr.Value))
	if err != nil {
		return err
	}
	*b = parsed
	return nil
}
