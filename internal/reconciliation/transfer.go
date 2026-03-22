package reconciliation

import (
	"encoding/json"
	"fmt"
	"strings"
)

type TransferProvider struct {
	// Simple token-based authentication for bank transfer alerts from an internal listener
	AuthToken string
}

func (t *TransferProvider) GetName() string {
	return "transfer"
}

func (t *TransferProvider) VerifyWebhook(payload []byte, headers map[string]string) (*NormalizedTransaction, error) {
	// Custom listener for bank NIP alerts
	auth := headers["authorization"]
	if auth != "Bearer "+t.AuthToken {
		return nil, fmt.Errorf("unauthorized bank alert")
	}

	var data struct {
		BankName     string  `json:"bankName"`
		SessionID    string  `json:"sessionId"`
		Amount       float64 `json:"amount"`
		Narration    string  `json:"narration"` // We expect the reference here
		CustomerName string  `json:"customerName"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bank alert: %w", err)
	}

	// We extract the reference from the narration (BD-XX-XXXXXX)
	reference := ""
	if strings.Contains(data.Narration, "BD-") {
		parts := strings.Split(data.Narration, " ")
		for _, p := range parts {
			if strings.HasPrefix(p, "BD-") {
				reference = p
				break
			}
		}
	}

	return &NormalizedTransaction{
		ExternalRef: data.SessionID,
		InternalRef: reference,
		Amount:      data.Amount,
		Fee:         0,
		HardwareID:  data.BankName,
		Status:      "SUCCESS", // If the alert exists, it's successful
		Currency:    "NGN",
		Raw:         string(payload),
	}, nil
}
