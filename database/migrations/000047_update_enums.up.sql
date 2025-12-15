-- Rename old types
ALTER TYPE public.transaction_status RENAME TO transaction_status_old;
ALTER TYPE public.payment_status RENAME TO payment_status_old;

-- Create new types
CREATE TYPE public.transaction_status AS ENUM (
    'PENDING',
    'PROCESSING',
    'SUCCESS',
    'FAILED'
);

CREATE TYPE public.payment_status AS ENUM (
    'UNPAID',
    'PAID',
    'FAILED',
    'EXPIRED',
    'REFUNDED'
);

-- Update column types
ALTER TABLE public.transactions 
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN status TYPE public.transaction_status 
  USING (
    CASE status::text
      WHEN 'PAID' THEN 'PROCESSING'::public.transaction_status
      WHEN 'EXPIRED' THEN 'FAILED'::public.transaction_status
      WHEN 'REFUNDED' THEN 'FAILED'::public.transaction_status
      ELSE status::text::public.transaction_status
    END
  ),
  ALTER COLUMN status SET DEFAULT 'PENDING'::public.transaction_status;

ALTER TABLE public.transactions
  ALTER COLUMN payment_status DROP DEFAULT,
  ALTER COLUMN payment_status TYPE public.payment_status
  USING payment_status::text::public.payment_status,
  ALTER COLUMN payment_status SET DEFAULT 'UNPAID'::public.payment_status;

-- Drop old types
DROP TYPE public.transaction_status_old;
DROP TYPE public.payment_status_old;
