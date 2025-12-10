-- Name: banner_regions; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.banner_regions (
    banner_id uuid NOT NULL,
    region_code public.region_code NOT NULL
);


ALTER TABLE public.banner_regions OWNER TO gate;

--

-- Name: banner_regions banner_regions_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.banner_regions
    ADD CONSTRAINT banner_regions_pkey PRIMARY KEY (banner_id, region_code);


--


-- SEED DATA --


INSERT INTO public.banner_regions (banner_id, region_code) VALUES
('7ef7d361-e377-4e8b-9585-82b64a82937e', 'ID'),
('7ef7d361-e377-4e8b-9585-82b64a82937e', 'MY'),
('7ef7d361-e377-4e8b-9585-82b64a82937e', 'PH'),
('7ef7d361-e377-4e8b-9585-82b64a82937e', 'SG'),
('7ef7d361-e377-4e8b-9585-82b64a82937e', 'TH'),
('85f97a4f-6748-4c05-b6d3-c5c0af2ebfb6', 'ID'),
('85f97a4f-6748-4c05-b6d3-c5c0af2ebfb6', 'MY'),
('85f97a4f-6748-4c05-b6d3-c5c0af2ebfb6', 'PH'),
('85f97a4f-6748-4c05-b6d3-c5c0af2ebfb6', 'SG'),
('85f97a4f-6748-4c05-b6d3-c5c0af2ebfb6', 'TH');