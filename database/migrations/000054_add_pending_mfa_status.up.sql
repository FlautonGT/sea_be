-- Add PENDING value to mfa_status enum
ALTER TYPE public.mfa_status ADD VALUE IF NOT EXISTS 'PENDING';

