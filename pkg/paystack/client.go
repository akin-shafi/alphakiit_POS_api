package paystack

import (
	"bytes"
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
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			Bin               string `json:"bin"`
			Last4             string `json:"last4"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			Channel           string `json:"channel"`
			CardType          string `json:"card_type"`
			Bank              string `json:"bank"`
			CountryCode       string `json:"country_code"`
			Brand             string `json:"brand"`
			Reusable          bool   `json:"reusable"`
			Signature         string `json:"signature"`
			AccountName       string `json:"account_name"`
		} `json:"authorization"`
	} `json:"data"`
}

func (c *PaystackClient) ChargeAuthorization(email string, amount float64, authorizationCode string) (*VerificationResponse, error) {
	if c.SecretKey == "" {
		return nil, errors.New("PAYSTACK_SECRET_KEY not set")
	}

	url := fmt.Sprintf("%s/transaction/charge_authorization", c.BaseURL)
	payload := map[string]interface{}{
		"email":              email,
		"amount":             fmt.Sprintf("%.0f", amount*100), // convert to kobo
		"authorization_code": authorizationCode,
	}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
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

	var result VerificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Status {
		return nil, errors.New(result.Message)
	}

	if result.Data.Status != "success" {
		return nil, fmt.Errorf("charge failed: status is %s", result.Data.Status)
	}

	return &result, nil
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
