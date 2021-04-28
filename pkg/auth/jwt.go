package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"math/big"
	"strings"
)

type OIDClient interface {
	GetJWK(kid string) (*JWK, error)
}

type jwtHandler struct {
	oidClient                   OIDClient
	authTokenValidationIssuer   string
	authTokenValidationAudience string
}

func NewJwtHandler(oidClient OIDClient, authTokenValidationIssuer string, authTokenValidationAudience string) *jwtHandler {
	jwtHandler := &jwtHandler{
		oidClient:                   oidClient,
		authTokenValidationIssuer:   authTokenValidationIssuer,
		authTokenValidationAudience: authTokenValidationAudience,
	}

	return jwtHandler
}

func (jh *jwtHandler) ValidateJWTToken(tokenString string) error {
	// Strip "Bearer " from token string
	authToken := strings.Replace(tokenString, "Bearer ", "", 1)

	// Parse authToken
	parsedToken, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		// Check for RSA signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			log.Printf("===== Unexpected signing method: %v =====", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Lookup and return signing key
		key, err := jh.oidClient.GetJWK(token.Header["kid"].(string))
		if err != nil {
			return nil, err
		}
		return decodePublicKey(key)
	})

	if err != nil {
		return fmt.Errorf("unable to validate JWT Token: %v", err)
	}

	return jh.validateClaims(parsedToken)
}

func (jh *jwtHandler) validateClaims(parsedToken *jwt.Token) error {
	claims := parsedToken.Claims.(jwt.MapClaims)
	log.Debugln("Claims: ", claims)
	validIssuer := jh.authTokenValidationIssuer
	log.Debugln("Issuer Check ", validIssuer, claims["iss"])
	if claims["iss"] != validIssuer {
		return fmt.Errorf("unauthorized issuer")
	}
	validAudience := jh.authTokenValidationAudience
	claimAudienceString := fmt.Sprintf("%v", claims["aud"])
	log.Debugln("Audience Check ", validAudience, claimAudienceString, !strings.Contains(claimAudienceString, validAudience))
	if !strings.Contains(claimAudienceString, validAudience) {
		return fmt.Errorf("unauthorized audience")
	}
	return nil
}

func decodePublicKey(jwk *JWK) (*rsa.PublicKey, error) {
	// decode exponent
	decodedE, err := safeDecode(jwk.E)
	if err != nil {
		return nil, errors.New("malformed JWK RSA key")
	}
	if len(decodedE) < 4 {
		ndata := make([]byte, 4)
		copy(ndata[4-len(decodedE):], decodedE)
		decodedE = ndata
	}

	pubKey := &rsa.PublicKey{
		N: &big.Int{},
		E: int(binary.BigEndian.Uint32(decodedE[:])),
	}

	decodedN, err := safeDecode(jwk.N)
	if err != nil {
		return nil, errors.New("malformed JWK RSA key")
	}
	pubKey.N.SetBytes(decodedN)

	return pubKey, nil
}

func safeDecode(str string) ([]byte, error) {
	lenMod4 := len(str) % 4
	if lenMod4 > 0 {
		str = str + strings.Repeat("=", 4-lenMod4)
	}

	return base64.URLEncoding.DecodeString(str)
}
