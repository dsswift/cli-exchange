package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsswift/cli-exchange/internal/config"
	"github.com/dsswift/cli-exchange/internal/graph"
	"github.com/dsswift/cli-exchange/internal/output"
	"github.com/dsswift/cli-exchange/internal/tz"
)

const maxAttachmentSize = 3 * 1024 * 1024 // 3MB

func cmdMailFolderList(f flags) int {
	client, _, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	folders, err := client.ListMailFolders()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Print(output.RenderFolderTable(folders))
	} else {
		output.PrintJSON(map[string]any{
			"count":   len(folders),
			"folders": folders,
		})
	}
	return 0
}

func buildMailListOpts(f flags, tzSvc *tz.Service, cfg *config.ExchangeConfig) (graph.ListMessagesOptions, error) {
	opts := graph.ListMessagesOptions{
		FolderID:       f.folder,
		Subject:        f.subject,
		IsRead:         f.isRead,
		HasAttachments: f.hasAttachments,
		Limit:          f.limit,
	}

	if f.sender != "" {
		addrs := expandSender(f.sender, cfg.DomainAliases)
		if addrs == nil {
			opts.Search = f.sender
		} else {
			parts := make([]string, len(addrs))
			for i, addr := range addrs {
				parts[i] = "from:" + addr
			}
			opts.Search = strings.Join(parts, " OR ")
		}
	}

	if f.start != "" {
		t, err := tzSvc.ParseDate(f.start)
		if err != nil {
			return opts, err
		}
		opts.StartDate = &t
	}
	if f.end != "" {
		t, err := tzSvc.ParseDate(f.end)
		if err != nil {
			return opts, err
		}
		opts.EndDate = &t
	}

	return opts, nil
}

func cmdMailList(f flags) int {
	client, tzSvc, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	cfg := config.LoadConfigPartial(configOverrides(f))
	opts, err := buildMailListOpts(f, tzSvc, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	messages, err := client.ListMessages(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Print(output.RenderMessageTable(messages))
	} else {
		output.PrintJSON(map[string]any{
			"count":    len(messages),
			"messages": messages,
		})
	}
	return 0
}

// expandSender applies smart detection to the sender value.
// Returns nil if the sender has no @ (use bare $search instead).
// Returns a list of addresses for KQL from: search terms.
func expandSender(sender string, aliases map[string][]string) []string {
	if !strings.Contains(sender, "@") {
		return nil
	}

	// Check for pipe syntax: user@domain1.com|domain2.com
	parts := strings.SplitN(sender, "@", 2)
	user := parts[0]
	domainPart := parts[1]

	if strings.Contains(domainPart, "|") {
		domains := strings.Split(domainPart, "|")
		addrs := make([]string, len(domains))
		for i, d := range domains {
			addrs[i] = user + "@" + d
		}
		return addrs
	}

	// Single address: check domain aliases
	addr := user + "@" + domainPart
	if aliases != nil {
		if aliasDomains, ok := aliases[domainPart]; ok && len(aliasDomains) > 1 {
			addrs := make([]string, len(aliasDomains))
			for i, d := range aliasDomains {
				addrs[i] = user + "@" + d
			}
			return addrs
		}
	}

	return []string{addr}
}

func hasMailFilters(f flags) bool {
	return f.folder != "" || f.sender != "" || f.subject != "" ||
		f.start != "" || f.end != "" || f.isRead != nil || f.hasAttachments != nil
}

func cmdMailShow(f flags) int {
	if len(f.ids) > 0 && f.batch > 0 {
		fmt.Fprintf(os.Stderr, "Error: --ids and --batch are mutually exclusive\n")
		return 1
	}

	if len(f.ids) == 0 && f.batch == 0 {
		fmt.Fprintf(os.Stderr, "Error: --ids or --batch required\n")
		return 1
	}

	if len(f.ids) > 0 && hasMailFilters(f) {
		fmt.Fprintf(os.Stderr, "Error: filter flags require --batch\n")
		return 1
	}

	if len(f.ids) > 0 {
		return cmdMailShowByIDs(f)
	}
	return cmdMailShowBatch(f)
}

func cmdMailShowByIDs(f flags) int {
	client, _, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	messages := make([]*graph.Message, 0, len(f.ids))
	for _, id := range f.ids {
		msg, err := client.GetMessageWithAttachments(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching %s: %s\n", id, err)
			return 1
		}
		messages = append(messages, msg)
	}

	if f.output == "table" {
		fmt.Print(output.RenderMessageBatch(messages))
	} else {
		if len(messages) == 1 {
			output.PrintJSON(messages[0])
		} else {
			output.PrintJSON(map[string]any{
				"count":    len(messages),
				"messages": messages,
			})
		}
	}
	return 0
}

func cmdMailShowBatch(f flags) int {
	client, tzSvc, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	cfg := config.LoadConfigPartial(configOverrides(f))
	opts, err := buildMailListOpts(f, tzSvc, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}
	opts.IncludeBody = true
	opts.Limit = f.batch

	messages, err := client.ListMessages(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	// Convert to pointer slice for detail rendering
	ptrs := make([]*graph.Message, len(messages))
	for i := range messages {
		ptrs[i] = &messages[i]
	}

	if f.output == "table" {
		fmt.Print(output.RenderMessageBatch(ptrs))
	} else {
		output.PrintJSON(map[string]any{
			"count":    len(messages),
			"messages": messages,
		})
	}
	return 0
}

func cmdMailArchive(f flags) int {
	if len(f.ids) == 0 {
		fmt.Fprintf(os.Stderr, "Error: --ids required\n")
		return 1
	}

	client, _, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	type archiveResult struct {
		Archived  bool   `json:"archived"`
		MessageID string `json:"messageId"`
		Subject   string `json:"subject"`
		Error     string `json:"error,omitempty"`
	}

	results := make([]archiveResult, 0, len(f.ids))
	failed := false

	for _, id := range f.ids {
		msg, err := client.MoveMessage(id, "archive")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error archiving %s: %s\n", id, err)
			results = append(results, archiveResult{MessageID: id, Error: err.Error()})
			failed = true
			continue
		}
		if f.output == "table" {
			fmt.Printf("Archived: %s\n", msg.Subject)
		}
		results = append(results, archiveResult{Archived: true, MessageID: msg.ID, Subject: msg.Subject})
	}

	if f.output != "table" {
		if len(results) == 1 {
			output.PrintJSON(results[0])
		} else {
			output.PrintJSON(results)
		}
	}

	if failed {
		return 1
	}
	return 0
}

func cmdMailDelete(f flags) int {
	if len(f.ids) == 0 {
		fmt.Fprintf(os.Stderr, "Error: --ids required\n")
		return 1
	}

	client, _, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	type deleteResult struct {
		Deleted   bool   `json:"deleted"`
		MessageID string `json:"messageId"`
		Subject   string `json:"subject"`
		Error     string `json:"error,omitempty"`
	}

	results := make([]deleteResult, 0, len(f.ids))
	failed := false

	for _, id := range f.ids {
		msg, err := client.GetMessage(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching %s: %s\n", id, err)
			results = append(results, deleteResult{MessageID: id, Error: err.Error()})
			failed = true
			continue
		}
		subject := msg.Subject

		if err := client.DeleteMessage(id); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting %s: %s\n", id, err)
			results = append(results, deleteResult{MessageID: id, Subject: subject, Error: err.Error()})
			failed = true
			continue
		}

		if f.output == "table" {
			fmt.Printf("Deleted: %s\n", subject)
		}
		results = append(results, deleteResult{Deleted: true, MessageID: id, Subject: subject})
	}

	if f.output != "table" {
		if len(results) == 1 {
			output.PrintJSON(results[0])
		} else {
			output.PrintJSON(results)
		}
	}

	if failed {
		return 1
	}
	return 0
}

func cmdMailDraftCreate(f flags) int {
	client, _, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	var attachments []graph.Attachment
	if len(f.attachFiles) > 0 {
		attachments, err = loadAttachments(f.attachFiles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
	}

	opts := graph.CreateDraftOptions{
		Subject:      f.subject,
		Body:         f.body,
		BodyType:     f.bodyType,
		ToRecipients: f.to,
		CcRecipients: f.cc,
		Importance:   f.importance,
		Attachments:  attachments,
	}

	msg, err := client.CreateDraft(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Printf("Draft created: %s (ID: %s)\n", msg.Subject, msg.ID)
	} else {
		output.PrintJSON(map[string]any{
			"created":   true,
			"messageId": msg.ID,
			"subject":   msg.Subject,
		})
	}
	return 0
}

func loadSendConfig(f flags) *config.ExchangeConfig {
	cfg := config.LoadConfigPartial(configOverrides(f))

	// Fallback: resolve user email from MSAL cache if not persisted
	if cfg.UserEmail == "" {
		if authenticator, err := newAuthenticator(cfg); err == nil {
			if email, err := authenticator.GetUserEmail(context.Background()); err == nil {
				cfg.UserEmail = email
			}
		}
	}
	return cfg
}

func cmdMailSend(f flags) int {
	if len(f.to) == 0 {
		fmt.Fprintf(os.Stderr, "Error: --to is required\n")
		return 1
	}

	cfg := loadSendConfig(f)
	allRecipients := append(f.to, f.cc...)
	if err := cfg.ValidateSendRecipients(allRecipients); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	client, _, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	var attachments []graph.Attachment
	if len(f.attachFiles) > 0 {
		attachments, err = loadAttachments(f.attachFiles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
	}

	opts := graph.SendMailOptions{
		Subject:         f.subject,
		Body:            f.body,
		BodyType:        f.bodyType,
		ToRecipients:    f.to,
		CcRecipients:    f.cc,
		Importance:      f.importance,
		Attachments:     attachments,
		SaveToSentItems: f.saveToSentItems,
	}

	if err := client.SendMail(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Printf("Sent: %s\n", f.subject)
	} else {
		output.PrintJSON(map[string]any{
			"sent":    true,
			"subject": f.subject,
		})
	}
	return 0
}

func cmdMailDraftSend(f flags) int {
	if f.id == "" {
		fmt.Fprintf(os.Stderr, "Error: message ID is required\n")
		return 1
	}

	client, _, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	// Fetch draft to validate recipients before sending
	cfg := loadSendConfig(f)
	draft, err := client.GetMessage(f.id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}
	var allRecipients []string
	for _, r := range draft.ToRecipients {
		allRecipients = append(allRecipients, r.EmailAddress.Address)
	}
	for _, r := range draft.CcRecipients {
		allRecipients = append(allRecipients, r.EmailAddress.Address)
	}
	if err := cfg.ValidateSendRecipients(allRecipients); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if err := client.SendDraft(f.id); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Printf("Sent draft: %s\n", f.id)
	} else {
		output.PrintJSON(map[string]any{
			"sent":      true,
			"messageId": f.id,
		})
	}
	return 0
}

func cmdMailDraftAttach(f flags) int {
	if f.id == "" {
		fmt.Fprintf(os.Stderr, "Error: message ID is required\n")
		return 1
	}
	if len(f.attachFiles) == 0 {
		fmt.Fprintf(os.Stderr, "Error: --attach is required\n")
		return 1
	}

	client, _, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	attachments, err := loadAttachments(f.attachFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	for _, att := range attachments {
		if err := client.AddAttachment(f.id, att); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
	}

	files := make([]string, len(attachments))
	for i, att := range attachments {
		files[i] = att.Name
	}

	if f.output == "table" {
		fmt.Printf("Attached %d file(s) to draft: %s\n", len(attachments), f.id)
	} else {
		output.PrintJSON(map[string]any{
			"attached":  true,
			"messageId": f.id,
			"files":     files,
		})
	}
	return 0
}

func filterAttachments(attachments []graph.AttachmentInfo, name string, noInline bool) []graph.AttachmentInfo {
	var result []graph.AttachmentInfo
	for _, a := range attachments {
		if noInline && a.IsInline {
			continue
		}
		if name != "" && !strings.Contains(strings.ToLower(a.Name), strings.ToLower(name)) {
			continue
		}
		result = append(result, a)
	}
	return result
}

func cmdMailAttachmentList(f flags) int {
	if f.messageID == "" {
		fmt.Fprintf(os.Stderr, "Error: --message-id is required\n")
		return 1
	}

	client, _, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	var attachments []graph.AttachmentInfo

	if f.id != "" {
		// Single attachment by ID
		att, err := client.GetAttachment(f.messageID, f.id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
		attachments = []graph.AttachmentInfo{*att}
	} else {
		attachments, err = client.ListAttachments(f.messageID, f.includeContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
		attachments = filterAttachments(attachments, f.name, f.noInline)
	}

	if f.output == "table" {
		fmt.Print(output.RenderAttachmentTable(attachments))
	} else {
		output.PrintJSON(map[string]any{
			"messageId":   f.messageID,
			"count":       len(attachments),
			"attachments": attachments,
		})
	}
	return 0
}

func resolveDownloadPath(dir, name string) string {
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); err != nil {
		return path
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	for i := 1; ; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", base, i, ext))
		if _, err := os.Stat(candidate); err != nil {
			return candidate
		}
	}
}

func cmdMailAttachmentDownload(f flags) int {
	if f.messageID == "" {
		fmt.Fprintf(os.Stderr, "Error: --message-id is required\n")
		return 1
	}

	dir := f.dir
	if dir == "" {
		dir = "."
	}

	client, _, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	var attachments []graph.AttachmentInfo

	if f.id != "" {
		// Download a single specific attachment by ID
		att, err := client.GetAttachment(f.messageID, f.id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
		attachments = []graph.AttachmentInfo{*att}
	} else {
		attachments, err = client.ListAttachments(f.messageID, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
		attachments = filterAttachments(attachments, f.name, f.noInline)
	}

	if len(attachments) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no attachments match the given filters\n")
		return 1
	}

	type downloadResult struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Size int    `json:"size"`
	}

	var downloaded []downloadResult

	for _, att := range attachments {
		if att.ContentBytes == "" {
			fmt.Fprintf(os.Stderr, "Warning: no content for attachment %s, skipping\n", att.Name)
			continue
		}

		data, err := base64.StdEncoding.DecodeString(att.ContentBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding attachment %s: %s\n", att.Name, err)
			return 1
		}

		destPath := resolveDownloadPath(dir, att.Name)
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %s\n", destPath, err)
			return 1
		}

		if f.output == "table" {
			fmt.Printf("Saved: %s -> %s (%s)\n", att.Name, destPath, output.FormatSize(len(data)))
		}

		downloaded = append(downloaded, downloadResult{
			Name: att.Name,
			Path: destPath,
			Size: len(data),
		})
	}

	if f.output != "table" {
		output.PrintJSON(map[string]any{
			"messageId":  f.messageID,
			"downloaded": downloaded,
		})
	}
	return 0
}

func loadAttachments(paths []string) ([]graph.Attachment, error) {
	attachments := make([]graph.Attachment, 0, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading attachment %s: %w", path, err)
		}
		if len(data) > maxAttachmentSize {
			return nil, fmt.Errorf("file exceeds 3MB limit: %s (%d bytes)", path, len(data))
		}

		name := filepath.Base(path)
		contentType := mime.TypeByExtension(filepath.Ext(path))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		attachments = append(attachments, graph.Attachment{
			Name:         name,
			ContentType:  contentType,
			ContentBytes: data,
		})
	}
	return attachments, nil
}
