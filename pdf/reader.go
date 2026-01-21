package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// ErrNoISDOCFound is returned when no ISDOC XML is found in the PDF.
var ErrNoISDOCFound = errors.New("no ISDOC XML found in PDF")

// Attachment represents an embedded file extracted from a PDF.
type Attachment struct {
	Name string
	Data []byte
}

// ReadResult contains the extracted ISDOC XML and any supplements.
type ReadResult struct {
	// XML is the extracted ISDOC XML content.
	XML []byte
	// Supplements contains any additional embedded files.
	Supplements []Attachment
}

// Reader extracts ISDOC XML from PDF files.
type Reader struct{}

// NewReader creates a new PDF reader.
func NewReader() *Reader {
	return &Reader{}
}

// ReadFile extracts ISDOC XML from a PDF file.
func (r *Reader) ReadFile(filename string) (*ReadResult, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	return r.Read(f)
}

// Read extracts ISDOC XML from a PDF reader.
func (r *Reader) Read(rs io.ReadSeeker) (*ReadResult, error) {
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	// Extract attachments using pdfcpu API
	attachments, err := pdfcpuapi.ExtractAttachmentsRaw(rs, "", nil, conf)
	if err != nil {
		return nil, fmt.Errorf("extracting attachments: %w", err)
	}

	var result ReadResult

	for _, att := range attachments {
		content := att.Reader
		if content == nil {
			continue
		}

		data, err := io.ReadAll(content)
		if err != nil {
			continue
		}

		// Check if this is the ISDOC XML
		if r.isISDOCXML(data) {
			result.XML = data
			continue
		}

		// Otherwise treat as a supplement
		result.Supplements = append(result.Supplements, Attachment{
			Name: att.FileName,
			Data: data,
		})
	}

	if result.XML == nil {
		return nil, ErrNoISDOCFound
	}

	return &result, nil
}

// isISDOCXML checks if the data appears to be ISDOC XML.
func (r *Reader) isISDOCXML(data []byte) bool {
	// Size sanity check - ISDOC shouldn't be larger than ~256KB
	if len(data) > 1<<18 {
		return false
	}

	// Check for XML signature
	if !bytes.HasPrefix(data, []byte("<?xml")) {
		return false
	}

	// Check for Invoice element
	return bytes.Contains(data, []byte("<Invoice"))
}

// ExtractXML is a convenience function to extract ISDOC XML from a file.
func ExtractXML(filename string) ([]byte, error) {
	r := NewReader()
	result, err := r.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return result.XML, nil
}
