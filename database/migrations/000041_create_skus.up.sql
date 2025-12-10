-- Name: skus; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.skus (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code character varying(100) NOT NULL,
    provider_sku_code character varying(100) NOT NULL,
    name character varying(200) NOT NULL,
    description text,
    image character varying(500),
    info text,
    product_id uuid NOT NULL,
    provider_id uuid NOT NULL,
    section_id uuid,
    process_time integer DEFAULT 0,
    is_active boolean DEFAULT true,
    is_featured boolean DEFAULT false,
    badge_text character varying(50),
    badge_color character varying(20),
    stock_status character varying(20) DEFAULT 'AVAILABLE'::character varying,
    total_sold integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.skus OWNER TO gate;

--

-- Name: skus skus_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.skus
    ADD CONSTRAINT skus_code_key UNIQUE (code);


--

-- Name: skus skus_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.skus
    ADD CONSTRAINT skus_pkey PRIMARY KEY (id);


--

-- Name: skus skus_provider_id_provider_sku_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.skus
    ADD CONSTRAINT skus_provider_id_provider_sku_code_key UNIQUE (provider_id, provider_sku_code);


--

-- Name: idx_skus_code; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_skus_code ON public.skus USING btree (code);


--

-- Name: idx_skus_is_active; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_skus_is_active ON public.skus USING btree (is_active);


--

-- Name: idx_skus_product; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_skus_product ON public.skus USING btree (product_id);


--

-- Name: idx_skus_provider; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_skus_provider ON public.skus USING btree (provider_id);


--

-- Name: idx_skus_section; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_skus_section ON public.skus USING btree (section_id);


--

-- Name: skus update_skus_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_skus_updated_at BEFORE UPDATE ON public.skus FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.skus (id, code, provider_sku_code, name, description, image, info, product_id, provider_id, section_id, process_time, is_active, is_featured, badge_text, badge_color, stock_status, total_sold, created_at, updated_at) VALUES
('7715e3b7-254c-4030-9bc8-567b7e6ed3f6', 'ff-50-dm', 'ff50', '50 Diamonds', '50 Diamonds Free Fire', 'https://cdn.gate.co.id/sku/ff-diamond.webp', NULL, 'e16042d2-0a9c-4e23-9ae0-c488a0900d88', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '8970', '2025-12-04 19:24:29.994031+00', '2025-12-04 19:24:29.994031+00'),
('ce6a590f-1761-4af5-bc72-509e1a096032', 'ff-100-dm', 'ff100', '100 Diamonds', '100 Diamonds Free Fire', 'https://cdn.gate.co.id/sku/ff-diamond.webp', NULL, 'e16042d2-0a9c-4e23-9ae0-c488a0900d88', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '11230', '2025-12-04 19:24:29.994031+00', '2025-12-04 19:24:29.994031+00'),
('7364818e-62b8-4694-b3b2-95ca71fb3a4e', 'ff-210-dm', 'ff210', '210 Diamonds', '210 Diamonds Free Fire', 'https://cdn.gate.co.id/sku/ff-diamond.webp', NULL, 'e16042d2-0a9c-4e23-9ae0-c488a0900d88', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 't', 'BEST SELLER', '#FF6B6B', 'AVAILABLE', '15670', '2025-12-04 19:24:29.994031+00', '2025-12-04 19:24:29.994031+00'),
('0fa7a173-59b3-4106-8d4f-67ae95a29df7', 'ff-520-dm', 'ff520', '520 Diamonds', '520 Diamonds Free Fire', 'https://cdn.gate.co.id/sku/ff-diamond.webp', NULL, 'e16042d2-0a9c-4e23-9ae0-c488a0900d88', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '7890', '2025-12-04 19:24:29.994031+00', '2025-12-04 19:24:29.994031+00'),
('7f42fead-0594-425f-8e07-3128e25c4306', 'ff-1060-dm', 'ff1060', '1060 Diamonds', '1060 Diamonds Free Fire', 'https://cdn.gate.co.id/sku/ff-diamond.webp', NULL, 'e16042d2-0a9c-4e23-9ae0-c488a0900d88', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '4560', '2025-12-04 19:24:29.994031+00', '2025-12-04 19:24:29.994031+00'),
('d5ed555e-a3fb-4289-9e1d-b13d44f5b837', 'ff-2180-dm', 'ff2180', '2180 Diamonds', '2180 Diamonds Free Fire', 'https://cdn.gate.co.id/sku/ff-diamond.webp', NULL, 'e16042d2-0a9c-4e23-9ae0-c488a0900d88', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '2340', '2025-12-04 19:24:29.994031+00', '2025-12-04 19:24:29.994031+00'),
('de70d09e-4761-4690-99ef-0130ec456739', 'ff-weekly-member', 'ff-wm', 'Weekly Membership', 'Weekly Membership Free Fire', 'https://cdn.gate.co.id/sku/ff-member.webp', NULL, 'e16042d2-0a9c-4e23-9ae0-c488a0900d88', '1c784573-f4ea-488a-9c25-1f87bb43d447', '838caff3-a279-47dd-a95e-a658425bf1d5', '300', 't', 'f', 'HEMAT', '#45B7D1', 'AVAILABLE', '3450', '2025-12-04 19:24:29.994031+00', '2025-12-04 19:24:29.994031+00'),
('3669a9e8-ef8c-40e9-855f-62c94dc27f3a', 'ff-monthly-member', 'ff-mm', 'Monthly Membership', 'Monthly Membership Free Fire', 'https://cdn.gate.co.id/sku/ff-member.webp', NULL, 'e16042d2-0a9c-4e23-9ae0-c488a0900d88', '1c784573-f4ea-488a-9c25-1f87bb43d447', 'cb63f4d0-2cc6-49a0-9001-885b7744fdc9', '300', 't', 'f', NULL, NULL, 'AVAILABLE', '1890', '2025-12-04 19:24:29.994031+00', '2025-12-04 19:24:29.994031+00'),
('42101729-a511-49d1-a55a-9b05d2fbcd54', 'pubgm-60-uc', 'pubgm60', '60 UC', '60 Unknown Cash PUBG Mobile', 'https://cdn.gate.co.id/sku/pubgm-uc.webp', NULL, 'e2552085-cffd-4a34-ab76-dca9e834fd16', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '5200', '2025-12-04 19:24:32.548749+00', '2025-12-04 19:24:32.548749+00'),
('920ac8c9-e6da-4fed-b708-d778b54b2610', 'pubgm-325-uc', 'pubgm325', '325 UC', '325 Unknown Cash PUBG Mobile', 'https://cdn.gate.co.id/sku/pubgm-uc.webp', NULL, 'e2552085-cffd-4a34-ab76-dca9e834fd16', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '8900', '2025-12-04 19:24:32.548749+00', '2025-12-04 19:24:32.548749+00'),
('3d2aff2a-f970-4c37-8765-73ba111fdb8f', 'pubgm-660-uc', 'pubgm660', '660 UC', '660 Unknown Cash PUBG Mobile', 'https://cdn.gate.co.id/sku/pubgm-uc.webp', NULL, 'e2552085-cffd-4a34-ab76-dca9e834fd16', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 't', 'BEST SELLER', '#FF6B6B', 'AVAILABLE', '12500', '2025-12-04 19:24:32.548749+00', '2025-12-04 19:24:32.548749+00'),
('052a80f9-b3f1-4ba8-8814-8a9130b10f71', 'pubgm-1800-uc', 'pubgm1800', '1800 UC', '1800 Unknown Cash PUBG Mobile', 'https://cdn.gate.co.id/sku/pubgm-uc.webp', NULL, 'e2552085-cffd-4a34-ab76-dca9e834fd16', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '4300', '2025-12-04 19:24:32.548749+00', '2025-12-04 19:24:32.548749+00'),
('fb95d9ac-d736-4575-8af0-aca350aa8a48', 'genshin-60-gc', 'genshin60', '60 Genesis Crystals', '60 Genesis Crystals Genshin Impact', 'https://cdn.gate.co.id/sku/genshin-gc.webp', NULL, 'f7d62818-0275-44f0-af91-d329b3be08a2', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '3400', '2025-12-04 19:24:32.579321+00', '2025-12-04 19:24:32.579321+00'),
('6c1884ec-3814-47c4-a54e-748f8aad3b99', 'genshin-330-gc', 'genshin330', '330 Genesis Crystals', '330 Genesis Crystals Genshin Impact', 'https://cdn.gate.co.id/sku/genshin-gc.webp', NULL, 'f7d62818-0275-44f0-af91-d329b3be08a2', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '5600', '2025-12-04 19:24:32.579321+00', '2025-12-04 19:24:32.579321+00'),
('aee26ec2-d6ba-48a8-a8de-6fa7f33b6ac8', 'genshin-1090-gc', 'genshin1090', '1090 Genesis Crystals', '1090 Genesis Crystals Genshin Impact', 'https://cdn.gate.co.id/sku/genshin-gc.webp', NULL, 'f7d62818-0275-44f0-af91-d329b3be08a2', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 't', 'POPULAR', '#4ECDC4', 'AVAILABLE', '8900', '2025-12-04 19:24:32.579321+00', '2025-12-04 19:24:32.579321+00'),
('214af75b-3870-49d2-9934-d30697453e8e', 'genshin-blessing', 'genshin-welkin', 'Blessing of the Welkin Moon', 'Blessing of the Welkin Moon (30 hari)', 'https://cdn.gate.co.id/sku/genshin-welkin.webp', NULL, 'f7d62818-0275-44f0-af91-d329b3be08a2', '1c784573-f4ea-488a-9c25-1f87bb43d447', 'cb63f4d0-2cc6-49a0-9001-885b7744fdc9', '300', 't', 'f', 'HEMAT', '#45B7D1', 'AVAILABLE', '6700', '2025-12-04 19:24:32.579321+00', '2025-12-04 19:24:32.579321+00'),
('8e1c9af1-88b4-4c3a-aa09-c1882283107f', 'mlbb-344-dm', 'mlbb344', '344 Diamonds', '344 Diamonds (312+32 Bonus)', 'https://gate.nos.jkt-1.neo.id/skus/mlbb-diamond.webp', NULL, 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 't', 'BEST SELLER', '#FF6B6B', 'AVAILABLE', '18920', '2025-12-04 19:24:29.98649+00', '2025-12-08 08:58:33.83239+00'),
('bbab0f19-c722-44f2-8170-ffa5c0a06fb7', 'mlbb-257-dm', 'mlbb257', '257 Diamonds', '257 Diamonds (234+23 Bonus)', 'https://gate.nos.jkt-1.neo.id/skus/mlbb-diamond.webp', NULL, 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '9870', '2025-12-04 19:24:29.98649+00', '2025-12-08 08:56:38.133716+00'),
('4045684c-ff9d-4cfd-a42e-8f39ec6ee1dc', 'mlbb-172-dm', 'mlbb172', '172 Diamonds', '172 Diamonds (156+16 Bonus)', 'https://gate.nos.jkt-1.neo.id/skus/mlbb-diamond.webp', NULL, 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '12350', '2025-12-04 19:24:29.98649+00', '2025-12-08 08:56:22.455385+00'),
('91fb3c49-00c7-45fc-9088-7033a8ff8d9e', 'mlbb-429-dm', 'mlbb429', '429 Diamonds', '429 Diamonds (390+39 Bonus)', 'https://gate.nos.jkt-1.neo.id/skus/mlbb-diamond.webp', NULL, 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '7650', '2025-12-04 19:24:29.98649+00', '2025-12-08 08:57:06.484289+00'),
('9bef3e8a-4b0c-4c28-821a-7f7ceca66fa3', 'mlbb-86-dm', 'mlbb86', '86 Diamonds', '86 Diamonds (78+8 Bonus)', 'https://gate.nos.jkt-1.neo.id/skus/mlbb-diamond.webp', NULL, 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '15420', '2025-12-04 19:24:29.98649+00', '2025-12-08 08:57:35.483628+00'),
('70e46f37-f68c-4fb1-ad1f-e5f8d0999f77', 'mlbb-514-dm', 'mlbb514', '514 Diamonds', '514 Diamonds (468+46 Bonus)', 'https://gate.nos.jkt-1.neo.id/skus/mlbb-diamond.webp', NULL, 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '6540', '2025-12-04 19:24:29.98649+00', '2025-12-08 08:58:43.315092+00'),
('7ae9b750-7c45-43f6-ba4a-88ae9b6f5521', 'mlbb-706-dm', 'mlbb706', '706 Diamonds', '706 Diamonds (642+64 Bonus)', 'https://gate.nos.jkt-1.neo.id/skus/mlbb-diamond.webp', NULL, 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '5430', '2025-12-04 19:24:29.98649+00', '2025-12-08 08:58:48.521791+00'),
('d13c61f3-c16d-45c1-81d2-f8a04cbd81c5', 'mlbb-878-dm', 'mlbb878', '878 Diamonds', '878 Diamonds (798+80 Bonus)', 'https://gate.nos.jkt-1.neo.id/skus/mlbb-diamond.webp', NULL, 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '4320', '2025-12-04 19:24:29.98649+00', '2025-12-08 08:59:00.246513+00'),
('2d5736f5-4ce6-4bb9-a7f9-b81a7e813ecf', 'mlbb-1050-dm', 'mlbb1050', '1050 Diamonds', '1050 Diamonds (955+95 Bonus)', 'https://gate.nos.jkt-1.neo.id/skus/mlbb-diamond.webp', NULL, 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 't', 'POPULAR', '#4ECDC4', 'AVAILABLE', '8760', '2025-12-04 19:24:29.98649+00', '2025-12-08 08:59:04.893185+00'),
('fb0948f9-4509-486f-9174-baae2834a9c3', 'mlbb-2195-dm', 'mlbb2195', '2195 Diamonds', '2195 Diamonds (1996+199 Bonus)', 'https://gate.nos.jkt-1.neo.id/skus/mlbb-diamond.webp', NULL, 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '1c784573-f4ea-488a-9c25-1f87bb43d447', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', '60', 't', 'f', NULL, NULL, 'AVAILABLE', '3210', '2025-12-04 19:24:29.98649+00', '2025-12-08 08:59:09.469591+00'),
('cee4f140-9c5a-4dba-a693-9120cc531ce0', 'mlbb-weekly-pass', 'mlbb-wp', 'Weekly Diamond Pass', 'Weekly Diamond Pass (Total 225 Diamonds)', 'https://gate.nos.jkt-1.neo.id/skus/wdp.webp', NULL, 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '1c784573-f4ea-488a-9c25-1f87bb43d447', '838caff3-a279-47dd-a95e-a658425bf1d5', '300', 't', 'f', 'HEMAT', '#45B7D1', 'AVAILABLE', '2340', '2025-12-04 19:24:29.98649+00', '2025-12-09 12:16:03.922804+00'),
('3d3e0643-bab2-4a66-a765-814edcce9457', 'mlbb-twilight-pass', 'mlbb-tp', 'Twilight Pass', 'Twilight Pass Season Ini', 'https://gate.nos.jkt-1.neo.id/skus/tiwlight.webp', NULL, 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', '1c784573-f4ea-488a-9c25-1f87bb43d447', 'cb63f4d0-2cc6-49a0-9001-885b7744fdc9', '300', 't', 't', 'EXCLUSIVE', '#9B59B6', 'AVAILABLE', '1890', '2025-12-04 19:24:29.98649+00', '2025-12-09 12:16:19.654668+00');