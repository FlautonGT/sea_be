-- Name: regions; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.regions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code public.region_code NOT NULL,
    country character varying(100) NOT NULL,
    currency public.currency_code NOT NULL,
    currency_symbol character varying(10) NOT NULL,
    image character varying(500),
    is_default boolean DEFAULT false,
    is_active boolean DEFAULT true,
    sort_order integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.regions OWNER TO gate;

--

-- Name: regions regions_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.regions
    ADD CONSTRAINT regions_code_key UNIQUE (code);


--

-- Name: regions regions_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.regions
    ADD CONSTRAINT regions_pkey PRIMARY KEY (id);


--

-- Name: regions update_regions_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_regions_updated_at BEFORE UPDATE ON public.regions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.regions (id, code, country, currency, currency_symbol, image, is_default, is_active, sort_order, created_at, updated_at) VALUES
('0c506105-70a5-4af9-a649-61632926f982', 'ID', 'Indonesia', 'IDR', 'Rp', 'https://gate.nos.jkt-1.neo.id/flags/ead217af-05dc-4c74-97b3-a95061a2bc27.svg', 't', 't', '1', '2025-12-04 19:24:29.715452+00', '2025-12-04 21:23:21.582553+00'),
('23ec7a7d-3212-4e43-95c1-c0b52c01d712', 'MY', 'Malaysia', 'MYR', 'RM', 'https://gate.nos.jkt-1.neo.id/flags/ec2b4833-c05a-4c06-86ea-7ffc3051a6be.svg', 'f', 't', '2', '2025-12-04 19:24:29.715452+00', '2025-12-04 21:23:34.57551+00'),
('abfd2347-b943-4fb3-8c24-10e8f28a8cce', 'PH', 'Philippines', 'PHP', '???', 'https://gate.nos.jkt-1.neo.id/flags/da6ffa81-5b1c-457a-91fd-d3197ffdfd09.svg', 'f', 't', '3', '2025-12-04 19:24:29.715452+00', '2025-12-04 21:23:40.038638+00'),
('f37ab772-2188-4822-8677-d773e1aa5c64', 'SG', 'Singapore', 'SGD', 'S$', 'https://gate.nos.jkt-1.neo.id/flags/812251c9-276a-4138-9b67-aa041f3013ba.svg', 'f', 't', '4', '2025-12-04 19:24:29.715452+00', '2025-12-04 21:23:45.302146+00'),
('6cb4eb90-c32b-4600-a2e0-aefea711303e', 'TH', 'Thailand', 'THB', '???', 'https://gate.nos.jkt-1.neo.id/flags/7162df54-a290-49c1-93d4-e063ca465d66.svg', 'f', 't', '5', '2025-12-04 19:24:29.715452+00', '2025-12-04 21:23:53.757225+00');