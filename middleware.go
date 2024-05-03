package main

import (
	"context"
	"log"
	"net/http"

	ory "github.com/ory/client-go"
)

type contextKey string

const (
	reqCookiesKey contextKey = "req.cookies"
	reqSessionKey contextKey = "req.session"
)

// save the cookies for any upstream calls to the Ory apis
func withCookies(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, reqCookiesKey, v)
}

func getCookies(ctx context.Context) string {
	return ctx.Value(reqCookiesKey).(string)
}

// save the session to display it on the dashboard
func withSession(ctx context.Context, v *ory.Session) context.Context {
	return context.WithValue(ctx, reqSessionKey, v)
}

func getSession(ctx context.Context) *ory.Session {
	return ctx.Value(reqSessionKey).(*ory.Session)
}

func (app *App) sessionMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		log.Printf("handling middleware request\n")

		// set the cookies on the ory client
		// merge variable declaration with assignment on next line
		cookies := request.Header.Get("Cookie")

		// check if we have a session
		session, _, err := app.ory.FrontendAPI.ToSession(request.Context()).Cookie(cookies).Execute()
		if (err != nil && session == nil) || (err == nil && !*session.Active) {
			// this will redirect the user to the managed Ory Login UI
			http.Redirect(writer, request, "/.ory/self-service/login/browser", http.StatusSeeOther)
			return
		}

		ctx := withCookies(request.Context(), cookies)
		ctx = withSession(ctx, session)

		// continue to the requested page (in our case the Dashboard)
		next.ServeHTTP(writer, request.WithContext(ctx))
	}
}
