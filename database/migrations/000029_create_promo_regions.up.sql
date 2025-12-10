-- Name: promo_regions; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.promo_regions (
    promo_id uuid NOT NULL,
    region_code public.region_code NOT NULL
);


ALTER TABLE public.promo_regions OWNER TO gate;

--

-- Name: promo_regions promo_regions_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.promo_regions
    ADD CONSTRAINT promo_regions_pkey PRIMARY KEY (promo_id, region_code);


--


-- SEED DATA --


INSERT INTO public.promo_regions (promo_id, region_code) VALUES
('0136662d-c470-4dba-bfd4-e4f26f40f3d7', 'ID'),
('939fff99-942e-4c04-80f6-d79c8e574dd6', 'ID'),
('e3e488e3-83e5-4fef-9703-f78108969401', 'ID'),
('4e0f4be6-73a6-4b81-bfe1-c039dcacd085', 'ID'),
('b7a452b5-e5bf-454e-88f9-a2b366b70c2f', 'ID'),
('e3e488e3-83e5-4fef-9703-f78108969401', 'MY'),
('7b546b04-c6d5-4e70-8d73-98e205cf8738', 'ID');