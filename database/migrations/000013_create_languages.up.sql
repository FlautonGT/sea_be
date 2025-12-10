-- Name: languages; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.languages (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code character varying(10) NOT NULL,
    name character varying(100) NOT NULL,
    country character varying(100) NOT NULL,
    image character varying(500),
    is_default boolean DEFAULT false,
    is_active boolean DEFAULT true,
    sort_order integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.languages OWNER TO gate;

--

-- Name: languages languages_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.languages
    ADD CONSTRAINT languages_code_key UNIQUE (code);


--

-- Name: languages languages_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.languages
    ADD CONSTRAINT languages_pkey PRIMARY KEY (id);


--

-- Name: languages update_languages_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_languages_updated_at BEFORE UPDATE ON public.languages FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.languages (id, code, name, country, image, is_default, is_active, sort_order, created_at, updated_at) VALUES
('06733d0e-d494-444c-b40b-b245549dfca7', 'id', 'Bahasa Indonesia', 'Indonesia', 'https://gate.nos.jkt-1.neo.id/flags/8378f017-a38f-4ad5-8c72-c8a94f311b41.svg', 't', 't', '1', '2025-12-04 19:24:29.721252+00', '2025-12-04 21:41:38.629544+00'),
('2f4eebaf-5321-4e39-8d34-a208cbf6306c', 'en', 'English', 'United States', 'https://gate.nos.jkt-1.neo.id/flags/1b710bd6-e038-455b-8049-323f691b50b8.svg', 'f', 't', '2', '2025-12-04 19:24:29.721252+00', '2025-12-04 21:41:44.207015+00');