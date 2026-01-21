package schema

import "github.com/xseman/isdoc/types"

// AccountingSupplierParty is the supplier/accounting entity.
type AccountingSupplierParty struct {
	Party Party `xml:"Party"`
}

// SellerSupplierParty is the supplier's invoicing address.
type SellerSupplierParty struct {
	Party Party `xml:"Party"`
}

// AccountingCustomerParty is the customer/accounting entity.
type AccountingCustomerParty struct {
	Party Party `xml:"Party"`
}

// BuyerCustomerParty is the purchaser's invoicing address.
type BuyerCustomerParty struct {
	Party Party `xml:"Party"`
}

// AnonymousCustomerParty is for simplified tax documents.
type AnonymousCustomerParty struct {
	ID       string `xml:"ID"`
	IDScheme string `xml:"IDScheme,omitempty"`
}

// Delivery contains delivery information.
type Delivery struct {
	Party Party `xml:"Party"`
}

// Party represents a business party (supplier, customer, etc.).
type Party struct {
	// PartyIdentification contains identifiers for the party.
	PartyIdentification PartyIdentification `xml:"PartyIdentification"`

	// PartyName contains the party's name.
	PartyName PartyName `xml:"PartyName"`

	// PostalAddress contains the party's address.
	PostalAddress PostalAddress `xml:"PostalAddress"`

	// PartyTaxScheme contains tax scheme information (can have multiple).
	PartyTaxScheme []PartyTaxScheme `xml:"PartyTaxScheme,omitempty"`

	// RegisterIdentification contains commercial register information.
	RegisterIdentification *RegisterIdentification `xml:"RegisterIdentification,omitempty"`

	// Contact contains contact information.
	Contact *Contact `xml:"Contact,omitempty"`
}

// PartyIdentification contains identifiers for a party.
type PartyIdentification struct {
	// UserID is an optional user identifier.
	UserID string `xml:"UserID,omitempty"`

	// CatalogFirmIdentification is a catalog identifier (e.g., EAN).
	CatalogFirmIdentification string `xml:"CatalogFirmIdentification,omitempty"`

	// ID is the company ID (IČO in Czech).
	ID string `xml:"ID"`
}

// PartyName contains the party's name.
type PartyName struct {
	Name string `xml:"Name"`
}

// PostalAddress contains address information.
type PostalAddress struct {
	StreetName     string  `xml:"StreetName"`
	BuildingNumber string  `xml:"BuildingNumber,omitempty"`
	CityName       string  `xml:"CityName"`
	PostalZone     string  `xml:"PostalZone"`
	Country        Country `xml:"Country"`
}

// Country contains country information.
type Country struct {
	IdentificationCode string `xml:"IdentificationCode"`
	Name               string `xml:"Name,omitempty"`
}

// PartyTaxScheme contains tax scheme information.
type PartyTaxScheme struct {
	// CompanyID is the VAT number (DIČ in Czech).
	CompanyID string `xml:"CompanyID"`

	// TaxScheme is either "VAT" or "TIN".
	TaxScheme string `xml:"TaxScheme"`
}

// RegisterIdentification contains commercial register information.
type RegisterIdentification struct {
	Preformatted    string     `xml:"Preformatted,omitempty"`
	RegisterKeptAt  string     `xml:"RegisterKeptAt,omitempty"`
	RegisterFileRef string     `xml:"RegisterFileRef,omitempty"`
	RegisterDate    types.Date `xml:"RegisterDate,omitempty"`
}

// Contact contains contact information.
type Contact struct {
	Name           string `xml:"Name,omitempty"`
	Telephone      string `xml:"Telephone,omitempty"`
	ElectronicMail string `xml:"ElectronicMail,omitempty"`
}
