package schema

import (
	"encoding/xml"

	"github.com/xseman/isdoc/types"
)

// CommonDocument is the root element of an ISDOC non-payment document.
// It represents contracts, certificates, and other non-tax documents.
// Unlike Invoice, CommonDocument has no VAT fields, line items, or payment info.
type CommonDocument struct {
	XMLName xml.Name `xml:"CommonDocument"`

	// Version is the ISDOC schema version (required attribute).
	Version string `xml:"version,attr"`

	// SubDocumentType is the document subtype (required).
	// Codelist is maintained by entity specified in SubDocumentTypeOrigin.
	SubDocumentType string `xml:"SubDocumentType"`

	// SubDocumentTypeOrigin identifies the maintainer of the subtype codelist (required).
	SubDocumentTypeOrigin string `xml:"SubDocumentTypeOrigin"`

	// TargetConsolidator identifies the target consolidator (B2C systems).
	TargetConsolidator string `xml:"TargetConsolidator,omitempty"`

	// ClientOnTargetConsolidator identifies the client in the issuer system.
	ClientOnTargetConsolidator string `xml:"ClientOnTargetConsolidator,omitempty"`

	// ClientBankAccount is the receiver's bank account number.
	ClientBankAccount string `xml:"ClientBankAccount,omitempty"`

	// ID is the human-readable document number (required).
	ID string `xml:"ID"`

	// UUID is the GUID identifier from the emitting system (required).
	UUID types.UUID `xml:"UUID"`

	// IssueDate is the document issue date (required).
	IssueDate types.Date `xml:"IssueDate"`

	// LastValidDate is the date until the document is valid.
	LastValidDate types.Date `xml:"LastValidDate,omitempty"`

	// Note is an optional document note.
	Note *Note `xml:"Note,omitempty"`

	// Extensions contains arbitrary user-defined elements.
	Extensions *Extensions `xml:"Extensions,omitempty"`

	// AccountingSupplierParty is the supplier/accounting entity (required).
	AccountingSupplierParty AccountingSupplierParty `xml:"AccountingSupplierParty"`

	// AccountingCustomerParty is the customer/accounting entity (required).
	AccountingCustomerParty AccountingCustomerParty `xml:"AccountingCustomerParty"`

	// SupplementsList is a collection of document attachments.
	SupplementsList *SupplementsList `xml:"SupplementsList,omitempty"`
}
