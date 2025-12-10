-- Name: promo_usages; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.promo_usages (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    promo_id uuid NOT NULL,
    user_id uuid,
    transaction_id uuid,
    device_id character varying(255),
    ip_address inet,
    discount_amount bigint NOT NULL,
    used_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.promo_usages OWNER TO gate;

--

-- Name: promo_usages promo_usages_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.promo_usages
    ADD CONSTRAINT promo_usages_pkey PRIMARY KEY (id);


--

-- Name: idx_promo_usages_promo; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_promo_usages_promo ON public.promo_usages USING btree (promo_id);


--

-- Name: idx_promo_usages_used_at; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_promo_usages_used_at ON public.promo_usages USING btree (used_at DESC);


--

-- Name: idx_promo_usages_user; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_promo_usages_user ON public.promo_usages USING btree (user_id);


--


-- SEED DATA --


INSERT INTO public.promo_usages (id, promo_id, user_id, transaction_id, device_id, ip_address, discount_amount, used_at) VALUES
('30d4d30d-17f4-4eb0-adf8-8fb408a208e6', '0136662d-c470-4dba-bfd4-e4f26f40f3d7', 'f5c92b21-fe15-463c-bfef-65fedad2ad9f', 'e98b3360-5d42-4348-9f64-946215c17656', NULL, NULL, '5000', '2025-12-04 19:24:33.464171+00'),
('9451a30a-322b-4917-bcde-78f7b8759a9c', '0136662d-c470-4dba-bfd4-e4f26f40f3d7', 'f5c92b21-fe15-463c-bfef-65fedad2ad9f', 'd6b4faef-6576-4c5c-85d1-9a97f77bf0b8', NULL, NULL, '5000', '2025-12-04 19:24:33.464171+00'),
('bf70b01d-ccab-4e4e-8eb7-0eb63bb56400', '0136662d-c470-4dba-bfd4-e4f26f40f3d7', 'f5c92b21-fe15-463c-bfef-65fedad2ad9f', 'fe1513ef-5e20-4772-bd98-cf17dae1398f', NULL, NULL, '5000', '2025-12-04 19:24:33.464171+00'),
('605729f8-7b3b-44c2-8ad0-b6c87c98c317', '0136662d-c470-4dba-bfd4-e4f26f40f3d7', 'fdf2b185-2e2d-4bfb-8e0c-07eac735fc69', '66574332-bb15-4110-9b99-c8de4c17dc94', NULL, NULL, '5000', '2025-12-04 19:24:33.464171+00'),
('c422bda4-392a-4750-a4d1-95f02f1c6f4d', '0136662d-c470-4dba-bfd4-e4f26f40f3d7', 'fdf2b185-2e2d-4bfb-8e0c-07eac735fc69', '349ea4ad-3ddb-4e38-a541-8f1c47712ad0', NULL, NULL, '5000', '2025-12-04 19:24:33.464171+00'),
('b9bffb0e-5fbd-4486-a87a-8131acdd067f', '0136662d-c470-4dba-bfd4-e4f26f40f3d7', 'fdf2b185-2e2d-4bfb-8e0c-07eac735fc69', 'd9b2d0c9-80b2-4252-b5fb-7c5e9a2aa493', NULL, NULL, '5000', '2025-12-04 19:24:33.464171+00'),
('9c435e1d-076e-4f70-8b3b-32ef696d26a1', '0136662d-c470-4dba-bfd4-e4f26f40f3d7', 'b2e1f30d-7303-4a45-821a-4038ca86abc4', 'ca8fc72c-0690-41c5-9c21-dfffd37584dc', NULL, NULL, '5000', '2025-12-04 19:24:33.464171+00'),
('a5f7dbf7-1934-40b9-a54a-1d7872fc0510', '0136662d-c470-4dba-bfd4-e4f26f40f3d7', 'b2e1f30d-7303-4a45-821a-4038ca86abc4', '210c05d7-cd7d-4ecf-a814-34462e3cb2bb', NULL, NULL, '5000', '2025-12-04 19:24:33.464171+00'),
('df3093e1-dd41-4221-926b-ea02930d49e5', '0136662d-c470-4dba-bfd4-e4f26f40f3d7', 'b2e1f30d-7303-4a45-821a-4038ca86abc4', '36a92a6e-1fa8-4d86-8636-4aba528fcf96', NULL, NULL, '5000', '2025-12-04 19:24:33.464171+00'),
('af292bb6-3c74-4ca4-98c8-dd32f1e4e8d1', '0136662d-c470-4dba-bfd4-e4f26f40f3d7', 'f5c92b21-fe15-463c-bfef-65fedad2ad9f', '815716dd-5a6e-4e95-802c-b745d69fdb0e', NULL, NULL, '5000', '2025-12-04 19:24:33.464171+00');