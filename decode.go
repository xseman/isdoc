package isdoc

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/xseman/isdoc/schema"
)

// Decoder decodes ISDOC XML documents.
type Decoder struct {
	reader io.Reader
}

// NewDecoder creates a new Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{reader: r}
}

// Decode decodes an ISDOC XML document and returns the Invoice.
// It performs two passes:
// 1. XML unmarshaling to populate struct fields
// 2. Reference resolution for id/ref attributes
func (d *Decoder) Decode() (*schema.Invoice, error) {
	// Read all data
	data, err := io.ReadAll(d.reader)
	if err != nil {
		return nil, NewDecodeError("", fmt.Errorf("reading input: %w", err))
	}

	return DecodeBytes(data)
}

// DecodeBytes decodes an ISDOC Invoice XML document from bytes.
//
// This is the primary function for parsing ISDOC invoices. It performs:
//  1. XML unmarshaling to populate struct fields
//  2. Reference resolution for id/ref attributes
//
// Returns (*Invoice, nil) on success, or (*Invoice, DecodeErrors) if there are
// reference resolution errors. The invoice may still be usable even with errors.
//
// Example:
//
//	data, _ := os.ReadFile("invoice.isdoc")
//	invoice, err := isdoc.DecodeBytes(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
func DecodeBytes(data []byte) (*schema.Invoice, error) {
	var invoice schema.Invoice

	// Pass 1: Unmarshal XML
	if err := xml.Unmarshal(data, &invoice); err != nil {
		return nil, NewDecodeError("", fmt.Errorf("XML parsing: %w", err))
	}

	// Pass 2: Resolve references
	if errs := resolveReferences(&invoice); len(errs) > 0 {
		// Return invoice with errors - caller can decide whether to use it
		return &invoice, errs
	}

	return &invoice, nil
}

// resolveReferences validates id/ref attribute linkages.
// Returns slice of errors for any unresolved references.
func resolveReferences(inv *schema.Invoice) DecodeErrors {
	var errs DecodeErrors

	// Build maps of header references by id
	orderRefs := make(map[string]bool)
	deliveryNoteRefs := make(map[string]bool)
	originalDocRefs := make(map[string]bool)
	contractRefs := make(map[string]bool)

	// Collect header reference IDs
	if inv.OrderReferences != nil {
		for i, ref := range inv.OrderReferences.OrderReference {
			if ref.ID != "" {
				if orderRefs[ref.ID] {
					errs = append(errs, NewDecodeError(
						fmt.Sprintf("Invoice.OrderReferences.OrderReference[%d]", i),
						fmt.Errorf("duplicate id %q", ref.ID),
					))
				}
				orderRefs[ref.ID] = true
			}
		}
	}

	if inv.DeliveryNoteReferences != nil {
		for i, ref := range inv.DeliveryNoteReferences.DeliveryNoteReference {
			if ref.ID != "" {
				if deliveryNoteRefs[ref.ID] {
					errs = append(errs, NewDecodeError(
						fmt.Sprintf("Invoice.DeliveryNoteReferences.DeliveryNoteReference[%d]", i),
						fmt.Errorf("duplicate id %q", ref.ID),
					))
				}
				deliveryNoteRefs[ref.ID] = true
			}
		}
	}

	if inv.OriginalDocumentReferences != nil {
		for i, ref := range inv.OriginalDocumentReferences.OriginalDocumentReference {
			if ref.ID != "" {
				if originalDocRefs[ref.ID] {
					errs = append(errs, NewDecodeError(
						fmt.Sprintf("Invoice.OriginalDocumentReferences.OriginalDocumentReference[%d]", i),
						fmt.Errorf("duplicate id %q", ref.ID),
					))
				}
				originalDocRefs[ref.ID] = true
			}
		}
	}

	if inv.ContractReferences != nil {
		for i, ref := range inv.ContractReferences.ContractReference {
			if ref.ID != "" {
				if contractRefs[ref.ID] {
					errs = append(errs, NewDecodeError(
						fmt.Sprintf("Invoice.ContractReferences.ContractReference[%d]", i),
						fmt.Errorf("duplicate id %q", ref.ID),
					))
				}
				contractRefs[ref.ID] = true
			}
		}
	}

	// Validate line references
	for i, line := range inv.InvoiceLines.InvoiceLine {
		path := fmt.Sprintf("Invoice.InvoiceLines.InvoiceLine[%d]", i)

		if line.OrderReference != nil && line.OrderReference.Ref != "" {
			if !orderRefs[line.OrderReference.Ref] {
				errs = append(errs, NewDecodeError(
					path+".OrderReference",
					fmt.Errorf("ref %q not found in header OrderReferences", line.OrderReference.Ref),
				))
			}
		}

		if line.DeliveryNoteReference != nil && line.DeliveryNoteReference.Ref != "" {
			if !deliveryNoteRefs[line.DeliveryNoteReference.Ref] {
				errs = append(errs, NewDecodeError(
					path+".DeliveryNoteReference",
					fmt.Errorf("ref %q not found in header DeliveryNoteReferences", line.DeliveryNoteReference.Ref),
				))
			}
		}

		if line.OriginalDocumentReference != nil && line.OriginalDocumentReference.Ref != "" {
			if !originalDocRefs[line.OriginalDocumentReference.Ref] {
				errs = append(errs, NewDecodeError(
					path+".OriginalDocumentReference",
					fmt.Errorf("ref %q not found in header OriginalDocumentReferences", line.OriginalDocumentReference.Ref),
				))
			}
		}

		if line.ContractReference != nil && line.ContractReference.Ref != "" {
			if !contractRefs[line.ContractReference.Ref] {
				errs = append(errs, NewDecodeError(
					path+".ContractReference",
					fmt.Errorf("ref %q not found in header ContractReferences", line.ContractReference.Ref),
				))
			}
		}
	}

	return errs
}

// DecodeCommonDocumentBytes decodes an ISDOC CommonDocument XML from bytes.
//
// CommonDocument is used for non-payment documents like contracts and certificates.
// Unlike Invoice, CommonDocument has no id/ref references, so decoding is simpler.
//
// Example:
//
//	data, _ := os.ReadFile("contract.isdoc")
//	doc, err := isdoc.DecodeCommonDocumentBytes(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
func DecodeCommonDocumentBytes(data []byte) (*schema.CommonDocument, error) {
	var doc schema.CommonDocument

	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, NewDecodeError("", fmt.Errorf("XML parsing: %w", err))
	}

	return &doc, nil
}
