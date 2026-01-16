package notify

import (
	"fmt"
	"sort"
	"strings"
)

// Status represents the status of a notification
type Status string

const (
	StatusInfo  Status = "info"
	StatusError Status = "error"
)

// MessageBuilder creates notifications with fixed title/metadata and dynamic content
type MessageBuilder struct {
	title        string
	baseMetadata map[string]string
}

// NewMessageBuilder creates a new MessageBuilder with fixed title and base metadata
func NewMessageBuilder(title string, baseMetadata map[string]string) *MessageBuilder {
	if baseMetadata == nil {
		baseMetadata = make(map[string]string)
	}
	return &MessageBuilder{
		title:        title,
		baseMetadata: baseMetadata,
	}
}

// Notification represents a single notification instance
type Notification struct {
	builder *MessageBuilder
	status  Status
	fields  map[string]string
	content string
}

// NewNotification creates a new notification with dynamic status, fields, and content
func (mb *MessageBuilder) NewNotification(status Status) *Notification {
	return &Notification{
		builder: mb,
		status:  status,
		fields:  make(map[string]string),
	}
}

// WithField adds a dynamic key-value field to the notification
func (n *Notification) WithField(key, value string) *Notification {
	n.fields[key] = value
	return n
}

// WithContent sets the content of the notification
func (n *Notification) WithContent(content string) *Notification {
	n.content = content
	return n
}

// GetTitle returns the formatted title with status icon
func (n *Notification) GetTitle() string {
	icon := "✅"
	if n.status == StatusError {
		icon = "❌"
	}
	return fmt.Sprintf("%s *%s*\n", icon, n.builder.title)
}

// GetBodyTelegram returns the Telegram-formatted body
func (n *Notification) GetBodyTelegram() string {
	var builder strings.Builder

	// Combine base metadata and dynamic fields
	allFields := make(map[string]string)
	for k, v := range n.builder.baseMetadata {
		allFields[k] = v
	}
	for k, v := range n.fields {
		allFields[k] = v
	}

	// Sort keys alphabetically
	keys := make([]string, 0, len(allFields))
	for key := range allFields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Format fields
	for _, key := range keys {
		value := allFields[key]
		builder.WriteString(fmt.Sprintf("- *%s*: `%s`\n", key, value))
	}

	// Content section
	if n.content != "" {
		builder.WriteString("\n*Content*\n")
		builder.WriteString("```text\n")
		builder.WriteString(n.content)
		builder.WriteString("```")
	}

	return builder.String()
}

// GetBodyMSTeams returns the MS Teams-formatted body
func (n *Notification) GetBodyMSTeams() string {
	var builder strings.Builder

	// Combine base metadata and dynamic fields
	allFields := make(map[string]string)
	for k, v := range n.builder.baseMetadata {
		allFields[k] = v
	}
	for k, v := range n.fields {
		allFields[k] = v
	}

	// Sort keys alphabetically
	keys := make([]string, 0, len(allFields))
	for key := range allFields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Format fields
	for _, key := range keys {
		value := allFields[key]
		builder.WriteString(fmt.Sprintf("**%s:** %s  \n", key, value))
	}

	// Content section
	if n.content != "" {
		builder.WriteString("\n---\n\n")
		builder.WriteString("**Details:**\n\n")
		builder.WriteString("```\n")
		builder.WriteString(n.content)
		builder.WriteString("\n```\n")
	}

	return builder.String()
}
