package mail

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time interface checks
var _ Sender = (*ResendSender)(nil)
var _ Sender = (*NoopSender)(nil)

func TestNoopSender_Send(t *testing.T) {
	s := &NoopSender{}
	err := s.Send(context.Background(), Message{
		To:      "test@example.com",
		Subject: "Test",
		HTML:    "<p>Hello</p>",
	})
	assert.NoError(t, err)
}

func TestRenderWelcome(t *testing.T) {
	html, err := RenderWelcome("My Blog", "user@example.com", "temp123")
	require.NoError(t, err)

	assert.True(t, strings.Contains(html, "My Blog"))
	assert.True(t, strings.Contains(html, "user@example.com"))
	assert.True(t, strings.Contains(html, "temp123"))
	assert.True(t, strings.Contains(html, "<!DOCTYPE html>"))
}

func TestRenderDisabled(t *testing.T) {
	html, err := RenderDisabled("My Blog", "user@example.com")
	require.NoError(t, err)

	assert.True(t, strings.Contains(html, "My Blog"))
	assert.True(t, strings.Contains(html, "user@example.com"))
	assert.True(t, strings.Contains(html, "Account Disabled"))
	assert.True(t, strings.Contains(html, "<!DOCTYPE html>"))
}

func TestNewResendSender(t *testing.T) {
	s := NewResendSender("re_test_key", "Sky Flux", "noreply@example.com")
	assert.NotNil(t, s.client)
	assert.Equal(t, "Sky Flux", s.fromName)
	assert.Equal(t, "noreply@example.com", s.fromEmail)
}
