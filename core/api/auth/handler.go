package auth

import (
	"errors"
	"github.com/redesblock/mop/core/api/jsonhttp"
	urlSigner "github.com/redesblock/mop/core/util/urlsigner"
	"net/http"
	"strings"
)

type auth interface {
	Enforce(string, string, string) (bool, error)
	SecretKey(string) (string, error)
}

func PermissionCheckHandler(auth auth) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqToken := r.Header.Get("Authorization")
			if !strings.HasPrefix(reqToken, "Bearer ") {
				jsonhttp.Forbidden(w, "Missing bearer token")
				return
			}

			keys := strings.Split(reqToken, "Bearer ")

			if len(keys) != 2 || strings.Trim(keys[1], " ") == "" {
				jsonhttp.Unauthorized(w, "Missing security token")
				return
			}

			apiKey := keys[1]

			allowed, err := auth.Enforce(apiKey, r.URL.Path, r.Method)
			if errors.Is(err, ErrTokenExpired) {
				jsonhttp.Unauthorized(w, "Token expired")
				return
			}

			if err != nil {
				jsonhttp.InternalServerError(w, "Error occurred while validating the security token")
				return
			}

			if !allowed {
				jsonhttp.Forbidden(w, "Provided security token does not grant access to the resource")
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

func URLSignCheckHandler(auth auth) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqToken := r.Header.Get("Authorization")
			if !strings.HasPrefix(reqToken, "Bearer ") {
				jsonhttp.Forbidden(w, "Missing bearer token")
				return
			}

			keys := strings.Split(reqToken, "Bearer ")

			if len(keys) != 2 || strings.Trim(keys[1], " ") == "" {
				jsonhttp.Unauthorized(w, "Missing security token")
				return
			}

			apiKey := keys[1]

			secretKey, err := auth.SecretKey(apiKey)
			if err != nil {
				jsonhttp.InternalServerError(w, "Error occurred while get secret from the security token")
				return
			}

			singer := urlSigner.New(secretKey)
			if !singer.VerifyTemporary(*r.URL) {
				jsonhttp.Forbidden(w, "Provided sign does not grant access to the resource")
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
