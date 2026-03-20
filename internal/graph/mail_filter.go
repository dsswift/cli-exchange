package graph

import "strings"

// messageFilter holds criteria for client-side message filtering.
// Used when the Graph API cannot combine $filter (dates) with $search (sender, subject, etc.).
type messageFilter struct {
	senderTerms    []string // email addresses or name fragments to match (OR logic)
	subject        string   // case-insensitive substring match
	isRead         *bool
	hasAttachments *bool
}

// buildFilter constructs a messageFilter from ListMessagesOptions search-related fields.
// Returns nil if no client-side filtering is needed.
func buildFilter(opts ListMessagesOptions) *messageFilter {
	if opts.Search == "" && opts.Subject == "" && opts.IsRead == nil && opts.HasAttachments == nil {
		return nil
	}

	f := &messageFilter{
		subject:        opts.Subject,
		isRead:         opts.IsRead,
		hasAttachments: opts.HasAttachments,
	}

	if opts.Search != "" {
		f.senderTerms = parseSenderTerms(opts.Search)
	}

	return f
}

// parseSenderTerms extracts sender match terms from the Search field.
// The Search field may be:
//   - A bare name like "cfavero" (no colon) -> match as name fragment
//   - A KQL string like "from:addr1 OR from:addr2" -> extract addresses
func parseSenderTerms(search string) []string {
	if !strings.Contains(search, ":") {
		// Bare name: match against both name and address
		return []string{search}
	}

	// Parse "from:addr1 OR from:addr2" pattern
	var terms []string
	parts := strings.Split(search, " OR ")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if idx := strings.Index(part, "from:"); idx >= 0 {
			addr := strings.TrimSpace(part[idx+5:])
			if addr != "" {
				terms = append(terms, addr)
			}
		}
	}
	return terms
}

// matches returns true if the message satisfies all filter criteria.
func (f *messageFilter) matches(msg Message) bool {
	if len(f.senderTerms) > 0 && !f.matchesSender(msg) {
		return false
	}
	if f.subject != "" && !f.matchesSubject(msg) {
		return false
	}
	if f.isRead != nil && msg.IsRead != *f.isRead {
		return false
	}
	if f.hasAttachments != nil && msg.HasAttachments != *f.hasAttachments {
		return false
	}
	return true
}

// matchesSender checks if the message's From field matches any sender term (OR logic).
// Each term is matched case-insensitively against both the email address and display name.
func (f *messageFilter) matchesSender(msg Message) bool {
	if msg.From == nil {
		return false
	}

	addr := strings.ToLower(msg.From.EmailAddress.Address)
	name := strings.ToLower(msg.From.EmailAddress.Name)

	for _, term := range f.senderTerms {
		t := strings.ToLower(term)
		if strings.Contains(addr, t) || strings.Contains(name, t) {
			return true
		}
	}
	return false
}

// matchesSubject checks if the message subject contains the filter substring (case-insensitive).
func (f *messageFilter) matchesSubject(msg Message) bool {
	return strings.Contains(strings.ToLower(msg.Subject), strings.ToLower(f.subject))
}
