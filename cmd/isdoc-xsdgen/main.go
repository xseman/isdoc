// Command isdoc-xsdgen parses ISDOC XSD schema and generates Go ordering maps.
//
// The schema is downloaded from the official ISDOC repository.
// To update to a new version, change the ISDOC_VERSION constant and regenerate.
//
// Usage:
//
//	go run ./cmd/isdoc-xsdgen -out internal/ordering/sequences.go
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"go/format"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ISDOC_VERSION defines the schema version to download and use.
// To update: change this version and run `go generate ./internal/ordering`
const ISDOC_VERSION = "6.0.2"

// ISDOC_SCHEMA_URL_TEMPLATE is the URL pattern for official schemas
const ISDOC_SCHEMA_URL_TEMPLATE = "https://isdoc.github.io/xsd/isdoc-invoice-%s.xsd"

// XSD structures for parsing
type Schema struct {
	XMLName      xml.Name      `xml:"schema"`
	ComplexTypes []ComplexType `xml:"complexType"`
	Elements     []Element     `xml:"element"`
	Groups       []Group       `xml:"group"`
}

type ComplexType struct {
	Name     string   `xml:"name,attr"`
	Sequence Sequence `xml:"sequence"`
	Choice   Choice   `xml:"choice"`
	All      All      `xml:"all"`
}

type Group struct {
	Name     string   `xml:"name,attr"`
	Sequence Sequence `xml:"sequence"`
}

type Sequence struct {
	Elements  []Element  `xml:"element"`
	Sequences []Sequence `xml:"sequence"`
	Choices   []Choice   `xml:"choice"`
	Groups    []GroupRef `xml:"group"`
}

type Choice struct {
	Elements  []Element  `xml:"element"`
	Sequences []Sequence `xml:"sequence"`
}

type All struct {
	Elements []Element `xml:"element"`
}

type Element struct {
	Name        string      `xml:"name,attr"`
	Type        string      `xml:"type,attr"`
	Ref         string      `xml:"ref,attr"`
	MinOccurs   string      `xml:"minOccurs,attr"`
	ComplexType ComplexType `xml:"complexType"`
}

type GroupRef struct {
	Ref string `xml:"ref,attr"`
}

func main() {
	outputPath := flag.String("out", "", "Output Go file path")
	flag.Parse()

	// Download schema
	schemaURL := fmt.Sprintf(ISDOC_SCHEMA_URL_TEMPLATE, ISDOC_VERSION)
	fmt.Fprintf(os.Stderr, "Downloading ISDOC v%s schema from %s...\n", ISDOC_VERSION, schemaURL)

	data, err := downloadSchema(schemaURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading schema: %v\n", err)
		os.Exit(1)
	}

	// Extract version from downloaded schema to verify
	downloadedVersion := extractVersion(data)
	if downloadedVersion != "" && downloadedVersion != ISDOC_VERSION {
		fmt.Fprintf(os.Stderr, "Warning: Downloaded schema version %s differs from expected %s\n", downloadedVersion, ISDOC_VERSION)
	}

	// Parse XSD
	var schema Schema
	if err := xml.Unmarshal(data, &schema); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing schema: %v\n", err)
		os.Exit(1)
	}

	// Extract element sequences
	sequences := make(map[string][]string)

	// Process named complex types
	for _, ct := range schema.ComplexTypes {
		if ct.Name == "" {
			continue
		}
		elements := extractElements(ct.Sequence, ct.Choice, ct.All)
		if len(elements) > 0 {
			sequences[ct.Name] = elements
		}
	}

	// Process root elements with inline complex types
	for _, elem := range schema.Elements {
		if elem.Name == "" {
			continue
		}
		elements := extractElements(elem.ComplexType.Sequence, elem.ComplexType.Choice, elem.ComplexType.All)
		if len(elements) > 0 {
			sequences[elem.Name] = elements
		}
	}

	// Process groups
	for _, g := range schema.Groups {
		if g.Name == "" {
			continue
		}
		elements := extractSequenceElements(g.Sequence)
		if len(elements) > 0 {
			sequences[g.Name] = elements
		}
	}

	// Generate Go code
	code := generateGoCode(sequences)

	// Format the code
	formatted, err := format.Source([]byte(code))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting code: %v\n", err)
		fmt.Println(code) // Print unformatted for debugging
		os.Exit(1)
	}

	// Output
	if *outputPath == "" {
		fmt.Print(string(formatted))
	} else {
		dir := filepath.Dir(*outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(*outputPath, formatted, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated %s\n", *outputPath)
	}
}

func downloadSchema(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return data, nil
}

func extractVersion(data []byte) string {
	// Extract version attribute from schema element
	re := regexp.MustCompile(`version="([0-9.]+)"`)
	matches := re.FindSubmatch(data)
	if len(matches) > 1 {
		return string(matches[1])
	}
	return ""
}

func extractElements(seq Sequence, choice Choice, all All) []string {
	var elements []string

	// From sequence
	elements = append(elements, extractSequenceElements(seq)...)

	// From choice (add all possible elements, they might appear)
	elements = append(elements, extractChoiceElements(choice)...)

	// From all
	for _, e := range all.Elements {
		if e.Name != "" {
			elements = append(elements, e.Name)
		}
	}

	return elements
}

func extractSequenceElements(seq Sequence) []string {
	var elements []string

	for _, e := range seq.Elements {
		if e.Name != "" {
			elements = append(elements, e.Name)
		}
	}

	// Nested sequences
	for _, nested := range seq.Sequences {
		elements = append(elements, extractSequenceElements(nested)...)
	}

	// Choices within sequence
	for _, c := range seq.Choices {
		elements = append(elements, extractChoiceElements(c)...)
	}

	return elements
}

func extractChoiceElements(choice Choice) []string {
	var elements []string

	for _, e := range choice.Elements {
		if e.Name != "" {
			elements = append(elements, e.Name)
		}
	}

	for _, seq := range choice.Sequences {
		elements = append(elements, extractSequenceElements(seq)...)
	}

	return elements
}

func generateGoCode(sequences map[string][]string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`// Code generated by isdoc-xsdgen. DO NOT EDIT.
// Source: ISDOC XSD schema v%s
// URL: %s

package ordering

// Sequence defines the required element ordering for each complex type.
// Elements must appear in this order when encoding to XML.
var Sequence = map[string][]string{
`, ISDOC_VERSION, fmt.Sprintf(ISDOC_SCHEMA_URL_TEMPLATE, ISDOC_VERSION)))

	// Sort type names for deterministic output
	typeNames := make([]string, 0, len(sequences))
	for name := range sequences {
		typeNames = append(typeNames, name)
	}
	sort.Strings(typeNames)

	for _, name := range typeNames {
		elements := sequences[name]
		if len(elements) == 0 {
			continue
		}

		b.WriteString(fmt.Sprintf("\t%q: {", name))
		for i, elem := range elements {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%q", elem))
		}
		b.WriteString("},\n")
	}

	b.WriteString("}\n")

	return b.String()
}
