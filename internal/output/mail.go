package output

import (
	"fmt"
	"strings"

	"github.com/dsswift/cli-exchange/internal/graph"
)

func FormatSize(bytes int) string {
	if bytes >= 1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
}

func RenderAttachmentTable(attachments []graph.AttachmentInfo) string {
	headers := []string{"Name", "Type", "Size", "Inline"}
	rows := make([][]string, len(attachments))
	for i, a := range attachments {
		inline := "no"
		if a.IsInline {
			inline = "yes"
		}
		rows[i] = []string{
			a.Name,
			a.ContentType,
			FormatSize(a.Size),
			inline,
		}
	}
	return BuildTable(headers, rows)
}

func RenderFolderTable(folders []graph.MailFolder) string {
	headers := []string{"Name", "Unread", "Total"}
	rows := make([][]string, len(folders))
	for i, f := range folders {
		rows[i] = []string{
			f.DisplayName,
			fmt.Sprintf("%d", f.UnreadItemCount),
			fmt.Sprintf("%d", f.TotalItemCount),
		}
	}
	return BuildTable(headers, rows)
}

func RenderMessageTable(messages []graph.Message) string {
	headers := []string{"Subject", "Sender", "Date", "Read"}
	rows := make([][]string, len(messages))
	for i, m := range messages {
		sender := ""
		if m.Sender != nil {
			sender = m.Sender.EmailAddress.Address
			if sender == "" {
				sender = m.Sender.EmailAddress.Name
			}
		}
		read := "no"
		if m.IsRead {
			read = "yes"
		}
		rows[i] = []string{
			truncate(m.Subject, 50),
			truncate(sender, 30),
			m.ReceivedDateTime,
			read,
		}
	}
	return BuildTable(headers, rows)
}

func RenderMessageDetail(m *graph.Message) string {
	var rows [][]string

	rows = append(rows, []string{"ID", m.ID})
	rows = append(rows, []string{"Subject", m.Subject})

	if m.Sender != nil {
		rows = append(rows, []string{"From", fmt.Sprintf("%s <%s>", m.Sender.EmailAddress.Name, m.Sender.EmailAddress.Address)})
	}

	if len(m.ToRecipients) > 0 {
		for j, r := range m.ToRecipients {
			label := ""
			if j == 0 {
				label = "To"
			}
			rows = append(rows, []string{label, fmt.Sprintf("%s <%s>", r.EmailAddress.Name, r.EmailAddress.Address)})
		}
	}

	if len(m.CcRecipients) > 0 {
		for j, r := range m.CcRecipients {
			label := ""
			if j == 0 {
				label = "CC"
			}
			rows = append(rows, []string{label, fmt.Sprintf("%s <%s>", r.EmailAddress.Name, r.EmailAddress.Address)})
		}
	}

	rows = append(rows, []string{"Date", m.ReceivedDateTime})
	rows = append(rows, []string{"Read", fmt.Sprintf("%t", m.IsRead)})
	rows = append(rows, []string{"Importance", m.Importance})

	if m.HasAttachments {
		if len(m.Attachments) > 0 {
			names := make([]string, len(m.Attachments))
			for i, a := range m.Attachments {
				names[i] = a.Name
			}
			rows = append(rows, []string{"Attachments", strings.Join(names, ", ")})
		} else {
			rows = append(rows, []string{"Attachments", "yes"})
		}
	}

	result := ""
	for _, row := range rows {
		if row[0] != "" {
			result += fmt.Sprintf("%-12s %s\n", row[0]+":", row[1])
		} else {
			result += fmt.Sprintf("%-12s %s\n", "", row[1])
		}
	}

	if m.Body != nil && m.Body.Content != "" {
		result += "\n" + m.Body.Content
	}

	return result
}

func RenderMessageBatch(messages []*graph.Message) string {
	if len(messages) == 0 {
		return "No messages found.\n"
	}

	var result string
	for i, m := range messages {
		if i > 0 {
			result += "\n" + strings.Repeat("-", 72) + "\n\n"
		}
		result += RenderMessageDetail(m)
	}
	return result
}
