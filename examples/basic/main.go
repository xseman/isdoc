// Example: Basic usage of the ISDOC library
//
// This example demonstrates the functional API for parsing, validating,
// and encoding ISDOC documents.
package main

import (
	"fmt"
	"log"

	"github.com/xseman/isdoc"
	"github.com/xseman/isdoc/schema"
	"github.com/xseman/isdoc/types"
)

func main() {
	// Example 1: Parse an existing ISDOC document
	parseExample()

	// Example 2: Create a new invoice from scratch
	createExample()
}

func parseExample() {
	fmt.Println("=== Parsing Example ===")

	// Sample ISDOC XML (minimal example)
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="http://isdoc.cz/namespace/2013" version="6.0.2">
  <DocumentType>1</DocumentType>
  <ID>FV-2024-001</ID>
  <UUID>12345678-1234-1234-1234-123456789012</UUID>
  <IssueDate>2024-01-15</IssueDate>
  <VATApplicable>true</VATApplicable>
  <LocalCurrencyCode>CZK</LocalCurrencyCode>
  <CurrRate>1</CurrRate>
  <RefCurrRate>1</RefCurrRate>
  <AccountingSupplierParty>
    <Party>
      <PartyIdentification><ID>12345678</ID></PartyIdentification>
      <PartyName><Name>Test Supplier s.r.o.</Name></PartyName>
      <PostalAddress>
        <StreetName>Hlavní 123</StreetName>
        <CityName>Praha</CityName>
        <PostalZone>11000</PostalZone>
        <Country><IdentificationCode>CZ</IdentificationCode></Country>
      </PostalAddress>
    </Party>
  </AccountingSupplierParty>
  <AccountingCustomerParty>
    <Party>
      <PartyIdentification><ID>87654321</ID></PartyIdentification>
      <PartyName><Name>Test Customer a.s.</Name></PartyName>
      <PostalAddress>
        <StreetName>Vedlejší 456</StreetName>
        <CityName>Brno</CityName>
        <PostalZone>60200</PostalZone>
        <Country><IdentificationCode>CZ</IdentificationCode></Country>
      </PostalAddress>
    </Party>
  </AccountingCustomerParty>
  <InvoiceLines>
    <InvoiceLine>
      <ID>1</ID>
      <InvoicedQuantity unitCode="C62">10</InvoicedQuantity>
      <LineExtensionAmount>1000.00</LineExtensionAmount>
      <LineExtensionAmountTaxInclusive>1210.00</LineExtensionAmountTaxInclusive>
      <LineExtensionTaxAmount>210.00</LineExtensionTaxAmount>
      <UnitPrice>100.00</UnitPrice>
      <UnitPriceTaxInclusive>121.00</UnitPriceTaxInclusive>
      <ClassifiedTaxCategory><Percent>21</Percent></ClassifiedTaxCategory>
      <Item><Description>Example Product</Description></Item>
    </InvoiceLine>
  </InvoiceLines>
  <TaxTotal>
    <TaxSubTotal>
      <TaxableAmount>1000.00</TaxableAmount>
      <TaxAmount>210.00</TaxAmount>
      <TaxInclusiveAmount>1210.00</TaxInclusiveAmount>
      <AlreadyClaimedTaxableAmount>0</AlreadyClaimedTaxableAmount>
      <AlreadyClaimedTaxAmount>0</AlreadyClaimedTaxAmount>
      <AlreadyClaimedTaxInclusiveAmount>0</AlreadyClaimedTaxInclusiveAmount>
      <DifferenceTaxableAmount>1000.00</DifferenceTaxableAmount>
      <DifferenceTaxAmount>210.00</DifferenceTaxAmount>
      <DifferenceTaxInclusiveAmount>1210.00</DifferenceTaxInclusiveAmount>
      <TaxCategory><Percent>21</Percent></TaxCategory>
    </TaxSubTotal>
    <TaxAmount>210.00</TaxAmount>
  </TaxTotal>
  <LegalMonetaryTotal>
    <TaxExclusiveAmount>1000.00</TaxExclusiveAmount>
    <TaxInclusiveAmount>1210.00</TaxInclusiveAmount>
    <AlreadyClaimedTaxExclusiveAmount>0</AlreadyClaimedTaxExclusiveAmount>
    <AlreadyClaimedTaxInclusiveAmount>0</AlreadyClaimedTaxInclusiveAmount>
    <DifferenceTaxExclusiveAmount>1000.00</DifferenceTaxExclusiveAmount>
    <DifferenceTaxInclusiveAmount>1210.00</DifferenceTaxInclusiveAmount>
    <PayableRoundingAmount>0</PayableRoundingAmount>
    <PaidDepositsAmount>0</PaidDepositsAmount>
    <PayableAmount>1210.00</PayableAmount>
  </LegalMonetaryTotal>
</Invoice>`)

	// Parse the document
	invoice, err := isdoc.DecodeBytes(xmlData)
	if err != nil {
		log.Fatalf("DecodeBytes error: %v", err)
	}

	fmt.Printf("Parsed invoice: %s\n", invoice.ID)
	fmt.Printf("  UUID: %s\n", invoice.UUID)
	fmt.Printf("  Issue Date: %s\n", invoice.IssueDate)
	fmt.Printf("  Supplier: %s\n", invoice.AccountingSupplierParty.Party.PartyName.Name)
	if invoice.AccountingCustomerParty != nil {
		fmt.Printf("  Customer: %s\n", invoice.AccountingCustomerParty.Party.PartyName.Name)
	}
	fmt.Printf("  Total: %s %s\n", invoice.LegalMonetaryTotal.PayableAmount, invoice.LocalCurrencyCode)
	fmt.Printf("  Lines: %d\n", len(invoice.InvoiceLines.InvoiceLine))

	// Validate the document
	validationErrors := isdoc.ValidateInvoice(invoice)
	if len(validationErrors) > 0 {
		fmt.Printf("Validation issues (%d):\n", len(validationErrors))
		for _, err := range validationErrors {
			fmt.Printf("  - [%s] %s: %s\n", err.Severity, err.Field, err.Msg)
		}
	} else {
		fmt.Println("Validation: OK")
	}

	fmt.Println()
}

func createExample() {
	fmt.Println("=== Creating Example ===")

	// Create a new invoice programmatically
	invoice := &schema.Invoice{
		Version:           "6.0.2",
		DocumentType:      1, // Invoice
		ID:                "FV-2024-002",
		UUID:              types.UUID("ABCDEF00-1234-5678-9ABC-DEF012345678"),
		IssueDate:         types.MustParseDate("2024-02-20"),
		VATApplicable:     types.Bool(true),
		LocalCurrencyCode: "CZK",
		CurrRate:          types.MustDecimal("1"),
		RefCurrRate:       types.MustDecimal("1"),
		AccountingSupplierParty: schema.AccountingSupplierParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{ID: "12345678"},
				PartyName:           schema.PartyName{Name: "My Company s.r.o."},
				PostalAddress: schema.PostalAddress{
					StreetName: "Business Street 100",
					CityName:   "Prague",
					PostalZone: "11000",
					Country:    schema.Country{IdentificationCode: "CZ"},
				},
				PartyTaxScheme: []schema.PartyTaxScheme{
					{CompanyID: "CZ12345678", TaxScheme: "VAT"},
				},
			},
		},
		AccountingCustomerParty: &schema.AccountingCustomerParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{ID: "87654321"},
				PartyName:           schema.PartyName{Name: "Customer Corp a.s."},
				PostalAddress: schema.PostalAddress{
					StreetName: "Customer Lane 50",
					CityName:   "Brno",
					PostalZone: "60200",
					Country:    schema.Country{IdentificationCode: "CZ"},
				},
			},
		},
		InvoiceLines: schema.InvoiceLines{
			InvoiceLine: []schema.InvoiceLine{
				{
					ID: "1",
					InvoicedQuantity: schema.Quantity{
						Value:    types.MustDecimal("5"),
						UnitCode: "C62",
					},
					LineExtensionAmount:             types.MustDecimal("500.00"),
					LineExtensionAmountTaxInclusive: types.MustDecimal("605.00"),
					LineExtensionTaxAmount:          types.MustDecimal("105.00"),
					UnitPrice:                       types.MustDecimal("100.00"),
					UnitPriceTaxInclusive:           types.MustDecimal("121.00"),
					ClassifiedTaxCategory: schema.ClassifiedTaxCategory{
						Percent:              types.MustDecimal("21"),
						VATCalculationMethod: 0,
					},
					Item: schema.Item{Description: "Widget A"},
				},
				{
					ID: "2",
					InvoicedQuantity: schema.Quantity{
						Value:    types.MustDecimal("3"),
						UnitCode: "C62",
					},
					LineExtensionAmount:             types.MustDecimal("300.00"),
					LineExtensionAmountTaxInclusive: types.MustDecimal("363.00"),
					LineExtensionTaxAmount:          types.MustDecimal("63.00"),
					UnitPrice:                       types.MustDecimal("100.00"),
					UnitPriceTaxInclusive:           types.MustDecimal("121.00"),
					ClassifiedTaxCategory: schema.ClassifiedTaxCategory{
						Percent:              types.MustDecimal("21"),
						VATCalculationMethod: 0,
					},
					Item: schema.Item{Description: "Widget B"},
				},
			},
		},
		TaxTotal: schema.TaxTotal{
			TaxSubTotal: []schema.TaxSubTotal{
				{
					TaxableAmount:                    types.MustDecimal("800.00"),
					TaxAmount:                        types.MustDecimal("168.00"),
					TaxInclusiveAmount:               types.MustDecimal("968.00"),
					AlreadyClaimedTaxableAmount:      types.MustDecimal("0"),
					AlreadyClaimedTaxAmount:          types.MustDecimal("0"),
					AlreadyClaimedTaxInclusiveAmount: types.MustDecimal("0"),
					DifferenceTaxableAmount:          types.MustDecimal("800.00"),
					DifferenceTaxAmount:              types.MustDecimal("168.00"),
					DifferenceTaxInclusiveAmount:     types.MustDecimal("968.00"),
					TaxCategory: schema.TaxCategory{
						Percent: types.MustDecimal("21"),
					},
				},
			},
			TaxAmount: types.MustDecimal("168.00"),
		},
		LegalMonetaryTotal: schema.LegalMonetaryTotal{
			TaxExclusiveAmount:               types.MustDecimal("800.00"),
			TaxInclusiveAmount:               types.MustDecimal("968.00"),
			AlreadyClaimedTaxExclusiveAmount: types.MustDecimal("0"),
			AlreadyClaimedTaxInclusiveAmount: types.MustDecimal("0"),
			DifferenceTaxExclusiveAmount:     types.MustDecimal("800.00"),
			DifferenceTaxInclusiveAmount:     types.MustDecimal("968.00"),
			PayableRoundingAmount:            types.MustDecimal("0"),
			PaidDepositsAmount:               types.MustDecimal("0"),
			PayableAmount:                    types.MustDecimal("968.00"),
		},
		PaymentMeans: &schema.PaymentMeans{
			Payment: []schema.Payment{
				{
					PaidAmount:       types.MustDecimal("968.00"),
					PaymentMeansCode: 42, // Bank transfer
					Details: &schema.PaymentDetails{
						PaymentDueDate: types.MustParseDate("2024-03-20"),
						VariableSymbol: "20240002",
						BankAccount: &schema.BankAccount{
							ID:       "1234567890",
							BankCode: "0100",
							IBAN:     "CZ6508000000001234567890",
						},
					},
				},
			},
		},
	}

	// Validate before encoding
	validationErrors := isdoc.ValidateInvoice(invoice)
	if validationErrors.HasErrors() {
		log.Fatalf("Validation errors: %v", validationErrors)
	}

	// Encode to XML
	xmlOutput, err := isdoc.EncodeBytes(invoice)
	if err != nil {
		log.Fatalf("EncodeBytes error: %v", err)
	}

	fmt.Printf("Created invoice %s with %d lines\n", invoice.ID, len(invoice.InvoiceLines.InvoiceLine))
	fmt.Printf("Total: %s CZK\n", invoice.LegalMonetaryTotal.PayableAmount)
	fmt.Printf("XML output length: %d bytes\n", len(xmlOutput))
	fmt.Println()

	// Print the first part of the XML
	preview := string(xmlOutput)
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	fmt.Printf("XML preview:\n%s\n", preview)
}
