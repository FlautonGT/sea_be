-- Name: sections; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.sections (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code character varying(50) NOT NULL,
    title character varying(100) NOT NULL,
    icon character varying(50),
    is_active boolean DEFAULT true,
    sort_order integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.sections OWNER TO gate;

--

-- Name: sections sections_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.sections
    ADD CONSTRAINT sections_code_key UNIQUE (code);


--

-- Name: sections sections_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.sections
    ADD CONSTRAINT sections_pkey PRIMARY KEY (id);


--

-- Name: sections update_sections_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_sections_updated_at BEFORE UPDATE ON public.sections FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.sections (id, code, title, icon, is_active, sort_order, created_at, updated_at) VALUES
('83e12b66-a25a-4e4e-8acf-bdbfd67e57df', 'topup-instant', 'Topup Instan', '≡ƒÄ«', 't', '1', '2025-12-04 19:24:29.948558+00', '2025-12-08 09:18:13.000257+00'),
('cb63f4d0-2cc6-49a0-9001-885b7744fdc9', 'monthly-pass', 'Monthly Pass', '≡ƒÅå', 't', '1', '2025-12-04 19:24:29.948558+00', '2025-12-08 09:19:00.425597+00'),
('838caff3-a279-47dd-a95e-a658425bf1d5', 'weekly-pass', 'Weekly Pass', '≡ƒôà', 't', '1', '2025-12-04 19:24:29.948558+00', '2025-12-09 12:17:27.57918+00'),
('95e9d0bc-55b1-477f-8629-0b7fc3e1ee9b', 'special-item', 'Spesial Item', '≡ƒöÑ', 't', '1', '2025-12-04 19:24:29.948558+00', '2025-12-09 12:17:32.787633+00');