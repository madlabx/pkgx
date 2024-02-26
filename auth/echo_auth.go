package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/errors"
)

type EchoAuth struct {
	Jwt         *JwtAuth
	StoreClaims []string
	TokenQuery  string //If set, it will try to get the query token from this query param
}

func (e *EchoAuth) Auth(c echo.Context) error {
	val := c.Request().Header.Get("Authorization")
	segs := strings.Split(val, "Bearer")
	var ts string
	if len(segs) < 2 {
		if len(e.TokenQuery) == 0 {
			return c.String(http.StatusUnauthorized, "The Bearer Authorization or Query Token param must be set and valid")
		}
		ts = c.QueryParam(e.TokenQuery)
	} else {
		ts = strings.Trim(segs[1], " \t")
	}
	if len(ts) == 0 {
		err := errors.New("The JWT token is empty")
		return c.String(http.StatusUnauthorized, err.Error())
	}
	token, err := e.Jwt.Verify(ts)
	if err != nil {
		err = errors.New("JWT token failed to be verified: " + err.Error())
		return c.String(http.StatusUnauthorized, err.Error())
	}
	c.Set("jwt", token)
	for _, sc := range e.StoreClaims {
		c.Set(sc, token.Claims.GetString(sc))
	}
	return nil
}

func (e *EchoAuth) AuthHandler(handle func(c echo.Context) error) func(c echo.Context) error {

	return func(c echo.Context) error {
		err := e.Auth(c)
		if err != nil {
			return err
		}
		return handle(c)
	}
}
