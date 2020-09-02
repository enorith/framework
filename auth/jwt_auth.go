package auth

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	. "github.com/enorith/framework/contracts"
	"github.com/enorith/framework/http/content"
	"github.com/enorith/framework/http/contract"
	"github.com/enorith/supports/str"
	"strconv"
	"time"
)

var (
	DefaultJwtAlg   = jwt.SigningMethodHS512
	JwtExpireSecond = 60 * 30
	JwtKey          = []byte("somerandomstring!!!!")
)

type JwtUser interface {
	User
	GetJwtClaims() map[string]interface{}
}

type JwtAuthenticator struct {
	GenericAuthenticator
}

func (j *JwtAuthenticator) Guard(r contract.RequestContract) (User, error) {
	token, tokenErr := r.BearerToken()
	if tokenErr != nil {
		return nil, tokenErr
	}

	claims := jwt.StandardClaims{}
	t, err := jwt.ParseWithClaims(string(token), &claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})
	if !t.Valid {
		return nil, errors.New("invalid token")
	}
	if err != nil {
		return nil, err
	}
	validErr := claims.Valid()
	if validErr != nil {
		return nil, validErr
	}

	id, _ := strconv.ParseUint(claims.Subject, 10, 64)
	return j.GetUserProvider().FindUserById(id)
}

func (j *JwtAuthenticator) Check(r contract.RequestContract) bool {
	u, err := j.Guard(r)

	return u != nil && err == nil
}

func (j *JwtAuthenticator) Parse(token []byte) jwt.MapClaims {
	claims := jwt.MapClaims{}
	jwt.ParseWithClaims(string(token), &claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})

	return claims
}

func (j *JwtAuthenticator) FromUser(u User) (string, error, int64) {
	now := time.Now().Unix()
	exp := now + int64(JwtExpireSecond)
	claims := jwt.MapClaims{
		"iss": "",
	}
	if ju, ok := u.(JwtUser); ok {
		customClaims := ju.GetJwtClaims()
		for k, v := range customClaims {
			claims[k] = v
		}
	}
	claims["jti"] = fmt.Sprintf("%d", u.UserIdentifier())
	claims["sub"] = fmt.Sprintf("%d", u.UserIdentifier())
	claims["iat"] = now
	claims["exp"] = exp
	claims["aud"] = str.RandString(16)

	token := jwt.NewWithClaims(DefaultJwtAlg, claims)

	tokenString, err := token.SignedString(JwtKey)
	return tokenString, err, exp
}

func (j *JwtAuthenticator) Auth(u User) contract.ResponseContract {
	tokenString, err, exp := j.FromUser(u)

	if err != nil {
		return content.ErrResponseFromError(err, 500, nil)
	}

	return content.JsonResponse(map[string]interface{}{
		"access_token": tokenString,
		"expire_in":    exp,
		"type":         "Bearer",
	}, 200, nil)
}

func NewJwtAuthFromDefault() *JwtAuthenticator {
	return &JwtAuthenticator{
		GenericAuthenticator{
			providerName: DefaultProvider,
		},
	}
}
