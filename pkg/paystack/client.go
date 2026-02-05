package paystack

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

type PaystackClient struct {
	SecretKey string
	BaseURL   string
}

func NewClient() *PaystackClient {
	return &PaystackClient{
		SecretKey: os.Getenv("PAYSTACK_SECRET_KEY"),
		BaseURL:   "https://api.paystack.co",
	}
}

type VerificationResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Amount    float64 `json:"amount"`
		Currency  string  `json:"currency"`
		Status    string  `json:"status"`
		Reference string  `json:"reference"`
		PaidAt    string  `json:"paid_at"`
		Customer  struct {
			Email string `json:"email"`
		} `json:"customer"`
	} `json:"data"`
}

func (c *PaystackClient) VerifyTransaction(reference string) (*VerificationResponse, error) {
	if c.SecretKey == "" {
		return nil, errors.New("PAYSTACK_SECRET_KEY not set")
	}

	url := fmt.Sprintf("%s/transaction/verify/%s", c.BaseURL, reference)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paystack returned status code %d", resp.StatusCode)
	}

	var result VerificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Status {
		return nil, errors.New(result.Message)
	}

	if result.Data.Status != "success" {
		return nil, fmt.Errorf("transaction not successful: status is %s", result.Data.Status)
	}

	return &result, nil
}
