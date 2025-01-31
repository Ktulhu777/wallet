CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE wallets (
    wallet_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    balance NUMERIC DEFAULT 0
);

CREATE TABLE transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID REFERENCES wallets(wallet_id),  -- Ссылка на кошелек
    operation_type VARCHAR(10) CHECK (operation_type IN ('DEPOSIT', 'WITHDRAW')),  -- Тип операции
    amount NUMERIC,  -- Сумма операции
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
