package stubpay

import (
	"context"
	"fmt"
	"net/url"
)

type Client struct {
	publicBaseURL string // например http://localhost:8080
}

func New(publicBaseURL string) *Client {
	return &Client{publicBaseURL: publicBaseURL}
}

// Соответствует services.PaymentProvider
func (c *Client) CreatePayment(ctx context.Context, userID string, amount string, currency string, returnURL string, description string) (providerPaymentID string, confirmationURL string, err error) {
	// providerPaymentID не нужен для заглушки
	providerPaymentID = ""

	// confirmationURL — просто страница-заглушка (может быть статичной)
	confirmationURL = fmt.Sprintf("%s/dev/payments/confirm?return=%s",
		c.publicBaseURL,
		url.QueryEscape(returnURL),
	)
	return providerPaymentID, confirmationURL, nil
}
