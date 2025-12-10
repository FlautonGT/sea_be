-- Name: popups; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.popups (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    region_code public.region_code NOT NULL,
    title character varying(200),
    content text,
    image character varying(500),
    href character varying(500),
    is_active boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.popups OWNER TO gate;

--

-- Name: popups popups_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.popups
    ADD CONSTRAINT popups_pkey PRIMARY KEY (id);


--

-- Name: popups popups_region_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.popups
    ADD CONSTRAINT popups_region_code_key UNIQUE (region_code);


--

-- Name: popups update_popups_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_popups_updated_at BEFORE UPDATE ON public.popups FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.popups (id, region_code, title, content, image, href, is_active, created_at, updated_at) VALUES
('7ad33173-7ce3-4b81-995c-5196bf43fe4e', 'ID', 'Selamat Datang di Gate!', '<p>ΓÜí Top Up Instan? Gate Solusinya! Isi pulsa, beli diamond game, hingga saldo e-wallet dalam satu platform cepat & aman.</p>', 'https://gate.nos.jkt-1.neo.id/popups/2d8bbc23-ad67-4ce9-ae7d-029dbf2968cc.png', '', 't', '2025-12-04 19:24:29.951783+00', '2025-12-04 20:31:46.447075+00'),
('bc6f38af-b748-4ca1-8207-3a8b1dec48d1', 'MY', 'Selamat Datang di Gate!', '<p>ΓÜí Isi Semula Segera? Gate adalah penyelesaiannya! Isi semula kredit telefon bimbit anda, beli berlian permainan, dan tambah baki dompet elektronik anda dalam satu platform yang pantas dan selamat.</p>', 'https://gate.nos.jkt-1.neo.id/popups/b4927919-44d7-450e-8d9d-66274011b006.png', '', 't', '2025-12-04 19:24:29.951783+00', '2025-12-04 20:33:47.857842+00'),
('24ede188-7049-4ab6-bbbb-a20981e8f843', 'PH', 'Maligayang pagdating sa Gate!', '<p>ΓÜí Mabilis na Top Up? Gate ang solusyon! Mag-top up ng mobile credit, bumili ng game diamonds, at magdagdag sa balanse ng iyong e-wallet sa isang mabilis at ligtas na plataporma.</p>', 'https://gate.nos.jkt-1.neo.id/popups/cffb6378-44d5-4ac2-ab3b-2dd8718a76b1.png', '', 't', '2025-12-04 19:24:29.951783+00', '2025-12-04 20:34:45.319228+00'),
('f30e88a1-8569-4242-9dbd-d9d2285ab5bd', 'SG', 'Welcome to Gate!', '<p>ΓÜí Instant Top Up? Gate is the solution! Top up your phone credit, buy game diamonds, and add to your e-wallet balance in one fast and secure platform.</p>', 'https://gate.nos.jkt-1.neo.id/popups/c97edfb5-f68c-4107-b2fe-981f28449a16.png', '', 't', '2025-12-04 19:24:29.951783+00', '2025-12-04 20:35:31.332675+00'),
('1fcdd188-e7d2-47a3-b43e-002a36da7ae9', 'TH', 'α╕óα╕┤α╕Öα╕öα╕╡α╕òα╣ëα╕¡α╕Öα╕úα╕▒α╕Üα╕¬α╕╣α╣êα╕¢α╕úα╕░α╕òα╕╣!', '<p>ΓÜí α╣Çα╕òα╕┤α╕íα╣Çα╕çα╕┤α╕Öα╕ùα╕▒α╕Öα╕ùα╕╡? Gate α╕äα╕╖α╕¡α╕äα╕│α╕òα╕¡α╕Ü! α╣Çα╕òα╕┤α╕íα╣Çα╕çα╕┤α╕Öα╣éα╕ùα╕úα╕¿α╕▒α╕₧α╕ùα╣î α╕ïα╕╖α╣ëα╕¡α╣Çα╕₧α╕èα╕úα╣âα╕Öα╣Çα╕üα╕í α╣üα╕Ñα╕░α╣üα╕íα╣ëα╕üα╕úα╕░α╕ùα╕▒α╣êα╕çα╣Çα╕òα╕┤α╕íα╣Çα╕çα╕┤α╕Öα╣âα╕Öα╕üα╕úα╕░α╣Çα╕¢α╣ïα╕▓α╣Çα╕çα╕┤α╕Öα╕¡α╕┤α╣Çα╕Ñα╣çα╕üα╕ùα╕úα╕¡α╕Öα╕┤α╕üα╕¬α╣îα╕éα╕¡α╕çα╕äα╕╕α╕ôα╕Üα╕Öα╣üα╕₧α╕Ñα╕òα╕ƒα╕¡α╕úα╣îα╕íα╣Çα╕öα╕╡α╕óα╕ºα╕ùα╕╡α╣êα╕úα╕ºα╕öα╣Çα╕úα╣çα╕ºα╣üα╕Ñα╕░α╕¢α╕Ñα╕¡α╕öα╕áα╕▒α╕ó</p>', 'https://gate.nos.jkt-1.neo.id/popups/762b0866-1173-4554-9d42-1df226ab5fed.png', '', 't', '2025-12-04 19:24:29.951783+00', '2025-12-04 20:37:03.329616+00');