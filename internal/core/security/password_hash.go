package security

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// ✅ MUHIIM: Furehan ayaa ah kan nidaamka oo dhan laga isticmaalayo.
// Iska hubi in kan iyo kan .env ku jira ay isku mid noqdaan haddii aad .env isticmaalayso.
var secretKey = "superaib_infrastructure_secret_key_2026"

// --- Password Hashing Functions ---

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// --- JWT Generation and Validation Functions ---

// GenerateToken: Wuxuu dhalinayaa Token-ka
func GenerateToken(userID string, role string) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 7 maalmood

	claims := jwt.MapClaims{
		"authorized": true,
		"user_id":    userID,
		"role":       role,
		"exp":        expirationTime.Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("could not sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT: ✅ KANI WAA MUHIIM!
// Middleware-ku hadda kan ayuu u yeeri doonaa si looga fogaado signature error.
func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
