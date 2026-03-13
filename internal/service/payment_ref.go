package service

import "crypto/rand"

// generatePaymentRef génère un code de virement unique de 8 caractères alphanumériques (A-Z0-9).
func generatePaymentRef() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	for i, c := range b {
		b[i] = chars[int(c)%len(chars)]
	}
	return string(b)
}
