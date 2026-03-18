package graph

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testClient(handler http.HandlerFunc) *GraphClient {
	server := httptest.NewServer(handler)
	return NewClient(server.URL, 5*time.Second, func() (string, error) {
		return "test-token", nil
	})
}

func decodeJSON(t *testing.T, r *http.Request, v any) {
	t.Helper()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		t.Fatalf("decoding request body: %v", err)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encoding response: %v", err)
	}
}

func TestSendMail(t *testing.T) {
	var gotBody map[string]any
	var gotPath string

	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		decodeJSON(t, r, &gotBody)
		w.WriteHeader(202)
	})

	err := client.SendMail(SendMailOptions{
		Subject:      "Test",
		Body:         "Hello",
		ToRecipients: []string{"user@example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/me/sendMail" {
		t.Errorf("expected path /me/sendMail, got %q", gotPath)
	}

	msg, ok := gotBody["message"].(map[string]any)
	if !ok {
		t.Fatal("expected message envelope in request body")
	}
	if msg["subject"] != "Test" {
		t.Errorf("expected subject Test, got %v", msg["subject"])
	}

	recipients, ok := msg["toRecipients"].([]any)
	if !ok || len(recipients) != 1 {
		t.Errorf("expected 1 toRecipient, got %v", msg["toRecipients"])
	}
}

func TestSendMail_SaveToSentItems(t *testing.T) {
	var gotBody map[string]any

	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		decodeJSON(t, r, &gotBody)
		w.WriteHeader(202)
	})

	save := false
	err := client.SendMail(SendMailOptions{
		Subject:         "Test",
		ToRecipients:    []string{"user@example.com"},
		SaveToSentItems: &save,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["saveToSentItems"] != false {
		t.Errorf("expected saveToSentItems false, got %v", gotBody["saveToSentItems"])
	}
}

func TestSendMail_WithAttachments(t *testing.T) {
	var gotBody map[string]any

	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		decodeJSON(t, r, &gotBody)
		w.WriteHeader(202)
	})

	err := client.SendMail(SendMailOptions{
		Subject:      "Test",
		ToRecipients: []string{"user@example.com"},
		Attachments: []Attachment{
			{Name: "test.pdf", ContentType: "application/pdf", ContentBytes: []byte("fake-pdf")},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := gotBody["message"].(map[string]any)
	atts, ok := msg["attachments"].([]any)
	if !ok || len(atts) != 1 {
		t.Fatalf("expected 1 attachment, got %v", msg["attachments"])
	}
	att := atts[0].(map[string]any)
	if att["@odata.type"] != "#microsoft.graph.fileAttachment" {
		t.Errorf("expected fileAttachment odata type, got %v", att["@odata.type"])
	}
	if att["name"] != "test.pdf" {
		t.Errorf("expected name test.pdf, got %v", att["name"])
	}
	if att["contentType"] != "application/pdf" {
		t.Errorf("expected contentType application/pdf, got %v", att["contentType"])
	}
}

func TestSendMail_Error(t *testing.T) {
	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		writeJSON(t, w, map[string]any{
			"error": map[string]string{"code": "BadRequest", "message": "bad request"},
		})
	})

	err := client.SendMail(SendMailOptions{
		Subject:      "Test",
		ToRecipients: []string{"user@example.com"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSendDraft(t *testing.T) {
	var gotPath, gotMethod string

	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(202)
	})

	err := client.SendDraft("msg-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/me/messages/msg-123/send" {
		t.Errorf("expected path /me/messages/msg-123/send, got %q", gotPath)
	}
	if gotMethod != "POST" {
		t.Errorf("expected POST, got %q", gotMethod)
	}
}

func TestSendDraft_Error(t *testing.T) {
	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		writeJSON(t, w, map[string]any{
			"error": map[string]string{"code": "ErrorItemNotFound", "message": "not found"},
		})
	})

	err := client.SendDraft("bad-id")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAddAttachment(t *testing.T) {
	var gotPath string
	var gotBody map[string]any

	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		decodeJSON(t, r, &gotBody)
		w.WriteHeader(201)
		writeJSON(t, w, map[string]string{"id": "att-1"})
	})

	err := client.AddAttachment("msg-123", Attachment{
		Name:         "doc.txt",
		ContentType:  "text/plain",
		ContentBytes: []byte("hello"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/me/messages/msg-123/attachments" {
		t.Errorf("expected path /me/messages/msg-123/attachments, got %q", gotPath)
	}
	if gotBody["name"] != "doc.txt" {
		t.Errorf("expected name doc.txt, got %v", gotBody["name"])
	}
	if gotBody["@odata.type"] != "#microsoft.graph.fileAttachment" {
		t.Errorf("expected fileAttachment odata type, got %v", gotBody["@odata.type"])
	}
}

func TestAddAttachment_Error(t *testing.T) {
	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		writeJSON(t, w, map[string]any{
			"error": map[string]string{"code": "ErrorItemNotFound", "message": "not found"},
		})
	})

	err := client.AddAttachment("bad-id", Attachment{
		Name:         "doc.txt",
		ContentType:  "text/plain",
		ContentBytes: []byte("hello"),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateDraft_WithAttachments(t *testing.T) {
	var gotBody map[string]any

	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		decodeJSON(t, r, &gotBody)
		w.WriteHeader(201)
		writeJSON(t, w, map[string]string{"id": "draft-1", "subject": "Test"})
	})

	msg, err := client.CreateDraft(CreateDraftOptions{
		Subject:      "Test",
		ToRecipients: []string{"user@example.com"},
		Attachments: []Attachment{
			{Name: "file.pdf", ContentType: "application/pdf", ContentBytes: []byte("data")},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != "draft-1" {
		t.Errorf("expected id draft-1, got %q", msg.ID)
	}

	atts, ok := gotBody["attachments"].([]any)
	if !ok || len(atts) != 1 {
		t.Fatalf("expected 1 attachment in request, got %v", gotBody["attachments"])
	}
}

func TestListAttachments(t *testing.T) {
	var gotPath string
	var gotQuery string

	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		writeJSON(t, w, map[string]any{
			"value": []map[string]any{
				{
					"id":                   "att-1",
					"name":                 "report.pdf",
					"contentType":          "application/pdf",
					"size":                 245760,
					"isInline":             false,
					"lastModifiedDateTime": "2026-03-15T10:30:00Z",
				},
			},
		})
	})

	attachments, err := client.ListAttachments("msg-123", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/me/messages/msg-123/attachments" {
		t.Errorf("expected path /me/messages/msg-123/attachments, got %q", gotPath)
	}
	if strings.Contains(gotQuery, "contentBytes") {
		t.Error("expected contentBytes to be excluded from $select when includeContent is false")
	}
	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	if attachments[0].Name != "report.pdf" {
		t.Errorf("expected name report.pdf, got %q", attachments[0].Name)
	}
	if attachments[0].ContentType != "application/pdf" {
		t.Errorf("expected contentType application/pdf, got %q", attachments[0].ContentType)
	}
	if attachments[0].Size != 245760 {
		t.Errorf("expected size 245760, got %d", attachments[0].Size)
	}
	if attachments[0].IsInline {
		t.Error("expected isInline false")
	}
}

func TestListAttachments_WithContent(t *testing.T) {
	rawText := "hello world"
	encoded := base64.StdEncoding.EncodeToString([]byte(rawText))

	var gotQuery string
	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		writeJSON(t, w, map[string]any{
			"value": []map[string]any{
				{
					"id":           "att-1",
					"name":         "notes.txt",
					"contentType":  "text/plain",
					"size":         11,
					"isInline":     false,
					"contentBytes": encoded,
				},
			},
		})
	})

	attachments, err := client.ListAttachments("msg-123", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// When includeContent is true, $select is omitted entirely so Graph returns
	// all fields including contentBytes (which can't be named in $select because
	// it lives on the fileAttachment subtype, not the base attachment type).
	if strings.Contains(gotQuery, "%24select") || strings.Contains(gotQuery, "$select") {
		t.Error("expected no $select when includeContent is true")
	}
	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	if attachments[0].ContentBytes != encoded {
		t.Errorf("expected contentBytes %q, got %q", encoded, attachments[0].ContentBytes)
	}
	if attachments[0].ContentText != rawText {
		t.Errorf("expected contentText %q, got %q", rawText, attachments[0].ContentText)
	}
}

func TestListAttachments_BinaryContentNoText(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("binary data"))

	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"value": []map[string]any{
				{
					"id":           "att-1",
					"name":         "image.png",
					"contentType":  "image/png",
					"size":         100,
					"isInline":     false,
					"contentBytes": encoded,
				},
			},
		})
	})

	attachments, err := client.ListAttachments("msg-123", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attachments[0].ContentText != "" {
		t.Errorf("expected no contentText for binary attachment, got %q", attachments[0].ContentText)
	}
}

func TestListAttachments_Error(t *testing.T) {
	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		writeJSON(t, w, map[string]any{
			"error": map[string]string{"code": "ErrorItemNotFound", "message": "not found"},
		})
	})

	_, err := client.ListAttachments("bad-id", false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetMessageWithAttachments_HasAttachments(t *testing.T) {
	callCount := 0

	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if strings.HasSuffix(r.URL.Path, "/attachments") {
			writeJSON(t, w, map[string]any{
				"value": []map[string]any{
					{"id": "att-1", "name": "report.pdf", "contentType": "application/pdf", "size": 1024, "isInline": false},
				},
			})
		} else {
			writeJSON(t, w, map[string]any{
				"id":             "msg-123",
				"subject":        "Test",
				"hasAttachments": true,
			})
		}
	})

	msg, err := client.GetMessageWithAttachments("msg-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (message + attachments), got %d", callCount)
	}
	if len(msg.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(msg.Attachments))
	}
	if msg.Attachments[0].Name != "report.pdf" {
		t.Errorf("expected attachment name report.pdf, got %q", msg.Attachments[0].Name)
	}
}

func TestGetMessageWithAttachments_NoAttachments(t *testing.T) {
	callCount := 0

	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		writeJSON(t, w, map[string]any{
			"id":             "msg-123",
			"subject":        "Test",
			"hasAttachments": false,
		})
	})

	msg, err := client.GetMessageWithAttachments("msg-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 API call (no attachment fetch when hasAttachments is false), got %d", callCount)
	}
	if len(msg.Attachments) != 0 {
		t.Errorf("expected no attachments, got %d", len(msg.Attachments))
	}
}

func TestGetAttachment_FileAttachment(t *testing.T) {
	rawText := "Hello World"
	encoded := base64.StdEncoding.EncodeToString([]byte(rawText))

	var gotPath string
	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		writeJSON(t, w, map[string]any{
			"id":           "att-1",
			"name":         "notes.txt",
			"contentType":  "text/plain",
			"size":         11,
			"isInline":     false,
			"contentBytes": encoded,
		})
	})

	att, err := client.GetAttachment("msg-123", "att-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/me/messages/msg-123/attachments/att-1" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if att.Name != "notes.txt" {
		t.Errorf("expected name notes.txt, got %q", att.Name)
	}
	if att.ContentBytes != encoded {
		t.Errorf("expected contentBytes %q, got %q", encoded, att.ContentBytes)
	}
	if att.ContentText != rawText {
		t.Errorf("expected contentText %q, got %q", rawText, att.ContentText)
	}
}

func TestGetAttachment_Error(t *testing.T) {
	client := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		writeJSON(t, w, map[string]any{
			"error": map[string]string{"code": "ErrorItemNotFound", "message": "not found"},
		})
	})

	_, err := client.GetAttachment("msg-123", "bad-att-id")
	if err == nil {
		t.Fatal("expected error")
	}
}
