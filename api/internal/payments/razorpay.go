package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

func VerifyCheckoutSignature(orderID, paymentID, signature, secret string) error {
	if secret == "" {
		return nil
	}
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(orderID + "|" + paymentID))
	expectedHex := hex.EncodeToString(m.Sum(nil))
	if !hmac.Equal([]byte(strings.ToLower(expectedHex)), []byte(strings.ToLower(signature))) {
		return errors.New("invalid razorpay checkout signature")
	}
	return nil
}

func VerifyWebhookSignature(body []byte, signature, secret string) error {
	if secret == "" {
		return nil
	}
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(body)
	expectedHex := hex.EncodeToString(m.Sum(nil))
	if !hmac.Equal([]byte(strings.ToLower(expectedHex)), []byte(strings.ToLower(signature))) {
		return errors.New("invalid razorpay webhook signature")
	}
	return nil
}
