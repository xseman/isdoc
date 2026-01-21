package schema

import "github.com/xseman/isdoc/types"

// PaymentMeans contains payment information.
type PaymentMeans struct {
	// Payment contains individual payment details.
	Payment []Payment `xml:"Payment"`

	// AlternateBankAccounts contains alternative bank accounts.
	AlternateBankAccounts *AlternateBankAccounts `xml:"AlternateBankAccounts,omitempty"`
}

// Payment contains payment details.
type Payment struct {
	// PaidAmount is the amount paid.
	PaidAmount types.Decimal `xml:"PaidAmount"`

	// PaymentMeansCode is the payment method code.
	// Values: 10 (cash), 20 (cheque), 31 (credit transfer), 42 (transfer),
	//         48 (card), 49 (direct debit), 50 (postgiro), 97 (composition)
	PaymentMeansCode int `xml:"PaymentMeansCode"`

	// Details contains payment details.
	Details *PaymentDetails `xml:"Details,omitempty"`
}

// PaymentDetails contains detailed payment information.
type PaymentDetails struct {
	// DocumentID is the payment document ID.
	DocumentID string `xml:"DocumentID,omitempty"`

	// IssueDate is the payment document issue date.
	IssueDate types.Date `xml:"IssueDate,omitempty"`

	// PaymentDueDate is the payment due date.
	PaymentDueDate types.Date `xml:"PaymentDueDate,omitempty"`

	// VariableSymbol is the variable symbol for payment.
	VariableSymbol string `xml:"VariableSymbol,omitempty"`

	// ConstantSymbol is the constant symbol for payment.
	ConstantSymbol string `xml:"ConstantSymbol,omitempty"`

	// SpecificSymbol is the specific symbol for payment.
	SpecificSymbol string `xml:"SpecificSymbol,omitempty"`

	// BankAccount contains bank account details.
	BankAccount *BankAccount `xml:"BankAccount,omitempty"`
}

// BankAccount contains bank account information.
type BankAccount struct {
	// ID is the bank account number.
	ID string `xml:"ID"`

	// BankCode is the bank code.
	BankCode string `xml:"BankCode,omitempty"`

	// Name is the account name.
	Name string `xml:"Name,omitempty"`

	// IBAN is the international bank account number.
	IBAN string `xml:"IBAN,omitempty"`

	// BIC is the bank identifier code (SWIFT).
	BIC string `xml:"BIC,omitempty"`
}

// AlternateBankAccounts is a collection of alternative bank accounts.
type AlternateBankAccounts struct {
	AlternateBankAccount []BankAccount `xml:"AlternateBankAccount"`
}

// SupplementsList is a collection of document attachments.
type SupplementsList struct {
	Supplement []Supplement `xml:"Supplement"`
}

// Supplement represents a document attachment.
type Supplement struct {
	// Filename is the attachment filename.
	Filename string `xml:"Filename"`

	// DigestMethod is the hash algorithm used.
	DigestMethod string `xml:"DigestMethod,omitempty"`

	// DigestValue is the hash value.
	DigestValue string `xml:"DigestValue,omitempty"`

	// Preview attribute indicates if this is the document preview.
	Preview types.Bool `xml:"preview,attr,omitempty"`
}
