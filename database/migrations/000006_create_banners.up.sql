-- Name: banners; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.banners (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    title character varying(200) NOT NULL,
    description text,
    href character varying(500),
    image character varying(500) NOT NULL,
    is_active boolean DEFAULT true,
    sort_order integer DEFAULT 0,
    start_at timestamp with time zone,
    expired_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.banners OWNER TO gate;

--

-- Name: banners banners_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.banners
    ADD CONSTRAINT banners_pkey PRIMARY KEY (id);


--

-- Name: banners update_banners_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_banners_updated_at BEFORE UPDATE ON public.banners FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.banners (id, title, description, href, image, is_active, sort_order, start_at, expired_at, created_at, updated_at) VALUES
('7ef7d361-e377-4e8b-9585-82b64a82937e', 'Mobile Legends x Transformers', 'Event kolaborasi terbaru sudah hadir!', '/products/mobile-legends', 'https://gate.nos.jkt-1.neo.id/banners/c1635119-5631-4da8-879e-6ab3fa805c73.png', 't', '2', '2025-12-04 05:24:00+00', '2025-12-18 05:24:00+00', '2025-12-04 19:24:30.022557+00', '2025-12-04 20:47:16.26755+00'),
('85f97a4f-6748-4c05-b6d3-c5c0af2ebfb6', 'Promo Tahun Baru 2025', 'Diskon hingga 20% untuk semua top up game!', '/mobile-legends', 'https://gate.nos.jkt-1.neo.id/banners/70d4ad44-1167-486e-bfa3-54e84214bb72.png', 't', '1', '2025-12-03 22:24:00+00', '2026-01-02 22:24:00+00', '2025-12-04 19:24:30.022557+00', '2025-12-08 14:51:15.096278+00');