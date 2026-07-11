package payments

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"os"
	"time"
)

type Service struct {
	db                        *pgxpool.Pool
	keyID, keySecret, baseURL string
	client                    *http.Client
}

func New(db *pgxpool.Pool) *Service {
	return &Service{db: db, keyID: os.Getenv("RAZORPAY_KEY_ID"), keySecret: os.Getenv("RAZORPAY_KEY_SECRET"), baseURL: "https://api.razorpay.com/v1", client: &http.Client{Timeout: 15 * time.Second}}
}

type Order struct {
	ID, RazorpayOrderID, RazorpayKeyID, Currency string
	AmountPaise                                  int64
}

func (s *Service) CreateOrder(ctx context.Context, userID, module string) (Order, error) {
	if s.keyID == "" || s.keySecret == "" {
		return Order{}, errors.New("razorpay is not configured")
	}
	receipt := randomHex(16)
	payload, _ := json.Marshal(map[string]any{"amount": 2000, "currency": "INR", "receipt": receipt, "notes": map[string]string{"user_id": userID, "module": module}})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/orders", bytes.NewReader(payload))
	if err != nil {
		return Order{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.keyID, s.keySecret)
	res, err := s.client.Do(req)
	if err != nil {
		return Order{}, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Order{}, errors.New("razorpay order creation failed")
	}
	var provider struct {
		ID       string `json:"id"`
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	}
	if err = json.NewDecoder(res.Body).Decode(&provider); err != nil {
		return Order{}, err
	}
	var internal string
	err = s.db.QueryRow(ctx, `INSERT INTO payment_orders(user_id,provider_order_id,receipt,amount_paise,currency,selected_module) VALUES($1,$2,$3,$4,$5,$6) RETURNING id::text`, userID, provider.ID, receipt, provider.Amount, provider.Currency, module).Scan(&internal)
	return Order{ID: internal, RazorpayOrderID: provider.ID, RazorpayKeyID: s.keyID, AmountPaise: provider.Amount, Currency: provider.Currency}, err
}
func randomHex(size int) string {
	b := make([]byte, size)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
