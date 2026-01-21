package schema

import "github.com/xseman/isdoc/types"

// OrderReferences is a collection of referenced purchase orders.
type OrderReferences struct {
	OrderReference []OrderReference `xml:"OrderReference"`
}

// OrderReference contains information about a referenced purchase order.
type OrderReference struct {
	// ID attribute for reference linking.
	ID string `xml:"id,attr,omitempty"`

	// SalesOrderID is the internal order identifier at the supplier.
	SalesOrderID string `xml:"SalesOrderID"`

	// ExternalOrderID is the external order number (from the buyer).
	ExternalOrderID string `xml:"ExternalOrderID,omitempty"`

	// IssueDate is the order issue date at the supplier.
	IssueDate types.Date `xml:"IssueDate,omitempty"`

	// ExternalOrderIssueDate is the order issue date at the buyer.
	ExternalOrderIssueDate types.Date `xml:"ExternalOrderIssueDate,omitempty"`

	// UUID is the unique GUID identifier.
	UUID types.UUID `xml:"UUID,omitempty"`

	// ISDS_ID is the ISDS message ID.
	ISDS_ID string `xml:"ISDS_ID,omitempty"`

	// FileReference is the file number.
	FileReference string `xml:"FileReference,omitempty"`

	// ReferenceNumber is the reference number.
	ReferenceNumber string `xml:"ReferenceNumber,omitempty"`
}

// OrderLineReference references a line on a purchase order.
type OrderLineReference struct {
	// Ref attribute linking to OrderReference.id.
	Ref string `xml:"ref,attr,omitempty"`

	// LineID is the line number on the order.
	LineID string `xml:"LineID,omitempty"`
}

// DeliveryNoteReferences is a collection of referenced delivery notes.
type DeliveryNoteReferences struct {
	DeliveryNoteReference []DeliveryNoteReference `xml:"DeliveryNoteReference"`
}

// DeliveryNoteReference contains information about a referenced delivery note.
type DeliveryNoteReference struct {
	// ID attribute for reference linking.
	ID string `xml:"id,attr,omitempty"`

	// ID is the delivery note identifier.
	DeliveryNoteID string `xml:"ID"`

	// IssueDate is the delivery note issue date.
	IssueDate types.Date `xml:"IssueDate,omitempty"`

	// UUID is the unique GUID identifier.
	UUID types.UUID `xml:"UUID,omitempty"`
}

// DeliveryNoteLineReference references a line on a delivery note.
type DeliveryNoteLineReference struct {
	// Ref attribute linking to DeliveryNoteReference.id.
	Ref string `xml:"ref,attr,omitempty"`

	// LineID is the line number on the delivery note.
	LineID string `xml:"LineID,omitempty"`
}

// OriginalDocumentReferences is a collection of referenced original documents.
type OriginalDocumentReferences struct {
	OriginalDocumentReference []OriginalDocumentReference `xml:"OriginalDocumentReference"`
}

// OriginalDocumentReference contains information about a referenced original document.
type OriginalDocumentReference struct {
	// ID attribute for reference linking.
	ID string `xml:"id,attr,omitempty"`

	// OriginalDocumentID is the original document identifier.
	OriginalDocumentID string `xml:"ID"`

	// IssueDate is the original document issue date.
	IssueDate types.Date `xml:"IssueDate,omitempty"`

	// UUID is the unique GUID identifier.
	UUID types.UUID `xml:"UUID,omitempty"`
}

// OriginalDocumentLineReference references a line on an original document.
type OriginalDocumentLineReference struct {
	// Ref attribute linking to OriginalDocumentReference.id.
	Ref string `xml:"ref,attr,omitempty"`

	// LineID is the line number on the original document.
	LineID string `xml:"LineID,omitempty"`
}

// ContractReferences is a collection of related contracts.
type ContractReferences struct {
	ContractReference []ContractReference `xml:"ContractReference"`
}

// ContractReference contains information about a related contract.
type ContractReference struct {
	// ID attribute for reference linking.
	ID string `xml:"id,attr,omitempty"`

	// ContractID is the human-readable contract identifier.
	ContractID string `xml:"ID"`

	// UUID is the unique GUID identifier.
	UUID types.UUID `xml:"UUID,omitempty"`

	// IssueDate is the contract signature date.
	IssueDate types.Date `xml:"IssueDate"`

	// LastValidDate is the contract end date.
	LastValidDate types.Date `xml:"LastValidDate,omitempty"`

	// LastValidDateUnbounded indicates an indefinite contract.
	LastValidDateUnbounded types.Bool `xml:"LastValidDateUnbounded,omitempty"`

	// ISDS_ID is the ISDS message ID.
	ISDS_ID string `xml:"ISDS_ID,omitempty"`

	// FileReference is the file number.
	FileReference string `xml:"FileReference,omitempty"`

	// ReferenceNumber is the reference number.
	ReferenceNumber string `xml:"ReferenceNumber,omitempty"`
}

// ContractLineReference references a related contract.
type ContractLineReference struct {
	// Ref attribute linking to ContractReference.id.
	Ref string `xml:"ref,attr,omitempty"`

	// ParagraphID is the contract paragraph identifier.
	ParagraphID string `xml:"ParagraphID,omitempty"`
}
