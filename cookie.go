package goblet

import (
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
)

// CookieIsMissing is returned when a cookie is missing.
var CookieIsMissing error = errors.New("Cookie is missing")

// SignedCookieIsMissing is returned when the signed cookie is missing.
var SignedCookieIsMissing error = errors.New("Signed cookie is missing")

// CookieNotValid is returned when the cookie and its signed counterpart do not match.
//
// I.e. the cookie value has been tampered with.
var CookieNotValid error = errors.New("Cookie and signed cookie do not match")

var (
	// SignedCookieFormat is the format string used to decide the name of the
	// signed cookie.
	SignedCookieFormat string = "%s_signed"
)

// toSignedCookieName gets the signed cookie name from the specified cookie name,
// by running it through the SignedCookieFormat string.
func toSignedCookieName(name string) string {
	return fmt.Sprintf(SignedCookieFormat, name)
}

// AddSignedCookie adds the specified cookie to the response and also adds an
// additional 'signed' cookie that is used to validate the cookies value when
// SignedCookie is called.
func (c *Context) AddSignedCookie(cookie *fasthttp.Cookie) (*fasthttp.Cookie, error) {

	// make the signed cookie
	signedCookie := new(fasthttp.Cookie)

	// copy the cookie settings
	signedCookie.SetPathBytes(cookie.Path())
	signedCookie.SetDomainBytes(cookie.Domain())
	signedCookie.SetExpire(cookie.Expire())
	// signedCookie.MaxAge = cookie.MaxAge
	// signedCookie.Secure = cookie.Secure
	// signedCookie.HttpOnly = cookie.HttpOnly
	// signedCookie.Raw = cookie.Raw
	signedCookie.SetKey(toSignedCookieName(string(cookie.Key())))
	// set the signed cookie specifics
	signedCookie.SetValue(c.Server.HashBytes(cookie.Value()))

	// add the cookies
	c.ctx.Response.Header.SetCookie(cookie)
	c.ctx.Response.Header.SetCookie(signedCookie)

	// return the new signed cookie (and no error)
	return signedCookie, nil

}

func (c *Context) AddCookie(cookie *fasthttp.Cookie) error {

	// add the cookies
	c.ctx.Response.Header.SetCookie(cookie)

	// return the new signed cookie (and no error)
	return nil

}

// Gets the cookie specified by name and validates that its value has not been
// tampered with by checking the signed cookie too.  Will return CookieNotValid error
// if it has been tampered with, otherwise, it will return the actual cookie.
func (c *Context) SignedCookie(name string) (*fasthttp.Cookie, error) {

	valid, validErr := c.cookieIsValid(name)
	if valid {
		bts := c.ctx.Request.Header.Cookie(name)
		cookie := new(fasthttp.Cookie)
		cookie.ParseBytes(bts)
		return cookie, nil
	} else if validErr != nil {
		return nil, validErr
	}

	return nil, CookieNotValid
}

// cookieIsValid checks to see if the cookie and its signed counterpart
// match.
func (c *Context) cookieIsValid(name string) (bool, error) {

	// get the cookies
	cookie := new(fasthttp.Cookie)
	parseCookieErr := cookie.ParseBytes(c.ctx.Request.Header.Cookie(name))
	signedCookie := new(fasthttp.Cookie)
	parseSCookieErr := signedCookie.ParseBytes(c.ctx.Request.Header.Cookie(toSignedCookieName(name)))

	// handle errors reading cookies

	if parseCookieErr != nil {
		if parseCookieErr.Error() == "no cookies found" {
			return false, CookieIsMissing
		}
		return false, parseCookieErr
	}
	if parseSCookieErr != nil {
		if parseSCookieErr.Error() == "no cookies found" {
			return false, SignedCookieIsMissing
		}
		return false, parseSCookieErr
	}

	// check the cookies
	if c.Server.HashBytes(cookie.Value()) != string(signedCookie.Value()) {
		return false, nil
	}

	// success - therefore valid
	return true, nil

}
