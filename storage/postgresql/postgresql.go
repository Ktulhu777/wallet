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

	var wallet Wallet

	err = stmt.QueryRow(wallet_uuid).Scan(&wallet.WalletID, &wallet.Balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Wallet{}, storage.ErrWalletNotFound
		}
		return Wallet{}, fmt.Errorf("%s: execute statement: %w", fn, err)
	}

	return wallet, nil
}


func (sp *StoragePostgresql) DepositWallet(walletID uuid.UUID, amount int64) (Wallet, error) {
	const fn = "storage.postgresql.DepositWallet"

	tx, err := sp.db.Begin()
	if err != nil {
		return Wallet{}, fmt.Errorf("%s: failed to start transaction: %w", fn, err)
	}

	stmt, err := tx.Prepare(`
		UPDATE wallets 
		SET balance = balance + $1 
		WHERE wallet_id = $2
		RETURNING wallet_id, balance;
	`)
	if err != nil {
		tx.Rollback()
		return Wallet{}, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	var wallet Wallet
	err = stmt.QueryRow(amount, walletID).Scan(&wallet.WalletID, &wallet.Balance)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return Wallet{}, fmt.Errorf("%s: wallet not found: %w", fn, storage.ErrWalletNotFound)
		}
		return Wallet{}, fmt.Errorf("%s: failed to execute statement: %w", fn, err)
	}

	if err := tx.Commit(); err != nil {
		return Wallet{}, fmt.Errorf("%s: failed to commit transaction: %w", fn, err)
	}

	return wallet, nil
}



func (sp *StoragePostgresql) WithdrawWallet(walletID uuid.UUID, amount int64) (Wallet, error) {
	const fn = "storage.postgresql.WithdrawWallet"

	tx, err := sp.db.Begin()
	if err != nil {
		return Wallet{}, fmt.Errorf("%s: failed to start transaction: %w", fn, err)
	}

	stmt, err := tx.Prepare(`
		UPDATE wallets 
		SET balance = balance - $1 
		WHERE wallet_id = $2 AND balance >= $3
		RETURNING wallet_id, balance;
	`)
	if err != nil {
		tx.Rollback()
		return Wallet{}, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	ok, err := sp.IsExistsWallet(walletID)
	if err != nil {
		return Wallet{}, err
	}
	
	if !ok {
		return Wallet{}, storage.ErrWalletNotFound
	}

	var wallet Wallet
	err = stmt.QueryRow(amount, walletID, amount).Scan(&wallet.WalletID, &wallet.Balance)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return Wallet{}, fmt.Errorf("%s: insufficient funds: %w", fn, storage.ErrInsufficientFunds)
		}
		return Wallet{}, fmt.Errorf("%s: failed to execute statement: %w", fn, err)
	}

	if err := tx.Commit(); err != nil {
		return Wallet{}, fmt.Errorf("%s: failed to commit transaction: %w", fn, err)
	}

	return wallet, nil
}


func (sp *StoragePostgresql) IsExistsWallet(walletID uuid.UUID) (bool, error) {
	const fn = "storage.postgresql.IsExistsWallet"

	stmt, err := sp.db.Prepare("SELECT EXISTS(SELECT 1 FROM wallets WHERE wallet_id = $1)")
	if err != nil {
		return false, fmt.Errorf("%s: prepare statement: %w", fn, err)
	}
	var exists bool
	
	err = stmt.QueryRow(walletID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: failed to check wallet existence: %w", fn, storage.ErrWalletNotFound)
	}

	return exists, nil
}
