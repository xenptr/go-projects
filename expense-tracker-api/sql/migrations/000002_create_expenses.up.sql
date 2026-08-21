CREATE TYPE expense_category AS ENUM (
    'Groceries',
    'Leisure',
    'Electronics',
    'Utilities',
    'Clothing',
    'Health',
    'Others'
);

CREATE TABLE expenses (
    id          BIGSERIAL        PRIMARY KEY,
    user_id     BIGINT           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT             NOT NULL,
    amount      NUMERIC(12, 2)   NOT NULL CHECK (amount > 0),
    category    expense_category NOT NULL DEFAULT 'Others',
    date        DATE             NOT NULL DEFAULT CURRENT_DATE,
    description TEXT             NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_expenses_user_id ON expenses(user_id);
CREATE INDEX idx_expenses_date    ON expenses(date);
