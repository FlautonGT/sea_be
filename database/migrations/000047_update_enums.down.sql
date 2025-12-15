-- Rename new types back (temporary)
ALTER TYPE public.transaction_status RENAME TO transaction_status_new;
ALTER TYPE public.payment_status RENAME TO payment_status_new;

-- Recreate old types
CREATE TYPE public.transaction_status AS ENUM (
    'PENDING',
    'PAID',
    'PROCESSING',
    'SUCCESS',
    'FAILED',
    'EXPIRED',
    'REFUNDED'
);

CREATE TYPE public.payment_status AS ENUM (
    'UNPAID',
    'PAID',
    'EXPIRED',
    'REFUNDED'
);

-- Revert columns
ALTER TABLE public.transactions 
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN status TYPE public.transaction_status 
  USING status::text::public.transaction_status,
  ALTER COLUMN status SET DEFAULT 'PENDING'::public.transaction_status;

ALTER TABLE public.transactions
  ALTER COLUMN payment_status DROP DEFAULT,
  ALTER COLUMN payment_status TYPE public.payment_status
  USING (
    CASE payment_status::text
      WHEN 'FAILED' THEN 'UNPAID'::public.payment_status -- Best effort fallback
      ELSE payment_status::text::public.payment_status
    END
  ),
  ALTER COLUMN payment_status SET DEFAULT 'UNPAID'::public.payment_status;

-- Drop new types
DROP TYPE public.transaction_status_new;
DROP TYPE public.payment_status_new;
