// Package archive provides ISDOCX archive (ZIP) support for ISDOC documents.
//
// ISDOCX is a ZIP archive format containing an ISDOC XML document and optional
// attachments. The archive structure is defined in the ISDOC 6.0.2 specification
// Section 3.3.
//
// Archive structure:
//   - manifest.xml (optional, recommended) - contains path to main document
//   - *.isdoc - the main ISDOC document
//   - attachments/ (optional) - supplementary files
//
// Legacy archives (pre-6.0) may not have a manifest.xml; in this case,
// the .isdoc file in the root directory is used as the main document.
package archive

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

// ManifestNamespace is the XML namespace for ISDOCX manifest files.
const ManifestNamespace = "http://isdoc.cz/namespace/2013/manifest"

// ManifestFilename is the standard name for the manifest file.
const ManifestFilename = "manifest.xml"

// Manifest represents the manifest.xml structure in an ISDOCX archive.
type Manifest struct {
	XMLName      xml.Name     `xml:"manifest"`
	MainDocument MainDocument `xml:"maindocument"`
}

// MainDocument contains the reference to the main ISDOC document.
type MainDocument struct {
	Filename string `xml:"filename,attr"`
}

// Archive represents an ISDOCX archive.
type Archive struct {
	// Manifest is the parsed manifest.xml (may be nil for legacy archives).
	Manifest *Manifest

	// MainDocumentPath is the path to the main ISDOC document within the archive.
	MainDocumentPath string

	// MainDocumentData is the raw XML content of the main ISDOC document.
	MainDocumentData []byte

	// Attachments contains paths and data of supplementary files.
	Attachments map[string][]byte
}

// ErrNoISDOCDocument is returned when no .isdoc file is found in the archive.
var ErrNoISDOCDocument = errors.New("no ISDOC document found in archive")

// ErrInvalidManifest is returned when the manifest.xml cannot be parsed.
var ErrInvalidManifest = errors.New("invalid manifest.xml")

// ReadFile reads an ISDOCX archive from a file path.
func ReadFile(path string) (*Archive, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat archive: %w", err)
	}

	return Read(f, stat.Size())
}

// Read reads an ISDOCX archive from an io.ReaderAt.
func Read(r io.ReaderAt, size int64) (*Archive, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	return readZip(zr)
}

// ReadBytes reads an ISDOCX archive from a byte slice.
func ReadBytes(data []byte) (*Archive, error) {
	r := bytes.NewReader(data)
	return Read(r, int64(len(data)))
}

// readZip extracts data from a zip.Reader.
func readZip(zr *zip.Reader) (*Archive, error) {
	arc := &Archive{
		Attachments: make(map[string][]byte),
	}

	// First pass: look for manifest.xml and collect all files
	var manifestFile *zip.File
	var isdocFiles []*zip.File

	for _, f := range zr.File {
		name := f.Name
		baseName := path.Base(name)

		if baseName == ManifestFilename && !f.FileInfo().IsDir() {
			manifestFile = f
		} else if strings.HasSuffix(strings.ToLower(name), ".isdoc") && !f.FileInfo().IsDir() {
			isdocFiles = append(isdocFiles, f)
		}
	}

	// Parse manifest if present
	if manifestFile != nil {
		data, err := readZipFile(manifestFile)
		if err != nil {
			return nil, fmt.Errorf("read manifest: %w", err)
		}

		manifest := &Manifest{}
		if err := xml.Unmarshal(data, manifest); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidManifest, err)
		}
		arc.Manifest = manifest
		arc.MainDocumentPath = manifest.MainDocument.Filename
	}

	// Find main document
	if arc.MainDocumentPath == "" {
		// Legacy mode: find .isdoc file in root
		for _, f := range isdocFiles {
			// Prefer file in root (no directory)
			if !strings.Contains(f.Name, "/") {
				arc.MainDocumentPath = f.Name
				break
			}
		}
		// If no root file, use first .isdoc found
		if arc.MainDocumentPath == "" && len(isdocFiles) > 0 {
			arc.MainDocumentPath = isdocFiles[0].Name
		}
	}

	if arc.MainDocumentPath == "" {
		return nil, ErrNoISDOCDocument
	}

	// Second pass: read all files
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}

		data, err := readZipFile(f)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", f.Name, err)
		}

		if f.Name == arc.MainDocumentPath {
			arc.MainDocumentData = data
		} else if f.Name != ManifestFilename {
			arc.Attachments[f.Name] = data
		}
	}

	if arc.MainDocumentData == nil {
		return nil, fmt.Errorf("main document %q not found in archive", arc.MainDocumentPath)
	}

	return arc, nil
}

// readZipFile reads the content of a zip file entry.
func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
}

// WriteFile writes an ISDOCX archive to a file path.
func (a *Archive) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	defer f.Close()

	return a.Write(f)
}

// Write writes an ISDOCX archive to an io.Writer.
func (a *Archive) Write(w io.Writer) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	// Write manifest.xml
	manifestPath := a.MainDocumentPath
	if manifestPath == "" {
		manifestPath = "invoice.isdoc"
	}

	manifest := &Manifest{
		XMLName: xml.Name{
			Space: ManifestNamespace,
			Local: "manifest",
		},
		MainDocument: MainDocument{
			Filename: manifestPath,
		},
	}

	manifestData, err := xml.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	manifestData = append([]byte(xml.Header), manifestData...)

	if err := writeZipFile(zw, ManifestFilename, manifestData); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	// Write main document
	docPath := manifestPath
	if a.MainDocumentPath != "" {
		docPath = a.MainDocumentPath
	}
	if err := writeZipFile(zw, docPath, a.MainDocumentData); err != nil {
		return fmt.Errorf("write main document: %w", err)
	}

	// Write attachments
	for name, data := range a.Attachments {
		if err := writeZipFile(zw, name, data); err != nil {
			return fmt.Errorf("write attachment %s: %w", name, err)
		}
	}

	return nil
}

// WriteBytes writes an ISDOCX archive to a byte slice.
func (a *Archive) WriteBytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := a.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// writeZipFile writes a single file to a zip archive.
func writeZipFile(zw *zip.Writer, name string, data []byte) error {
	// Use UTF-8 filename encoding (Language encoding flag)
	header := &zip.FileHeader{
		Name:   name,
		Method: zip.Deflate,
	}
	header.SetMode(0644)

	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// NewArchive creates a new Archive with the given ISDOC document data.
func NewArchive(isdocData []byte, filename string) *Archive {
	if filename == "" {
		filename = "invoice.isdoc"
	}
	return &Archive{
		MainDocumentPath: filename,
		MainDocumentData: isdocData,
		Attachments:      make(map[string][]byte),
	}
}

// AddAttachment adds an attachment to the archive.
func (a *Archive) AddAttachment(name string, data []byte) {
	if a.Attachments == nil {
		a.Attachments = make(map[string][]byte)
	}
	a.Attachments[name] = data
}
