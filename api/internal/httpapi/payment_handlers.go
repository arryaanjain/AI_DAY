package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/arryaanjain/AI_DAY/internal/auth"
	"github.com/arryaanjain/AI_DAY/internal/config"
	"github.com/arryaanjain/AI_DAY/internal/payments"
)

type orderRequest struct {
	Module   string `json:"module"`
	Quantity int    `json:"quantity"`
}

type verifyPaymentRequest struct {
	RazorpayOrderID   string `json:"razorpayOrderId"`
	RazorpayPaymentID string `json:"razorpayPaymentId"`
	RazorpaySignature string `json:"razorpaySignature"`
}

func createPaymentOrder(cfg config.Config, a *auth.Service, p *payments.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, a)
		if !ok {
			return
		}
		var request orderRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || (request.Module != "pixel_portrait" && request.Module != "comic") || request.Quantity < 1 {
			errorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "A valid module and quantity are required.")
			return
		}
		order, err := p.CreateOrder(r.Context(), user.ID, request.Module, request.Quantity, cfg.PricePerGenerationPaise)
		if err != nil {
			errorResponse(w, http.StatusServiceUnavailable, "PAYMENT_PROVIDER_UNAVAILABLE", "Unable to create payment order.")
			return
		}
		respond(w, http.StatusCreated, map[string]any{
			"orderId":          order.ID,
			"razorpayOrderId":  order.RazorpayOrderID,
			"razorpayKeyId":    order.RazorpayKeyID,
			"amountPaise":      order.AmountPaise,
			"currency":         order.Currency,
			"creditsPurchased": order.CreditsPurchased,
		})
	}
}

func verifyPaymentHandler(a *auth.Service, p *payments.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, a)
		if !ok {
			return
		}
		var request verifyPaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.RazorpayOrderID == "" || request.RazorpayPaymentID == "" {
			errorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Razorpay order ID and payment ID are required.")
			return
		}
		err := p.VerifyAndCapturePayment(r.Context(), user.ID, request.RazorpayOrderID, request.RazorpayPaymentID, request.RazorpaySignature)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, "PAYMENT_VERIFICATION_FAILED", err.Error())
			return
		}
		respond(w, http.StatusOK, map[string]any{
			"status":  "success",
			"message": "Payment verified and credits awarded.",
		})
	}
}
