package mailer

import "context"

type Mailer interface {
	SendOTP(ctx context.Context, toEmail string, code string) error
}
