package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/xseman/isdoc"
	"github.com/xseman/isdoc/schema"
)

// Version of the library
const version = "0.1.0"

//export isdoc_parse
func isdoc_parse(xmlData *C.char) *C.char {
	input := C.GoString(xmlData)

	invoice, err := isdoc.DecodeBytes([]byte(input))
	if err != nil {
		errJSON := fmt.Sprintf(`{"error": "DecodeBytes error: %s"}`, err.Error())
		return C.CString(errJSON)
	}

	output, err := json.Marshal(invoice)
	if err != nil {
		errJSON := fmt.Sprintf(`{"error": "JSON marshal error: %s"}`, err.Error())
		return C.CString(errJSON)
	}

	return C.CString(string(output))
}

//export isdoc_parse_common_document
func isdoc_parse_common_document(xmlData *C.char) *C.char {
	input := C.GoString(xmlData)

	doc, err := isdoc.DecodeCommonDocumentBytes([]byte(input))
	if err != nil {
		errJSON := fmt.Sprintf(`{"error": "DecodeBytes error: %s"}`, err.Error())
		return C.CString(errJSON)
	}

	output, err := json.Marshal(doc)
	if err != nil {
		errJSON := fmt.Sprintf(`{"error": "JSON marshal error: %s"}`, err.Error())
		return C.CString(errJSON)
	}

	return C.CString(string(output))
}

//export isdoc_validate
func isdoc_validate(xmlData *C.char) *C.char {
	input := C.GoString(xmlData)

	invoice, err := isdoc.DecodeBytes([]byte(input))
	if err != nil {
		errJSON := fmt.Sprintf(`{"error": "DecodeBytes error: %s"}`, err.Error())
		return C.CString(errJSON)
	}

	errs := isdoc.ValidateInvoice(invoice)

	type ValidationResult struct {
		Valid  bool     `json:"valid"`
		Errors []string `json:"errors,omitempty"`
	}

	result := ValidationResult{
		Valid:  !errs.HasErrors(),
		Errors: make([]string, 0),
	}

	for _, e := range errs.Errors() {
		result.Errors = append(result.Errors, e.Error())
	}

	output, err := json.Marshal(result)
	if err != nil {
		errJSON := fmt.Sprintf(`{"error": "JSON marshal error: %s"}`, err.Error())
		return C.CString(errJSON)
	}

	return C.CString(string(output))
}

//export isdoc_marshal
func isdoc_marshal(jsonData *C.char) *C.char {
	input := C.GoString(jsonData)

	var invoice schema.Invoice
	if err := json.Unmarshal([]byte(input), &invoice); err != nil {
		errJSON := fmt.Sprintf(`{"error": "JSON parse error: %s"}`, err.Error())
		return C.CString(errJSON)
	}

	xmlOutput, err := isdoc.EncodeBytes(&invoice)
	if err != nil {
		errJSON := fmt.Sprintf(`{"error": "EncodeBytes error: %s"}`, err.Error())
		return C.CString(errJSON)
	}

	return C.CString(string(xmlOutput))
}

//export isdoc_free
func isdoc_free(ptr *C.char) {
	C.free(unsafe.Pointer(ptr))
}

//export isdoc_version
func isdoc_version() *C.char {
	return C.CString(version)
}

func main() {}
