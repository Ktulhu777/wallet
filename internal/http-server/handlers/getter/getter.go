package getter

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	resp "wallet/internal/lib/api/response"
	"wallet/internal/lib/logger/sl"
	"wallet/storage"
	"wallet/storage/postgresql"
)

type GetterWallet interface {
	GetWallet(wallet_uuid uuid.UUID) (postgresql.Wallet, error)
}


type Request struct {
	WalletID uuid.UUID `json:"wallet_id" validate:"required"`
}


type Response struct {
	resp.Response
	WalletID uuid.UUID  `json:"wallet_id"`
	Balance  int 		`json:"balance"`
}

func FetchWallet(log *slog.Logger, getterWallet GetterWallet) http.HandlerFunc {
	if log == nil {
        log = slog.Default() // Используем дефолтный логгер, если передан nil
    }
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.getter.FetchWallet"

		log := log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)	

		walletUUIDStr := chi.URLParam(r, "WALLET_UUID")

		if walletUUIDStr == "" {
			log.Error("wallet_uuid is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		walletUUID, err := uuid.Parse(walletUUIDStr)
		if err != nil {
			log.Error("wallet_uuid is invalid", slog.String("wallet_uuid", walletUUIDStr))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid UUID"))
			return
		}

		resWallet, err := getterWallet.GetWallet(walletUUID)

		if errors.Is(err, storage.ErrWalletNotFound) {
			log.Error("does not exist", sl.Err(err), slog.String("wallet_uuid", walletUUIDStr))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, resp.Error("wallet does not exist"))
			return
		}

		if err != nil {
			log.Error("failed to fetch wallet", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, resp.Error("failed to fetch wallet"))
			return
		}

		render.JSON(w, r, Response{
			Response: resp.OK(),
			WalletID: resWallet.WalletID,
			Balance: resWallet.Balance,
		})
	}
}