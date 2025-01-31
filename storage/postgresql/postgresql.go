package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"wallet/storage"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type StoragePostgresql struct {
	db *sql.DB
}

type Wallet struct {
	WalletID uuid.UUID
	Balance int

}

func NewStorage(dbURL string) (*StoragePostgresql, error) {
	const fn = "storage.postgresql.NewStorage"

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %s: %v", fn, storage.ErrOpenDBConnection, err)
	}


	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %s: %v", fn, storage.ErrPingDB, err)
	}


	return &StoragePostgresql{db: db}, nil
}	

func (sp *StoragePostgresql) GetWallet(wallet_uuid uuid.UUID) (Wallet, error) {
	const fn = "storage.postgresql.GetWallet"

	stmt, err := sp.db.Prepare("SELECT wallet_id, balance FROM wallets WHERE wallet_id = $1")
	if err != nil {
		return Wallet{}, fmt.Errorf("%s: prepare statement: %w", fn, err)
	}

	var wallet_id uuid.UUID
	var balance int

	err = stmt.QueryRow(wallet_uuid).Scan(&wallet_id, &balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Wallet{}, storage.ErrWalletNotFound
		}
		return Wallet{}, fmt.Errorf("%s: execute statement: %w", fn, err)
	}

	return Wallet{
		WalletID: wallet_id,
		Balance: balance,
	}, nil
}