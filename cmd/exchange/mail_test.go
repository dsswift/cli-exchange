package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dsswift/cli-exchange/internal/graph"
)

func writeTestFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("writing test file %s: %v", path, err)
	}
}

func TestLoadAttachments(t *testing.T) {
	dir := t.TempDir()

	pdfPath := filepath.Join(dir, "report.pdf")
	writeTestFile(t, pdfPath, []byte("fake-pdf-content"))

	csvPath := filepath.Join(dir, "data.csv")
	writeTestFile(t, csvPath, []byte("a,b,c"))

	attachments, err := loadAttachments([]string{pdfPath, csvPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attachments) != 2 {
		t.Fatalf("expected 2 attachments, got %d", len(attachments))
	}

	if attachments[0].Name != "report.pdf" {
		t.Errorf("expected name report.pdf, got %q", attachments[0].Name)
	}
	if attachments[0].ContentType != "application/pdf" {
		t.Errorf("expected content type application/pdf, got %q", attachments[0].ContentType)
	}
	if string(attachments[0].ContentBytes) != "fake-pdf-content" {
		t.Errorf("unexpected content bytes")
	}

	if attachments[1].Name != "data.csv" {
		t.Errorf("expected name data.csv, got %q", attachments[1].Name)
	}
	if !strings.HasPrefix(attachments[1].ContentType, "text/csv") {
		t.Errorf("expected content type starting with text/csv, got %q", attachments[1].ContentType)
	}
}

func TestLoadAttachments_UnknownExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.qzx")
	writeTestFile(t, path, []byte("data"))

	attachments, err := loadAttachments([]string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attachments[0].ContentType != "application/octet-stream" {
		t.Errorf("expected application/octet-stream, got %q", attachments[0].ContentType)
	}
}

func TestLoadAttachments_TooLarge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.bin")

	data := make([]byte, 3*1024*1024+1)
	writeTestFile(t, path, data)

	_, err := loadAttachments([]string{path})
	if err == nil {
		t.Fatal("expected error for oversized file")
	}
	if !strings.Contains(err.Error(), "exceeds 3MB limit") {
		t.Errorf("expected 3MB limit error, got: %v", err)
	}
}

func TestLoadAttachments_FileNotFound(t *testing.T) {
	_, err := loadAttachments([]string{"/nonexistent/file.pdf"})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadAttachments_ExactlyAtLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exact.bin")

	data := make([]byte, 3*1024*1024)
	writeTestFile(t, path, data)

	_, err := loadAttachments([]string{path})
	if err != nil {
		t.Fatalf("expected no error for exactly 3MB file, got: %v", err)
	}
}

func TestFilterAttachments_ByName(t *testing.T) {
	attachments := []graph.AttachmentInfo{
		{Name: "report.pdf", IsInline: false},
		{Name: "invoice.pdf", IsInline: false},
		{Name: "logo.png", IsInline: true},
	}

	result := filterAttachments(attachments, "report", false)
	if len(result) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(result))
	}
	if result[0].Name != "report.pdf" {
		t.Errorf("expected report.pdf, got %q", result[0].Name)
	}
}

func TestFilterAttachments_NoInline(t *testing.T) {
	attachments := []graph.AttachmentInfo{
		{Name: "report.pdf", IsInline: false},
		{Name: "logo.png", IsInline: true},
	}

	result := filterAttachments(attachments, "", true)
	if len(result) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(result))
	}
	if result[0].Name != "report.pdf" {
		t.Errorf("expected report.pdf, got %q", result[0].Name)
	}
}

func TestFilterAttachments_CaseInsensitiveName(t *testing.T) {
	attachments := []graph.AttachmentInfo{
		{Name: "Report.PDF", IsInline: false},
	}

	result := filterAttachments(attachments, "report", false)
	if len(result) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(result))
	}
}

func TestFilterAttachments_NoMatch(t *testing.T) {
	attachments := []graph.AttachmentInfo{
		{Name: "report.pdf", IsInline: false},
	}

	result := filterAttachments(attachments, "invoice", false)
	if len(result) != 0 {
		t.Errorf("expected 0 attachments, got %d", len(result))
	}
}

func TestResolveDownloadPath_NoConflict(t *testing.T) {
	dir := t.TempDir()
	path := resolveDownloadPath(dir, "report.pdf")
	expected := filepath.Join(dir, "report.pdf")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestResolveDownloadPath_Conflict(t *testing.T) {
	dir := t.TempDir()

	// Create the original file
	original := filepath.Join(dir, "report.pdf")
	writeTestFile(t, original, []byte("original"))

	path := resolveDownloadPath(dir, "report.pdf")
	expected := filepath.Join(dir, "report (1).pdf")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestResolveDownloadPath_MultipleConflicts(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, filepath.Join(dir, "report.pdf"), []byte("original"))
	writeTestFile(t, filepath.Join(dir, "report (1).pdf"), []byte("first"))

	path := resolveDownloadPath(dir, "report.pdf")
	expected := filepath.Join(dir, "report (2).pdf")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestResolveDownloadPath_NoExtension(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, filepath.Join(dir, "README"), []byte("original"))

	path := resolveDownloadPath(dir, "README")
	expected := filepath.Join(dir, "README (1)")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}
