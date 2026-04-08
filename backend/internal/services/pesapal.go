package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/warui/event-ticketing-api/internal/config"
)

type PesapalService struct {
	cfg        *config.Config
	httpClient *http.Client
	baseURL    string
	token      string
	tokenExp   time.Time
}

// Pesapal Authentication
type PesapalAuthRequest struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
}

type PesapalAuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiryDate"`
	Error     string    `json:"error,omitempty"`
	Status    string    `json:"status,omitempty"`
	Message   string    `json:"message,omitempty"`
}

// Pesapal Submit Order Request
type PesapalSubmitOrderRequest struct {
	ID              string                 `json:"id"`
	Currency        string                 `json:"currency"`
	Amount          float64                `json:"amount"`
	Description     string                 `json:"description"`
	CallbackURL     string                 `json:"callback_url"`
	NotificationID  string                 `json:"notification_id,omitempty"` // Optional
	BillingAddress  PesapalBillingAddress  `json:"billing_address"`
}

// Pesapal IPN Registration
type PesapalIPNRequest struct {
	URL            string `json:"url"`
	IPNType        string `json:"ipn_notification_type"` // "GET" or "POST"
}

type PesapalIPNResponse struct {
	URL                     string `json:"url"`
	CreatedDate             string `json:"created_date"`
	IPNID                   string `json:"ipn_id"`
	Error                   string `json:"error,omitempty"`
	Status                  string `json:"status,omitempty"`
	IPNNotificationType     string `json:"ipn_notification_type"`
}

type PesapalIPNListResponse struct {
	IPNs []PesapalIPNResponse `json:"ipn_list"`
}

type PesapalBillingAddress struct {
	EmailAddress string `json:"email_address"`
	PhoneNumber  string `json:"phone_number,omitempty"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
}

type PesapalSubmitOrderResponse struct {
	OrderTrackingID string `json:"order_tracking_id"`
	MerchantReference string `json:"merchant_reference"`
	RedirectURL     string `json:"redirect_url"`
	Error           string `json:"error,omitempty"`
	Status          string `json:"status,omitempty"`
	Message         string `json:"message,omitempty"`
}

// Pesapal Transaction Status
type PesapalTransactionStatusResponse struct {
	PaymentMethod         string    `json:"payment_method"`
	Amount                float64   `json:"amount"`
	CreatedDate           time.Time `json:"created_date"`
	ConfirmationCode      string    `json:"confirmation_code"`
	PaymentStatusDescription string `json:"payment_status_description"`
	Description           string    `json:"description"`
	Message               string    `json:"message"`
	PaymentAccount        string    `json:"payment_account"`
	CallBackURL           string    `json:"call_back_url"`
	StatusCode            int       `json:"status_code"`
	MerchantReference     string    `json:"merchant_reference"`
	PaymentStatusCode     string    `json:"payment_status_code"` // "1" = completed, "2" = failed, "3" = reversed
	Currency              string    `json:"currency"`
	Error                 string    `json:"error,omitempty"`
	Status                string    `json:"status,omitempty"`
}

func NewPesapalService(cfg *config.Config) *PesapalService {
	baseURL := "https://pay.pesapal.com/v3"
	if cfg.PesapalEnvironment == "sandbox" {
		baseURL = "https://cybqa.pesapal.com/pesapalv3"
	}

	return &PesapalService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				MaxIdleConns:          10,
				MaxIdleConnsPerHost:   5,
				IdleConnTimeout:       90 * time.Second,
			},
		},
		baseURL: baseURL,
	}
}

// authenticate gets or refreshes the access token
func (p *PesapalService) authenticate() error {
	// Check if token is still valid
	if p.token != "" && time.Now().Before(p.tokenExp) {
		return nil
	}

	if p.cfg.PesapalConsumerKey == "" || p.cfg.PesapalConsumerSecret == "" {
		return fmt.Errorf("pesapal not configured")
	}

	reqBody := PesapalAuthRequest{
		ConsumerKey:    p.cfg.PesapalConsumerKey,
		ConsumerSecret: p.cfg.PesapalConsumerSecret,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	req, err := http.NewRequest("POST", p.baseURL+"/api/Auth/RequestToken", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send auth request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	// Log auth response for debugging
	fmt.Printf("Pesapal Auth Response Status: %d\n", resp.StatusCode)
	fmt.Printf("Pesapal Auth Response Body: %s\n", string(body))

	var result PesapalAuthResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to unmarshal auth response: %w", err)
	}

	if result.Error != "" {
		return fmt.Errorf("pesapal auth error: %s", result.Error)
	}

	if result.Token == "" {
		return fmt.Errorf("no token received from pesapal")
	}

	p.token = result.Token
	p.tokenExp = result.ExpiresAt
	return nil
}

// InitializeTransaction initializes a payment transaction
func (p *PesapalService) InitializeTransaction(email string, amount float64, reference string, metadata map[string]interface{}) (*PesapalSubmitOrderResponse, error) {
	if err := p.authenticate(); err != nil {
		return nil, err
	}

	// Extract customer info from metadata if available
	firstName := ""
	lastName := ""
	if metadata != nil {
		if fn, ok := metadata["first_name"].(string); ok {
			firstName = fn
		}
		if ln, ok := metadata["last_name"].(string); ok {
			lastName = ln
		}
	}

	reqBody := PesapalSubmitOrderRequest{
		ID:             reference,
		Currency:       p.cfg.Currency,
		Amount:         amount,
		Description:    "Event ticket purchase",
		CallbackURL:    p.cfg.PesapalCallbackURL,
		NotificationID: p.cfg.PesapalIPNID, // Can be empty for testing
		BillingAddress: PesapalBillingAddress{
			EmailAddress: email,
			FirstName:    firstName,
			LastName:     lastName,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.baseURL+"/api/Transactions/SubmitOrderRequest", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log response for debugging
	fmt.Printf("Pesapal Response Status: %d\n", resp.StatusCode)
	fmt.Printf("Pesapal Response Body: %s\n", string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result PesapalSubmitOrderResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("pesapal error: %s", result.Error)
	}

	if result.RedirectURL == "" {
		return nil, fmt.Errorf("no redirect URL received from pesapal")
	}

	return &result, nil
}

// VerifyTransaction verifies a payment transaction
func (p *PesapalService) VerifyTransaction(orderTrackingID string) (*PesapalTransactionStatusResponse, error) {
	if err := p.authenticate(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", p.baseURL+"/api/Transactions/GetTransactionStatus?orderTrackingId="+orderTrackingID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result PesapalTransactionStatusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("pesapal error: %s", result.Error)
	}

	return &result, nil
}

// IsTransactionSuccessful checks if a transaction was successful
func (p *PesapalService) IsTransactionSuccessful(verification *PesapalTransactionStatusResponse) bool {
	// Status code 1 = completed/successful
	return verification.PaymentStatusCode == "1"
}

// GetTransactionAmount returns the transaction amount
func (p *PesapalService) GetTransactionAmount(verification *PesapalTransactionStatusResponse) float64 {
	return verification.Amount
}

// ListIPNs retrieves all registered IPNs
func (p *PesapalService) ListIPNs() ([]PesapalIPNResponse, error) {
	if err := p.authenticate(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", p.baseURL+"/api/URLSetup/GetIpnList", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// The response is directly an array of IPNs
	var ipns []PesapalIPNResponse
	if err := json.Unmarshal(body, &ipns); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return ipns, nil
}

// RegisterIPN registers a new IPN URL
func (p *PesapalService) RegisterIPN(url string, ipnType string) (*PesapalIPNResponse, error) {
	if err := p.authenticate(); err != nil {
		return nil, err
	}

	if ipnType == "" {
		ipnType = "GET"
	}

	reqBody := PesapalIPNRequest{
		URL:     url,
		IPNType: ipnType,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.baseURL+"/api/URLSetup/RegisterIPN", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result PesapalIPNResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("pesapal error: %s", result.Error)
	}

	return &result, nil
}
