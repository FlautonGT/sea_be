-- Name: contacts; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.contacts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    email character varying(255),
    phone character varying(50),
    whatsapp character varying(500),
    instagram character varying(500),
    facebook character varying(500),
    x character varying(500),
    youtube character varying(500),
    telegram character varying(500),
    discord character varying(500),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.contacts OWNER TO gate;

--

-- Name: contacts contacts_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.contacts
    ADD CONSTRAINT contacts_pkey PRIMARY KEY (id);


--


-- SEED DATA --


INSERT INTO public.contacts (id, email, phone, whatsapp, instagram, facebook, x, youtube, telegram, discord, updated_at) VALUES
('61dbbf00-e008-421e-b33f-5c6466097134', 'support@gate.co.id', '+6281234567890', 'https://wa.me/6281234567890', 'https://instagram.com/gate.official', 'https://facebook.com/gate.official', 'https://x.com/gate_official', 'https://youtube.com/@gateofficial', 'https://t.me/gate_official', 'https://discord.gg/gate', '2025-12-04 19:24:29.95517+00');