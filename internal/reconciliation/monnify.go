package reconciliation

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type MonnifyProvider struct {
	ApiKey    string
	SecretKey string
}

func (m *MonnifyProvider) GetName() string {
	return "monnify"
}

func (m *MonnifyProvider) VerifyWebhook(payload []byte, headers map[string]string) (*NormalizedTransaction, error) {
	// Monnify uses Monnify-Signature header
	// Computed using the Client Secret and the raw payload
	signature := headers["monnify-signature"]
	if signature == "" {
		return nil, fmt.Errorf("missing monnify signature")
	}

	h := hmac.New(sha512.New, []byte(m.SecretKey))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, fmt.Errorf("invalid monnify signature")
	}

	var data struct {
		EventType string `json:"eventType"`
		EventData struct {
			Amount           float64 `json:"amountPaid"`
			Reference        string  `json:"paymentReference"`
			ServiceReference string  `json:"transactionReference"`
			PaymentStatus    string  `json:"paymentStatus"`
			PaymentMethod    string  `json:"paymentMethod"`
			InternalRef      string  `json:"metaData.internal_ref"`
		} `json:"eventData"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal monnify payload: %w", err)
	}

	status := "FAILED"
	if strings.ToUpper(data.EventData.PaymentStatus) == "PAID" {
		status = "SUCCESS"
	}

	// Calculate fee (0.5% max 100)
	fee := data.EventData.Amount * 0.005
	if fee > 100 {
		fee = 100
	}

	return &NormalizedTransaction{
		ExternalRef: data.EventData.Reference,
		InternalRef: data.EventData.InternalRef, // Custom meta data we should send
		Amount:      data.EventData.Amount,
		Fee:         fee,
		HardwareID:  data.EventData.ServiceReference, // Often contains terminal SN
		Status:      status,
		Currency:    "NGN",
		Raw:         string(payload),
	}, nil
}
