package isdoc

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/xseman/isdoc/internal/ordering"
	"github.com/xseman/isdoc/schema"
)

// Encoder encodes ISDOC invoices to XML.
type Encoder struct {
	writer     io.Writer
	indent     string
	addXMLDecl bool
}

// NewEncoder creates a new Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer:     w,
		indent:     "  ",
		addXMLDecl: true,
	}
}

// SetIndent sets the indentation string. Default is two spaces.
func (e *Encoder) SetIndent(indent string) {
	e.indent = indent
}

// SetXMLDeclaration controls whether to add XML declaration. Default is true.
func (e *Encoder) SetXMLDeclaration(add bool) {
	e.addXMLDecl = add
}

// Encode encodes an invoice to XML.
func (e *Encoder) Encode(inv *schema.Invoice) error {
	var buf bytes.Buffer

	if e.addXMLDecl {
		buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	}

	// Write root element with namespace
	buf.WriteString(fmt.Sprintf(`<Invoice xmlns="%s" version="%s">`,
		schema.Namespace, inv.Version))
	buf.WriteString("\n")

	// Encode child elements in XSD order
	if err := e.encodeInvoiceContent(&buf, inv, 1); err != nil {
		return err
	}

	buf.WriteString("</Invoice>\n")

	_, err := e.writer.Write(buf.Bytes())
	return err
}

// encodeInvoiceContent encodes the content of an Invoice element.
func (e *Encoder) encodeInvoiceContent(buf *bytes.Buffer, inv *schema.Invoice, depth int) error {
	// Get element order from XSD
	order := ordering.Sequence["Invoice"]

	// Create a map of field values by XML element name
	fields := e.extractFields(reflect.ValueOf(inv).Elem())

	// Write elements in order
	for _, elemName := range order {
		if val, ok := fields[elemName]; ok {
			if err := e.encodeValue(buf, elemName, val, depth); err != nil {
				return err
			}
		}
	}

	return nil
}

// extractFields extracts field values by their XML element name.
func (e *Encoder) extractFields(v reflect.Value) map[string]reflect.Value {
	fields := make(map[string]reflect.Value)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		xmlTag := field.Tag.Get("xml")
		if xmlTag == "" || xmlTag == "-" {
			continue
		}

		// Parse XML tag: "ElementName,omitempty"
		parts := strings.Split(xmlTag, ",")
		elemName := parts[0]

		// Skip attributes and special fields
		if strings.HasPrefix(elemName, "@") || elemName == "xmlns" {
			continue
		}
		if elemName == "" {
			continue
		}

		fields[elemName] = v.Field(i)
	}

	return fields
}

// encodeValue encodes a single value as XML.
func (e *Encoder) encodeValue(buf *bytes.Buffer, name string, v reflect.Value, depth int) error {
	indent := strings.Repeat(e.indent, depth)

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	// Handle interfaces
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	// Check if value should be omitted
	if e.isZero(v) {
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" {
			return nil
		}
		buf.WriteString(indent)
		buf.WriteString("<")
		buf.WriteString(name)
		buf.WriteString(">")
		xml.EscapeText(buf, []byte(s))
		buf.WriteString("</")
		buf.WriteString(name)
		buf.WriteString(">\n")

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf.WriteString(indent)
		buf.WriteString("<")
		buf.WriteString(name)
		buf.WriteString(">")
		buf.WriteString(fmt.Sprintf("%d", v.Int()))
		buf.WriteString("</")
		buf.WriteString(name)
		buf.WriteString(">\n")

	case reflect.Bool:
		buf.WriteString(indent)
		buf.WriteString("<")
		buf.WriteString(name)
		buf.WriteString(">")
		if v.Bool() {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		buf.WriteString("</")
		buf.WriteString(name)
		buf.WriteString(">\n")

	case reflect.Slice:
		// For slices, encode each element
		for i := 0; i < v.Len(); i++ {
			if err := e.encodeValue(buf, name, v.Index(i), depth); err != nil {
				return err
			}
		}

	case reflect.Struct:
		// Check if type implements xml.Marshaler (like Date, Decimal, etc.)
		if _, ok := v.Interface().(xml.Marshaler); ok {
			// For types with custom marshaling, we need to wrap them with our element name
			buf.WriteString(indent)
			buf.WriteString("<")
			buf.WriteString(name)
			buf.WriteString(">")

			// Get the string value using Stringer interface
			if stringer, ok := v.Interface().(fmt.Stringer); ok {
				xml.EscapeText(buf, []byte(stringer.String()))
			}

			buf.WriteString("</")
			buf.WriteString(name)
			buf.WriteString(">\n")
		} else {
			// For complex structs (like AccountingSupplierParty), encode fields recursively
			buf.WriteString(indent)
			buf.WriteString("<")
			buf.WriteString(name)
			buf.WriteString(">\n")

			// Encode struct fields - get ordered fields for this type
			typeName := v.Type().Name()
			order, hasOrder := ordering.Sequence[typeName]

			fields := e.extractFields(v)

			if hasOrder {
				// Use XSD order
				for _, elemName := range order {
					if val, ok := fields[elemName]; ok {
						if err := e.encodeValue(buf, elemName, val, depth+1); err != nil {
							return err
						}
					}
				}
			} else {
				// Encode fields in struct order
				for i := 0; i < v.NumField(); i++ {
					field := v.Type().Field(i)
					xmlTag := field.Tag.Get("xml")
					if xmlTag == "" || xmlTag == "-" {
						continue
					}
					parts := strings.Split(xmlTag, ",")
					elemName := parts[0]
					if elemName == "" || strings.Contains(elemName, ",attr") {
						continue
					}
					if err := e.encodeValue(buf, elemName, v.Field(i), depth+1); err != nil {
						return err
					}
				}
			}

			buf.WriteString(indent)
			buf.WriteString("</")
			buf.WriteString(name)
			buf.WriteString(">\n")
		}

	default:
		// For other types (like custom types), try String() method first
		if stringer, ok := v.Interface().(fmt.Stringer); ok {
			s := stringer.String()
			if s == "" {
				return nil
			}
			buf.WriteString(indent)
			buf.WriteString("<")
			buf.WriteString(name)
			buf.WriteString(">")
			xml.EscapeText(buf, []byte(s))
			buf.WriteString("</")
			buf.WriteString(name)
			buf.WriteString(">\n")
		} else {
			// Fall back to xml.Marshal
			data, err := xml.MarshalIndent(v.Interface(), strings.Repeat(e.indent, depth-1), e.indent)
			if err != nil {
				return fmt.Errorf("encoding %s: %w", name, err)
			}
			buf.Write(data)
			buf.WriteString("\n")
		}
	}

	return nil
}

// isZero checks if a value is the zero value for its type.
func (e *Encoder) isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return false // Always emit booleans
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Struct:
		// Check if struct has IsZero method
		if method := v.MethodByName("IsZero"); method.IsValid() {
			result := method.Call(nil)
			if len(result) == 1 && result[0].Kind() == reflect.Bool {
				return result[0].Bool()
			}
		}
		return false
	default:
		return false
	}
}

// EncodeBytes encodes an ISDOC Invoice to XML and returns the bytes.
//
// The encoding maintains proper element ordering according to the XSD schema
// and includes the XML declaration. The output is formatted with 2-space indentation.
//
// Example:
//
//	invoice := &schema.Invoice{
//	    Version: "6.0.2",
//	    DocumentType: 1,
//	    ID: "FV-2025-001",
//	    // ... other fields
//	}
//	xmlData, err := isdoc.EncodeBytes(invoice)
//	if err != nil {
//	    log.Fatal(err)
//	}
func EncodeBytes(inv *schema.Invoice) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(inv); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeCommonDocument encodes a CommonDocument to XML.
func (e *Encoder) EncodeCommonDocument(doc *schema.CommonDocument) error {
	var buf bytes.Buffer

	if e.addXMLDecl {
		buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	}

	// Write root element with namespace
	buf.WriteString(fmt.Sprintf(`<CommonDocument xmlns="%s" version="%s">`,
		schema.Namespace, doc.Version))
	buf.WriteString("\n")

	// Encode child elements in XSD order
	if err := e.encodeCommonDocumentContent(&buf, doc, 1); err != nil {
		return err
	}

	buf.WriteString("</CommonDocument>\n")

	_, err := e.writer.Write(buf.Bytes())
	return err
}

// encodeCommonDocumentContent encodes the content of a CommonDocument element.
func (e *Encoder) encodeCommonDocumentContent(buf *bytes.Buffer, doc *schema.CommonDocument, depth int) error {
	// Get element order from XSD
	order := ordering.Sequence["CommonDocument"]

	// Create a map of field values by XML element name
	fields := e.extractFields(reflect.ValueOf(doc).Elem())

	// Write elements in order
	for _, elemName := range order {
		if val, ok := fields[elemName]; ok {
			if err := e.encodeValue(buf, elemName, val, depth); err != nil {
				return err
			}
		}
	}

	return nil
}

// EncodeCommonDocumentBytes encodes an ISDOC CommonDocument to XML and returns the bytes.
//
// CommonDocument is used for non-payment documents like contracts and certificates.
// The encoding maintains proper element ordering and includes the XML declaration.
//
// Example:
//
//	doc := &schema.CommonDocument{
//	    Version: "6.0.2",
//	    SubDocumentType: "CONTRACT",
//	    ID: "DOC-2025-001",
//	    // ... other fields
//	}
//	xmlData, err := isdoc.EncodeCommonDocumentBytes(doc)
//	if err != nil {
//	    log.Fatal(err)
//	}
func EncodeCommonDocumentBytes(doc *schema.CommonDocument) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.EncodeCommonDocument(doc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
