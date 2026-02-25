package mail

import (
	"bytes"
	"fmt"
	"html/template"
)

var welcomeTpl = template.Must(template.New("welcome").Parse(`<!DOCTYPE html>
<html><body>
<h2>Welcome to {{.SiteName}}</h2>
<p>Your account has been created. Here are your login credentials:</p>
<p><strong>Email:</strong> {{.Email}}</p>
<p><strong>Temporary Password:</strong> {{.Password}}</p>
<p>Please change your password after first login.</p>
</body></html>`))

var disabledTpl = template.Must(template.New("disabled").Parse(`<!DOCTYPE html>
<html><body>
<h2>Account Disabled</h2>
<p>Your account ({{.Email}}) on {{.SiteName}} has been disabled by an administrator.</p>
<p>If you believe this is a mistake, please contact support.</p>
</body></html>`))

// RenderWelcome renders the welcome email template with the given parameters.
func RenderWelcome(siteName, email, password string) (string, error) {
	var buf bytes.Buffer
	err := welcomeTpl.Execute(&buf, map[string]string{
		"SiteName": siteName,
		"Email":    email,
		"Password": password,
	})
	return buf.String(), err
}

// RenderDisabled renders the account disabled email template.
func RenderDisabled(siteName, email string) (string, error) {
	var buf bytes.Buffer
	err := disabledTpl.Execute(&buf, map[string]string{
		"SiteName": siteName,
		"Email":    email,
	})
	return buf.String(), err
}

var passwordResetTpl = template.Must(template.New("password_reset").Parse(`<!DOCTYPE html>
<html><body>
<h2>Password Reset</h2>
<p>You requested a password reset on {{.SiteName}}.</p>
<p>Click the link below to reset your password (expires in {{.ExpiryMinutes}} minutes):</p>
<p><a href="{{.ResetURL}}">Reset Password</a></p>
<p>If you did not request this, please ignore this email.</p>
</body></html>`))

// RenderPasswordReset renders the password reset email template.
func RenderPasswordReset(siteName, resetURL string, expiryMinutes int) (string, error) {
	var buf bytes.Buffer
	err := passwordResetTpl.Execute(&buf, map[string]any{
		"SiteName":      siteName,
		"ResetURL":      resetURL,
		"ExpiryMinutes": fmt.Sprintf("%d", expiryMinutes),
	})
	return buf.String(), err
}

var newCommentTpl = template.Must(template.New("new_comment").Parse(`<!DOCTYPE html>
<html><body>
<h2>New Comment on {{.SiteName}}</h2>
<p>A new comment was posted on <strong>{{.PostTitle}}</strong> by <strong>{{.AuthorName}}</strong>:</p>
<blockquote>{{.Content}}</blockquote>
<p>Log in to your admin panel to moderate this comment.</p>
</body></html>`))

// RenderNewComment renders the new comment notification email template.
func RenderNewComment(siteName, postTitle, authorName, content string) (string, error) {
	var buf bytes.Buffer
	err := newCommentTpl.Execute(&buf, map[string]string{
		"SiteName":   siteName,
		"PostTitle":  postTitle,
		"AuthorName": authorName,
		"Content":    content,
	})
	return buf.String(), err
}

var commentReplyTpl = template.Must(template.New("comment_reply").Parse(`<!DOCTYPE html>
<html><body>
<h2>Reply to Your Comment on {{.SiteName}}</h2>
<p><strong>{{.ReplierName}}</strong> replied to your comment on <strong>{{.PostTitle}}</strong>:</p>
<blockquote>{{.Content}}</blockquote>
</body></html>`))

// RenderCommentReply renders the comment reply notification email template.
func RenderCommentReply(siteName, postTitle, replierName, content string) (string, error) {
	var buf bytes.Buffer
	err := commentReplyTpl.Execute(&buf, map[string]string{
		"SiteName":    siteName,
		"PostTitle":   postTitle,
		"ReplierName": replierName,
		"Content":     content,
	})
	return buf.String(), err
}
