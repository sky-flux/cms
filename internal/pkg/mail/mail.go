package mail

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v3"
)

// Sender defines the interface for sending emails.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// Message represents an email to be sent.
type Message struct {
	To      string
	Subject string
	HTML    string
}

// ResendSender sends emails via the Resend API.
type ResendSender struct {
	client    *resend.Client
	fromName  string
	fromEmail string
}

// NewResendSender creates a new ResendSender with the given API key and sender info.
func NewResendSender(apiKey, fromName, fromEmail string) *ResendSender {
	return &ResendSender{
		client:    resend.NewClient(apiKey),
		fromName:  fromName,
		fromEmail: fromEmail,
	}
}

// Send delivers an email message via Resend.
func (s *ResendSender) Send(ctx context.Context, msg Message) error {
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{msg.To},
		Subject: msg.Subject,
		Html:    msg.HTML,
	}
	_, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("resend send: %w", err)
	}
	return nil
}

// NoopSender is a no-op implementation of Sender for testing and development.
type NoopSender struct{}

// Send does nothing and returns nil.
func (n *NoopSender) Send(_ context.Context, _ Message) error { return nil }
