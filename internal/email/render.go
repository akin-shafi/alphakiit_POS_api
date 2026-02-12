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
	VerificationURL string
	Message         string // for security alerts, etc.

	// Daily Report Data
	BusinessName     string
	Currency         string
	TotalSales       string
	TransactionCount int
	CashSales        string
	CardSales        string
	TransferSales    string
	Date             string
	LowStockItems    []EmailInventoryItem
	SecurityAlerts   []string
}

type EmailInventoryItem struct {
	Name  string
	Stock int
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
	tmpl, err := template.ParseFS(templateFS,
		append(baseTemplates, contentPath)...,
	)
	if err != nil {
		return "", "", err
	}

	var buf bytes.Buffer
	// Execute the specific content template using its full path
	if err := tmpl.ExecuteTemplate(&buf, contentPath, data); err != nil {
		return "", "", err
	}

	return data.Subject, buf.String(), nil
}
