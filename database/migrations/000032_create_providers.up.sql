-- Name: providers; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.providers (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code character varying(50) NOT NULL,
    name character varying(100) NOT NULL,
    base_url character varying(500) NOT NULL,
    webhook_url character varying(500),
    is_active boolean DEFAULT true,
    priority integer DEFAULT 0,
    supported_types text[],
    health_status public.health_status DEFAULT 'HEALTHY'::public.health_status,
    last_health_check timestamp with time zone,
    api_config jsonb DEFAULT '{"timeout": 30000, "retryDelay": 1000, "retryAttempts": 3}'::jsonb,
    status_mapping jsonb DEFAULT '{}'::jsonb,
    env_credential_keys jsonb DEFAULT '{}'::jsonb,
    total_skus integer DEFAULT 0,
    active_skus integer DEFAULT 0,
    success_rate numeric(5,2) DEFAULT 0,
    avg_response_time integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.providers OWNER TO gate;

--

-- Name: providers providers_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.providers
    ADD CONSTRAINT providers_code_key UNIQUE (code);


--

-- Name: providers providers_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.providers
    ADD CONSTRAINT providers_pkey PRIMARY KEY (id);


--

-- Name: providers update_providers_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_providers_updated_at BEFORE UPDATE ON public.providers FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.providers (id, code, name, base_url, webhook_url, is_active, priority, supported_types, health_status, last_health_check, api_config, status_mapping, env_credential_keys, total_skus, active_skus, success_rate, avg_response_time, created_at, updated_at) VALUES
('1c784573-f4ea-488a-9c25-1f87bb43d447', 'DIGIFLAZZ', 'Digiflazz', 'https://api.digiflazz.com/v1', 'https://gateway.gate.id/webhooks/digiflazz', 't', '1', '{PULSA,DATA,GAME,EWALLET,PLN}', 'HEALTHY', '2025-12-04 19:24:33.504156+00', '{"timeout": 30000, "retryDelay": 1000, "retryAttempts": 3}', '{}', '{"apiKey": "DIGIFLAZZ_API_KEY", "username": "DIGIFLAZZ_USERNAME", "webhookSecret": "DIGIFLAZZ_WEBHOOK_SECRET"}', '28', '28', '98.50', '1200', '2025-12-04 19:24:29.938285+00', '2025-12-04 19:24:33.504156+00'),
('3a497ae6-690b-48ac-9c8f-17e0f01471ab', 'VIPRESELLER', 'VIP Reseller', 'https://vip-reseller.co.id/api', 'https://gateway.gate.id/webhooks/vipreseller', 't', '2', '{GAME,VOUCHER}', 'HEALTHY', '2025-12-04 19:24:33.508194+00', '{"timeout": 30000, "retryDelay": 1000, "retryAttempts": 3}', '{}', '{"apiId": "VIPRESELLER_API_ID", "apiKey": "VIPRESELLER_API_KEY"}', '0', '0', '97.80', '1500', '2025-12-04 19:24:29.938285+00', '2025-12-04 19:24:33.508194+00'),
('f5ef4de3-d2f0-4416-8da8-b591cde5f46e', 'BANGJEFF', 'BangJeff', 'https://api.bangjeff.com', 'https://gateway.gate.id/webhooks/bangjeff', 't', '3', '{GAME,STREAMING}', 'DEGRADED', '2025-12-04 19:24:33.5115+00', '{"timeout": 30000, "retryDelay": 1000, "retryAttempts": 3}', '{}', '{"memberId": "BANGJEFF_MEMBER_ID", "secretKey": "BANGJEFF_SECRET_KEY"}', '0', '0', '95.20', '2100', '2025-12-04 19:24:29.938285+00', '2025-12-04 19:24:33.5115+00');