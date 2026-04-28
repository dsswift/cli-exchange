package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// --- Input structs ---

type SearchEmailsInput struct {
	Folder         string `json:"folder,omitempty"          jsonschema:"Mail folder name or ID"`
	Sender         string `json:"sender,omitempty"          jsonschema:"Sender email or alias"`
	Subject        string `json:"subject,omitempty"         jsonschema:"Subject substring filter"`
	StartDate      string `json:"start_date,omitempty"      jsonschema:"After date (YYYY-MM-DD)"`
	EndDate        string `json:"end_date,omitempty"        jsonschema:"Before date (YYYY-MM-DD)"`
	Unread         *bool  `json:"unread,omitempty"          jsonschema:"Filter to unread only"`
	HasAttachments *bool  `json:"has_attachments,omitempty" jsonschema:"Filter to messages with attachments"`
	Read           *bool  `json:"read,omitempty"            jsonschema:"Filter to read only"`
	Limit          int    `json:"limit,omitempty"           jsonschema:"Max results (default 25)"`
}

type GetEmailInput struct {
	IDs string `json:"ids" jsonschema:"Comma-separated message IDs"`
}

type SendEmailInput struct {
	To         []string `json:"to"                    jsonschema:"Recipient email addresses"`
	Subject    string   `json:"subject"               jsonschema:"Email subject"`
	Body       string   `json:"body"                  jsonschema:"Email body text"`
	CC         []string `json:"cc,omitempty"           jsonschema:"CC email addresses"`
	BodyType   string   `json:"body_type,omitempty"    jsonschema:"Body type: text or html"`
	Importance string   `json:"importance,omitempty"   jsonschema:"Importance: low normal or high"`
}

type ArchiveEmailInput struct {
	IDs string `json:"ids" jsonschema:"Comma-separated message IDs to archive"`
}

type DeleteEmailInput struct {
	IDs string `json:"ids" jsonschema:"Comma-separated message IDs to delete"`
}

type CreateDraftInput struct {
	To       []string `json:"to"                  jsonschema:"Recipient email addresses"`
	Subject  string   `json:"subject"             jsonschema:"Email subject"`
	Body     string   `json:"body"                jsonschema:"Email body text"`
	CC       []string `json:"cc,omitempty"         jsonschema:"CC email addresses"`
	BodyType string   `json:"body_type,omitempty"  jsonschema:"Body type: text or html"`
}

type SendDraftInput struct {
	ID string `json:"id" jsonschema:"Draft message ID"`
}

type AttachToDraftInput struct {
	ID    string   `json:"id"    jsonschema:"Draft message ID"`
	Files []string `json:"files" jsonschema:"File paths to attach"`
}

type EmptyInput struct{}

type ListAttachmentsInput struct {
	MessageID      string `json:"message_id"               jsonschema:"Message ID"`
	IncludeContent *bool  `json:"include_content,omitempty" jsonschema:"Include decoded text content inline"`
	Name           string `json:"name,omitempty"            jsonschema:"Filter by attachment name substring"`
	NoInline       *bool  `json:"no_inline,omitempty"       jsonschema:"Exclude inline attachments"`
}

type ListEventsInput struct {
	CalendarID string `json:"calendar_id,omitempty" jsonschema:"Calendar ID (default: primary)"`
	StartDate  string `json:"start_date,omitempty"  jsonschema:"Start date (YYYY-MM-DD)"`
	EndDate    string `json:"end_date,omitempty"    jsonschema:"End date (YYYY-MM-DD)"`
	Limit      int    `json:"limit,omitempty"       jsonschema:"Max results"`
}

type GetEventInput struct {
	ID string `json:"id" jsonschema:"Event ID"`
}

type GetFreeBusyInput struct {
	Emails          string `json:"emails"                    jsonschema:"Comma-separated email addresses"`
	Start           string `json:"start,omitempty"            jsonschema:"Start time (YYYY-MM-DDTHH:MM:SS)"`
	End             string `json:"end,omitempty"              jsonschema:"End time (YYYY-MM-DDTHH:MM:SS)"`
	Timespan        string `json:"timespan,omitempty"         jsonschema:"Relative timespan (e.g. 2h 1d 1w)"`
	Interval        int    `json:"interval,omitempty"         jsonschema:"Interval in minutes"`
	BusinessHours   string `json:"business_hours,omitempty"   jsonschema:"Business hours range (HH:MM-HH:MM)"`
	IncludeWeekends *bool  `json:"include_weekends,omitempty" jsonschema:"Include weekends in availability"`
}

// --- Tool registration ---

func registerTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_emails",
		Description: "Search and list emails with optional filters for folder, sender, subject, date range, read status, and attachments",
	}, handleSearchEmails)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_email",
		Description: "Get full email details including body for one or more message IDs",
	}, handleGetEmail)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "send_email",
		Description: "Send an email. Subject to sender whitelist restrictions.",
	}, handleSendEmail)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "archive_email",
		Description: "Archive one or more emails by message ID",
	}, handleArchiveEmail)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_email",
		Description: "Permanently delete one or more emails by message ID",
	}, handleDeleteEmail)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_draft",
		Description: "Create a draft email. Use this when send is blocked by whitelist restrictions.",
	}, handleCreateDraft)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "send_draft",
		Description: "Send an existing draft email by its message ID",
	}, handleSendDraft)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "attach_to_draft",
		Description: "Attach files to an existing draft email",
	}, handleAttachToDraft)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_mail_folders",
		Description: "List all mail folders in the mailbox",
	}, handleListMailFolders)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_attachments",
		Description: "List attachments on a message, optionally including decoded text content",
	}, handleListAttachments)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_calendars",
		Description: "List all calendars available to the user",
	}, handleListCalendars)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_events",
		Description: "List calendar events with optional date range and calendar filters",
	}, handleListEvents)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_event",
		Description: "Get full details of a calendar event by ID",
	}, handleGetEvent)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_free_busy",
		Description: "Check availability/free-busy schedule for one or more users",
	}, handleGetFreeBusy)
}

// --- Handlers ---

func handleSearchEmails(_ context.Context, _ *mcp.CallToolRequest, input SearchEmailsInput) (*mcp.CallToolResult, any, error) {
	args := []string{"mail", "list", "-o", "json"}
	if input.Folder != "" {
		args = append(args, "--folder", input.Folder)
	}
	if input.Sender != "" {
		args = append(args, "--sender", input.Sender)
	}
	if input.Subject != "" {
		args = append(args, "--subject", input.Subject)
	}
	if input.StartDate != "" {
		args = append(args, "--start", input.StartDate)
	}
	if input.EndDate != "" {
		args = append(args, "--end", input.EndDate)
	}
	if input.Unread != nil && *input.Unread {
		args = append(args, "--unread")
	}
	if input.Read != nil && *input.Read {
		args = append(args, "--read")
	}
	if input.HasAttachments != nil && *input.HasAttachments {
		args = append(args, "--has-attachments")
	}
	if input.Limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", input.Limit))
	}
	return execTool(args...)
}

func handleGetEmail(_ context.Context, _ *mcp.CallToolRequest, input GetEmailInput) (*mcp.CallToolResult, any, error) {
	return execTool("mail", "show", "-o", "json", "--ids", input.IDs)
}

func handleSendEmail(_ context.Context, _ *mcp.CallToolRequest, input SendEmailInput) (*mcp.CallToolResult, any, error) {
	args := []string{"mail", "send", "-o", "json"}
	for _, to := range input.To {
		args = append(args, "--to", to)
	}
	args = append(args, "--subject", input.Subject, "--body", input.Body)
	for _, cc := range input.CC {
		args = append(args, "--cc", cc)
	}
	if input.BodyType != "" {
		args = append(args, "--body-type", input.BodyType)
	}
	if input.Importance != "" {
		args = append(args, "--importance", input.Importance)
	}
	return execTool(args...)
}

func handleArchiveEmail(_ context.Context, _ *mcp.CallToolRequest, input ArchiveEmailInput) (*mcp.CallToolResult, any, error) {
	return execTool("mail", "archive", "-o", "json", "--ids", input.IDs)
}

func handleDeleteEmail(_ context.Context, _ *mcp.CallToolRequest, input DeleteEmailInput) (*mcp.CallToolResult, any, error) {
	return execTool("mail", "delete", "-o", "json", "--ids", input.IDs)
}

func handleCreateDraft(_ context.Context, _ *mcp.CallToolRequest, input CreateDraftInput) (*mcp.CallToolResult, any, error) {
	args := []string{"mail", "draft", "create", "-o", "json"}
	for _, to := range input.To {
		args = append(args, "--to", to)
	}
	args = append(args, "--subject", input.Subject, "--body", input.Body)
	for _, cc := range input.CC {
		args = append(args, "--cc", cc)
	}
	if input.BodyType != "" {
		args = append(args, "--body-type", input.BodyType)
	}
	return execTool(args...)
}

func handleSendDraft(_ context.Context, _ *mcp.CallToolRequest, input SendDraftInput) (*mcp.CallToolResult, any, error) {
	return execTool("mail", "draft", "send", "-o", "json", input.ID)
}

func handleAttachToDraft(_ context.Context, _ *mcp.CallToolRequest, input AttachToDraftInput) (*mcp.CallToolResult, any, error) {
	args := []string{"mail", "draft", "attach", "-o", "json", input.ID}
	for _, f := range input.Files {
		args = append(args, "--attach", f)
	}
	return execTool(args...)
}

func handleListMailFolders(_ context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, any, error) {
	return execTool("mail", "folder", "list", "-o", "json")
}

func handleListAttachments(_ context.Context, _ *mcp.CallToolRequest, input ListAttachmentsInput) (*mcp.CallToolResult, any, error) {
	args := []string{"mail", "attachment", "list", "-o", "json", "--message-id", input.MessageID}
	if input.IncludeContent != nil && *input.IncludeContent {
		args = append(args, "--include-content")
	}
	if input.Name != "" {
		args = append(args, "--name", input.Name)
	}
	if input.NoInline != nil && *input.NoInline {
		args = append(args, "--no-inline")
	}
	return execTool(args...)
}

func handleListCalendars(_ context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, any, error) {
	return execTool("calendar", "list", "-o", "json")
}

func handleListEvents(_ context.Context, _ *mcp.CallToolRequest, input ListEventsInput) (*mcp.CallToolResult, any, error) {
	args := []string{"calendar", "event", "list", "-o", "json"}
	if input.CalendarID != "" {
		args = append(args, "--calendar", input.CalendarID)
	}
	if input.StartDate != "" {
		args = append(args, "--start", input.StartDate)
	}
	if input.EndDate != "" {
		args = append(args, "--end", input.EndDate)
	}
	if input.Limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", input.Limit))
	}
	return execTool(args...)
}

func handleGetEvent(_ context.Context, _ *mcp.CallToolRequest, input GetEventInput) (*mcp.CallToolResult, any, error) {
	return execTool("calendar", "event", "show", "-o", "json", input.ID)
}

func handleGetFreeBusy(_ context.Context, _ *mcp.CallToolRequest, input GetFreeBusyInput) (*mcp.CallToolResult, any, error) {
	args := []string{"calendar", "availability", "check", "-o", "json", "--emails", input.Emails}
	if input.Start != "" {
		args = append(args, "--start", input.Start)
	}
	if input.End != "" {
		args = append(args, "--end", input.End)
	}
	if input.Timespan != "" {
		args = append(args, "--timespan", input.Timespan)
	}
	if input.Interval > 0 {
		args = append(args, "--interval", fmt.Sprintf("%d", input.Interval))
	}
	if input.BusinessHours != "" {
		args = append(args, "--business-hours", input.BusinessHours)
	}
	if input.IncludeWeekends != nil && *input.IncludeWeekends {
		args = append(args, "--include-weekends")
	}
	return execTool(args...)
}
