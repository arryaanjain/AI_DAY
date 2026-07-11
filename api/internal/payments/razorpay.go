package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

func VerifyCheckoutSignature(orderID, paymentID, signature, secret string) error {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(orderID + "|" + paymentID))
	expected := m.Sum(nil)
	actual, err := hex.DecodeString(signature)
	if err != nil || !hmac.Equal(expected, actual) {
		return errors.New("invalid razorpay checkout signature")
	}
	return nil
}
func VerifyWebhookSignature(body []byte, signature, secret string) error {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(body)
	expected := m.Sum(nil)
	actual, err := hex.DecodeString(signature)
	if err != nil || !hmac.Equal(expected, actual) {
		return errors.New("invalid razorpay webhook signature")
	}
	return nil
}
