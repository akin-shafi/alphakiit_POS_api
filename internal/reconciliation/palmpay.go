package reconciliation

import (
	"crypto/hmac"
	"crypto/sha256" // PalmPay sometimes uses SHA256
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type PalmPayProvider struct {
	SecretKey string
}

func (p *PalmPayProvider) GetName() string {
	return "palmpay"
}

func (p *PalmPayProvider) VerifyWebhook(payload []byte, headers map[string]string) (*NormalizedTransaction, error) {
	// PalmPay typically uses a signature based on the raw payload + secret
	// Note: Exact signature header and algorithm vary by PalmPay API version.
	signature := headers["palmpay-signature"]
	if signature == "" {
		return nil, fmt.Errorf("missing palmpay signature")
	}

	h := hmac.New(sha256.New, []byte(p.SecretKey))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, fmt.Errorf("invalid palmpay signature")
	}

	var data struct {
		Event      string  `json:"eventType"`
		Reference  string  `json:"orderNo"`
		ExtRef     string  `json:"palmpayNo"`
		Amount     float64 `json:"amount"`
		Currency   string  `json:"currency"`
		Status     string  `json:"status"`
		TerminalSN string  `json:"sn"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal palmpay payload: %w", err)
	}

	status := "FAILED"
	if strings.ToUpper(data.Status) == "SUCCESS" {
		status = "SUCCESS"
	}

	// Calculate fee (0.5% max 100)
	fee := data.Amount * 0.005
	if fee > 100 {
		fee = 100
	}

	return &NormalizedTransaction{
		ExternalRef: data.ExtRef,
		InternalRef: data.Reference,
		Amount:      data.Amount,
		Fee:         fee,
		HardwareID:  data.TerminalSN,
		Status:      status,
		Currency:    data.Currency,
		Raw:         string(payload),
	}, nil
}
