package httpapi

import (
	"encoding/json"
	"github.com/arryaanjain/AI_DAY/internal/auth"
	"github.com/arryaanjain/AI_DAY/internal/payments"
	"net/http"
)

type orderRequest struct {
	Module   string `json:"module"`
	Quantity int    `json:"quantity"`
}

func createPaymentOrder(a *auth.Service, p *payments.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, a)
		if !ok {
			return
		}
		var request orderRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || (request.Module != "pixel_portrait" && request.Module != "comic") || request.Quantity != 1 {
			errorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "A valid module and quantity of one are required.")
			return
		}
		order, err := p.CreateOrder(r.Context(), user.ID, request.Module)
		if err != nil {
			errorResponse(w, http.StatusServiceUnavailable, "PAYMENT_PROVIDER_UNAVAILABLE", "Unable to create payment order.")
			return
		}
		respond(w, http.StatusCreated, map[string]any{"orderId": order.ID, "razorpayOrderId": order.RazorpayOrderID, "razorpayKeyId": order.RazorpayKeyID, "amountPaise": order.AmountPaise, "currency": order.Currency})
	}
}
