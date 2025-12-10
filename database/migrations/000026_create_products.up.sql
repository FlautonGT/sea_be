-- Name: products; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.products (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code character varying(50) NOT NULL,
    slug character varying(100) NOT NULL,
    title character varying(200) NOT NULL,
    subtitle character varying(200),
    description text,
    publisher character varying(200),
    thumbnail character varying(500),
    banner character varying(500),
    category_id uuid NOT NULL,
    is_active boolean DEFAULT true,
    is_popular boolean DEFAULT false,
    features jsonb DEFAULT '[]'::jsonb,
    how_to_order jsonb DEFAULT '[]'::jsonb,
    tags text[],
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    inquiry_slug character varying(100)
);


ALTER TABLE public.products OWNER TO gate;

--

-- Name: products products_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_code_key UNIQUE (code);


--

-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--

-- Name: products products_slug_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_slug_key UNIQUE (slug);


--

-- Name: idx_products_category; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_products_category ON public.products USING btree (category_id);


--

-- Name: idx_products_code; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_products_code ON public.products USING btree (code);


--

-- Name: idx_products_inquiry_slug; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_products_inquiry_slug ON public.products USING btree (inquiry_slug) WHERE (inquiry_slug IS NOT NULL);


--

-- Name: idx_products_is_active; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_products_is_active ON public.products USING btree (is_active);


--

-- Name: idx_products_is_popular; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_products_is_popular ON public.products USING btree (is_popular);


--

-- Name: idx_products_slug; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_products_slug ON public.products USING btree (slug);


--

-- Name: products update_products_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON public.products FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.products (id, code, slug, title, subtitle, description, publisher, thumbnail, banner, category_id, is_active, is_popular, features, how_to_order, tags, created_at, updated_at, inquiry_slug) VALUES
('d3fd0451-8ad0-4dd2-99f2-96e451b49a6d', 'MLBB', 'mobile-legends', 'Mobile Legends: Bang Bang', 'Top Up Diamond ML Murah & Cepat', NULL, 'Moonton', 'https://gate.nos.jkt-1.neo.id/products/f49fe655-7403-4915-be9d-2fb3d3631603.webp', 'https://gate.nos.jkt-1.neo.id/products/65c9e56f-c8be-4a77-a1cd-d0ffebf776c3.webp', 'a9341837-8c47-48ed-be60-c2baee160c1c', 't', 't', '[]', '[]', '{}', '2025-12-04 19:24:29.962406+00', '2025-12-05 20:53:48.428038+00', 'mobile-legends-gp'),
('f7d62818-0275-44f0-af91-d329b3be08a2', 'GENSHIN', 'genshin-impact', 'Genshin Impact', 'Top Up Genesis Crystal Murah', NULL, 'miHoYo', 'https://gate.nos.jkt-1.neo.id/products/229320de-7de6-4d74-8046-654dc6aea14f.webp', 'https://gate.nos.jkt-1.neo.id/products/941d00b6-b9ad-445f-b0c3-e926d3751f09.webp', 'a9341837-8c47-48ed-be60-c2baee160c1c', 't', 't', '[]', '[]', '{}', '2025-12-04 19:24:29.962406+00', '2025-12-07 21:13:28.714893+00', NULL),
('e2552085-cffd-4a34-ab76-dca9e834fd16', 'PUBGM', 'pubg-mobile', 'PUBG Mobile', 'Top Up UC PUBG Mobile Murah', NULL, 'Krafton', 'https://gate.nos.jkt-1.neo.id/products/02ced4c9-7ee9-46bd-949a-3d5719c37e5a.webp', 'https://gate.nos.jkt-1.neo.id/products/322e3b06-8963-4d8e-9d1d-ef1592043d2b.webp', 'a9341837-8c47-48ed-be60-c2baee160c1c', 't', 't', '[]', '[]', '{}', '2025-12-04 19:24:29.962406+00', '2025-12-07 21:13:42.859171+00', 'pubg-mobile-gp'),
('dd85e668-7ab8-4e5c-94e8-87cb867409dd', 'VALORANT', 'valorant', 'Valorant', 'Top Up Valorant Points Murah', NULL, 'Riot Games', 'https://gate.nos.jkt-1.neo.id/products/9731ae2c-63ee-4827-9bcd-54c193b387f5.webp', 'https://gate.nos.jkt-1.neo.id/products/5a5f1fd3-484c-4fc4-b3eb-780185dda815.webp', 'a9341837-8c47-48ed-be60-c2baee160c1c', 't', 't', '[]', '[]', '{}', '2025-12-04 19:24:29.962406+00', '2025-12-08 14:48:53.906161+00', NULL),
('0725bcb9-e733-4921-99b0-2ffd6763873d', 'HSR', 'honkai-star-rail', 'Honkai Star Rail', 'Top Up Oneiric Shard Murah', NULL, 'HoYoverse', 'https://gate.nos.jkt-1.neo.id/products/1cf1ff78-7471-4728-96ea-b8f5f79bb310.webp', 'https://gate.nos.jkt-1.neo.id/products/f198848b-e629-49cb-9c90-2cd5aa9a886e.webp', 'a9341837-8c47-48ed-be60-c2baee160c1c', 't', 't', '[]', '[]', '{}', '2025-12-04 19:24:29.962406+00', '2025-12-08 14:48:59.511835+00', NULL),
('e16042d2-0a9c-4e23-9ae0-c488a0900d88', 'FF', 'free-fire', 'Free Fire', 'Top Up Diamond FF Murah & Cepat', NULL, 'Garena', 'https://gate.nos.jkt-1.neo.id/products/d78e6709-df12-4918-8743-cc4377f25ae6.webp', 'https://gate.nos.jkt-1.neo.id/products/b1915e40-3a9b-4e6e-b353-0541947a388e.webp', 'a9341837-8c47-48ed-be60-c2baee160c1c', 't', 't', '[]', '[]', '{}', '2025-12-04 19:24:29.962406+00', '2025-12-09 20:03:08.681508+00', 'free-fire-dg');