package auth

import (
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/madlabx/pkgx/log"
	"github.com/madlabx/pkgx/typex"
)

type JwtConf struct {
	Secret []byte
	Method jwt.SigningMethod
}

type JwtToken struct {
	Method    jwt.SigningMethod
	Header    typex.JsonMap
	Claims    typex.JsonMap
	Signature string
}

type JwtAuth struct {
	Conf JwtConf
}

func NewJwtAuth(secret []byte) *JwtAuth {
	if len(secret) == 0 {
		log.Fatal("The JWT secret must not be empty")
	}
	return &JwtAuth{
		Conf: JwtConf{
			Secret: secret,
			Method: jwt.SigningMethodHS256,
		},
	}
}

func (a *JwtAuth) Verify(tokenString string) (*JwtToken, error) {

	token, err := jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
		return a.Conf.Secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("JWT token is not valid")
	}
	c, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Invalid JWT Claims type: %T", token.Claims))
	}
	jt := &JwtToken{
		Method:    token.Method,
		Header:    typex.JsonMap(token.Header),
		Claims:    typex.JsonMap(c),
		Signature: token.Signature,
	}
	return jt, nil
}

func (a *JwtAuth) GenToken(claims typex.JsonMap) (string, error) {

	var token *jwt.Token
	if len(claims) > 0 {
		token = jwt.NewWithClaims(a.Conf.Method, jwt.MapClaims(claims))
	} else {
		token = jwt.New(a.Conf.Method)
	}
	ts, err := token.SignedString(a.Conf.Secret)
	return ts, err
}
