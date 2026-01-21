package types

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

// uuidPattern validates UUID format per ISDOC XSD UUIDType.
var uuidPattern = regexp.MustCompile(`^[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$`)

// UUID represents an ISDOC UUID (GUID) with format validation.
type UUID string

// NewUUID creates a UUID from a string, validating the format.
func NewUUID(s string) (UUID, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	if !uuidPattern.MatchString(s) {
		return "", fmt.Errorf("invalid UUID format: %q", s)
	}
	return UUID(s), nil
}

// MustUUID creates a UUID from a string, panicking on invalid format.
func MustUUID(s string) UUID {
	u, err := NewUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}

// String returns the UUID as a string.
func (u UUID) String() string {
	return string(u)
}

// IsZero returns true if the UUID is empty.
func (u UUID) IsZero() bool {
	return u == ""
}

// MarshalXML implements xml.Marshaler for UUID.
func (u UUID) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(string(u), start)
}

// UnmarshalXML implements xml.Unmarshaler for UUID.
func (u *UUID) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := dec.DecodeElement(&s, &start); err != nil {
		return err
	}
	parsed, err := NewUUID(s)
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}

// MarshalXMLAttr implements xml.MarshalerAttr for UUID.
func (u UUID) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: string(u)}, nil
}

// UnmarshalXMLAttr implements xml.UnmarshalerAttr for UUID.
func (u *UUID) UnmarshalXMLAttr(attr xml.Attr) error {
	parsed, err := NewUUID(attr.Value)
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}
