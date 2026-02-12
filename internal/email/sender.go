// internal/email/sender.go
package email

import (
	"gopkg.in/mail.v2"
)

type Sender struct {
	config Config
	dialer *mail.Dialer
}

func NewSender(cfg Config) *Sender {
	return &Sender{
		config: cfg,
		dialer: cfg.Dialer(),
	}
}

type EmailVerificationData struct {
	Name            string // optional: user's first name
	VerificationURL string // full URL with token, e.g. https://app.posfiber.com/verify?token=abc123
}

func (s *Sender) SendPasswordResetOTP(toEmail, otp string) error {
	data := EmailData{
		AppName:      s.config.AppName,
		AppURL:       s.config.AppURL,
		SupportEmail: s.config.SupportEmail,
		Subject:      "Your Password Reset OTP",
		OTP:          otp,
	}

	subject, htmlBody, err := RenderTemplate("reset_password_otp.html", data)
	if err != nil {
		return err
	}

	m := mail.NewMessage()
	m.SetHeader("From", s.config.SMTPFrom)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	return s.dialer.DialAndSend(m)
}

func (s *Sender) SendWelcomeEmail(toEmail, name string) error {
	data := EmailData{
		AppName:      s.config.AppName,
		AppURL:       s.config.AppURL,
		SupportEmail: s.config.SupportEmail,
		Subject:      "Welcome to " + s.config.AppName,
		Name:         name,
	}
	subject, htmlBody, err := RenderTemplate("welcome.html", data)
	if err != nil {
		return err
	}
	m := mail.NewMessage()
	m.SetHeader("From", s.config.SMTPFrom)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)
	return s.dialer.DialAndSend(m)
}

func (s *Sender) SendCustomEmail(toEmail, subject, htmlBody string) error {
	m := mail.NewMessage()
	m.SetHeader("From", s.config.SMTPFrom)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)
	return s.dialer.DialAndSend(m)
}

func (s *Sender) SendEmailVerification(toEmail, name, verificationURL string) error {
	data := EmailData{
		AppName:         s.config.AppName,
		AppURL:          s.config.AppURL,
		SupportEmail:    s.config.SupportEmail,
		Subject:         "Verify Your Email Address",
		Name:            name,
		VerificationURL: verificationURL,
	}

	subject, htmlBody, err := RenderTemplate("verification.html", data)
	if err != nil {
		return err
	}

	m := mail.NewMessage()
	m.SetHeader("From", s.config.SMTPFrom)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	return s.dialer.DialAndSend(m)
}

func (s *Sender) SendSecurityEmail(toEmail, name, subject, message string) error {
	data := EmailData{
		AppName:      s.config.AppName,
		AppURL:       s.config.AppURL,
		SupportEmail: s.config.SupportEmail,
		Subject:      subject,
		Name:         name,
		Message:      message,
	}

	renderedSubject, htmlBody, err := RenderTemplate("security_alert.html", data)
	if err != nil {
		return err
	}

	m := mail.NewMessage()
	m.SetHeader("From", s.config.SMTPFrom)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", renderedSubject)
	m.SetBody("text/html", htmlBody)

	return s.dialer.DialAndSend(m)
}
func (s *Sender) SendDailyReport(toEmail string, data EmailData) error {
	data.AppName = s.config.AppName
	data.AppURL = s.config.AppURL
	data.SupportEmail = s.config.SupportEmail
	data.Subject = "Daily Business Report: " + data.BusinessName

	renderedSubject, htmlBody, err := RenderTemplate("daily_report.html", data)
	if err != nil {
		return err
	}

	m := mail.NewMessage()
	m.SetHeader("From", s.config.SMTPFrom)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", renderedSubject)
	m.SetBody("text/html", htmlBody)

	return s.dialer.DialAndSend(m)
}
