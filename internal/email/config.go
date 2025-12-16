// internal/email/config.go
package email

import (
	"os"
	"strconv"

	"gopkg.in/mail.v2"
)

type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	UseTLS       bool

	AppName      string
	AppURL       string
	SupportEmail string
}

func LoadConfig() Config {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))

	useTLS := true
	if os.Getenv("SMTP_TLS") == "false" {
		useTLS = false
	}

	return Config{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     port,
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     os.Getenv("SMTP_FROM"),

		UseTLS: useTLS,

		AppName:      getEnv("APP_NAME", "POS Fiber App"),
		AppURL:       getEnv("APP_URL", "https://app.posfiber.com"),
		SupportEmail: getEnv("SUPPORT_EMAIL", "support@posfiber.com"),
	}
}

// Helper to provide fallback defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (cfg Config) Dialer() *mail.Dialer {
	d := mail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)

	if cfg.UseTLS {
		d.StartTLSPolicy = mail.MandatoryStartTLS
	}
	// d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // only for dev/testing

	return d
}
