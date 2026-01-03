package memory

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"
	"sync"
	"time"
)

type OTPRecord struct {
	CodeHash     string
	ExpiresAt    time.Time
	AttemptsLeft int
}

type OTPStore struct {
	mu sync.Mutex
	m  map[string]OTPRecord // key=email(lower)
}

func NewOTPStore() *OTPStore {
	s := &OTPStore{m: make(map[string]OTPRecord)}

	// GC раз в минуту
	go func() {
		t := time.NewTicker(1 * time.Minute)
		defer t.Stop()
		for range t.C {
			s.cleanup()
		}
	}()

	return s
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *OTPStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for k, v := range s.m {
		if now.After(v.ExpiresAt) {
			delete(s.m, k)
		}
	}
}

func hashCode(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}

func genDigits(n int) string {
	const digits = "0123456789"
	out := make([]byte, n)

	for i := 0; i < n; i++ {
		x, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		out[i] = digits[x.Int64()]
	}
	return string(out)
}

func (s *OTPStore) Create(email string, codeLen int, ttl time.Duration, maxAttempts int) (code string) {
	email = normalizeEmail(email)
	code = genDigits(codeLen)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.m[email] = OTPRecord{
		CodeHash:     hashCode(code),
		ExpiresAt:    time.Now().Add(ttl),
		AttemptsLeft: maxAttempts,
	}
	return code
}

func (s *OTPStore) Verify(email string, code string) (ok bool) {
	email = normalizeEmail(email)
	code = strings.TrimSpace(code)

	s.mu.Lock()
	defer s.mu.Unlock()

	rec, exists := s.m[email]
	if !exists {
		return false
	}
	if time.Now().After(rec.ExpiresAt) || rec.AttemptsLeft <= 0 {
		delete(s.m, email)
		return false
	}

	if rec.CodeHash != hashCode(code) {
		rec.AttemptsLeft--
		s.m[email] = rec
		return false
	}

	delete(s.m, email)
	return true
}
