-- Name: product_sections; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.product_sections (
    product_id uuid NOT NULL,
    section_id uuid NOT NULL
);


ALTER TABLE public.product_sections OWNER TO gate;

--

-- Name: product_sections product_sections_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.product_sections
    ADD CONSTRAINT product_sections_pkey PRIMARY KEY (product_id, section_id);


--


-- SEED DATA --


INSERT INTO public.product_sections (product_id, section_id) VALUES
('d3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df'),
('e16042d2-0a9c-4e23-9ae0-c488a0900d88', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df'),
('f7d62818-0275-44f0-af91-d329b3be08a2', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df'),
('dd85e668-7ab8-4e5c-94e8-87cb867409dd', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df'),
('e2552085-cffd-4a34-ab76-dca9e834fd16', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df'),
('d3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '838caff3-a279-47dd-a95e-a658425bf1d5'),
('d3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '95e9d0bc-55b1-477f-8629-0b7fc3e1ee9b');