package loom

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type FlashKey struct{}

const flashCookieName = "loom_flash_message"

func FlashMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		msg := get(c)
		if msg != nil {
			ctx = context.WithValue(ctx, FlashKey{}, msg)
			clear(c)
		}

		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}

func newMessage(msg string, t string, kv ...string) *FlashMessage {
	m := FlashMessage{
		Message: msg,
		Type:    t,
		Params:  make(map[string]string),
	}

	for i := 0; i < len(kv); i += 2 {
		m.Params[kv[i]] = kv[i+1]
	}

	return &m
}

type FlashMessage struct {
	Message string
	Type    string
	Params  map[string]string
}

func (m FlashMessage) Encode() string {
	data, _ := json.Marshal(m)

	return encode(data)
}

func FlashSuccessNow(c echo.Context, msg string, kv ...string) {
	setNow(c, newMessage(msg, "success", kv...))
}

func FlashWarningNow(c echo.Context, msg string, kv ...string) {
	setNow(c, newMessage(msg, "warning", kv...))
}

func FlashErrorNow(c echo.Context, msg string, kv ...string) {
	setNow(c, newMessage(msg, "error", kv...))
}

func setNow(c echo.Context, msg *FlashMessage) {
	ctx := context.WithValue(c.Request().Context(), FlashKey{}, msg)
	c.SetRequest(c.Request().WithContext(ctx))
}

func FlashSuccess(c echo.Context, msg string, kv ...string) {
	set(c, newMessage(msg, "success", kv...))
}

func FlashWarning(c echo.Context, msg string, kv ...string) {
	set(c, newMessage(msg, "warning", kv...))
}

func FlashError(c echo.Context, msg string, kv ...string) {
	set(c, newMessage(msg, "error", kv...))
}

func set(c echo.Context, msg *FlashMessage) {
	c.SetCookie(&http.Cookie{
		Name:     flashCookieName,
		Value:    msg.Encode(),
		Path:     "/",
		HttpOnly: true,
	})
}

func get(c echo.Context) *FlashMessage {
	cookie, err := c.Cookie(flashCookieName)
	if err != nil {
		return nil
	}

	if cookie == nil {
		return nil
	}

	data, err := decode(cookie.Value)
	if err != nil {
		return nil
	}

	var m FlashMessage

	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil
	}

	return &m
}

func clear(c echo.Context) {
	c.SetCookie(
		&http.Cookie{
			Name:     flashCookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
		},
	)
}

func encode(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

func decode(src string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(src)
}
