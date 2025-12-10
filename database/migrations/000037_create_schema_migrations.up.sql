-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO gate;

--

-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--


-- SEED DATA --


INSERT INTO public.schema_migrations (version, dirty) VALUES
('1', 't');