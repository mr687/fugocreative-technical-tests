package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/httplog/v3"
)

func Logger(skipPaths ...string) func(next http.Handler) http.Handler {
	isProduction := false
	logFormat := httplog.SchemaECS.Concise(!isProduction)
	logger := slog.Default()
	return func(next http.Handler) http.Handler {
		return httplog.RequestLogger(logger, &httplog.Options{
			Level:         slog.LevelInfo,
			Schema:        logFormat,
			RecoverPanics: true,
			Skip: func(req *http.Request, respStatus int) bool {
				for _, path := range skipPaths {
					if strings.HasPrefix(req.URL.Path, path) {
						return true
					}
				}
				return respStatus == 405
			},
			LogRequestHeaders:  []string{"Origin"},
			LogResponseHeaders: []string{},
		})(next)
	}
}
