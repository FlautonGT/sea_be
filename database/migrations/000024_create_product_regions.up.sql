-- Name: product_regions; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.product_regions (
    product_id uuid NOT NULL,
    region_code public.region_code NOT NULL
);


ALTER TABLE public.product_regions OWNER TO gate;

--

-- Name: product_regions product_regions_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.product_regions
    ADD CONSTRAINT product_regions_pkey PRIMARY KEY (product_id, region_code);


--


-- SEED DATA --


INSERT INTO public.product_regions (product_id, region_code) VALUES
('d3fd0451-8ad0-4dd2-99f2-96e451b49a6d', 'ID'),
('f7d62818-0275-44f0-af91-d329b3be08a2', 'ID'),
('e2552085-cffd-4a34-ab76-dca9e834fd16', 'ID'),
('dd85e668-7ab8-4e5c-94e8-87cb867409dd', 'ID'),
('0725bcb9-e733-4921-99b0-2ffd6763873d', 'ID'),
('0725bcb9-e733-4921-99b0-2ffd6763873d', 'MY'),
('0725bcb9-e733-4921-99b0-2ffd6763873d', 'PH'),
('0725bcb9-e733-4921-99b0-2ffd6763873d', 'SG'),
('0725bcb9-e733-4921-99b0-2ffd6763873d', 'TH'),
('e16042d2-0a9c-4e23-9ae0-c488a0900d88', 'ID');