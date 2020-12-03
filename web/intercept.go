package web

import (
	"context"
	"fmt"
	"github.com/gabrielmorenobrc/go-lib/auth"
	"github.com/gabrielmorenobrc/go-lib/sql"
	"net/http"
	"strings"
)

func HandleDefault(path string, f func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(path, InterceptFatal(InterceptCORS(f)))
}

func HandleTransactional(path string, databaseConfig sql.DatabaseConfig, f func(txContext *sql.TxCtx, w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(path, InterceptFatal(InterceptCORS(sql.InterceptTransactional(&databaseConfig, InterceptAuditable(f)))))
}

func HandleAuthenticatedTransactional(path string, sessionManager *auth.SessionManager, databaseConfig sql.DatabaseConfig, f func(txContext *sql.TxCtx, w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(path, InterceptFatal(InterceptCORS(InterceptAuth(sessionManager, sql.InterceptTransactional(&databaseConfig, InterceptAuditable(f))))))
}

func InterceptAuth(sessionManager *auth.SessionManager, delegate func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenEntry := validateToken(r, sessionManager)
		if tokenEntry == nil {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			ctx := context.WithValue(r.Context(), "sessionEntry", tokenEntry)
			delegate(w, r.WithContext(ctx))
		}
	}
}

func validateToken(r *http.Request, sessionManager *auth.SessionManager) *auth.SessionEntry {
	token, ok := resolveToken(r)
	if ok {
		return sessionManager.ValidateToken(token)
	} else {
		return nil
	}
}

func resolveToken(r *http.Request) (string, bool) {
	c, err := r.Cookie("authToken")
	if err == nil {
		return c.Value, true
	}
	value := r.Header.Get("authToken")
	if len(value) > 0 {
		return value, true
	}
	value = r.Header.Get("Authorization")
	if len(value) > 0 {
		parts := strings.Split(value, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1], true
		}
	}
	values, ok := r.URL.Query()["authToken"]
	if ok {
		return values[0], true
	}
	return "", false
}

func InterceptBasicAuth(delegate func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inUsername, inPw, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			ctx := context.WithValue(r.Context(), "username", inUsername)
			ctx = context.WithValue(ctx, "password", inPw)
			delegate(w, r.WithContext(ctx))
		}
	}
}

func InterceptAuditable(delegate func(txCtx *sql.TxCtx, w http.ResponseWriter, r *http.Request)) func(txCtx *sql.TxCtx, w http.ResponseWriter, r *http.Request) {
	return func(txCtx *sql.TxCtx, w http.ResponseWriter, r *http.Request) {
		sessionEntry := r.Context().Value("sessionEntry")
		if sessionEntry != nil {
			txCtx.ExecSql(fmt.Sprintf("set local fanaticus.user_name to '%d';", sessionEntry.(*auth.SessionEntry).UserId))
		}
		txCtx.ExecSql(fmt.Sprintf("set local fanaticus.context to '%s';", r.URL.Path))
		delegate(txCtx, w, r)
	}
}
