// internal/email/render.go
package email

import (
	"bytes"
	"embed"
	"html/template"
	"time"
)

//go:embed templates/*
var templateFS embed.FS

// Global data available to all emails
type EmailData struct {
	AppName      string
	AppURL       string
	SupportEmail string
	Year         int // current year for footer
	// Specific data for each email type
	Subject         string
	OTP             string // for password reset
	Name            string // for welcome, etc.
	VerificationURL string // <-- add this
	// Add more as needed
}

var baseTemplates = []string{"templates/header.html", "templates/footer.html"}

func RenderTemplate(templateName string, data EmailData) (subject string, htmlBody string, err error) {
	// Default subject if not provided
	if data.Subject == "" {
		data.Subject = "Notification from " + data.AppName
	}
	// Default year
	if data.Year == 0 {
		data.Year = time.Now().Year()
	}

	// Full path to the content template
	contentPath := "templates/" + templateName

	// Parse all templates: header, footer, and the specific content
	tmpl, err := template.New(templateName).ParseFS(templateFS,
		append(baseTemplates, contentPath)...,
	)
	if err != nil {
		return "", "", err
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", "", err
	}

	return data.Subject, buf.String(), nil
}
