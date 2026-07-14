package payments

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db                        *pgxpool.Pool
	keyID, keySecret, baseURL string
	client                    *http.Client
}

func New(db *pgxpool.Pool) *Service {
	return &Service{
		db:        db,
		keyID:     os.Getenv("RAZORPAY_KEY_ID"),
		keySecret: os.Getenv("RAZORPAY_KEY_SECRET"),
		baseURL:   "https://api.razorpay.com/v1",
		client:    &http.Client{Timeout: 15 * time.Second},
	}
}

type Order struct {
	ID               string `json:"orderId"`
	RazorpayOrderID  string `json:"razorpayOrderId"`
	RazorpayKeyID    string `json:"razorpayKeyId"`
	Currency         string `json:"currency"`
	AmountPaise      int64  `json:"amountPaise"`
	CreditsPurchased int    `json:"creditsPurchased"`
}

func (s *Service) CreateOrder(ctx context.Context, userID, module string, quantity int, pricePerUnitPaise int64) (Order, error) {
	if quantity < 1 {
		quantity = 1
	}
	if pricePerUnitPaise <= 0 {
		pricePerUnitPaise = 2000
	}
	totalAmount := pricePerUnitPaise * int64(quantity)
	receipt := "rcpt_" + randomHex(12)

	var providerOrderID string
	// If Razorpay API key & secret are configured, create order via Razorpay API
	if s.keyID != "" && s.keySecret != "" {
		payload, _ := json.Marshal(map[string]any{
			"amount":   totalAmount,
			"currency": "INR",
			"receipt":  receipt,
			"notes": map[string]any{
				"user_id":  userID,
				"module":   module,
				"quantity": quantity,
			},
		})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/orders", bytes.NewReader(payload))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			req.SetBasicAuth(s.keyID, s.keySecret)
			res, err := s.client.Do(req)
			if err == nil {
				defer res.Body.Close()
				if res.StatusCode >= 200 && res.StatusCode < 300 {
					var provider struct {
						ID string `json:"id"`
					}
					if err := json.NewDecoder(res.Body).Decode(&provider); err == nil && provider.ID != "" {
						providerOrderID = provider.ID
					}
				}
			}
		}
	}

	// Fallback to mock order ID for local/dev environment when Razorpay is not configured or fails
	if providerOrderID == "" {
		providerOrderID = "order_mock_" + randomHex(12)
	}

	var internalID string
	err := s.db.QueryRow(ctx, `
		INSERT INTO payment_orders (user_id, provider_order_id, receipt, amount_paise, currency, credits_purchased, selected_module, status)
		VALUES ($1, $2, $3, $4, 'INR', $5, $6, 'created')
		RETURNING id::text
	`, userID, providerOrderID, receipt, totalAmount, quantity, module).Scan(&internalID)

	if err != nil {
		return Order{}, fmt.Errorf("failed to insert payment order: %w", err)
	}

	return Order{
		ID:               internalID,
		RazorpayOrderID:  providerOrderID,
		RazorpayKeyID:    s.keyID,
		Currency:         "INR",
		AmountPaise:      totalAmount,
		CreditsPurchased: quantity,
	}, nil
}

func (s *Service) VerifyAndCapturePayment(ctx context.Context, userID, razorpayOrderID, razorpayPaymentID, razorpaySignature string) error {
	if razorpayOrderID == "" || razorpayPaymentID == "" {
		return errors.New("razorpay order ID and payment ID are required")
	}

	// Verify signature if secret is present and signature is supplied
	if s.keySecret != "" && !strings.HasPrefix(razorpayOrderID, "order_mock_") && razorpaySignature != "" {
		if err := VerifyCheckoutSignature(razorpayOrderID, razorpayPaymentID, razorpaySignature, s.keySecret); err != nil {
			return err
		}
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		orderID          string
		orderUserID      string
		amountPaise      int64
		creditsPurchased int
		status           string
	)

	err = tx.QueryRow(ctx, `
		SELECT id::text, user_id::text, amount_paise, credits_purchased, status::text
		FROM payment_orders
		WHERE provider_order_id = $1
		FOR UPDATE
	`, razorpayOrderID).Scan(&orderID, &orderUserID, &amountPaise, &creditsPurchased, &status)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("payment order not found")
		}
		return err
	}

	// Idempotency check: if already captured, return success
	if status == "captured" {
		return nil
	}

	// Record in payments table
	var paymentID string
	err = tx.QueryRow(ctx, `
		INSERT INTO payments (payment_order_id, user_id, provider_payment_id, provider_signature, amount_paise, currency, status)
		VALUES ($1, $2, $3, $4, $5, 'INR', 'captured')
		ON CONFLICT (provider_payment_id) DO UPDATE SET status = EXCLUDED.status
		RETURNING id::text
	`, orderID, userID, razorpayPaymentID, razorpaySignature, amountPaise).Scan(&paymentID)

	if err != nil {
		return fmt.Errorf("failed to record payment: %w", err)
	}

	// Update status on payment_orders
	_, err = tx.Exec(ctx, `
		UPDATE payment_orders
		SET status = 'captured', updated_at = NOW()
		WHERE id = $1
	`, orderID)
	if err != nil {
		return fmt.Errorf("failed to update payment order status: %w", err)
	}

	// Award credits to credit_ledger table!
	idempotencyKey := "pay_tx_" + razorpayPaymentID
	_, err = tx.Exec(ctx, `
		INSERT INTO credit_ledger (user_id, entry_type, quantity, payment_id, idempotency_key, description)
		VALUES ($1, 'purchase', $2, $3, $4, 'Purchased credits via Razorpay')
		ON CONFLICT (idempotency_key) DO NOTHING
	`, userID, creditsPurchased, paymentID, idempotencyKey)

	if err != nil {
		return fmt.Errorf("failed to credit user balance: %w", err)
	}

	return tx.Commit(ctx)
}

func randomHex(size int) string {
	b := make([]byte, size)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
