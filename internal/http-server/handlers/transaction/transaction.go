package transaction

import (
	"errors"
	"log/slog"
	"net/http"
	resp "wallet/internal/lib/api/response"
	"wallet/internal/lib/api/sender"
	"wallet/internal/lib/logger/sl"
	"wallet/storage"
	"wallet/storage/postgresql"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Operation interface {
	DepositWallet(walletID uuid.UUID, amount int64) (postgresql.Wallet, error)
	WithdrawWallet(walletID uuid.UUID, amount int64) (postgresql.Wallet, error)
}

type Request struct {
	WalletID     uuid.UUID `json:"valletId" validate:"required"`
	Operation    string    `json:"operationType"` // можно было и так  validate:"required,oneof=DEPOSIT WITHDRAW", но я сделал слегка по другому))
	Amount       int64     `json:"amount" validate:"required,min=1"`
}

type Response struct {
	resp.Response
	WalletID uuid.UUID `json:"walletId"`
	Balance  int64     `json:"balance"`
}

func WalletOperation(log *slog.Logger, operation Operation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
			const fn = "handlers.transaction.WalletOperation"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			sender.SendError(w, r, log, http.StatusBadRequest, "failed to decode request body", err)
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("invalid request", sl.Err(err))
			validatorErr := err.(validator.ValidationErrors)
			render.JSON(w,r, resp.ValidationError(validatorErr))
			return
		}


		switch req.Operation {
		case "DEPOSIT":
			res, err := operation.DepositWallet(req.WalletID, req.Amount)
			if err != nil {
				if errors.Is(err, storage.ErrWalletNotFound) {
					sender.SendError(w, r, log, http.StatusNotFound, "walletId not found", err)
					return
				}
				sender.SendError(w, r, log, http.StatusBadRequest, "operation failed", err)				
				return
			}
			log.Info("wallet found, operation - WITHDRAW", slog.String("walletID", req.WalletID.String()))
			render.JSON(w, r, Response{
				Response: resp.OK(),
				WalletID: res.WalletID,
				Balance: int64(res.Balance),
			})
			return
		case "WITHDRAW":
			res, err := operation.WithdrawWallet(req.WalletID, req.Amount)
			if err != nil {
				if errors.Is(err, storage.ErrWalletNotFound) {
					sender.SendError(w, r, log, http.StatusNotFound, "walletId not found", err)
					return
				} else if errors.Is(err, storage.ErrInsufficientFunds) {
					sender.SendError(w, r, log, http.StatusNotFound, "insufficient funds", err)
					return
				}
				sender.SendError(w, r, log, http.StatusBadRequest, "operation failed", err)				
				return
			}
			log.Info("wallet found, operation - WITHDRAW", slog.String("walletID", req.WalletID.String()))
			render.JSON(w, r, Response{
				Response: resp.OK(),
				WalletID: res.WalletID,
				Balance: int64(res.Balance),
			})
		default:
			sender.SendError(w, r, log, http.StatusBadRequest, "unsupported operation", errors.New("unsupported operation"))
			return
		}
	}
}

