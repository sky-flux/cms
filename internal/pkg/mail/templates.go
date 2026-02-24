package mail

import (
	"bytes"
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
