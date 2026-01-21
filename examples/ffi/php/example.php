<?php
$libExt = match (PHP_OS_FAMILY) {
    'Darwin' => 'dylib',
    'Windows' => 'dll',
    default => 'so',
};

$ffi = FFI::cdef("
    char* isdoc_parse(char* xmlData);
    char* isdoc_validate(char* xmlData);
    void isdoc_free(char* ptr);
", __DIR__ . "/../../../bin/libisdoc.{$libExt}");

$xml = <<<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="http://isdoc.cz/namespace/2013" version="6.0.2">
  <DocumentType>1</DocumentType>
  <ID>FV-2025-001</ID>
  <UUID>550e8400-e29b-41d4-a716-446655440000</UUID>
  <IssueDate>2025-01-19</IssueDate>
  <TaxPointDate>2025-01-19</TaxPointDate>
  <VATApplicable>true</VATApplicable>
  <LocalCurrencyCode>CZK</LocalCurrencyCode>
  <CurrRate>1</CurrRate>
  <RefCurrRate>1</RefCurrRate>
  <AccountingSupplierParty>
    <Party>
      <PartyIdentification><ID>12345678</ID></PartyIdentification>
      <PartyName><Name>Test Supplier</Name></PartyName>
      <PostalAddress>
        <StreetName>Test Street</StreetName>
        <BuildingNumber>123</BuildingNumber>
        <CityName>Prague</CityName>
        <PostalZone>11000</PostalZone>
        <Country><IdentificationCode>CZ</IdentificationCode></Country>
      </PostalAddress>
    </Party>
  </AccountingSupplierParty>
  <AccountingCustomerParty>
    <Party>
      <PartyIdentification><ID>87654321</ID></PartyIdentification>
      <PartyName><Name>Test Customer</Name></PartyName>
      <PostalAddress>
        <StreetName>Customer Road</StreetName>
        <BuildingNumber>456</BuildingNumber>
        <CityName>Brno</CityName>
        <PostalZone>60200</PostalZone>
        <Country><IdentificationCode>CZ</IdentificationCode></Country>
      </PostalAddress>
    </Party>
  </AccountingCustomerParty>
  <InvoiceLines>
    <InvoiceLine>
      <ID>1</ID>
      <InvoicedQuantity unitCode="PCE">1</InvoicedQuantity>
      <LineExtensionAmount>1000.00</LineExtensionAmount>
      <LineExtensionAmountTaxInclusive>1210.00</LineExtensionAmountTaxInclusive>
      <LineExtensionTaxAmount>210.00</LineExtensionTaxAmount>
      <UnitPrice>1000.00</UnitPrice>
      <UnitPriceTaxInclusive>1210.00</UnitPriceTaxInclusive>
      <ClassifiedTaxCategory>
        <Percent>21</Percent>
        <VATCalculationMethod>0</VATCalculationMethod>
      </ClassifiedTaxCategory>
      <Item><Description>Test Product</Description></Item>
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
</Invoice>
XML;

$result = $ffi->isdoc_parse($xml);
$parsed = FFI::string($result);
$ffi->isdoc_free($result);

$data = json_decode($parsed, true);
echo "Parsed Invoice ID: " . ($data['ID'] ?? 'N/A') . "\n";

$result = $ffi->isdoc_validate($xml);
$validated = FFI::string($result);
$ffi->isdoc_free($result);

$validation = json_decode($validated, true);
echo "Valid: " . ($validation['valid'] ? 'true' : 'false') . "\n";
