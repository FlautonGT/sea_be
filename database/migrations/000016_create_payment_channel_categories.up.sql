-- Name: payment_channel_categories; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.payment_channel_categories (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code character varying(50) NOT NULL,
    title character varying(100) NOT NULL,
    icon character varying(500),
    is_active boolean DEFAULT true,
    sort_order integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.payment_channel_categories OWNER TO gate;

--

-- Name: payment_channel_categories payment_channel_categories_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.payment_channel_categories
    ADD CONSTRAINT payment_channel_categories_code_key UNIQUE (code);


--

-- Name: payment_channel_categories payment_channel_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.payment_channel_categories
    ADD CONSTRAINT payment_channel_categories_pkey PRIMARY KEY (id);


--


-- SEED DATA --


INSERT INTO public.payment_channel_categories (id, code, title, icon, is_active, sort_order, created_at, updated_at) VALUES
('25b0365c-8eba-4daa-b8a7-36cf9ccf3c66', 'E_WALLET', 'E-Wallet', 'https://cdn.gate.co.id/icons/wallet.svg', 't', '1', '2025-12-04 19:24:29.931493+00', '2025-12-04 19:24:29.931493+00'),
('695ccf5c-dcc1-414b-be4c-ccc1338357bd', 'VIRTUAL_ACCOUNT', 'Virtual Account', 'https://cdn.gate.co.id/icons/bank.svg', 't', '2', '2025-12-04 19:24:29.931493+00', '2025-12-04 19:24:29.931493+00'),
('a2e9f28d-d545-4d77-a29c-e28f9908914e', 'RETAIL', 'Convenience Store', 'https://cdn.gate.co.id/icons/store.svg', 't', '3', '2025-12-04 19:24:29.931493+00', '2025-12-04 19:24:29.931493+00'),
('276989c5-baaf-44f8-9a78-034f9a24cff0', 'CARD', 'Credit or Debit Card', 'https://cdn.gate.co.id/icons/card.svg', 't', '4', '2025-12-04 19:24:29.931493+00', '2025-12-04 19:24:29.931493+00');