-- Name: payment_channel_regions; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.payment_channel_regions (
    channel_id uuid NOT NULL,
    region_code public.region_code NOT NULL
);


ALTER TABLE public.payment_channel_regions OWNER TO gate;

--

-- Name: payment_channel_regions payment_channel_regions_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.payment_channel_regions
    ADD CONSTRAINT payment_channel_regions_pkey PRIMARY KEY (channel_id, region_code);


--


-- SEED DATA --


INSERT INTO public.payment_channel_regions (channel_id, region_code) VALUES
('9457bd7d-216c-4ae4-a654-02272dec87d8', 'ID'),
('6b754f8e-c124-4f00-9776-36d9c73fde9e', 'ID'),
('f72f575c-4cd8-42fa-b1a3-6bf78bfd6a0c', 'ID'),
('a45b214c-cd62-4be9-862c-83c034ddf4f2', 'ID'),
('2162d8a6-906b-4c03-a2eb-c905f1228879', 'ID'),
('d5606ef6-ee54-46df-9739-64220fdc1b28', 'ID'),
('48604997-4746-47d8-862e-f49efd23bdd7', 'ID'),
('c708c6d0-844d-47ac-8475-fb2f3b81b196', 'ID'),
('0d28ee79-9767-4fbf-ad0b-732fb42a9e05', 'ID'),
('a7251199-b644-4a34-be74-298819a2841e', 'ID'),
('cb1707a4-285f-4260-9d4b-5908be50d087', 'ID');