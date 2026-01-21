package schema

import "github.com/xseman/isdoc/types"

// TaxTotal contains tax recapitulation.
type TaxTotal struct {
	// TaxSubTotal contains breakdown by tax category.
	TaxSubTotal []TaxSubTotal `xml:"TaxSubTotal"`

	// TaxAmountCurr is the total tax amount in foreign currency.
	TaxAmountCurr types.Decimal `xml:"TaxAmountCurr,omitempty"`

	// TaxAmount is the total tax amount.
	TaxAmount types.Decimal `xml:"TaxAmount"`
}

// TaxSubTotal contains tax breakdown for a specific rate.
type TaxSubTotal struct {
	// TaxableAmountCurr is the taxable amount in foreign currency.
	TaxableAmountCurr types.Decimal `xml:"TaxableAmountCurr,omitempty"`

	// TaxableAmount is the taxable amount.
	TaxableAmount types.Decimal `xml:"TaxableAmount"`

	// TaxAmountCurr is the tax amount in foreign currency.
	TaxAmountCurr types.Decimal `xml:"TaxAmountCurr,omitempty"`

	// TaxAmount is the tax amount for this category.
	TaxAmount types.Decimal `xml:"TaxAmount"`

	// TaxInclusiveAmountCurr is the tax-inclusive amount in foreign currency.
	TaxInclusiveAmountCurr types.Decimal `xml:"TaxInclusiveAmountCurr,omitempty"`

	// TaxInclusiveAmount is the tax-inclusive amount.
	TaxInclusiveAmount types.Decimal `xml:"TaxInclusiveAmount"`

	// AlreadyClaimedTaxableAmountCurr (for credit notes).
	AlreadyClaimedTaxableAmountCurr types.Decimal `xml:"AlreadyClaimedTaxableAmountCurr,omitempty"`

	// AlreadyClaimedTaxableAmount (for credit notes).
	AlreadyClaimedTaxableAmount types.Decimal `xml:"AlreadyClaimedTaxableAmount,omitempty"`

	// AlreadyClaimedTaxAmountCurr (for credit notes).
	AlreadyClaimedTaxAmountCurr types.Decimal `xml:"AlreadyClaimedTaxAmountCurr,omitempty"`

	// AlreadyClaimedTaxAmount (for credit notes).
	AlreadyClaimedTaxAmount types.Decimal `xml:"AlreadyClaimedTaxAmount,omitempty"`

	// AlreadyClaimedTaxInclusiveAmountCurr (for credit notes).
	AlreadyClaimedTaxInclusiveAmountCurr types.Decimal `xml:"AlreadyClaimedTaxInclusiveAmountCurr,omitempty"`

	// AlreadyClaimedTaxInclusiveAmount (for credit notes).
	AlreadyClaimedTaxInclusiveAmount types.Decimal `xml:"AlreadyClaimedTaxInclusiveAmount,omitempty"`

	// DifferenceTaxableAmountCurr (for credit notes).
	DifferenceTaxableAmountCurr types.Decimal `xml:"DifferenceTaxableAmountCurr,omitempty"`

	// DifferenceTaxableAmount (for credit notes).
	DifferenceTaxableAmount types.Decimal `xml:"DifferenceTaxableAmount,omitempty"`

	// DifferenceTaxAmountCurr (for credit notes).
	DifferenceTaxAmountCurr types.Decimal `xml:"DifferenceTaxAmountCurr,omitempty"`

	// DifferenceTaxAmount (for credit notes).
	DifferenceTaxAmount types.Decimal `xml:"DifferenceTaxAmount,omitempty"`

	// DifferenceTaxInclusiveAmountCurr (for credit notes).
	DifferenceTaxInclusiveAmountCurr types.Decimal `xml:"DifferenceTaxInclusiveAmountCurr,omitempty"`

	// DifferenceTaxInclusiveAmount (for credit notes).
	DifferenceTaxInclusiveAmount types.Decimal `xml:"DifferenceTaxInclusiveAmount,omitempty"`

	// TaxCategory contains the tax category information.
	TaxCategory TaxCategory `xml:"TaxCategory"`
}

// TaxCategory contains tax category information in tax totals.
type TaxCategory struct {
	// Percent is the VAT rate percentage.
	Percent types.Decimal `xml:"Percent"`

	// TaxScheme is the tax scheme (typically "VAT").
	TaxScheme string `xml:"TaxScheme,omitempty"`

	// VATApplicable indicates whether VAT is applicable.
	VATApplicable types.Bool `xml:"VATApplicable,omitempty"`

	// LocalReverseChargeFlag indicates local reverse charge.
	LocalReverseChargeFlag types.Bool `xml:"LocalReverseChargeFlag,omitempty"`
}

// LegalMonetaryTotal contains document totals.
type LegalMonetaryTotal struct {
	// TaxExclusiveAmount is the total without tax.
	TaxExclusiveAmount types.Decimal `xml:"TaxExclusiveAmount"`

	// TaxExclusiveAmountCurr is the total without tax in foreign currency.
	TaxExclusiveAmountCurr types.Decimal `xml:"TaxExclusiveAmountCurr,omitempty"`

	// TaxInclusiveAmount is the total with tax.
	TaxInclusiveAmount types.Decimal `xml:"TaxInclusiveAmount"`

	// TaxInclusiveAmountCurr is the total with tax in foreign currency.
	TaxInclusiveAmountCurr types.Decimal `xml:"TaxInclusiveAmountCurr,omitempty"`

	// AlreadyClaimedTaxExclusiveAmount (for credit notes).
	AlreadyClaimedTaxExclusiveAmount types.Decimal `xml:"AlreadyClaimedTaxExclusiveAmount,omitempty"`

	// AlreadyClaimedTaxExclusiveAmountCurr (for credit notes).
	AlreadyClaimedTaxExclusiveAmountCurr types.Decimal `xml:"AlreadyClaimedTaxExclusiveAmountCurr,omitempty"`

	// AlreadyClaimedTaxInclusiveAmount (for credit notes).
	AlreadyClaimedTaxInclusiveAmount types.Decimal `xml:"AlreadyClaimedTaxInclusiveAmount,omitempty"`

	// AlreadyClaimedTaxInclusiveAmountCurr (for credit notes).
	AlreadyClaimedTaxInclusiveAmountCurr types.Decimal `xml:"AlreadyClaimedTaxInclusiveAmountCurr,omitempty"`

	// DifferenceTaxExclusiveAmount (for credit notes).
	DifferenceTaxExclusiveAmount types.Decimal `xml:"DifferenceTaxExclusiveAmount,omitempty"`

	// DifferenceTaxExclusiveAmountCurr (for credit notes).
	DifferenceTaxExclusiveAmountCurr types.Decimal `xml:"DifferenceTaxExclusiveAmountCurr,omitempty"`

	// DifferenceTaxInclusiveAmount (for credit notes).
	DifferenceTaxInclusiveAmount types.Decimal `xml:"DifferenceTaxInclusiveAmount,omitempty"`

	// DifferenceTaxInclusiveAmountCurr (for credit notes).
	DifferenceTaxInclusiveAmountCurr types.Decimal `xml:"DifferenceTaxInclusiveAmountCurr,omitempty"`

	// PayableRoundingAmount is the rounding amount.
	PayableRoundingAmount types.Decimal `xml:"PayableRoundingAmount,omitempty"`

	// PayableRoundingAmountCurr is the rounding amount in foreign currency.
	PayableRoundingAmountCurr types.Decimal `xml:"PayableRoundingAmountCurr,omitempty"`

	// PaidDepositsAmount is the amount of paid deposits.
	PaidDepositsAmount types.Decimal `xml:"PaidDepositsAmount,omitempty"`

	// PaidDepositsAmountCurr is the amount of paid deposits in foreign currency.
	PaidDepositsAmountCurr types.Decimal `xml:"PaidDepositsAmountCurr,omitempty"`

	// PayableAmount is the final payable amount.
	PayableAmount types.Decimal `xml:"PayableAmount"`

	// PayableAmountCurr is the final payable amount in foreign currency.
	PayableAmountCurr types.Decimal `xml:"PayableAmountCurr,omitempty"`
}

// NonTaxedDeposits is a collection of proforma invoices (without VAT).
type NonTaxedDeposits struct {
	NonTaxedDeposit []NonTaxedDeposit `xml:"NonTaxedDeposit"`
}

// NonTaxedDeposit represents a proforma invoice deposit.
type NonTaxedDeposit struct {
	ID                string        `xml:"ID"`
	VariableSymbol    string        `xml:"VariableSymbol,omitempty"`
	DepositAmountCurr types.Decimal `xml:"DepositAmountCurr,omitempty"`
	DepositAmount     types.Decimal `xml:"DepositAmount"`
}

// TaxedDeposits is a collection of taxed deposits.
type TaxedDeposits struct {
	TaxedDeposit []TaxedDeposit `xml:"TaxedDeposit"`
}

// TaxedDeposit represents a taxed deposit (advance invoice).
type TaxedDeposit struct {
	ID                            string                `xml:"ID"`
	VariableSymbol                string                `xml:"VariableSymbol,omitempty"`
	TaxableDepositAmountCurr      types.Decimal         `xml:"TaxableDepositAmountCurr,omitempty"`
	TaxableDepositAmount          types.Decimal         `xml:"TaxableDepositAmount"`
	TaxInclusiveDepositAmountCurr types.Decimal         `xml:"TaxInclusiveDepositAmountCurr,omitempty"`
	TaxInclusiveDepositAmount     types.Decimal         `xml:"TaxInclusiveDepositAmount"`
	ClassifiedTaxCategory         ClassifiedTaxCategory `xml:"ClassifiedTaxCategory"`
}
