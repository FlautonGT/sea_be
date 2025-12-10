-- Name: payment_gateways; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.payment_gateways (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code character varying(50) NOT NULL,
    name character varying(100) NOT NULL,
    base_url character varying(500) NOT NULL,
    callback_url character varying(500),
    is_active boolean DEFAULT true,
    supported_methods text[],
    supported_types public.payment_type[],
    health_status public.health_status DEFAULT 'HEALTHY'::public.health_status,
    last_health_check timestamp with time zone,
    api_config jsonb DEFAULT '{"timeout": 30000, "retryAttempts": 3}'::jsonb,
    status_mapping jsonb DEFAULT '{}'::jsonb,
    env_credential_keys jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.payment_gateways OWNER TO gate;

--

-- Name: payment_gateways payment_gateways_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.payment_gateways
    ADD CONSTRAINT payment_gateways_code_key UNIQUE (code);


--

-- Name: payment_gateways payment_gateways_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.payment_gateways
    ADD CONSTRAINT payment_gateways_pkey PRIMARY KEY (id);


--

-- Name: payment_gateways update_payment_gateways_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_payment_gateways_updated_at BEFORE UPDATE ON public.payment_gateways FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.payment_gateways (id, code, name, base_url, callback_url, is_active, supported_methods, supported_types, health_status, last_health_check, api_config, status_mapping, env_credential_keys, created_at, updated_at) VALUES
('b012905e-b4b4-4f74-8bfb-de3a895d8d20', 'LINKQU', 'LinkQu', 'https://api.linkqu.id', 'https://gateway.gate.id/callbacks/linkqu', 't', '{QRIS}', '{purchase,deposit}', 'HEALTHY', NULL, '{"timeout": 30000, "retryAttempts": 3}', '{}', '{"pin": "LINKQU_PIN", "clientId": "LINKQU_CLIENT_ID", "username": "LINKQU_USERNAME", "clientSecret": "LINKQU_CLIENT_SECRET"}', '2025-12-04 19:24:29.935121+00', '2025-12-04 19:24:29.935121+00'),
('4a7124cf-a073-4cee-adbe-afb047f6a314', 'BCA_DIRECT', 'BCA Direct API', 'https://sandbox.bca.co.id', 'https://gateway.gate.id/callbacks/bca', 't', '{BCA_VA}', '{purchase,deposit}', 'HEALTHY', NULL, '{"timeout": 30000, "retryAttempts": 3}', '{}', '{"apiKey": "BCA_API_KEY", "clientId": "BCA_CLIENT_ID", "apiSecret": "BCA_API_SECRET", "clientSecret": "BCA_CLIENT_SECRET"}', '2025-12-04 19:24:29.935121+00', '2025-12-04 19:24:29.935121+00'),
('84ff7eef-e31a-4a13-ad84-8eae3bb1ec43', 'BRI_DIRECT', 'BRI Direct API', 'https://sandbox.bri.co.id', 'https://gateway.gate.id/callbacks/bri', 't', '{BRI_VA}', '{purchase,deposit}', 'HEALTHY', NULL, '{"timeout": 30000, "retryAttempts": 3}', '{}', '{"clientId": "BRI_CLIENT_ID", "clientSecret": "BRI_CLIENT_SECRET"}', '2025-12-04 19:24:29.935121+00', '2025-12-04 19:24:29.935121+00'),
('4e8cbd84-4eb2-403a-a365-3cce2f9b5dbc', 'XENDIT', 'Xendit', 'https://api.xendit.co', 'https://gateway.gate.id/callbacks/xendit', 't', '{PERMATA_VA,MANDIRI_VA,CARD}', '{purchase,deposit}', 'HEALTHY', NULL, '{"timeout": 30000, "retryAttempts": 3}', '{}', '{"secretKey": "XENDIT_SECRET_KEY", "callbackToken": "XENDIT_CALLBACK_TOKEN"}', '2025-12-04 19:24:29.935121+00', '2025-12-04 19:24:29.935121+00'),
('47f1ccd5-df1c-4848-bf32-67324c60f18b', 'MIDTRANS', 'Midtrans', 'https://api.midtrans.com', 'https://gateway.gate.id/callbacks/midtrans', 't', '{GOPAY,SHOPEEPAY}', '{purchase}', 'HEALTHY', NULL, '{"timeout": 30000, "retryAttempts": 3}', '{}', '{"clientKey": "MIDTRANS_CLIENT_KEY", "serverKey": "MIDTRANS_SERVER_KEY"}', '2025-12-04 19:24:29.935121+00', '2025-12-04 19:24:29.935121+00'),
('04bc68be-f58c-475b-947f-1d4301e99fbf', 'DANA_DIRECT', 'DANA Direct', 'https://api.dana.id', 'https://gateway.gate.id/callbacks/dana', 't', '{DANA}', '{purchase}', 'HEALTHY', NULL, '{"timeout": 30000, "retryAttempts": 3}', '{}', '{"clientId": "DANA_CLIENT_ID", "clientSecret": "DANA_CLIENT_SECRET"}', '2025-12-04 19:24:29.935121+00', '2025-12-04 19:24:29.935121+00');