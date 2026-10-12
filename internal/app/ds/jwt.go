package ds

import "github.com/golang-jwt/jwt"

type JWTClaims struct {
	jwt.StandardClaims
	EngineerID uint `json:"engineer_id"`
}
