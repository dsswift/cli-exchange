package graph

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

var messageSelectFields = "id,subject,sender,from,toRecipients,ccRecipients,receivedDateTime,isRead,hasAttachments,importance,bodyPreview,webLink"
var messageDetailFields = messageSelectFields + ",body"

func (c *GraphClient) ListMailFolders() ([]MailFolder, error) {
	params := url.Values{}
	params.Set("$select", "id,displayName,unreadItemCount,totalItemCount,parentFolderId")
	params.Set("$top", "100")

	data, err := c.do("GET", "/me/mailFolders", params, nil)
	if err != nil {
		return nil, err
	}

	var resp listResponse[MailFolder]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing mail folders: %w", err)
	}
	return resp.Value, nil
}

type ListMessagesOptions struct {
	FolderID       string
	Search         string
	Subject        string
	StartDate      *time.Time
	EndDate        *time.Time
	IsRead         *bool
	HasAttachments *bool
	Limit          int
	IncludeBody    bool
}

func (c *GraphClient) ListMessages(opts ListMessagesOptions) ([]Message, error) {
	path := "/me/messages"
	if opts.FolderID == "all" {
		// keep /me/messages (all folders)
	} else {
		folderID := opts.FolderID
		if folderID == "" {
			folderID = "inbox"
		}
		path = fmt.Sprintf("/me/mailFolders/%s/messages", folderID)
	}

	params := url.Values{}
	selectFields := messageSelectFields
	if opts.IncludeBody {
		selectFields = messageDetailFields
	}
	params.Set("$select", selectFields)

	// Build KQL $search terms for fields that don't work with $filter
	var searchTerms []string
	if opts.Search != "" {
		search := opts.Search
		if !strings.Contains(search, ":") {
			search = "from:" + search
		}
		searchTerms = append(searchTerms, search)
	}
	if opts.Subject != "" {
		searchTerms = append(searchTerms, fmt.Sprintf("subject:%s", opts.Subject))
	}
	if opts.IsRead != nil {
		searchTerms = append(searchTerms, fmt.Sprintf("isread:%t", *opts.IsRead))
	}
	if opts.HasAttachments != nil {
		searchTerms = append(searchTerms, fmt.Sprintf("hasattachment:%t", *opts.HasAttachments))
	}

	useSearch := len(searchTerms) > 0
	if !useSearch {
		params.Set("$orderby", "receivedDateTime desc")
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	params.Set("$top", fmt.Sprintf("%d", limit))

	var filters []string

	if opts.StartDate != nil {
		filters = append(filters, fmt.Sprintf("receivedDateTime ge %s", opts.StartDate.Format(time.RFC3339)))
	}
	if opts.EndDate != nil {
		filters = append(filters, fmt.Sprintf("receivedDateTime le %s", opts.EndDate.Format(time.RFC3339)))
	}
	if len(filters) > 0 {
		params.Set("$filter", strings.Join(filters, " and "))
	}

	if useSearch {
		params.Set("$search", fmt.Sprintf(`"%s"`, strings.Join(searchTerms, " AND ")))
	}

	data, err := c.do("GET", path, params, nil)
	if err != nil {
		return nil, err
	}

	var resp listResponse[Message]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing messages: %w", err)
	}

	// Client-side sort when server-side $orderby was omitted
	if useSearch {
		sort.Slice(resp.Value, func(i, j int) bool {
			return resp.Value[i].ReceivedDateTime > resp.Value[j].ReceivedDateTime
		})
	}

	return resp.Value, nil
}

func (c *GraphClient) GetMessage(id string) (*Message, error) {
	params := url.Values{}
	params.Set("$select", messageDetailFields)

	data, err := c.do("GET", fmt.Sprintf("/me/messages/%s", id), params, nil)
	if err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("parsing message: %w", err)
	}
	return &msg, nil
}

func (c *GraphClient) ListAttachments(messageID string, includeContent bool) ([]AttachmentInfo, error) {
	var params url.Values
	if !includeContent {
		// $select only works on the base attachment type; contentBytes is on the
		// fileAttachment subtype and cannot be named in $select. When we don't
		// need content, restrict fields to avoid pulling large payloads.
		// When we do need content, omit $select and let Graph return everything.
		params = url.Values{}
		params.Set("$select", "id,name,contentType,size,isInline,lastModifiedDateTime")
	}

	data, err := c.do("GET", fmt.Sprintf("/me/messages/%s/attachments", messageID), params, nil)
	if err != nil {
		return nil, err
	}

	var resp listResponse[AttachmentInfo]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing attachments: %w", err)
	}

	if includeContent {
		for i := range resp.Value {
			att := &resp.Value[i]
			if strings.HasPrefix(att.ContentType, "text/") && att.ContentBytes != "" {
				decoded, err := base64.StdEncoding.DecodeString(att.ContentBytes)
				if err == nil {
					att.ContentText = string(decoded)
				}
			}
		}
	}

	return resp.Value, nil
}

func (c *GraphClient) GetAttachment(messageID, attachmentID string) (*AttachmentInfo, error) {
	data, err := c.do("GET", fmt.Sprintf("/me/messages/%s/attachments/%s", messageID, attachmentID), nil, nil)
	if err != nil {
		return nil, err
	}

	var att AttachmentInfo
	if err := json.Unmarshal(data, &att); err != nil {
		return nil, fmt.Errorf("parsing attachment: %w", err)
	}

	if strings.HasPrefix(att.ContentType, "text/") && att.ContentBytes != "" {
		decoded, err := base64.StdEncoding.DecodeString(att.ContentBytes)
		if err == nil {
			att.ContentText = string(decoded)
		}
	}

	return &att, nil
}

func (c *GraphClient) GetMessageWithAttachments(id string) (*Message, error) {
	msg, err := c.GetMessage(id)
	if err != nil {
		return nil, err
	}

	if msg.HasAttachments {
		attachments, err := c.ListAttachments(id, false)
		if err != nil {
			return nil, err
		}
		msg.Attachments = attachments
	}

	return msg, nil
}

func (c *GraphClient) MoveMessage(id, destinationID string) (*Message, error) {
	body := map[string]string{"destinationId": destinationID}
	data, err := c.do("POST", fmt.Sprintf("/me/messages/%s/move", id), nil, body)
	if err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("parsing moved message: %w", err)
	}
	return &msg, nil
}

func (c *GraphClient) DeleteMessage(id string) error {
	_, err := c.do("DELETE", fmt.Sprintf("/me/messages/%s", id), nil, nil)
	return err
}

// Attachment represents a file attachment for Graph API messages.
type Attachment struct {
	Name         string // filename
	ContentType  string // MIME type
	ContentBytes []byte // raw file bytes
}

type CreateDraftOptions struct {
	Subject      string
	Body         string
	BodyType     string
	ToRecipients []string
	CcRecipients []string
	Importance   string
	Attachments  []Attachment
}

type SendMailOptions struct {
	Subject         string
	Body            string
	BodyType        string
	ToRecipients    []string
	CcRecipients    []string
	Importance      string
	Attachments     []Attachment
	SaveToSentItems *bool
}

func (c *GraphClient) CreateDraft(opts CreateDraftOptions) (*Message, error) {
	msg := map[string]any{}

	if opts.Subject != "" {
		msg["subject"] = opts.Subject
	}
	if opts.Body != "" {
		bodyType := opts.BodyType
		if bodyType == "" {
			bodyType = "text"
		}
		msg["body"] = map[string]string{
			"contentType": bodyType,
			"content":     opts.Body,
		}
	}
	if opts.Importance != "" {
		msg["importance"] = opts.Importance
	}

	if len(opts.ToRecipients) > 0 {
		msg["toRecipients"] = buildRecipients(opts.ToRecipients)
	}
	if len(opts.CcRecipients) > 0 {
		msg["ccRecipients"] = buildRecipients(opts.CcRecipients)
	}
	if len(opts.Attachments) > 0 {
		msg["attachments"] = buildAttachments(opts.Attachments)
	}

	data, err := c.do("POST", "/me/messages", nil, msg)
	if err != nil {
		return nil, err
	}

	var result Message
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing draft: %w", err)
	}
	return &result, nil
}

func buildRecipients(emails []string) []map[string]any {
	recipients := make([]map[string]any, len(emails))
	for i, email := range emails {
		recipients[i] = map[string]any{
			"emailAddress": map[string]string{
				"address": email,
			},
		}
	}
	return recipients
}

func buildAttachments(attachments []Attachment) []map[string]any {
	result := make([]map[string]any, len(attachments))
	for i, att := range attachments {
		result[i] = map[string]any{
			"@odata.type":  "#microsoft.graph.fileAttachment",
			"name":         att.Name,
			"contentType":  att.ContentType,
			"contentBytes": base64.StdEncoding.EncodeToString(att.ContentBytes),
		}
	}
	return result
}

func buildMessageBody(subject, body, bodyType, importance string, to, cc []string, attachments []Attachment) map[string]any {
	msg := map[string]any{}
	if subject != "" {
		msg["subject"] = subject
	}
	if body != "" {
		bt := bodyType
		if bt == "" {
			bt = "text"
		}
		msg["body"] = map[string]string{
			"contentType": bt,
			"content":     body,
		}
	}
	if importance != "" {
		msg["importance"] = importance
	}
	if len(to) > 0 {
		msg["toRecipients"] = buildRecipients(to)
	}
	if len(cc) > 0 {
		msg["ccRecipients"] = buildRecipients(cc)
	}
	if len(attachments) > 0 {
		msg["attachments"] = buildAttachments(attachments)
	}
	return msg
}

func (c *GraphClient) SendMail(opts SendMailOptions) error {
	msg := buildMessageBody(opts.Subject, opts.Body, opts.BodyType, opts.Importance,
		opts.ToRecipients, opts.CcRecipients, opts.Attachments)

	envelope := map[string]any{
		"message": msg,
	}
	if opts.SaveToSentItems != nil {
		envelope["saveToSentItems"] = *opts.SaveToSentItems
	}

	_, err := c.do("POST", "/me/sendMail", nil, envelope)
	return err
}

func (c *GraphClient) SendDraft(messageID string) error {
	_, err := c.do("POST", fmt.Sprintf("/me/messages/%s/send", messageID), nil, nil)
	return err
}

func (c *GraphClient) AddAttachment(messageID string, att Attachment) error {
	body := map[string]any{
		"@odata.type":  "#microsoft.graph.fileAttachment",
		"name":         att.Name,
		"contentType":  att.ContentType,
		"contentBytes": base64.StdEncoding.EncodeToString(att.ContentBytes),
	}
	_, err := c.do("POST", fmt.Sprintf("/me/messages/%s/attachments", messageID), nil, body)
	return err
}
