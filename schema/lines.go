package schema

import "github.com/xseman/isdoc/types"

// InvoiceLines is a collection of invoice line items.
type InvoiceLines struct {
	InvoiceLine []InvoiceLine `xml:"InvoiceLine"`
}

// InvoiceLine represents a single line item on the invoice.
type InvoiceLine struct {
	// ID is the line identifier (max 36 chars).
	ID string `xml:"ID"`

	// OrderReference links to a header OrderReference.
	OrderReference *OrderLineReference `xml:"OrderReference,omitempty"`

	// DeliveryNoteReference links to a header DeliveryNoteReference.
	DeliveryNoteReference *DeliveryNoteLineReference `xml:"DeliveryNoteReference,omitempty"`

	// OriginalDocumentReference links to a header OriginalDocumentReference.
	OriginalDocumentReference *OriginalDocumentLineReference `xml:"OriginalDocumentReference,omitempty"`

	// ContractReference links to a header ContractReference.
	ContractReference *ContractLineReference `xml:"ContractReference,omitempty"`

	// EgovClassifier is an optional document classifier.
	EgovClassifier string `xml:"EgovClassifier,omitempty"`

	// InvoicedQuantity is the quantity with optional unit code.
	InvoicedQuantity Quantity `xml:"InvoicedQuantity"`

	// LineExtensionAmountCurr is the line amount in foreign currency.
	LineExtensionAmountCurr types.Decimal `xml:"LineExtensionAmountCurr,omitempty"`

	// LineExtensionAmount is the line amount (without tax).
	LineExtensionAmount types.Decimal `xml:"LineExtensionAmount"`

	// LineExtensionAmountBeforeDiscount is the amount before discount.
	LineExtensionAmountBeforeDiscount types.Decimal `xml:"LineExtensionAmountBeforeDiscount,omitempty"`

	// LineExtensionAmountTaxInclusiveCurr is the tax-inclusive amount in foreign currency.
	LineExtensionAmountTaxInclusiveCurr types.Decimal `xml:"LineExtensionAmountTaxInclusiveCurr,omitempty"`

	// LineExtensionAmountTaxInclusive is the tax-inclusive line amount.
	LineExtensionAmountTaxInclusive types.Decimal `xml:"LineExtensionAmountTaxInclusive"`

	// LineExtensionAmountTaxInclusiveBeforeDiscount is the tax-inclusive amount before discount.
	LineExtensionAmountTaxInclusiveBeforeDiscount types.Decimal `xml:"LineExtensionAmountTaxInclusiveBeforeDiscount,omitempty"`

	// LineExtensionTaxAmount is the tax amount for this line.
	LineExtensionTaxAmount types.Decimal `xml:"LineExtensionTaxAmount"`

	// UnitPrice is the price per unit (without tax).
	UnitPrice types.Decimal `xml:"UnitPrice"`

	// UnitPriceTaxInclusive is the price per unit (with tax).
	UnitPriceTaxInclusive types.Decimal `xml:"UnitPriceTaxInclusive"`

	// ClassifiedTaxCategory contains tax category information.
	ClassifiedTaxCategory ClassifiedTaxCategory `xml:"ClassifiedTaxCategory"`

	// Note is an optional line note.
	Note string `xml:"Note,omitempty"`

	// VATNote is an optional VAT-related note.
	VATNote string `xml:"VATNote,omitempty"`

	// Item contains item details.
	Item Item `xml:"Item"`

	// Extensions contains arbitrary user-defined elements.
	Extensions *Extensions `xml:"Extensions,omitempty"`
}

// Quantity represents a quantity with optional unit code.
type Quantity struct {
	Value    types.Decimal `xml:",chardata"`
	UnitCode string        `xml:"unitCode,attr,omitempty"`
}

// ClassifiedTaxCategory contains tax category information for a line.
type ClassifiedTaxCategory struct {
	// Percent is the VAT rate percentage.
	Percent types.Decimal `xml:"Percent"`

	// VATCalculationMethod is 0 (from bottom) or 1 (from top).
	VATCalculationMethod int `xml:"VATCalculationMethod,omitempty"`

	// VATApplicable indicates whether VAT is applicable.
	VATApplicable types.Bool `xml:"VATApplicable,omitempty"`

	// LocalReverseCharge contains reverse charge information.
	LocalReverseCharge *LocalReverseCharge `xml:"LocalReverseCharge,omitempty"`
}

// LocalReverseCharge contains local reverse charge information.
type LocalReverseCharge struct {
	LocalReverseChargeCode     string        `xml:"LocalReverseChargeCode"`
	LocalReverseChargeQuantity types.Decimal `xml:"LocalReverseChargeQuantity,omitempty"`
}

// Item contains item details.
type Item struct {
	// Description is the item description.
	Description string `xml:"Description"`

	// CatalogueItemIdentification contains catalog identifiers.
	CatalogueItemIdentification *ItemIdentification `xml:"CatalogueItemIdentification,omitempty"`

	// SellersItemIdentification is the seller's item identifier.
	SellersItemIdentification *ItemIdentification `xml:"SellersItemIdentification,omitempty"`

	// SecondarySellersItemIdentification is a secondary seller's identifier.
	SecondarySellersItemIdentification *ItemIdentification `xml:"SecondarySellersItemIdentification,omitempty"`

	// TertiarySellersItemIdentification is a tertiary seller's identifier.
	TertiarySellersItemIdentification *ItemIdentification `xml:"TertiarySellersItemIdentification,omitempty"`

	// BuyersItemIdentification is the buyer's item identifier.
	BuyersItemIdentification *ItemIdentification `xml:"BuyersItemIdentification,omitempty"`

	// StoreBatches contains batch/serial number information.
	StoreBatches *StoreBatches `xml:"StoreBatches,omitempty"`
}

// ItemIdentification contains an item identifier.
type ItemIdentification struct {
	ID string `xml:"ID"`
}

// StoreBatches is a collection of store batches.
type StoreBatches struct {
	StoreBatch []StoreBatch `xml:"StoreBatch"`
}

// StoreBatch contains batch/serial number information.
type StoreBatch struct {
	Name                string     `xml:"Name,omitempty"`
	Note                string     `xml:"Note,omitempty"`
	ExpirationDate      types.Date `xml:"ExpirationDate,omitempty"`
	Specification       string     `xml:"Specification,omitempty"`
	Quantity            Quantity   `xml:"Quantity,omitempty"`
	BatchOrSerialNumber string     `xml:"BatchOrSerialNumber,omitempty"` // "B" or "S"
	SealSeriesID        string     `xml:"SealSeriesID,omitempty"`
}
