-- Name: category_regions; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.category_regions (
    category_id uuid NOT NULL,
    region_code public.region_code NOT NULL
);


ALTER TABLE public.category_regions OWNER TO gate;

--

-- Name: category_regions category_regions_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.category_regions
    ADD CONSTRAINT category_regions_pkey PRIMARY KEY (category_id, region_code);


--


-- SEED DATA --


INSERT INTO public.category_regions (category_id, region_code) VALUES
('a9341837-8c47-48ed-be60-c2baee160c1c', 'TH'),
('a9341837-8c47-48ed-be60-c2baee160c1c', 'SG'),
('a9341837-8c47-48ed-be60-c2baee160c1c', 'PH'),
('a9341837-8c47-48ed-be60-c2baee160c1c', 'MY'),
('a9341837-8c47-48ed-be60-c2baee160c1c', 'ID'),
('0421994d-df98-4d13-a527-8276a4b6a606', 'TH'),
('0421994d-df98-4d13-a527-8276a4b6a606', 'SG'),
('0421994d-df98-4d13-a527-8276a4b6a606', 'PH'),
('0421994d-df98-4d13-a527-8276a4b6a606', 'MY'),
('0421994d-df98-4d13-a527-8276a4b6a606', 'ID'),
('c3042b7e-2b19-47a1-9170-f719ea7c143d', 'ID'),
('c3042b7e-2b19-47a1-9170-f719ea7c143d', 'MY'),
('c3042b7e-2b19-47a1-9170-f719ea7c143d', 'PH'),
('c3042b7e-2b19-47a1-9170-f719ea7c143d', 'SG'),
('c3042b7e-2b19-47a1-9170-f719ea7c143d', 'TH');