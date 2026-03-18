package graph

type EmailAddress struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Recipient struct {
	EmailAddress EmailAddress `json:"emailAddress"`
}

type ItemBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type MailFolder struct {
	ID              string `json:"id"`
	DisplayName     string `json:"displayName"`
	UnreadItemCount int    `json:"unreadItemCount"`
	TotalItemCount  int    `json:"totalItemCount"`
	ParentFolderID  string `json:"parentFolderId"`
}

type AttachmentInfo struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	ContentType          string `json:"contentType"`
	Size                 int    `json:"size"`
	IsInline             bool   `json:"isInline"`
	LastModifiedDateTime string `json:"lastModifiedDateTime"`
	ContentBytes         string `json:"contentBytes,omitempty"`
	ContentText          string `json:"contentText,omitempty"`
}

type Message struct {
	ID               string           `json:"id"`
	Subject          string           `json:"subject"`
	Sender           *Recipient       `json:"sender"`
	From             *Recipient       `json:"from"`
	ToRecipients     []Recipient      `json:"toRecipients"`
	CcRecipients     []Recipient      `json:"ccRecipients"`
	ReceivedDateTime string           `json:"receivedDateTime"`
	IsRead           bool             `json:"isRead"`
	HasAttachments   bool             `json:"hasAttachments"`
	Importance       string           `json:"importance"`
	BodyPreview      string           `json:"bodyPreview"`
	Body             *ItemBody        `json:"body"`
	WebLink          string           `json:"webLink"`
	Attachments      []AttachmentInfo `json:"attachments,omitempty"`
}

type DateTimeTimeZone struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type Location struct {
	DisplayName string `json:"displayName"`
}

type ResponseStatus struct {
	Response string `json:"response"`
	Time     string `json:"time"`
}

type Attendee struct {
	EmailAddress EmailAddress   `json:"emailAddress"`
	Type         string         `json:"type"`
	Status       ResponseStatus `json:"status"`
}

type PatternedRecurrence struct {
	Pattern map[string]any `json:"pattern"`
	Range   map[string]any `json:"range"`
}

type Calendar struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	IsDefaultCalendar bool          `json:"isDefaultCalendar"`
	Owner             *EmailAddress `json:"owner"`
	Color             string        `json:"color"`
}

type Event struct {
	ID               string               `json:"id"`
	Subject          string               `json:"subject"`
	Start            *DateTimeTimeZone    `json:"start"`
	End              *DateTimeTimeZone    `json:"end"`
	Location         *Location            `json:"location"`
	IsAllDay         bool                 `json:"isAllDay"`
	ShowAs           string               `json:"showAs"`
	Organizer        *Recipient           `json:"organizer"`
	Attendees        []Attendee           `json:"attendees"`
	Body             *ItemBody            `json:"body"`
	Recurrence       *PatternedRecurrence `json:"recurrence"`
	WebLink          string               `json:"webLink"`
	OnlineMeetingURL string               `json:"onlineMeetingUrl"`
}

type ScheduleItem struct {
	Status    string            `json:"status"`
	Start     *DateTimeTimeZone `json:"start"`
	End       *DateTimeTimeZone `json:"end"`
	Subject   string            `json:"subject"`
	Location  string            `json:"location"`
	IsPrivate bool              `json:"isPrivate"`
}

type ScheduleInformation struct {
	ScheduleID       string         `json:"scheduleId"`
	AvailabilityView string         `json:"availabilityView"`
	ScheduleItems    []ScheduleItem `json:"scheduleItems"`
	Error            *GraphError    `json:"error,omitempty"`
}

// Response wrappers for paginated Graph API responses
type listResponse[T any] struct {
	Value []T `json:"value"`
}
