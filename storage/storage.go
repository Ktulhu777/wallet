package storage

import "errors"


var (
	ErrOpenDBConnection = errors.New("failed to open database connection")
	ErrPingDB           = errors.New("failed to ping database")
)


var (
	ErrWalletNotFound = errors.New("wallet not found")
)