-- Name: categories; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.categories (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code character varying(50) NOT NULL,
    title character varying(100) NOT NULL,
    description text,
    icon character varying(500),
    is_active boolean DEFAULT true,
    sort_order integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.categories OWNER TO gate;

--

-- Name: categories categories_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_code_key UNIQUE (code);


--

-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--

-- Name: categories update_categories_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON public.categories FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.categories (id, code, title, description, icon, is_active, sort_order, created_at, updated_at) VALUES
('a9341837-8c47-48ed-be60-c2baee160c1c', 'top-up-game', 'Top Up Game', 'Top up diamond, UC, dan in-game currency lainnya', '≡ƒÄ«', 't', '1', '2025-12-04 19:24:29.941649+00', '2025-12-08 09:09:21.983702+00'),
('0421994d-df98-4d13-a527-8276a4b6a606', 'streaming', 'Streaming', 'Langganan Netflix, Spotify, Disney+ dan lainnya', '≡ƒô▒', 't', '1', '2025-12-04 19:24:29.941649+00', '2025-12-08 09:09:35.655524+00'),
('c3042b7e-2b19-47a1-9170-f719ea7c143d', 'voucher', 'Voucher', 'Voucher game dan digital content', '≡ƒÄƒ∩╕Å', 't', '1', '2025-12-04 19:24:29.941649+00', '2025-12-08 09:09:49.237387+00');