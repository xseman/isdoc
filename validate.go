package isdoc

import (
	"fmt"

	"github.com/xseman/isdoc/schema"
	"github.com/xseman/isdoc/types"
)

// ValidateOptions configures validation behavior.
type ValidateOptions struct {
	// Strict mode treats all issues as errors. Default is false (warnings for minor issues).
	Strict bool

	// AllowRoundingTolerance permits small differences in totals. Default is true.
	AllowRoundingTolerance bool

	// Tolerance is the maximum allowed difference for total mismatches. Default is 0.01.
	Tolerance types.Decimal
}

// DefaultValidateOptions returns sensible defaults for validation.
func DefaultValidateOptions() ValidateOptions {
	return ValidateOptions{
		Strict:                 false,
		AllowRoundingTolerance: true,
		Tolerance:              types.MustDecimal("0.01"),
	}
}

// ValidateInvoice validates an ISDOC Invoice with default options.
//
// Performs three-layer validation:
//  1. Structural validation: Required fields, XML structure
//  2. Semantic validation: Business logic, calculations, totals
//  3. Schematron validation: Official ISDOC business rules
//
// Returns ValidationErrors which can contain both errors and warnings.
// Use HasErrors() to check if there are blocking errors.
//
// Example:
//
//	errs := isdoc.ValidateInvoice(invoice)
//	if errs.HasErrors() {
//	    for _, err := range errs.Errors() {
//	        fmt.Println(err)
//	    }
//	}
func ValidateInvoice(inv *schema.Invoice) ValidationErrors {
	return ValidateInvoiceWithOptions(inv, DefaultValidateOptions())
}

// ValidateInvoiceWithOptions validates an ISDOC Invoice with custom options.
//
// Use this function when you need to customize validation behavior:
//   - Strict mode: Treat all issues as errors
//   - Rounding tolerance: Allow small differences in totals
//
// Example:
//
//	opts := isdoc.ValidateOptions{
//	    Strict: true,
//	    AllowRoundingTolerance: false,
//	    Tolerance: types.MustDecimal("0.01"),
//	}
//	errs := isdoc.ValidateInvoiceWithOptions(invoice, opts)
func ValidateInvoiceWithOptions(inv *schema.Invoice, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors

	// Structural validation
	errs = append(errs, validateStructural(inv, opts)...)

	// Semantic validation
	errs = append(errs, validateSemantic(inv, opts)...)

	return errs
}

// validateStructural checks required fields and structure.
func validateStructural(inv *schema.Invoice, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors
	severity := SeverityWarning
	if opts.Strict {
		severity = SeverityError
	}

	// Required fields
	if inv.Version == "" {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.@version",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError, // Always error for version
			Msg:      "version attribute is required",
		})
	}

	if inv.DocumentType < 1 || inv.DocumentType > 7 {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.DocumentType",
			Code:     ErrCodeInvalidEnum,
			Severity: SeverityError,
			Msg:      fmt.Sprintf("DocumentType must be 1-7, got %d", inv.DocumentType),
		})
	}

	if inv.ID == "" {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.ID",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "ID is required",
		})
	}

	if inv.UUID.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.UUID",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "UUID is required",
		})
	}

	if inv.IssueDate.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.IssueDate",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "IssueDate is required",
		})
	}

	if inv.LocalCurrencyCode == "" {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LocalCurrencyCode",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "LocalCurrencyCode is required",
		})
	} else if len(inv.LocalCurrencyCode) != 3 {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LocalCurrencyCode",
			Code:     ErrCodeInvalidLength,
			Severity: severity,
			Msg:      fmt.Sprintf("LocalCurrencyCode must be 3 characters, got %d", len(inv.LocalCurrencyCode)),
		})
	}

	if inv.CurrRate.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.CurrRate",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "CurrRate is required",
		})
	}

	if inv.RefCurrRate.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.RefCurrRate",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "RefCurrRate is required",
		})
	}

	// Supplier
	errs = append(errs, validateParty("Invoice.AccountingSupplierParty.Party",
		&inv.AccountingSupplierParty.Party, opts)...)

	// Customer (either AccountingCustomerParty or AnonymousCustomerParty for simplified docs)
	if inv.AccountingCustomerParty == nil && inv.AnonymousCustomerParty == nil {
		// For document type 7 (simplified), AnonymousCustomerParty is acceptable
		if inv.DocumentType != 7 {
			errs = append(errs, &ValidationError{
				Field:    "Invoice.AccountingCustomerParty",
				Code:     ErrCodeRequiredField,
				Severity: SeverityError,
				Msg:      "AccountingCustomerParty is required (or AnonymousCustomerParty for simplified tax documents)",
			})
		}
	} else if inv.AccountingCustomerParty != nil {
		errs = append(errs, validateParty("Invoice.AccountingCustomerParty.Party",
			&inv.AccountingCustomerParty.Party, opts)...)
	}

	// Invoice lines
	if len(inv.InvoiceLines.InvoiceLine) == 0 {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.InvoiceLines",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "at least one InvoiceLine is required",
		})
	}

	for i, line := range inv.InvoiceLines.InvoiceLine {
		errs = append(errs, validateInvoiceLine(
			fmt.Sprintf("Invoice.InvoiceLines.InvoiceLine[%d]", i),
			&line, opts)...)
	}

	// TaxTotal
	if len(inv.TaxTotal.TaxSubTotal) == 0 {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.TaxTotal.TaxSubTotal",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "at least one TaxSubTotal is expected",
		})
	}

	// LegalMonetaryTotal required fields
	if inv.LegalMonetaryTotal.TaxExclusiveAmount.IsZero() && inv.LegalMonetaryTotal.TaxInclusiveAmount.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "TaxExclusiveAmount or TaxInclusiveAmount should be set",
		})
	}

	return errs
}

// validateParty validates a Party structure.
func validateParty(path string, party *schema.Party, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors
	severity := SeverityWarning
	if opts.Strict {
		severity = SeverityError
	}

	if party.PartyIdentification.ID == "" {
		errs = append(errs, &ValidationError{
			Field:    path + ".PartyIdentification.ID",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "PartyIdentification.ID (IČO) is required",
		})
	}

	if party.PartyName.Name == "" {
		errs = append(errs, &ValidationError{
			Field:    path + ".PartyName.Name",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "PartyName.Name is required",
		})
	}

	if party.PostalAddress.CityName == "" {
		errs = append(errs, &ValidationError{
			Field:    path + ".PostalAddress.CityName",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "PostalAddress.CityName is expected",
		})
	}

	if party.PostalAddress.Country.IdentificationCode == "" {
		errs = append(errs, &ValidationError{
			Field:    path + ".PostalAddress.Country.IdentificationCode",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "Country.IdentificationCode is expected",
		})
	}

	return errs
}

// validateInvoiceLine validates an invoice line.
func validateInvoiceLine(path string, line *schema.InvoiceLine, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors
	severity := SeverityWarning
	if opts.Strict {
		severity = SeverityError
	}

	if line.ID == "" {
		errs = append(errs, &ValidationError{
			Field:    path + ".ID",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "ID is required",
		})
	} else if len(line.ID) > 36 {
		errs = append(errs, &ValidationError{
			Field:    path + ".ID",
			Code:     ErrCodeInvalidLength,
			Severity: severity,
			Msg:      fmt.Sprintf("ID must be max 36 characters, got %d", len(line.ID)),
		})
	}

	if line.InvoicedQuantity.Value.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    path + ".InvoicedQuantity",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "InvoicedQuantity is expected",
		})
	}

	if line.LineExtensionAmount.IsZero() && !opts.Strict {
		// Some invoices may have zero amounts legitimately
	}

	if line.Item.Description == "" {
		errs = append(errs, &ValidationError{
			Field:    path + ".Item.Description",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "Item.Description is expected",
		})
	}

	return errs
}

// validateSemantic checks business logic consistency.
func validateSemantic(inv *schema.Invoice, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors

	// Schematron business rules from isdoc-6.0.2.sch
	errs = append(errs, validateOriginalDocumentLink(inv, opts)...)
	errs = append(errs, validateCurrencyConsistency(inv, opts)...)
	errs = append(errs, validateVATConsistency(inv, opts)...)
	errs = append(errs, validateItemIdentificationHierarchy(inv, opts)...)
	errs = append(errs, validateStoreBatches(inv, opts)...)

	return errs
}

// validateOriginalDocumentLink checks that DocumentType 2,3,6 have OriginalDocumentReferences.
// Schematron rule: "Vazba na původní doklad"
func validateOriginalDocumentLink(inv *schema.Invoice, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors

	// DocumentType 2=credit note, 3=debit note, 6=credit of advance invoice
	if inv.DocumentType == 2 || inv.DocumentType == 3 || inv.DocumentType == 6 {
		if inv.OriginalDocumentReferences == nil || len(inv.OriginalDocumentReferences.OriginalDocumentReference) == 0 {
			errs = append(errs, &ValidationError{
				Field:    "Invoice.OriginalDocumentReferences",
				Code:     ErrCodeRequiredField,
				Severity: SeverityError,
				Msg:      fmt.Sprintf("DocumentType %d requires OriginalDocumentReferences with at least one OriginalDocumentReference", inv.DocumentType),
			})
		}
	}

	return errs
}

// validateCurrencyConsistency validates foreign/domestic currency field consistency.
// Schematron rules: "Konzistentní uvádění cizí měny", "Konzistentní uvádění tuzemské měny",
// "Tuzemská a zahraniční měna musí být rozdílná"
func validateCurrencyConsistency(inv *schema.Invoice, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors
	severity := SeverityWarning
	if opts.Strict {
		severity = SeverityError
	}

	hasForeignCurrency := inv.ForeignCurrencyCode != ""

	if hasForeignCurrency {
		// Rule: Foreign and local currency must be different
		if inv.ForeignCurrencyCode == inv.LocalCurrencyCode {
			errs = append(errs, &ValidationError{
				Field:    "Invoice.ForeignCurrencyCode",
				Code:     ErrCodeSchemaViolation,
				Severity: SeverityError,
				Msg:      "ForeignCurrencyCode must differ from LocalCurrencyCode",
			})
		}

		// Rule: When ForeignCurrencyCode exists, all *Curr fields should be present
		errs = append(errs, validateForeignCurrencyFieldsPresent(inv, opts)...)
	} else {
		// Rule: When no ForeignCurrencyCode, CurrRate and RefCurrRate must be 1
		one := types.MustDecimal("1")
		if !inv.CurrRate.IsZero() && !inv.CurrRate.Equal(one) {
			errs = append(errs, &ValidationError{
				Field:    "Invoice.CurrRate",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      fmt.Sprintf("CurrRate must be 1 when no ForeignCurrencyCode, got %s", inv.CurrRate.String()),
			})
		}
		if !inv.RefCurrRate.IsZero() && !inv.RefCurrRate.Equal(one) {
			errs = append(errs, &ValidationError{
				Field:    "Invoice.RefCurrRate",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      fmt.Sprintf("RefCurrRate must be 1 when no ForeignCurrencyCode, got %s", inv.RefCurrRate.String()),
			})
		}

		// Rule: No *Curr fields should exist when no ForeignCurrencyCode
		errs = append(errs, validateNoForeignCurrencyFields(inv, opts)...)
	}

	return errs
}

// validateForeignCurrencyFieldsPresent checks that *Curr fields are present when ForeignCurrencyCode exists.
// Schematron rule: "Konzistentní uvádění cizí měny"
func validateForeignCurrencyFieldsPresent(inv *schema.Invoice, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors
	severity := SeverityWarning
	if opts.Strict {
		severity = SeverityError
	}

	// Check LegalMonetaryTotal currency fields
	lmt := inv.LegalMonetaryTotal
	if !lmt.TaxExclusiveAmount.IsZero() && lmt.TaxExclusiveAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.TaxExclusiveAmountCurr",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "TaxExclusiveAmountCurr required when ForeignCurrencyCode is set",
		})
	}
	if !lmt.TaxInclusiveAmount.IsZero() && lmt.TaxInclusiveAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.TaxInclusiveAmountCurr",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "TaxInclusiveAmountCurr required when ForeignCurrencyCode is set",
		})
	}
	if !lmt.PayableAmount.IsZero() && lmt.PayableAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.PayableAmountCurr",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "PayableAmountCurr required when ForeignCurrencyCode is set",
		})
	}
	// Extended currency fields in LegalMonetaryTotal
	if !lmt.PayableRoundingAmount.IsZero() && lmt.PayableRoundingAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.PayableRoundingAmountCurr",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "PayableRoundingAmountCurr required when ForeignCurrencyCode is set",
		})
	}
	if !lmt.PaidDepositsAmount.IsZero() && lmt.PaidDepositsAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.PaidDepositsAmountCurr",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "PaidDepositsAmountCurr required when ForeignCurrencyCode is set",
		})
	}
	// Credit note fields
	if !lmt.AlreadyClaimedTaxExclusiveAmount.IsZero() && lmt.AlreadyClaimedTaxExclusiveAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.AlreadyClaimedTaxExclusiveAmountCurr",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "AlreadyClaimedTaxExclusiveAmountCurr required when ForeignCurrencyCode is set",
		})
	}
	if !lmt.AlreadyClaimedTaxInclusiveAmount.IsZero() && lmt.AlreadyClaimedTaxInclusiveAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.AlreadyClaimedTaxInclusiveAmountCurr",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "AlreadyClaimedTaxInclusiveAmountCurr required when ForeignCurrencyCode is set",
		})
	}
	if !lmt.DifferenceTaxExclusiveAmount.IsZero() && lmt.DifferenceTaxExclusiveAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.DifferenceTaxExclusiveAmountCurr",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "DifferenceTaxExclusiveAmountCurr required when ForeignCurrencyCode is set",
		})
	}
	if !lmt.DifferenceTaxInclusiveAmount.IsZero() && lmt.DifferenceTaxInclusiveAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.DifferenceTaxInclusiveAmountCurr",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "DifferenceTaxInclusiveAmountCurr required when ForeignCurrencyCode is set",
		})
	}

	// Check TaxTotal currency fields
	if !inv.TaxTotal.TaxAmount.IsZero() && inv.TaxTotal.TaxAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.TaxTotal.TaxAmountCurr",
			Code:     ErrCodeRequiredField,
			Severity: severity,
			Msg:      "TaxAmountCurr required when ForeignCurrencyCode is set",
		})
	}

	// Check TaxSubTotal currency fields
	for i, sub := range inv.TaxTotal.TaxSubTotal {
		path := fmt.Sprintf("Invoice.TaxTotal.TaxSubTotal[%d]", i)
		if !sub.TaxableAmount.IsZero() && sub.TaxableAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".TaxableAmountCurr",
				Code:     ErrCodeRequiredField,
				Severity: severity,
				Msg:      "TaxableAmountCurr required when ForeignCurrencyCode is set",
			})
		}
		if !sub.TaxAmount.IsZero() && sub.TaxAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".TaxAmountCurr",
				Code:     ErrCodeRequiredField,
				Severity: severity,
				Msg:      "TaxAmountCurr required when ForeignCurrencyCode is set",
			})
		}
		if !sub.TaxInclusiveAmount.IsZero() && sub.TaxInclusiveAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".TaxInclusiveAmountCurr",
				Code:     ErrCodeRequiredField,
				Severity: severity,
				Msg:      "TaxInclusiveAmountCurr required when ForeignCurrencyCode is set",
			})
		}
		// Credit note fields in TaxSubTotal
		if !sub.AlreadyClaimedTaxableAmount.IsZero() && sub.AlreadyClaimedTaxableAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".AlreadyClaimedTaxableAmountCurr",
				Code:     ErrCodeRequiredField,
				Severity: severity,
				Msg:      "AlreadyClaimedTaxableAmountCurr required when ForeignCurrencyCode is set",
			})
		}
		if !sub.AlreadyClaimedTaxAmount.IsZero() && sub.AlreadyClaimedTaxAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".AlreadyClaimedTaxAmountCurr",
				Code:     ErrCodeRequiredField,
				Severity: severity,
				Msg:      "AlreadyClaimedTaxAmountCurr required when ForeignCurrencyCode is set",
			})
		}
		if !sub.AlreadyClaimedTaxInclusiveAmount.IsZero() && sub.AlreadyClaimedTaxInclusiveAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".AlreadyClaimedTaxInclusiveAmountCurr",
				Code:     ErrCodeRequiredField,
				Severity: severity,
				Msg:      "AlreadyClaimedTaxInclusiveAmountCurr required when ForeignCurrencyCode is set",
			})
		}
		if !sub.DifferenceTaxableAmount.IsZero() && sub.DifferenceTaxableAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".DifferenceTaxableAmountCurr",
				Code:     ErrCodeRequiredField,
				Severity: severity,
				Msg:      "DifferenceTaxableAmountCurr required when ForeignCurrencyCode is set",
			})
		}
		if !sub.DifferenceTaxAmount.IsZero() && sub.DifferenceTaxAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".DifferenceTaxAmountCurr",
				Code:     ErrCodeRequiredField,
				Severity: severity,
				Msg:      "DifferenceTaxAmountCurr required when ForeignCurrencyCode is set",
			})
		}
		if !sub.DifferenceTaxInclusiveAmount.IsZero() && sub.DifferenceTaxInclusiveAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".DifferenceTaxInclusiveAmountCurr",
				Code:     ErrCodeRequiredField,
				Severity: severity,
				Msg:      "DifferenceTaxInclusiveAmountCurr required when ForeignCurrencyCode is set",
			})
		}
	}

	// Check InvoiceLine currency fields
	for i, line := range inv.InvoiceLines.InvoiceLine {
		path := fmt.Sprintf("Invoice.InvoiceLines.InvoiceLine[%d]", i)
		if !line.LineExtensionAmount.IsZero() && line.LineExtensionAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".LineExtensionAmountCurr",
				Code:     ErrCodeRequiredField,
				Severity: severity,
				Msg:      "LineExtensionAmountCurr required when ForeignCurrencyCode is set",
			})
		}
		if !line.LineExtensionAmountTaxInclusive.IsZero() && line.LineExtensionAmountTaxInclusiveCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".LineExtensionAmountTaxInclusiveCurr",
				Code:     ErrCodeRequiredField,
				Severity: severity,
				Msg:      "LineExtensionAmountTaxInclusiveCurr required when ForeignCurrencyCode is set",
			})
		}
	}

	// Check NonTaxedDeposits currency fields
	if inv.NonTaxedDeposits != nil {
		for i, dep := range inv.NonTaxedDeposits.NonTaxedDeposit {
			path := fmt.Sprintf("Invoice.NonTaxedDeposits.NonTaxedDeposit[%d]", i)
			if !dep.DepositAmount.IsZero() && dep.DepositAmountCurr.IsZero() {
				errs = append(errs, &ValidationError{
					Field:    path + ".DepositAmountCurr",
					Code:     ErrCodeRequiredField,
					Severity: severity,
					Msg:      "DepositAmountCurr required when ForeignCurrencyCode is set",
				})
			}
		}
	}

	// Check TaxedDeposits currency fields
	if inv.TaxedDeposits != nil {
		for i, dep := range inv.TaxedDeposits.TaxedDeposit {
			path := fmt.Sprintf("Invoice.TaxedDeposits.TaxedDeposit[%d]", i)
			if !dep.TaxableDepositAmount.IsZero() && dep.TaxableDepositAmountCurr.IsZero() {
				errs = append(errs, &ValidationError{
					Field:    path + ".TaxableDepositAmountCurr",
					Code:     ErrCodeRequiredField,
					Severity: severity,
					Msg:      "TaxableDepositAmountCurr required when ForeignCurrencyCode is set",
				})
			}
			if !dep.TaxInclusiveDepositAmount.IsZero() && dep.TaxInclusiveDepositAmountCurr.IsZero() {
				errs = append(errs, &ValidationError{
					Field:    path + ".TaxInclusiveDepositAmountCurr",
					Code:     ErrCodeRequiredField,
					Severity: severity,
					Msg:      "TaxInclusiveDepositAmountCurr required when ForeignCurrencyCode is set",
				})
			}
		}
	}

	return errs
}

// validateNoForeignCurrencyFields checks that no *Curr fields exist when no ForeignCurrencyCode.
// Schematron rule: "Konzistentní uvádění tuzemské měny"
func validateNoForeignCurrencyFields(inv *schema.Invoice, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors
	severity := SeverityWarning
	if opts.Strict {
		severity = SeverityError
	}

	// Check LegalMonetaryTotal
	lmt := inv.LegalMonetaryTotal
	if !lmt.TaxExclusiveAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.TaxExclusiveAmountCurr",
			Code:     ErrCodeSchemaViolation,
			Severity: severity,
			Msg:      "TaxExclusiveAmountCurr must not be set without ForeignCurrencyCode",
		})
	}
	if !lmt.TaxInclusiveAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.TaxInclusiveAmountCurr",
			Code:     ErrCodeSchemaViolation,
			Severity: severity,
			Msg:      "TaxInclusiveAmountCurr must not be set without ForeignCurrencyCode",
		})
	}
	if !lmt.PayableAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.PayableAmountCurr",
			Code:     ErrCodeSchemaViolation,
			Severity: severity,
			Msg:      "PayableAmountCurr must not be set without ForeignCurrencyCode",
		})
	}
	// Extended currency fields in LegalMonetaryTotal
	if !lmt.PayableRoundingAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.PayableRoundingAmountCurr",
			Code:     ErrCodeSchemaViolation,
			Severity: severity,
			Msg:      "PayableRoundingAmountCurr must not be set without ForeignCurrencyCode",
		})
	}
	if !lmt.PaidDepositsAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.PaidDepositsAmountCurr",
			Code:     ErrCodeSchemaViolation,
			Severity: severity,
			Msg:      "PaidDepositsAmountCurr must not be set without ForeignCurrencyCode",
		})
	}
	// Credit note fields
	if !lmt.AlreadyClaimedTaxExclusiveAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.AlreadyClaimedTaxExclusiveAmountCurr",
			Code:     ErrCodeSchemaViolation,
			Severity: severity,
			Msg:      "AlreadyClaimedTaxExclusiveAmountCurr must not be set without ForeignCurrencyCode",
		})
	}
	if !lmt.AlreadyClaimedTaxInclusiveAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.AlreadyClaimedTaxInclusiveAmountCurr",
			Code:     ErrCodeSchemaViolation,
			Severity: severity,
			Msg:      "AlreadyClaimedTaxInclusiveAmountCurr must not be set without ForeignCurrencyCode",
		})
	}
	if !lmt.DifferenceTaxExclusiveAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.DifferenceTaxExclusiveAmountCurr",
			Code:     ErrCodeSchemaViolation,
			Severity: severity,
			Msg:      "DifferenceTaxExclusiveAmountCurr must not be set without ForeignCurrencyCode",
		})
	}
	if !lmt.DifferenceTaxInclusiveAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.LegalMonetaryTotal.DifferenceTaxInclusiveAmountCurr",
			Code:     ErrCodeSchemaViolation,
			Severity: severity,
			Msg:      "DifferenceTaxInclusiveAmountCurr must not be set without ForeignCurrencyCode",
		})
	}

	// Check TaxTotal
	if !inv.TaxTotal.TaxAmountCurr.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "Invoice.TaxTotal.TaxAmountCurr",
			Code:     ErrCodeSchemaViolation,
			Severity: severity,
			Msg:      "TaxAmountCurr must not be set without ForeignCurrencyCode",
		})
	}

	// Check TaxSubTotal currency fields
	for i, sub := range inv.TaxTotal.TaxSubTotal {
		path := fmt.Sprintf("Invoice.TaxTotal.TaxSubTotal[%d]", i)
		if !sub.TaxableAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".TaxableAmountCurr",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      "TaxableAmountCurr must not be set without ForeignCurrencyCode",
			})
		}
		if !sub.TaxAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".TaxAmountCurr",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      "TaxAmountCurr must not be set without ForeignCurrencyCode",
			})
		}
		if !sub.TaxInclusiveAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".TaxInclusiveAmountCurr",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      "TaxInclusiveAmountCurr must not be set without ForeignCurrencyCode",
			})
		}
		// Credit note fields in TaxSubTotal
		if !sub.AlreadyClaimedTaxableAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".AlreadyClaimedTaxableAmountCurr",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      "AlreadyClaimedTaxableAmountCurr must not be set without ForeignCurrencyCode",
			})
		}
		if !sub.AlreadyClaimedTaxAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".AlreadyClaimedTaxAmountCurr",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      "AlreadyClaimedTaxAmountCurr must not be set without ForeignCurrencyCode",
			})
		}
		if !sub.AlreadyClaimedTaxInclusiveAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".AlreadyClaimedTaxInclusiveAmountCurr",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      "AlreadyClaimedTaxInclusiveAmountCurr must not be set without ForeignCurrencyCode",
			})
		}
		if !sub.DifferenceTaxableAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".DifferenceTaxableAmountCurr",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      "DifferenceTaxableAmountCurr must not be set without ForeignCurrencyCode",
			})
		}
		if !sub.DifferenceTaxAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".DifferenceTaxAmountCurr",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      "DifferenceTaxAmountCurr must not be set without ForeignCurrencyCode",
			})
		}
		if !sub.DifferenceTaxInclusiveAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".DifferenceTaxInclusiveAmountCurr",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      "DifferenceTaxInclusiveAmountCurr must not be set without ForeignCurrencyCode",
			})
		}
	}

	// Check InvoiceLines
	for i, line := range inv.InvoiceLines.InvoiceLine {
		path := fmt.Sprintf("Invoice.InvoiceLines.InvoiceLine[%d]", i)
		if !line.LineExtensionAmountCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".LineExtensionAmountCurr",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      "LineExtensionAmountCurr must not be set without ForeignCurrencyCode",
			})
		}
		if !line.LineExtensionAmountTaxInclusiveCurr.IsZero() {
			errs = append(errs, &ValidationError{
				Field:    path + ".LineExtensionAmountTaxInclusiveCurr",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      "LineExtensionAmountTaxInclusiveCurr must not be set without ForeignCurrencyCode",
			})
		}
	}

	// Check NonTaxedDeposits currency fields
	if inv.NonTaxedDeposits != nil {
		for i, dep := range inv.NonTaxedDeposits.NonTaxedDeposit {
			path := fmt.Sprintf("Invoice.NonTaxedDeposits.NonTaxedDeposit[%d]", i)
			if !dep.DepositAmountCurr.IsZero() {
				errs = append(errs, &ValidationError{
					Field:    path + ".DepositAmountCurr",
					Code:     ErrCodeSchemaViolation,
					Severity: severity,
					Msg:      "DepositAmountCurr must not be set without ForeignCurrencyCode",
				})
			}
		}
	}

	// Check TaxedDeposits currency fields
	if inv.TaxedDeposits != nil {
		for i, dep := range inv.TaxedDeposits.TaxedDeposit {
			path := fmt.Sprintf("Invoice.TaxedDeposits.TaxedDeposit[%d]", i)
			if !dep.TaxableDepositAmountCurr.IsZero() {
				errs = append(errs, &ValidationError{
					Field:    path + ".TaxableDepositAmountCurr",
					Code:     ErrCodeSchemaViolation,
					Severity: severity,
					Msg:      "TaxableDepositAmountCurr must not be set without ForeignCurrencyCode",
				})
			}
			if !dep.TaxInclusiveDepositAmountCurr.IsZero() {
				errs = append(errs, &ValidationError{
					Field:    path + ".TaxInclusiveDepositAmountCurr",
					Code:     ErrCodeSchemaViolation,
					Severity: severity,
					Msg:      "TaxInclusiveDepositAmountCurr must not be set without ForeignCurrencyCode",
				})
			}
		}
	}

	return errs
}

// validateVATConsistency checks VAT applicability consistency.
// Schematron rule: "Nedaňový doklad nesmí obsahovat řádkové položky podléhající DPH"
func validateVATConsistency(inv *schema.Invoice, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors

	// If invoice VATApplicable is false, all line items must also be non-VAT
	if !inv.VATApplicable.Bool() {
		for i, line := range inv.InvoiceLines.InvoiceLine {
			if line.ClassifiedTaxCategory.VATApplicable.Bool() {
				errs = append(errs, &ValidationError{
					Field:    fmt.Sprintf("Invoice.InvoiceLines.InvoiceLine[%d].ClassifiedTaxCategory.VATApplicable", i),
					Code:     ErrCodeVATMismatch,
					Severity: SeverityError,
					Msg:      "Invoice VATApplicable is false, but line item has VATApplicable=true",
				})
			}
		}
	}

	return errs
}

// validateItemIdentificationHierarchy checks item identification hierarchy.
// Schematron rules: Secondary requires Primary, Tertiary requires Secondary and Primary
func validateItemIdentificationHierarchy(inv *schema.Invoice, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors
	severity := SeverityWarning
	if opts.Strict {
		severity = SeverityError
	}

	for i, line := range inv.InvoiceLines.InvoiceLine {
		path := fmt.Sprintf("Invoice.InvoiceLines.InvoiceLine[%d].Item", i)

		// SecondarySellersItemIdentification requires SellersItemIdentification
		if line.Item.SecondarySellersItemIdentification != nil && line.Item.SellersItemIdentification == nil {
			errs = append(errs, &ValidationError{
				Field:    path + ".SecondarySellersItemIdentification",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      "SecondarySellersItemIdentification requires SellersItemIdentification",
			})
		}

		// TertiarySellersItemIdentification requires both Secondary and Primary
		if line.Item.TertiarySellersItemIdentification != nil {
			if line.Item.SellersItemIdentification == nil || line.Item.SecondarySellersItemIdentification == nil {
				errs = append(errs, &ValidationError{
					Field:    path + ".TertiarySellersItemIdentification",
					Code:     ErrCodeSchemaViolation,
					Severity: severity,
					Msg:      "TertiarySellersItemIdentification requires both SellersItemIdentification and SecondarySellersItemIdentification",
				})
			}
		}
	}

	return errs
}

// validateStoreBatches checks store batch structure.
// Schematron rules:
// - "Jednotky jednotlivých šarží" - All batch unitCodes must match line unitCode
// - "Součet množství za jednotlivé šarže" - Batch quantities must sum to InvoicedQuantity
func validateStoreBatches(inv *schema.Invoice, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors
	severity := SeverityWarning
	if opts.Strict {
		severity = SeverityError
	}

	for i, line := range inv.InvoiceLines.InvoiceLine {
		if line.Item.StoreBatches == nil || len(line.Item.StoreBatches.StoreBatch) == 0 {
			continue
		}

		path := fmt.Sprintf("Invoice.InvoiceLines.InvoiceLine[%d]", i)
		lineUnitCode := line.InvoicedQuantity.UnitCode

		// Collect all batch unit codes and validate consistency
		var batchQuantitySum float64
		var seenUnitCodes []string

		for j, batch := range line.Item.StoreBatches.StoreBatch {
			batchPath := fmt.Sprintf("%s.Item.StoreBatches.StoreBatch[%d]", path, j)
			batchUnitCode := batch.Quantity.UnitCode

			// Track unique unit codes
			if batchUnitCode != "" {
				found := false
				for _, uc := range seenUnitCodes {
					if uc == batchUnitCode {
						found = true
						break
					}
				}
				if !found {
					seenUnitCodes = append(seenUnitCodes, batchUnitCode)
				}
			}

			// Validate unit code matches line's unit code (if line has one)
			if lineUnitCode != "" && batchUnitCode != "" && batchUnitCode != lineUnitCode {
				errs = append(errs, &ValidationError{
					Field:    batchPath + ".Quantity",
					Code:     ErrCodeSchemaViolation,
					Severity: severity,
					Msg:      fmt.Sprintf("StoreBatch unitCode %q must match InvoicedQuantity unitCode %q", batchUnitCode, lineUnitCode),
				})
			}

			// Sum up batch quantities
			batchQuantitySum += batch.Quantity.Value.Float64()
		}

		// Check all batch unit codes are the same (if there are multiple with unit codes)
		if len(seenUnitCodes) > 1 {
			errs = append(errs, &ValidationError{
				Field:    path + ".Item.StoreBatches",
				Code:     ErrCodeSchemaViolation,
				Severity: severity,
				Msg:      fmt.Sprintf("All StoreBatch Quantity unitCodes must be the same, found: %v", seenUnitCodes),
			})
		}

		// Validate batch quantities sum to InvoicedQuantity
		invoicedQty := line.InvoicedQuantity.Value.Float64()
		// Use a small tolerance for floating-point comparison
		tolerance := 0.0001
		if diff := batchQuantitySum - invoicedQty; diff > tolerance || diff < -tolerance {
			errs = append(errs, &ValidationError{
				Field:    path + ".Item.StoreBatches",
				Code:     ErrCodeTotalMismatch,
				Severity: severity,
				Msg:      fmt.Sprintf("StoreBatch Quantity sum (%.4f) must equal InvoicedQuantity (%.4f)", batchQuantitySum, invoicedQty),
			})
		}
	}

	return errs
}

// -----------------------------------------------------------------------------
// CommonDocument Validation
// -----------------------------------------------------------------------------

// ValidateCommonDocument validates an ISDOC CommonDocument with default options.
//
// CommonDocument is for non-payment documents like contracts and certificates.
// Validation checks structure, required fields, and business logic.
//
// Example:
//
//	errs := isdoc.ValidateCommonDocument(doc)
//	if errs.HasErrors() {
//	    log.Fatal(errs)
//	}
func ValidateCommonDocument(doc *schema.CommonDocument) ValidationErrors {
	return ValidateCommonDocumentWithOptions(doc, DefaultValidateOptions())
}

// ValidateCommonDocumentWithOptions validates a CommonDocument with custom options.
func ValidateCommonDocumentWithOptions(doc *schema.CommonDocument, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors

	// Structural validation
	errs = append(errs, validateCommonDocumentStructural(doc, opts)...)

	return errs
}

// validateCommonDocumentStructural checks required fields for CommonDocument.
func validateCommonDocumentStructural(doc *schema.CommonDocument, opts ValidateOptions) ValidationErrors {
	var errs ValidationErrors

	// Required fields
	if doc.Version == "" {
		errs = append(errs, &ValidationError{
			Field:    "CommonDocument.@version",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "version attribute is required",
		})
	}

	if doc.SubDocumentType == "" {
		errs = append(errs, &ValidationError{
			Field:    "CommonDocument.SubDocumentType",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "SubDocumentType is required",
		})
	}

	if doc.SubDocumentTypeOrigin == "" {
		errs = append(errs, &ValidationError{
			Field:    "CommonDocument.SubDocumentTypeOrigin",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "SubDocumentTypeOrigin is required",
		})
	}

	if doc.ID == "" {
		errs = append(errs, &ValidationError{
			Field:    "CommonDocument.ID",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "ID is required",
		})
	}

	if doc.UUID.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "CommonDocument.UUID",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "UUID is required",
		})
	}

	if doc.IssueDate.IsZero() {
		errs = append(errs, &ValidationError{
			Field:    "CommonDocument.IssueDate",
			Code:     ErrCodeRequiredField,
			Severity: SeverityError,
			Msg:      "IssueDate is required",
		})
	}

	// Supplier party
	errs = append(errs, validateParty("CommonDocument.AccountingSupplierParty.Party",
		&doc.AccountingSupplierParty.Party, opts)...)

	// Customer party
	errs = append(errs, validateParty("CommonDocument.AccountingCustomerParty.Party",
		&doc.AccountingCustomerParty.Party, opts)...)

	return errs
}
