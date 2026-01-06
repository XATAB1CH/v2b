package services

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/XATAB1CH/v2b/internal/clients/mailer"
	"github.com/XATAB1CH/v2b/internal/store/memory"
	"github.com/XATAB1CH/v2b/internal/store/postgres"
)

type AuthService struct {
	otp *memory.OTPStore

	users *postgres.UsersRepo
	ent   *postgres.EntitlementsRepo

	jwtSecret []byte
	accessTTL time.Duration

	otpTTL         time.Duration
	otpLen         int
	otpMaxAttempts int

	mailer mailer.Mailer
	env    string

	freeLimit int
}

func NewAuthService(
	otp *memory.OTPStore,
	users *postgres.UsersRepo,
	ent *postgres.EntitlementsRepo,
	jwtSecret string,
	accessTTLMins int,
	otpTTLMins int,
	otpLen int,
	otpMaxAttempts int,
	freeLimit int,
	m mailer.Mailer,
	env string,
) *AuthService {
	return &AuthService{
		otp:            otp,
		users:          users,
		ent:            ent,
		jwtSecret:      []byte(jwtSecret),
		accessTTL:      time.Duration(accessTTLMins) * time.Minute,
		otpTTL:         time.Duration(otpTTLMins) * time.Minute,
		otpLen:         otpLen,
		otpMaxAttempts: otpMaxAttempts,
		freeLimit:      freeLimit,
		mailer:         m,
		env:            env,
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *AuthService) RequestOTP(email string) (code string, err error) {
	email = normalizeEmail(email)
	if email == "" || !strings.Contains(email, "@") {
		return "", errors.New("invalid email")
	}

	code = s.otp.Create(email, s.otpLen, s.otpTTL, s.otpMaxAttempts)

	// local fallback: если SMTP не настроен — логируем
	if s.mailer == nil {
		if s.env == "local" {
			log.Printf("[DEV OTP] email=%s code=%s", email, code)
			return code, nil
		}
		return "", errors.New("mailer is not configured")
	}

	if err := s.mailer.SendOTP(context.Background(), email, code); err != nil {
		return "", err
	}

	// В проде код не возвращаем (handler и так отвечает {ok:true})
	return "", nil
}

type Tokens struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (s *AuthService) VerifyOTP(ctx context.Context, email string, code string) (Tokens, error) {
	email = normalizeEmail(email)
	if email == "" {
		return Tokens{}, errors.New("invalid email")
	}

	if ok := s.otp.Verify(email, code); !ok {
		return Tokens{}, errors.New("invalid code")
	}

	u, err := s.users.GetOrCreateByEmail(ctx, email)
	if err != nil {
		return Tokens{}, err
	}
	_ = s.users.TouchLogin(ctx, u.ID)
	_ = s.ent.Ensure(ctx, u.ID, s.freeLimit)

	now := time.Now().UTC()
	exp := now.Add(s.accessTTL)

	claims := jwt.MapClaims{
		"sub":   u.ID,
		"email": u.Email,
		"iat":   now.Unix(),
		"exp":   exp.Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := t.SignedString(s.jwtSecret)
	if err != nil {
		return Tokens{}, err
	}

	return Tokens{
		AccessToken: tokenStr,
		ExpiresIn:   int(s.accessTTL.Seconds()),
	}, nil
}
