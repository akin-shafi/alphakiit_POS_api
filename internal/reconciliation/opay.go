package reconciliation

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type OPayProvider struct {
	SecretKey string
}

func (o *OPayProvider) GetName() string {
	return "opay"
}

func (o *OPayProvider) VerifyWebhook(payload []byte, headers map[string]string) (*NormalizedTransaction, error) {
	// 1. Signature Verification
	// OPay typically sends a signature in 'X-Opay-Signature' or similar. 
	// The signature is HMAC-SHA512 of the raw body using the merchant secret.
	signature := headers["x-opay-signature"]
	if signature == "" {
		return nil, fmt.Errorf("missing OPay signature header")
	}

	h := hmac.New(sha512.New, []byte(o.SecretKey))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, fmt.Errorf("invalid OPay signature. expected %s, got %s", expectedSignature, signature)
	}

	var data struct {
		Event      string `json:"event"`
		Status     string `json:"status"`
		OrderNo    string `json:"reference"`
		Amount     float64 `json:"amount"`
		OutOrderNo string `json:"orderNo"` // Our Reference
		Currency   string `json:"currency"`
		TerminalID string `json:"terminalId"` 
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OPay payload: %w", err)
	}

	// 2. Normalize status
	status := "FAILED"
	if strings.ToLower(data.Status) == "success" {
		status = "SUCCESS"
	}

	// Calculate fee (0.5% max 100)
	fee := data.Amount * 0.005
	if fee > 100 {
		fee = 100
	}

	return &NormalizedTransaction{
		ExternalRef: data.OrderNo,
		InternalRef: data.OutOrderNo,
		Amount:      data.Amount,
		Fee:         fee,
		HardwareID:  data.TerminalID,
		Status:      status,
		Currency:    data.Currency,
		Raw:         string(payload),
	}, nil
}
