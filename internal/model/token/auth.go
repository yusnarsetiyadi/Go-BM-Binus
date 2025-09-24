package token

import (
	"bm_binus/internal/config"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/golang-jwt/jwt/v4"
)

type AuthToken struct {
	token *jwt.Token
}

type AuthEksternalToken struct {
	UserId int `json:"user_id"`
}

func NewAuthToken(claims *TokenClaims) *AuthToken {
	return &AuthToken{token: jwt.NewWithClaims(jwt.SigningMethodHS256, claims)}
}

func (t *AuthToken) Token() (string, error) {
	signedString, err := t.token.SignedString([]byte(config.Get().JWT.SecretKey))
	if err != nil {
		return "", err
	}
	return signedString, nil
}

func (data *AuthEksternalToken) GenerateTokenEksternal() (*string, error) {
	sha1 := sha1.New()
	io.WriteString(sha1, config.Get().JWT.SecretKeyEksternal)

	salt := string(sha1.Sum(nil))[0:16]
	block, err := aes.NewCipher([]byte(salt))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	out, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	seal := gcm.Seal(nonce, nonce, out, nil)
	token := base64.URLEncoding.EncodeToString(seal)

	return &token, nil
}

func ValidateTokenEksternal(token string) (data AuthEksternalToken, err error) {
	sha1 := sha1.New()
	io.WriteString(sha1, config.Get().JWT.SecretKeyEksternal)

	salt := string(sha1.Sum(nil))[0:16]
	block, err := aes.NewCipher([]byte(salt))
	if err != nil {
		return data, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return data, err
	}

	decode, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return data, err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := decode[:nonceSize], decode[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(plain, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}
