package output

import (
	"strings"
	"testing"

	"github.com/dsswift/cli-exchange/internal/graph"
)

func TestFormatSize_KB(t *testing.T) {
	result := FormatSize(1024)
	if result != "1.0 KB" {
		t.Errorf("expected '1.0 KB', got %q", result)
	}
}

func TestFormatSize_MB(t *testing.T) {
	result := FormatSize(1024 * 1024)
	if result != "1.0 MB" {
		t.Errorf("expected '1.0 MB', got %q", result)
	}
}

func TestFormatSize_FractionalKB(t *testing.T) {
	result := FormatSize(512)
	if result != "0.5 KB" {
		t.Errorf("expected '0.5 KB', got %q", result)
	}
}

func TestFormatSize_FractionalMB(t *testing.T) {
	result := FormatSize(1536 * 1024) // 1.5 MB
	if result != "1.5 MB" {
		t.Errorf("expected '1.5 MB', got %q", result)
	}
}

func TestRenderAttachmentTable_Basic(t *testing.T) {
	attachments := []graph.AttachmentInfo{
		{Name: "report.pdf", ContentType: "application/pdf", Size: 245760, IsInline: false},
		{Name: "logo.png", ContentType: "image/png", Size: 1024, IsInline: true},
	}

	result := RenderAttachmentTable(attachments)

	if !strings.Contains(result, "report.pdf") {
		t.Error("expected result to contain 'report.pdf'")
	}
	if !strings.Contains(result, "application/pdf") {
		t.Error("expected result to contain 'application/pdf'")
	}
	if !strings.Contains(result, "logo.png") {
		t.Error("expected result to contain 'logo.png'")
	}
	if !strings.Contains(result, "Name") {
		t.Error("expected result to contain header 'Name'")
	}
	if !strings.Contains(result, "Inline") {
		t.Error("expected result to contain header 'Inline'")
	}
}

func TestRenderAttachmentTable_Empty(t *testing.T) {
	result := RenderAttachmentTable([]graph.AttachmentInfo{})
	// Should produce just headers or empty output without panic
	_ = result
}

func TestRenderMessageDetail_AttachmentNames(t *testing.T) {
	msg := &graph.Message{
		ID:             "msg-1",
		Subject:        "Test",
		HasAttachments: true,
		Attachments: []graph.AttachmentInfo{
			{Name: "report.pdf"},
			{Name: "data.csv"},
		},
	}

	result := RenderMessageDetail(msg)

	if !strings.Contains(result, "report.pdf") {
		t.Error("expected result to contain 'report.pdf'")
	}
	if !strings.Contains(result, "data.csv") {
		t.Error("expected result to contain 'data.csv'")
	}
	if strings.Contains(result, "Attachments:     yes") {
		t.Error("expected attachment names, not just 'yes'")
	}
}

func TestRenderMessageDetail_AttachmentYesWhenNoMetadata(t *testing.T) {
	msg := &graph.Message{
		ID:             "msg-1",
		Subject:        "Test",
		HasAttachments: true,
	}

	result := RenderMessageDetail(msg)

	if !strings.Contains(result, "yes") {
		t.Error("expected 'yes' when attachments are present but metadata is not loaded")
	}
}
