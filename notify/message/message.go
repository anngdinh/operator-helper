package notify

import (
	"fmt"
	"sort"
	"strings"
)

// MessageType represents the type of message
type MessageType string

const (
	// MessageTypeInfo represents an informational message
	MessageTypeInfo MessageType = "info"
	// MessageTypeError represents an error message
	MessageTypeError MessageType = "error"
)

func NewMessage(_type MessageType, title, content string) *Message {
	return &Message{
		_type:      _type,
		title:      title,
		content:    content,
		attributes: make(map[string]string),
	}
}

// Message represents a message to be sent
type Message struct {
	_type      MessageType
	title      string
	attributes map[string]string
	content    string
}

func (m *Message) WithAttribute(key, value string) *Message {
	if m.attributes == nil {
		m.attributes = map[string]string{}
	}
	m.attributes[key] = value
	return m
}

func (m *Message) GetTitle() string {
	var builder strings.Builder

	// Set icon based on message type
	icon := "✅"
	if m._type == MessageTypeError {
		icon = "❌"
	}

	// title
	builder.WriteString(fmt.Sprintf("%s *%s*\n", icon, m.title))
	return builder.String()
}

// GetBodyTelegram converts the message to a Telegram-formatted string
func (m *Message) GetBodyTelegram() string {
	var builder strings.Builder

	// // Set icon based on message type
	// icon := "✅"
	// if m._type == MessageTypeError {
	// 	icon = "❌"
	// }

	// // title
	// builder.WriteString(fmt.Sprintf("%s %s\n", icon, m.title))

	// order attributes alphabetically
	keys := make([]string, 0, len(m.attributes))
	for key := range m.attributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	// attributes
	for _, key := range keys {
		value := m.attributes[key]
		builder.WriteString(fmt.Sprintf("- *%s*: `%s`\n", key, value))
	}

	// content
	if m.content != "" {
		builder.WriteString("\n*Content*\n")
		builder.WriteString("```text\n")
		builder.WriteString(m.content)
		builder.WriteString("```")
	}

	return builder.String()
}

func (m *Message) GetBodyMSTeams() string {
	var builder strings.Builder

	// // Icon based on type
	// icon := "✅"
	// if m._type == MessageTypeError {
	// 	icon = "❌"
	// }

	// // title
	// builder.WriteString(fmt.Sprintf("**%s %s**\n\n", icon, m.title))

	// order attributes alphabetically
	keys := make([]string, 0, len(m.attributes))
	for key := range m.attributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	// attributes
	for _, key := range keys {
		value := m.attributes[key]
		builder.WriteString(fmt.Sprintf("**%s:** %s  \n", key, value)) // double space for line break
	}

	// content section
	if m.content != "" {
		builder.WriteString("\n---\n\n")
		builder.WriteString("**Details:**\n\n")
		builder.WriteString("```\n")
		builder.WriteString(m.content)
		builder.WriteString("\n```\n")
	}

	return builder.String()
}
