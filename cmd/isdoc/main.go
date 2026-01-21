// Command isdoc provides CLI operations for ISDOC documents.
//
// Usage:
//
//	isdoc extract input.pdf [output.isdoc]  - Extract ISDOC XML from PDF
//	isdoc embed input.pdf invoice.isdoc [output.pdf] - Embed ISDOC into PDF
//	isdoc validate input.isdoc              - Validate ISDOC XML
//	isdoc convert input.isdoc output.json   - Convert ISDOC to JSON
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xseman/isdoc"
	"github.com/xseman/isdoc/pdf"
)

const (
	exitSuccess = 0
	exitError   = 1
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return exitError
	}

	cmd := args[0]

	if cmd == "-h" || cmd == "--help" || cmd == "help" {
		printUsage(stdout)
		return exitSuccess
	}

	switch cmd {
	case "extract":
		return cmdExtract(args[1:], stdout, stderr)
	case "embed":
		return cmdEmbed(args[1:], stderr)
	case "validate":
		return cmdValidate(args[1:], stdin, stdout, stderr)
	case "convert":
		return cmdConvert(args[1:], stdin, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", cmd)
		printUsage(stderr)
		return exitError
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `Usage: isdoc <command> [arguments]

Commands:
  extract   Extract ISDOC XML from a PDF file
  embed     Embed ISDOC XML into a PDF file
  validate  Validate an ISDOC XML document
  convert   Convert ISDOC XML to JSON format

Use "isdoc <command> -h" for more information about a command.`)
}

func cmdExtract(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("extract", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, `Usage: isdoc extract <input.pdf> [output.isdoc]

Extract ISDOC XML from a PDF file.

Arguments:
  input.pdf     PDF file containing embedded ISDOC XML
  output.isdoc  Output file for extracted XML (optional, defaults to stdout)`)
	}

	if err := fs.Parse(args); err != nil {
		return exitError
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(stderr, "error: missing input PDF file")
		fs.Usage()
		return exitError
	}

	inputPath := fs.Arg(0)
	outputPath := fs.Arg(1)

	xmlData, err := pdf.ExtractXML(inputPath)
	if err != nil {
		fmt.Fprintf(stderr, "error: extracting ISDOC from PDF: %v\n", err)
		return exitError
	}

	if outputPath == "" || outputPath == "-" {
		if _, err := stdout.Write(xmlData); err != nil {
			fmt.Fprintf(stderr, "error: writing to stdout: %v\n", err)
			return exitError
		}
	} else {
		if err := os.WriteFile(outputPath, xmlData, 0644); err != nil {
			fmt.Fprintf(stderr, "error: writing output file: %v\n", err)
			return exitError
		}
		fmt.Fprintf(stderr, "extracted ISDOC XML to %s\n", outputPath)
	}

	return exitSuccess
}

func cmdEmbed(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("embed", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, `Usage: isdoc embed <input.pdf> <invoice.isdoc> [output.pdf]

Embed ISDOC XML into a PDF file.

Arguments:
  input.pdf      Source PDF file
  invoice.isdoc  ISDOC XML file to embed
  output.pdf     Output PDF file (optional, defaults to input.pdf)`)
	}

	if err := fs.Parse(args); err != nil {
		return exitError
	}

	if fs.NArg() < 2 {
		fmt.Fprintln(stderr, "error: missing required arguments")
		fs.Usage()
		return exitError
	}

	inputPDF := fs.Arg(0)
	inputISDOC := fs.Arg(1)
	outputPDF := fs.Arg(2)

	if outputPDF == "" {
		outputPDF = inputPDF
	}

	var xmlData []byte
	var err error
	if inputISDOC == "-" {
		xmlData, err = io.ReadAll(os.Stdin)
	} else {
		xmlData, err = os.ReadFile(inputISDOC)
	}
	if err != nil {
		fmt.Fprintf(stderr, "error: reading ISDOC XML: %v\n", err)
		return exitError
	}

	if err := pdf.EmbedXML(inputPDF, outputPDF, xmlData); err != nil {
		fmt.Fprintf(stderr, "error: embedding ISDOC into PDF: %v\n", err)
		return exitError
	}

	fmt.Fprintf(stderr, "embedded ISDOC XML into %s\n", outputPDF)
	return exitSuccess
}

func cmdValidate(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	strict := fs.Bool("strict", false, "Enable strict validation mode")
	fs.Usage = func() {
		fmt.Fprintln(stderr, `Usage: isdoc validate [options] <input.isdoc>

Validate an ISDOC XML document.

Options:
  -strict  Enable strict validation mode`)
	}

	if err := fs.Parse(args); err != nil {
		return exitError
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(stderr, "error: missing input file")
		fs.Usage()
		return exitError
	}

	inputPath := fs.Arg(0)

	var data []byte
	var err error
	if inputPath == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(inputPath)
	}
	if err != nil {
		fmt.Fprintf(stderr, "error: reading input: %v\n", err)
		return exitError
	}

	invoice, err := isdoc.DecodeBytes(data)
	if err != nil {
		fmt.Fprintf(stderr, "error: parsing ISDOC: %v\n", err)
		return exitError
	}

	opts := isdoc.DefaultValidateOptions()
	opts.Strict = *strict

	errs := isdoc.ValidateInvoiceWithOptions(invoice, opts)

	if len(errs) == 0 {
		fmt.Fprintln(stdout, "validation passed")
		return exitSuccess
	}

	hasErrors := false
	for _, e := range errs {
		if e.Severity == isdoc.SeverityError {
			hasErrors = true
		}
		fmt.Fprintf(stdout, "[%s] %s: %s\n", e.Severity, e.Field, e.Msg)
	}

	if hasErrors {
		fmt.Fprintf(stdout, "\nvalidation failed with %d error(s)\n", len(errs.Errors()))
		return exitError
	}

	fmt.Fprintf(stdout, "\nvalidation passed with %d warning(s)\n", len(errs))
	return exitSuccess
}

func cmdConvert(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	fs.SetOutput(stderr)
	indent := fs.Bool("indent", true, "Indent JSON output")
	fs.Usage = func() {
		fmt.Fprintln(stderr, `Usage: isdoc convert [options] <input.isdoc> [output.json]

Convert ISDOC XML to JSON format.

Options:
  -indent  Indent JSON output (default: true)`)
	}

	if err := fs.Parse(args); err != nil {
		return exitError
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(stderr, "error: missing input file")
		fs.Usage()
		return exitError
	}

	inputPath := fs.Arg(0)
	outputPath := fs.Arg(1)

	var data []byte
	var err error
	if inputPath == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(inputPath)
	}
	if err != nil {
		fmt.Fprintf(stderr, "error: reading input: %v\n", err)
		return exitError
	}

	invoice, err := isdoc.DecodeBytes(data)
	if err != nil {
		fmt.Fprintf(stderr, "error: parsing ISDOC: %v\n", err)
		return exitError
	}

	var jsonData []byte
	if *indent {
		jsonData, err = json.MarshalIndent(invoice, "", "  ")
	} else {
		jsonData, err = json.Marshal(invoice)
	}
	if err != nil {
		fmt.Fprintf(stderr, "error: converting to JSON: %v\n", err)
		return exitError
	}

	if outputPath == "" || outputPath == "-" {
		stdout.Write(jsonData)
		fmt.Fprintln(stdout)
	} else {
		if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
			fmt.Fprintf(stderr, "error: writing output file: %v\n", err)
			return exitError
		}
		baseName := filepath.Base(inputPath)
		if inputPath == "-" {
			baseName = "stdin"
		}
		ext := strings.TrimPrefix(filepath.Ext(outputPath), ".")
		fmt.Fprintf(stderr, "converted %s to %s\n", baseName, ext)
	}

	return exitSuccess
}
