package sender

import (
	"log/slog"
	"net/http"
	"wallet/internal/lib/logger/sl"

	"github.com/go-chi/render"
    resp "wallet/internal/lib/api/response"
)

func SendError(w http.ResponseWriter, r *http.Request, log *slog.Logger, statusCode int, message string, err error) {
	log.Error(message, sl.Err(err))
	w.WriteHeader(statusCode)
	render.JSON(w, r, resp.Error(message))
}
