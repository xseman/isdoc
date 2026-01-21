// Example: Advanced usage of the ISDOC library
//
// This example demonstrates advanced workflows including
// validation options and round-trip processing.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/xseman/isdoc"
	"github.com/xseman/isdoc/schema"
	"github.com/xseman/isdoc/types"
)

func main() {
	// Example 1: Direct API usage
	directAPIExample()

	// Example 2: Strict validation
	strictValidationExample()

	// Example 3: Round-trip processing
	roundTripExample()
}

func directAPIExample() {
	fmt.Println("=== Direct API Example ===")

	// Create an invoice
	invoice := &schema.Invoice{
		Version:           "6.0.2",
		DocumentType:      1,
		ID:                "FV-2024-003",
		UUID:              types.UUID("12345678-AAAA-BBBB-CCCC-123456789012"),
		IssueDate:         types.MustParseDate("2024-03-01"),
		VATApplicable:     types.Bool(true),
		LocalCurrencyCode: "CZK",
		CurrRate:          types.MustDecimal("1"),
		RefCurrRate:       types.MustDecimal("1"),
		AccountingSupplierParty: schema.AccountingSupplierParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{ID: "12345678"},
				PartyName:           schema.PartyName{Name: "Advanced Corp s.r.o."},
				PostalAddress: schema.PostalAddress{
					StreetName: "Tech Park 42",
					CityName:   "Prague",
					PostalZone: "14000",
					Country:    schema.Country{IdentificationCode: "CZ"},
				},
			},
		},
		AccountingCustomerParty: &schema.AccountingCustomerParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{ID: "87654321"},
				PartyName:           schema.PartyName{Name: "Enterprise Ltd a.s."},
				PostalAddress: schema.PostalAddress{
					StreetName: "Business Center 100",
					CityName:   "Brno",
					PostalZone: "60200",
					Country:    schema.Country{IdentificationCode: "CZ"},
				},
			},
		},
		InvoiceLines: schema.InvoiceLines{
			InvoiceLine: []schema.InvoiceLine{
				{
					ID:                              "1",
					InvoicedQuantity:                schema.Quantity{Value: types.MustDecimal("1"), UnitCode: "C62"},
					LineExtensionAmount:             types.MustDecimal("10000.00"),
					LineExtensionAmountTaxInclusive: types.MustDecimal("12100.00"),
					LineExtensionTaxAmount:          types.MustDecimal("2100.00"),
					UnitPrice:                       types.MustDecimal("10000.00"),
					UnitPriceTaxInclusive:           types.MustDecimal("12100.00"),
					ClassifiedTaxCategory:           schema.ClassifiedTaxCategory{Percent: types.MustDecimal("21")},
					Item:                            schema.Item{Description: "Enterprise Software License"},
				},
			},
		},
		TaxTotal: schema.TaxTotal{
			TaxSubTotal: []schema.TaxSubTotal{
				{
					TaxableAmount:                    types.MustDecimal("10000.00"),
					TaxAmount:                        types.MustDecimal("2100.00"),
					TaxInclusiveAmount:               types.MustDecimal("12100.00"),
					AlreadyClaimedTaxableAmount:      types.MustDecimal("0"),
					AlreadyClaimedTaxAmount:          types.MustDecimal("0"),
					AlreadyClaimedTaxInclusiveAmount: types.MustDecimal("0"),
					DifferenceTaxableAmount:          types.MustDecimal("10000.00"),
					DifferenceTaxAmount:              types.MustDecimal("2100.00"),
					DifferenceTaxInclusiveAmount:     types.MustDecimal("12100.00"),
					TaxCategory:                      schema.TaxCategory{Percent: types.MustDecimal("21")},
				},
			},
			TaxAmount: types.MustDecimal("2100.00"),
		},
		LegalMonetaryTotal: schema.LegalMonetaryTotal{
			TaxExclusiveAmount:               types.MustDecimal("10000.00"),
			TaxInclusiveAmount:               types.MustDecimal("12100.00"),
			AlreadyClaimedTaxExclusiveAmount: types.MustDecimal("0"),
			AlreadyClaimedTaxInclusiveAmount: types.MustDecimal("0"),
			DifferenceTaxExclusiveAmount:     types.MustDecimal("10000.00"),
			DifferenceTaxInclusiveAmount:     types.MustDecimal("12100.00"),
			PayableRoundingAmount:            types.MustDecimal("0"),
			PaidDepositsAmount:               types.MustDecimal("0"),
			PayableAmount:                    types.MustDecimal("12100.00"),
		},
	}

	// Access the invoice directly
	fmt.Printf("Document ID: %s\n", invoice.ID)
	fmt.Printf("Document UUID: %s\n", invoice.UUID)

	// Validate with default options
	errors := isdoc.ValidateInvoice(invoice)
	if len(errors) == 0 {
		fmt.Println("Validation: OK")
	} else {
		fmt.Printf("Validation found %d issues\n", len(errors))
	}

	// Convert to XML
	xml, err := isdoc.EncodeBytes(invoice)
	if err != nil {
		log.Fatalf("EncodeBytes error: %v", err)
	}
	fmt.Printf("Generated XML: %d bytes\n\n", len(xml))
}

func strictValidationExample() {
	fmt.Println("=== Strict Validation Example ===")

	// Create a minimal/incomplete invoice
	invoice := &schema.Invoice{
		Version:           "6.0.2",
		DocumentType:      1,
		ID:                "INCOMPLETE-001",
		UUID:              types.UUID("00000000-0000-0000-0000-000000000001"),
		IssueDate:         types.MustParseDate("2024-01-01"),
		VATApplicable:     types.Bool(true),
		LocalCurrencyCode: "CZK",
		CurrRate:          types.MustDecimal("1"),
		RefCurrRate:       types.MustDecimal("1"),
		// Supplier is required
		AccountingSupplierParty: schema.AccountingSupplierParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{ID: "12345678"},
				PartyName:           schema.PartyName{Name: "Minimal Supplier"},
				PostalAddress: schema.PostalAddress{
					StreetName: "Minimal Street",
					CityName:   "Prague",
					PostalZone: "10000",
					Country:    schema.Country{IdentificationCode: "CZ"},
				},
			},
		},
		// AccountingCustomerParty is missing - will trigger warning
		InvoiceLines: schema.InvoiceLines{
			InvoiceLine: []schema.InvoiceLine{
				{
					ID:                              "1",
					LineExtensionAmount:             types.MustDecimal("100.00"),
					LineExtensionAmountTaxInclusive: types.MustDecimal("121.00"),
					LineExtensionTaxAmount:          types.MustDecimal("21.00"),
					UnitPrice:                       types.MustDecimal("100.00"),
					UnitPriceTaxInclusive:           types.MustDecimal("121.00"),
					ClassifiedTaxCategory:           schema.ClassifiedTaxCategory{Percent: types.MustDecimal("21")},
					// Missing: InvoicedQuantity, Item - will trigger warnings
				},
			},
		},
		TaxTotal: schema.TaxTotal{
			// Missing TaxSubTotal - will trigger warning
			TaxAmount: types.MustDecimal("21.00"),
		},
		LegalMonetaryTotal: schema.LegalMonetaryTotal{
			TaxExclusiveAmount:               types.MustDecimal("100.00"),
			TaxInclusiveAmount:               types.MustDecimal("121.00"),
			AlreadyClaimedTaxExclusiveAmount: types.MustDecimal("0"),
			AlreadyClaimedTaxInclusiveAmount: types.MustDecimal("0"),
			DifferenceTaxExclusiveAmount:     types.MustDecimal("100.00"),
			DifferenceTaxInclusiveAmount:     types.MustDecimal("121.00"),
			PayableRoundingAmount:            types.MustDecimal("0"),
			PaidDepositsAmount:               types.MustDecimal("0"),
			PayableAmount:                    types.MustDecimal("121.00"),
		},
	}

	// Validate with default options (non-strict)
	fmt.Println("Default validation:")
	defaultErrors := isdoc.ValidateInvoice(invoice)
	for _, err := range defaultErrors {
		fmt.Printf("  [%s] %s: %s\n", err.Severity, err.Field, err.Msg)
	}
	fmt.Printf("  Total: %d issues, HasErrors: %v\n\n", len(defaultErrors), defaultErrors.HasErrors())

	// Validate with strict options
	fmt.Println("Strict validation:")
	strictErrors := isdoc.ValidateInvoiceWithOptions(invoice, isdoc.ValidateOptions{
		Strict:                 true,
		AllowRoundingTolerance: false,
	})
	for _, err := range strictErrors {
		fmt.Printf("  [%s] %s: %s\n", err.Severity, err.Field, err.Msg)
	}
	fmt.Printf("  Total: %d issues, HasErrors: %v\n\n", len(strictErrors), strictErrors.HasErrors())
}

func roundTripExample() {
	fmt.Println("=== Round-Trip Example ===")

	// Sample ISDOC XML
	originalXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="http://isdoc.cz/namespace/2013" version="6.0.2">
  <DocumentType>1</DocumentType>
  <ID>RT-001</ID>
  <UUID>AAAABBBB-CCCC-DDDD-EEEE-FFFFFFFFFFFF</UUID>
  <IssueDate>2024-04-15</IssueDate>
  <VATApplicable>true</VATApplicable>
  <LocalCurrencyCode>CZK</LocalCurrencyCode>
  <CurrRate>1</CurrRate>
  <RefCurrRate>1</RefCurrRate>
  <AccountingSupplierParty>
    <Party>
      <PartyIdentification><ID>11111111</ID></PartyIdentification>
      <PartyName><Name>Round-Trip Company</Name></PartyName>
      <PostalAddress>
        <StreetName>Test Avenue</StreetName>
        <CityName>Prague</CityName>
        <PostalZone>12000</PostalZone>
        <Country><IdentificationCode>CZ</IdentificationCode></Country>
      </PostalAddress>
    </Party>
  </AccountingSupplierParty>
  <AccountingCustomerParty>
    <Party>
      <PartyIdentification><ID>22222222</ID></PartyIdentification>
      <PartyName><Name>Test Customer</Name></PartyName>
      <PostalAddress>
        <StreetName>Customer Road</StreetName>
        <CityName>Brno</CityName>
        <PostalZone>60000</PostalZone>
        <Country><IdentificationCode>CZ</IdentificationCode></Country>
      </PostalAddress>
    </Party>
  </AccountingCustomerParty>
  <InvoiceLines>
    <InvoiceLine>
      <ID>1</ID>
      <InvoicedQuantity unitCode="C62">2</InvoicedQuantity>
      <LineExtensionAmount>200.00</LineExtensionAmount>
      <LineExtensionAmountTaxInclusive>242.00</LineExtensionAmountTaxInclusive>
      <LineExtensionTaxAmount>42.00</LineExtensionTaxAmount>
      <UnitPrice>100.00</UnitPrice>
      <UnitPriceTaxInclusive>121.00</UnitPriceTaxInclusive>
      <ClassifiedTaxCategory><Percent>21</Percent></ClassifiedTaxCategory>
      <Item><Description>Test Product</Description></Item>
    </InvoiceLine>
  </InvoiceLines>
  <TaxTotal>
    <TaxSubTotal>
      <TaxableAmount>200.00</TaxableAmount>
      <TaxAmount>42.00</TaxAmount>
      <TaxInclusiveAmount>242.00</TaxInclusiveAmount>
      <AlreadyClaimedTaxableAmount>0</AlreadyClaimedTaxableAmount>
      <AlreadyClaimedTaxAmount>0</AlreadyClaimedTaxAmount>
      <AlreadyClaimedTaxInclusiveAmount>0</AlreadyClaimedTaxInclusiveAmount>
      <DifferenceTaxableAmount>200.00</DifferenceTaxableAmount>
      <DifferenceTaxAmount>42.00</DifferenceTaxAmount>
      <DifferenceTaxInclusiveAmount>242.00</DifferenceTaxInclusiveAmount>
      <TaxCategory><Percent>21</Percent></TaxCategory>
    </TaxSubTotal>
    <TaxAmount>42.00</TaxAmount>
  </TaxTotal>
  <LegalMonetaryTotal>
    <TaxExclusiveAmount>200.00</TaxExclusiveAmount>
    <TaxInclusiveAmount>242.00</TaxInclusiveAmount>
    <AlreadyClaimedTaxExclusiveAmount>0</AlreadyClaimedTaxExclusiveAmount>
    <AlreadyClaimedTaxInclusiveAmount>0</AlreadyClaimedTaxInclusiveAmount>
    <DifferenceTaxExclusiveAmount>200.00</DifferenceTaxExclusiveAmount>
    <DifferenceTaxInclusiveAmount>242.00</DifferenceTaxInclusiveAmount>
    <PayableRoundingAmount>0</PayableRoundingAmount>
    <PaidDepositsAmount>0</PaidDepositsAmount>
    <PayableAmount>242.00</PayableAmount>
  </LegalMonetaryTotal>
</Invoice>`)

	// Parse original XML
	invoice, err := isdoc.DecodeBytes(originalXML)
	if err != nil {
		log.Fatalf("DecodeBytes error: %v", err)
	}

	// Validate
	validationErrors := isdoc.ValidateInvoice(invoice)

	// Encode back
	outputXML, err := isdoc.EncodeBytes(invoice)
	if err != nil {
		log.Fatalf("EncodeBytes error: %v", err)
	}

	fmt.Printf("Original XML: %d bytes\n", len(originalXML))
	fmt.Printf("Output XML: %d bytes\n", len(outputXML))
	fmt.Printf("Validation issues: %d\n", len(validationErrors))

	// Optionally save to file
	outputPath := "/tmp/roundtrip-output.isdoc"
	if err := os.WriteFile(outputPath, outputXML, 0644); err != nil {
		log.Printf("Could not write output: %v", err)
	} else {
		fmt.Printf("Output saved to: %s\n", outputPath)
	}
}
