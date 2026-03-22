package reconciliation

import (
	"errors"
)

type NormalizedTransaction struct {
	ExternalRef string
	InternalRef string
	Amount      float64
	Fee         float64 // Commission charged by the provider
	HardwareID  string  // The specific POS hardware SN/ID
	Status      string  // SUCCESS, FAILED
	Currency    string
	Raw         string
}

type PaymentProvider interface {
	GetName() string
	VerifyWebhook(payload []byte, headers map[string]string) (*NormalizedTransaction, error)
}

var (
	ErrInvalidSignature = errors.New("invalid webhook signature")
	ErrProviderNotFound = errors.New("payment provider not found")
)
