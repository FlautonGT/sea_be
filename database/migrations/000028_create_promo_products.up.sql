-- Name: promo_products; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.promo_products (
    promo_id uuid NOT NULL,
    product_id uuid NOT NULL
);


ALTER TABLE public.promo_products OWNER TO gate;

--

-- Name: promo_products promo_products_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.promo_products
    ADD CONSTRAINT promo_products_pkey PRIMARY KEY (promo_id, product_id);


--


-- SEED DATA --


INSERT INTO public.promo_products (promo_id, product_id) VALUES
('939fff99-942e-4c04-80f6-d79c8e574dd6', 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d'),
('e3e488e3-83e5-4fef-9703-f78108969401', 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d'),
('4e0f4be6-73a6-4b81-bfe1-c039dcacd085', 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d'),
('e3e488e3-83e5-4fef-9703-f78108969401', 'e16042d2-0a9c-4e23-9ae0-c488a0900d88'),
('4e0f4be6-73a6-4b81-bfe1-c039dcacd085', 'e16042d2-0a9c-4e23-9ae0-c488a0900d88'),
('e3e488e3-83e5-4fef-9703-f78108969401', 'e2552085-cffd-4a34-ab76-dca9e834fd16'),
('4e0f4be6-73a6-4b81-bfe1-c039dcacd085', 'e2552085-cffd-4a34-ab76-dca9e834fd16'),
('e3e488e3-83e5-4fef-9703-f78108969401', 'f7d62818-0275-44f0-af91-d329b3be08a2'),
('4e0f4be6-73a6-4b81-bfe1-c039dcacd085', 'f7d62818-0275-44f0-af91-d329b3be08a2'),
('e3e488e3-83e5-4fef-9703-f78108969401', 'dd85e668-7ab8-4e5c-94e8-87cb867409dd'),
('4e0f4be6-73a6-4b81-bfe1-c039dcacd085', 'dd85e668-7ab8-4e5c-94e8-87cb867409dd'),
('e3e488e3-83e5-4fef-9703-f78108969401', '0725bcb9-e733-4921-99b0-2ffd6763873d'),
('4e0f4be6-73a6-4b81-bfe1-c039dcacd085', '0725bcb9-e733-4921-99b0-2ffd6763873d');