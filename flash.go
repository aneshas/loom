package loom

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type FlashSuccessKey struct{}
type FlashWarningKey struct{}
type FlashErrorKey struct{}

const (
	flashSuccess = "hyper_flash_success"
	flashWarning = "hyper_flash_warning"
	flashErr     = "hyper_flash_error"
)

func newMessage(msg string, kv ...string) FlashMessage {
	m := FlashMessage{
		Message: msg,
		Params:  make(map[string]string),
	}

	for i := 0; i < len(kv); i += 2 {
		m.Params[kv[i]] = kv[i+1]
	}

	return m
}

type FlashMessage struct {
	Message string
	Params  map[string]string
}

func (m FlashMessage) Encode() string {
	data, _ := json.Marshal(m)

	return encode(data)
}

func FlashSuccess(c echo.Context, msg string, kv ...string) {
	set(c, flashSuccess, msg, kv...)
}

func FlashWarning(c echo.Context, msg string, kv ...string) {
	set(c, flashWarning, msg, kv...)
}

func FlashError(c echo.Context, msg string, kv ...string) {
	set(c, flashErr, msg, kv...)
}

func set(c echo.Context, name, msg string, kv ...string) {
	m := newMessage(msg, kv...)

	c.SetCookie(&http.Cookie{
		Name:  name,
		Value: m.Encode(),
		Path:  "/",
	})

	ctx := c.Request().Context()

	switch name {
	case flashSuccess:
		ctx = context.WithValue(ctx, FlashSuccessKey{}, &m)
	case flashWarning:
		ctx = context.WithValue(ctx, FlashWarningKey{}, &m)
	case flashErr:
		ctx = context.WithValue(ctx, FlashErrorKey{}, &m)
	}

	c.SetRequest(c.Request().WithContext(ctx))
}

func WithFlashMessages(ctx context.Context, c echo.Context) context.Context {
	if msg := getSuccess(c); msg != nil {
		ctx = context.WithValue(ctx, FlashSuccessKey{}, msg)
	}

	if msg := getWarning(c); msg != nil {
		ctx = context.WithValue(ctx, FlashWarningKey{}, msg)
	}

	if msg := getError(c); msg != nil {
		ctx = context.WithValue(ctx, FlashErrorKey{}, msg)
	}

	return ctx
}

func getSuccess(c echo.Context) *FlashMessage {
	return get(c, flashSuccess)
}

func getWarning(c echo.Context) *FlashMessage {
	return get(c, flashWarning)
}

func getError(c echo.Context) *FlashMessage {
	return get(c, flashErr)
}

func get(c echo.Context, name string) *FlashMessage {
	// resp := c.Response()
	// kk := resp.Header().Get("Set-Cookie")

	var cookie *http.Cookie
	var err error

	// if kk == "" {
	cookie, err = c.Cookie(name)
	if err != nil {
		return nil
	}
	// } else {
	// 	cookies, err := http.ParseCookie(kk)
	// 	if err != nil {
	// 		cookie, err = c.Cookie(name)
	// 		if err != nil {
	// 			return nil
	// 		}
	// 	}

	// 	for _, c := range cookies {
	// 		if c.Name == name {
	// 			cookie = c
	// 			resp.Header().Del("Set-Cookie")
	// 			break
	// 		}
	// 	}
	// }

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

	c.SetCookie(
		&http.Cookie{
			Name:    name,
			MaxAge:  -1,
			Expires: time.Unix(1, 0),
			Value:   "",
			Path:    "/",
		},
	)

	return &m
}

func encode(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

func decode(src string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(src)
}
