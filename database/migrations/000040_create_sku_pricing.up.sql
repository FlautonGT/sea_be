-- Name: sku_pricing; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.sku_pricing (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    sku_id uuid NOT NULL,
    region_code public.region_code NOT NULL,
    currency public.currency_code NOT NULL,
    buy_price bigint NOT NULL,
    sell_price bigint NOT NULL,
    original_price bigint NOT NULL,
    margin_percentage numeric(5,2) GENERATED ALWAYS AS (
CASE
    WHEN (buy_price > 0) THEN ((((sell_price - buy_price))::numeric / (buy_price)::numeric) * (100)::numeric)
    ELSE (0)::numeric
END) STORED,
    discount_percentage numeric(5,2) GENERATED ALWAYS AS (
CASE
    WHEN (original_price > 0) THEN ((((original_price - sell_price))::numeric / (original_price)::numeric) * (100)::numeric)
    ELSE (0)::numeric
END) STORED,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.sku_pricing OWNER TO gate;

--

-- Name: sku_pricing sku_pricing_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.sku_pricing
    ADD CONSTRAINT sku_pricing_pkey PRIMARY KEY (id);


--

-- Name: sku_pricing sku_pricing_sku_id_region_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.sku_pricing
    ADD CONSTRAINT sku_pricing_sku_id_region_code_key UNIQUE (sku_id, region_code);


--

-- Name: idx_sku_pricing_region; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_sku_pricing_region ON public.sku_pricing USING btree (region_code);


--

-- Name: idx_sku_pricing_sku; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_sku_pricing_sku ON public.sku_pricing USING btree (sku_id);


--

-- Name: sku_pricing update_sku_pricing_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_sku_pricing_updated_at BEFORE UPDATE ON public.sku_pricing FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.sku_pricing (id, sku_id, region_code, currency, buy_price, sell_price, original_price, is_active, created_at, updated_at) VALUES
('e5588eae-a180-4877-a3bb-dfca14f03729', '7715e3b7-254c-4030-9bc8-567b7e6ed3f6', 'ID', 'IDR', '7000', '8000', '10000', 't', '2025-12-04 19:24:30.005943+00', '2025-12-04 19:24:30.005943+00'),
('7fe79960-980e-419d-8b79-9fef45157676', 'ce6a590f-1761-4af5-bc72-509e1a096032', 'ID', 'IDR', '14000', '16000', '20000', 't', '2025-12-04 19:24:30.005943+00', '2025-12-04 19:24:30.005943+00'),
('a59eea90-10e0-4915-8cb7-4654dbcd2979', '7364818e-62b8-4694-b3b2-95ca71fb3a4e', 'ID', 'IDR', '28000', '32000', '40000', 't', '2025-12-04 19:24:30.005943+00', '2025-12-04 19:24:30.005943+00'),
('8da83ee2-315b-48e7-bafe-292da49a6647', '0fa7a173-59b3-4106-8d4f-67ae95a29df7', 'ID', 'IDR', '70000', '79000', '95000', 't', '2025-12-04 19:24:30.005943+00', '2025-12-04 19:24:30.005943+00'),
('4832c4c3-9a85-42cd-9c4a-81312373410d', '7f42fead-0594-425f-8e07-3128e25c4306', 'ID', 'IDR', '140000', '155000', '190000', 't', '2025-12-04 19:24:30.005943+00', '2025-12-04 19:24:30.005943+00'),
('53db6b52-4a53-4ce4-af75-b9b8e366b1d6', 'd5ed555e-a3fb-4289-9e1d-b13d44f5b837', 'ID', 'IDR', '280000', '310000', '380000', 't', '2025-12-04 19:24:30.005943+00', '2025-12-04 19:24:30.005943+00'),
('522c93d9-8503-4dcf-bb9a-3a379cda3416', 'de70d09e-4761-4690-99ef-0130ec456739', 'ID', 'IDR', '18000', '22000', '25000', 't', '2025-12-04 19:24:30.005943+00', '2025-12-04 19:24:30.005943+00'),
('62502f02-f04f-4429-b8fd-58b98ce9d711', '3669a9e8-ef8c-40e9-855f-62c94dc27f3a', 'ID', 'IDR', '85000', '95000', '110000', 't', '2025-12-04 19:24:30.005943+00', '2025-12-04 19:24:30.005943+00'),
('631d0ca1-651d-43f3-b2f3-44baee35da78', '42101729-a511-49d1-a55a-9b05d2fbcd54', 'ID', 'IDR', '15000', '17000', '19000', 't', '2025-12-04 19:24:32.564774+00', '2025-12-04 19:24:32.564774+00'),
('6d0f5a25-1d81-48f2-8426-922e058c0098', '920ac8c9-e6da-4fed-b708-d778b54b2610', 'ID', 'IDR', '75000', '85000', '95000', 't', '2025-12-04 19:24:32.569146+00', '2025-12-04 19:24:32.569146+00'),
('4f7dd25f-a1ec-4506-9ac9-c527ec08b719', '3d2aff2a-f970-4c37-8765-73ba111fdb8f', 'ID', 'IDR', '150000', '165000', '180000', 't', '2025-12-04 19:24:32.572659+00', '2025-12-04 19:24:32.572659+00'),
('3881d720-7280-4e25-8dca-40857905a040', '052a80f9-b3f1-4ba8-8814-8a9130b10f71', 'ID', 'IDR', '400000', '440000', '480000', 't', '2025-12-04 19:24:32.576387+00', '2025-12-04 19:24:32.576387+00'),
('b8ff978d-facf-449c-8183-add5196eda85', 'fb95d9ac-d736-4575-8af0-aca350aa8a48', 'ID', 'IDR', '15000', '17000', '19000', 't', '2025-12-04 19:24:32.583633+00', '2025-12-04 19:24:32.583633+00'),
('05c29f78-4126-49d4-a18c-6ef8d4cc3c34', '6c1884ec-3814-47c4-a54e-748f8aad3b99', 'ID', 'IDR', '80000', '89000', '99000', 't', '2025-12-04 19:24:32.586786+00', '2025-12-04 19:24:32.586786+00'),
('7813ca0b-a1e9-4ab9-bc86-cdd1f301606c', 'aee26ec2-d6ba-48a8-a8de-6fa7f33b6ac8', 'ID', 'IDR', '250000', '275000', '299000', 't', '2025-12-04 19:24:32.59004+00', '2025-12-04 19:24:32.59004+00'),
('02c6b28c-e69c-448d-8dbe-81e638043d3a', '214af75b-3870-49d2-9934-d30697453e8e', 'ID', 'IDR', '75000', '82000', '89000', 't', '2025-12-04 19:24:32.592913+00', '2025-12-04 19:24:32.592913+00'),
('ae0def3c-0ad6-4da3-961a-a94097945aa8', '4045684c-ff9d-4cfd-a42e-8f39ec6ee1dc', 'ID', 'IDR', '37000', '40000', '44000', 't', '2025-12-04 19:24:29.99912+00', '2025-12-08 08:56:22.455385+00'),
('4b39c26a-80d8-46d0-9d95-d5c6796397d6', 'bbab0f19-c722-44f2-8170-ffa5c0a06fb7', 'ID', 'IDR', '55500', '60000', '66000', 't', '2025-12-04 19:24:29.99912+00', '2025-12-08 08:56:38.133716+00'),
('31c6c42d-abd0-4f9c-835b-ea792de2de2b', '91fb3c49-00c7-45fc-9088-7033a8ff8d9e', 'ID', 'IDR', '92500', '100000', '110000', 't', '2025-12-04 19:24:29.99912+00', '2025-12-08 08:57:06.484289+00'),
('c8abaa10-14d9-4485-ba5d-a3f95235d2b6', '9bef3e8a-4b0c-4c28-821a-7f7ceca66fa3', 'ID', 'IDR', '18500', '20000', '22000', 't', '2025-12-04 19:24:29.99912+00', '2025-12-08 08:57:35.483628+00'),
('8f92af0b-669a-4b0c-9a43-8e3092cd0fb8', '8e1c9af1-88b4-4c3a-aa09-c1882283107f', 'ID', 'IDR', '74000', '80000', '88000', 't', '2025-12-04 19:24:29.99912+00', '2025-12-08 08:58:33.83239+00'),
('97281e2e-a624-474b-a262-3dc6c8dc3eda', '70e46f37-f68c-4fb1-ad1f-e5f8d0999f77', 'ID', 'IDR', '111000', '120000', '132000', 't', '2025-12-04 19:24:29.99912+00', '2025-12-08 08:58:43.315092+00'),
('ac1e1bc6-c046-4cad-9b9f-1abef7f90efd', '7ae9b750-7c45-43f6-ba4a-88ae9b6f5521', 'ID', 'IDR', '148000', '160000', '176000', 't', '2025-12-04 19:24:29.99912+00', '2025-12-08 08:58:48.521791+00'),
('cae0c23e-6eb8-4aff-be2a-62132305f763', 'd13c61f3-c16d-45c1-81d2-f8a04cbd81c5', 'ID', 'IDR', '185000', '200000', '220000', 't', '2025-12-04 19:24:29.99912+00', '2025-12-08 08:59:00.246513+00'),
('6844c6bc-66a7-48a1-8401-9dbdfe5fda16', '2d5736f5-4ce6-4bb9-a7f9-b81a7e813ecf', 'ID', 'IDR', '222000', '240000', '264000', 't', '2025-12-04 19:24:29.99912+00', '2025-12-08 08:59:04.893185+00'),
('b8c3b844-bc9c-4dfa-9c13-5d4a4b8cb4b7', 'fb0948f9-4509-486f-9174-baae2834a9c3', 'ID', 'IDR', '463000', '500000', '550000', 't', '2025-12-04 19:24:29.99912+00', '2025-12-08 08:59:09.469591+00'),
('78e4d82f-9b38-4d3c-9310-49fd6d2da350', 'cee4f140-9c5a-4dba-a693-9120cc531ce0', 'ID', 'IDR', '25000', '28000', '30000', 't', '2025-12-04 19:24:29.99912+00', '2025-12-09 12:16:03.922804+00'),
('956697cf-4ba9-4a52-9602-7479ae2429c8', '3d3e0643-bab2-4a66-a765-814edcce9457', 'ID', 'IDR', '115000', '125000', '140000', 't', '2025-12-04 19:24:29.99912+00', '2025-12-09 12:16:19.654668+00');