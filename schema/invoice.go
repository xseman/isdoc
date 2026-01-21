// Package schema contains Go struct models for ISDOC v6.0.2.
//
// These models are based on the official ISDOC XSD schema and use
// custom types from the types package for proper validation and XML handling.
package schema

import (
	"encoding/xml"

	"github.com/xseman/isdoc/types"
)

// Namespace is the ISDOC XML namespace.
const Namespace = "http://isdoc.cz/namespace/2013"

// Invoice is the root element of an ISDOC document.
type Invoice struct {
	XMLName xml.Name `xml:"Invoice"`

	// Version is the ISDOC schema version (required attribute).
	Version string `xml:"version,attr"`

	// DocumentType specifies the type of document (1-7).
	DocumentType int `xml:"DocumentType"`

	// SubDocumentType is an optional document subtype.
	SubDocumentType string `xml:"SubDocumentType,omitempty"`

	// SubDocumentTypeOrigin identifies the maintainer of the subtype codelist.
	SubDocumentTypeOrigin string `xml:"SubDocumentTypeOrigin,omitempty"`

	// TargetConsolidator identifies the target consolidator (B2C systems).
	TargetConsolidator string `xml:"TargetConsolidator,omitempty"`

	// ClientOnTargetConsolidator identifies the client in the issuer system.
	ClientOnTargetConsolidator string `xml:"ClientOnTargetConsolidator,omitempty"`

	// ClientBankAccount is the receiver's bank account number.
	ClientBankAccount string `xml:"ClientBankAccount,omitempty"`

	// ID is the human-readable document number.
	ID string `xml:"ID"`

	// UUID is the GUID identifier from the emitting system.
	UUID types.UUID `xml:"UUID"`

	// EgovFlag is the state-governed document flag.
	EgovFlag types.Bool `xml:"EgovFlag,omitempty"`

	// ISDS_ID is the unique identifier in the ISDS system.
	ISDS_ID string `xml:"ISDS_ID,omitempty"`

	// FileReference is the file number from the issuer's records.
	FileReference string `xml:"FileReference,omitempty"`

	// ReferenceNumber is the reference number from the issuer's records.
	ReferenceNumber string `xml:"ReferenceNumber,omitempty"`

	// EgovClassifiers is a collection of document classifiers.
	EgovClassifiers *EgovClassifiers `xml:"EgovClassifiers,omitempty"`

	// IssuingSystem identifies the system generating the invoice.
	IssuingSystem string `xml:"IssuingSystem,omitempty"`

	// IssueDate is the document issue date.
	IssueDate types.Date `xml:"IssueDate"`

	// TaxPointDate is the tax point date.
	TaxPointDate types.Date `xml:"TaxPointDate,omitempty"`

	// VATApplicable indicates whether VAT is applicable.
	VATApplicable types.Bool `xml:"VATApplicable"`

	// ElectronicPossibilityAgreementReference references the agreement for electronic invoicing.
	ElectronicPossibilityAgreementReference Note `xml:"ElectronicPossibilityAgreementReference"`

	// Note is an optional document note.
	Note *Note `xml:"Note,omitempty"`

	// LocalCurrencyCode is the local currency code (e.g., "CZK").
	LocalCurrencyCode string `xml:"LocalCurrencyCode"`

	// ForeignCurrencyCode is the foreign currency code.
	ForeignCurrencyCode string `xml:"ForeignCurrencyCode,omitempty"`

	// CurrRate is the foreign currency exchange rate (or 1 if not used).
	CurrRate types.Decimal `xml:"CurrRate"`

	// RefCurrRate is the reference currency rate (usually 1).
	RefCurrRate types.Decimal `xml:"RefCurrRate"`

	// Extensions contains arbitrary user-defined elements.
	Extensions *Extensions `xml:"Extensions,omitempty"`

	// AccountingSupplierParty is the supplier/accounting entity.
	AccountingSupplierParty AccountingSupplierParty `xml:"AccountingSupplierParty"`

	// SellerSupplierParty is the supplier's invoicing address.
	SellerSupplierParty *SellerSupplierParty `xml:"SellerSupplierParty,omitempty"`

	// AnonymousCustomerParty is for simplified tax documents.
	AnonymousCustomerParty *AnonymousCustomerParty `xml:"AnonymousCustomerParty,omitempty"`

	// AccountingCustomerParty is the customer/accounting entity.
	AccountingCustomerParty *AccountingCustomerParty `xml:"AccountingCustomerParty,omitempty"`

	// BuyerCustomerParty is the purchaser's invoicing address.
	BuyerCustomerParty *BuyerCustomerParty `xml:"BuyerCustomerParty,omitempty"`

	// OrderReferences is a collection of referenced purchase orders.
	OrderReferences *OrderReferences `xml:"OrderReferences,omitempty"`

	// DeliveryNoteReferences is a collection of referenced delivery notes.
	DeliveryNoteReferences *DeliveryNoteReferences `xml:"DeliveryNoteReferences,omitempty"`

	// OriginalDocumentReferences is a collection of referenced original documents.
	OriginalDocumentReferences *OriginalDocumentReferences `xml:"OriginalDocumentReferences,omitempty"`

	// ContractReferences is a collection of related contracts.
	ContractReferences *ContractReferences `xml:"ContractReferences,omitempty"`

	// Delivery contains delivery information.
	Delivery *Delivery `xml:"Delivery,omitempty"`

	// InvoiceLines is the collection of invoice line items.
	InvoiceLines InvoiceLines `xml:"InvoiceLines"`

	// NonTaxedDeposits is a collection of proforma invoices (without VAT).
	NonTaxedDeposits *NonTaxedDeposits `xml:"NonTaxedDeposits,omitempty"`

	// TaxedDeposits is a collection of taxed deposits.
	TaxedDeposits *TaxedDeposits `xml:"TaxedDeposits,omitempty"`

	// TaxTotal contains tax recapitulation.
	TaxTotal TaxTotal `xml:"TaxTotal"`

	// LegalMonetaryTotal contains document totals.
	LegalMonetaryTotal LegalMonetaryTotal `xml:"LegalMonetaryTotal"`

	// PaymentMeans contains payment information.
	PaymentMeans *PaymentMeans `xml:"PaymentMeans,omitempty"`

	// SupplementsList is a collection of document attachments.
	SupplementsList *SupplementsList `xml:"SupplementsList,omitempty"`
}

// Note represents a text note with optional language identifier.
type Note struct {
	Value      string `xml:",chardata"`
	LanguageID string `xml:"languageID,attr,omitempty"`
}

// Extensions contains arbitrary user-defined XML elements.
type Extensions struct {
	Raw []byte `xml:",innerxml"`
}

// EgovClassifiers is a collection of document classifiers.
type EgovClassifiers struct {
	EgovClassifier []EgovClassifier `xml:"EgovClassifier"`
}

// EgovClassifier represents a document classifier.
type EgovClassifier struct {
	Value string `xml:",chardata"`
}
