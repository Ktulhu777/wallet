CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE wallets (
    wallet_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    balance NUMERIC DEFAULT 0
);
