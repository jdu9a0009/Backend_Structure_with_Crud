package middleware

import (
	"context"
	"errors"
	"net/http"
	"project/foundation/web"
	"project/internal/auth"
	"strings"
)

func Authenticate(a *auth.Auth, role ...string) web.Middleware {
	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(c *web.Context) error {

			// Expecting: Bearer <token>
			authStr := c.Request.Header.Get("authorization")

			// Parse the authorization header.
			parts := strings.Split(authStr, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.New("expected authorization header format: Bearer <token>")
				return c.RespondError(web.NewRequestError(err, http.StatusUnauthorized))
			}

			// Validate the token is signed by us.
			claims, err := a.ValidateToken(parts[1])
			if err != nil {
				return c.RespondError(web.NewRequestError(err, http.StatusUnauthorized))
			}

			//check role inside token data
			if ok := claims.Authorized(role...); !ok && (len(role) > 0) {
				return c.RespondError(web.NewRequestError(errors.New("attempted action is not allowed"), http.StatusUnauthorized))
			}

			// check if claims from database
			//if err = a.CheckClaimsDataFromDatabase(c.Ctx, claims); err != nil {
			//	return c.RespondError(err)
			//}

			// Add claims to the context so that they can be retrieved later.
			c.Ctx = context.WithValue(c.Ctx, auth.Key, claims)

			// Call the next handler.
			return handler(c)
		}

		return h
	}

	return m
}
