package email

import "time"

// Email is exactly what it sounds like
type Email struct {
	// From: REQUIRED The sender email address. Must have a registered and confirmed Sender Signature.
	From string `json:",omitempty"`
	// To: REQUIRED Recipient email address. Multiple addresses are comma separated. Max 50.
	To string `json:",omitempty"`
	// Subject: Email subject
	Subject string `json:",omitempty"`
	// HtmlBody: HTML email message. REQUIRED, If no TextBody specified
	HtmlBody string `json:",omitempty"`
	// TextBody: Plain text email message. REQUIRED, If no HtmlBody specified
	TextBody string `json:",omitempty"`
	// MessageStream is the way Postmark separates emails
	MessageStream string `json:",omitempty"`
	// Attachments: List of attachments
	Attachments []Attachment `json:",omitempty"`
}

// Header - an email header
type Header struct {
	// Name: header name
	Name string
	// Value: header value
	Value string
}

// Attachment is an optional encoded file to send along with an email
type Attachment struct {
	// Name: attachment name
	Name string
	// Content: Base64 encoded attachment data
	Content string
	// ContentType: attachment MIME type
	ContentType string
	// ContentId: populate for inlining images with the images cid
	ContentID string `json:",omitempty"`
}

// Response holds info in response to a send/send-batch request
// Even if API request comes back successful, check the ErrorCode to see if there might be a delivery problem
type Response struct {
	// To: Recipient email address
	To string
	// SubmittedAt: Timestamp
	SubmittedAt time.Time
	// MessageID: ID of message
	MessageID string
	// ErrorCode: API Error Codes
	ErrorCode int64
	// Message: Response message
	Message string
}
