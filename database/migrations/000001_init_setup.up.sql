--
-- PostgreSQL database dump
--

\restrict c3QP85HytuRcK9VXINTPw6aS0oD3Nq9I2uA9HUC4QCwqO7cVPLyTSmd2pjIpcbh

-- Dumped from database version 15.15
-- Dumped by pg_dump version 15.15

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

---- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--

-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--

-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--

-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--

-- Name: admin_role; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.admin_role AS ENUM (
    'SUPERADMIN',
    'ADMIN',
    'FINANCE',
    'CS_LEAD',
    'CS'
);


ALTER TYPE public.admin_role OWNER TO gate;

--

-- Name: audit_action; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.audit_action AS ENUM (
    'CREATE',
    'UPDATE',
    'DELETE',
    'LOGIN',
    'LOGOUT'
);


ALTER TYPE public.audit_action OWNER TO gate;

--

-- Name: currency_code; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.currency_code AS ENUM (
    'IDR',
    'MYR',
    'PHP',
    'SGD',
    'THB'
);


ALTER TYPE public.currency_code OWNER TO gate;

--

-- Name: deposit_status; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.deposit_status AS ENUM (
    'PENDING',
    'SUCCESS',
    'FAILED',
    'EXPIRED',
    'REFUNDED'
);


ALTER TYPE public.deposit_status OWNER TO gate;

--

-- Name: fee_type; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.fee_type AS ENUM (
    'FIXED',
    'PERCENTAGE',
    'MIXED'
);


ALTER TYPE public.fee_type OWNER TO gate;

--

-- Name: field_type; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.field_type AS ENUM (
    'text',
    'number',
    'email',
    'select',
    'phone'
);


ALTER TYPE public.field_type OWNER TO gate;

--

-- Name: health_status; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.health_status AS ENUM (
    'HEALTHY',
    'DEGRADED',
    'UNHEALTHY'
);


ALTER TYPE public.health_status OWNER TO gate;

--

-- Name: membership_level; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.membership_level AS ENUM (
    'CLASSIC',
    'PRESTIGE',
    'ROYAL'
);


ALTER TYPE public.membership_level OWNER TO gate;

--

-- Name: mfa_status; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.mfa_status AS ENUM (
    'ACTIVE',
    'INACTIVE'
);


ALTER TYPE public.mfa_status OWNER TO gate;

--

-- Name: mutation_type; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.mutation_type AS ENUM (
    'CREDIT',
    'DEBIT'
);


ALTER TYPE public.mutation_type OWNER TO gate;

--

-- Name: payment_status; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.payment_status AS ENUM (
    'UNPAID',
    'PAID',
    'EXPIRED',
    'REFUNDED'
);


ALTER TYPE public.payment_status OWNER TO gate;

--

-- Name: payment_type; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.payment_type AS ENUM (
    'purchase',
    'deposit'
);


ALTER TYPE public.payment_type OWNER TO gate;

--

-- Name: region_code; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.region_code AS ENUM (
    'ID',
    'MY',
    'PH',
    'SG',
    'TH'
);


ALTER TYPE public.region_code OWNER TO gate;

--

-- Name: transaction_status; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.transaction_status AS ENUM (
    'PENDING',
    'PAID',
    'PROCESSING',
    'SUCCESS',
    'FAILED',
    'EXPIRED',
    'REFUNDED'
);


ALTER TYPE public.transaction_status OWNER TO gate;

--

-- Name: user_status; Type: TYPE; Schema: public; Owner: seaply
--

CREATE TYPE public.user_status AS ENUM (
    'ACTIVE',
    'INACTIVE',
    'SUSPENDED'
);


ALTER TYPE public.user_status OWNER TO gate;

--

-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: seaply
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO gate;

SET default_tablespace = '';

SET default_table_access_method = heap;

--

-- Name: COLUMN payment_channels.gateway_code; Type: COMMENT; Schema: public; Owner: seaply
--

COMMENT ON COLUMN public.payment_channels.gateway_code IS 'Gateway-specific code for this payment channel (e.g., bank code 002 for BRI)';


--