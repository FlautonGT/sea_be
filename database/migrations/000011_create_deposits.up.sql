-- Name: deposits; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.deposits (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    invoice_number character varying(50) NOT NULL,
    user_id uuid NOT NULL,
    amount bigint NOT NULL,
    payment_fee bigint DEFAULT 0,
    total_amount bigint NOT NULL,
    currency public.currency_code NOT NULL,
    payment_channel_id uuid NOT NULL,
    payment_gateway_id uuid,
    payment_gateway_ref_id character varying(255),
    payment_data jsonb,
    status public.deposit_status DEFAULT 'PENDING'::public.deposit_status,
    balance_before bigint,
    balance_after bigint,
    region public.region_code NOT NULL,
    ip_address inet,
    user_agent text,
    confirmed_by uuid,
    confirmed_at timestamp with time zone,
    cancelled_by uuid,
    cancelled_at timestamp with time zone,
    cancel_reason text,
    paid_at timestamp with time zone,
    expired_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.deposits OWNER TO gate;

--

-- Name: deposits deposits_invoice_number_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.deposits
    ADD CONSTRAINT deposits_invoice_number_key UNIQUE (invoice_number);


--

-- Name: deposits deposits_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.deposits
    ADD CONSTRAINT deposits_pkey PRIMARY KEY (id);


--

-- Name: idx_deposits_created_at; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_deposits_created_at ON public.deposits USING btree (created_at DESC);


--

-- Name: idx_deposits_invoice; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_deposits_invoice ON public.deposits USING btree (invoice_number);


--

-- Name: idx_deposits_status; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_deposits_status ON public.deposits USING btree (status);


--

-- Name: idx_deposits_user; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_deposits_user ON public.deposits USING btree (user_id);


--

-- Name: deposits update_deposits_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_deposits_updated_at BEFORE UPDATE ON public.deposits FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.deposits (id, invoice_number, user_id, amount, payment_fee, total_amount, currency, payment_channel_id, payment_gateway_id, payment_gateway_ref_id, payment_data, status, balance_before, balance_after, region, ip_address, user_agent, confirmed_by, confirmed_at, cancelled_by, cancelled_at, cancel_reason, paid_at, expired_at, created_at, updated_at) VALUES
('b1ee7120-80d5-4af8-910d-2d41b5521701', 'DEP2907B0F9F881DA6D10AA', 'f5c92b21-fe15-463c-bfef-65fedad2ad9f', '500000', '0', '500000', 'IDR', '2162d8a6-906b-4c03-a2eb-c905f1228879', NULL, NULL, NULL, 'SUCCESS', NULL, NULL, 'ID', '103.123.48.104', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2025-12-04 19:24:33.480405+00', '2025-12-04 19:24:33.480405+00'),
('0d014694-b436-41d2-8f34-3d835fa833d9', 'DEPF5EC463FE8B6F2E87E67', 'fdf2b185-2e2d-4bfb-8e0c-07eac735fc69', '100000', '0', '1000000', 'IDR', '2162d8a6-906b-4c03-a2eb-c905f1228879', NULL, NULL, NULL, 'PENDING', NULL, NULL, 'ID', '103.123.78.126', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2025-12-04 19:24:33.480405+00', '2025-12-04 19:24:33.480405+00'),
('693296a1-49f7-4e73-9c8a-ff870b9fb65a', 'DEP7A45A7CE0FE2B64D0C69', '6f672cb5-12ae-4003-8e56-a9ac013f04d5', '50000', '0', '1000000', 'IDR', '2162d8a6-906b-4c03-a2eb-c905f1228879', NULL, NULL, NULL, 'SUCCESS', NULL, NULL, 'ID', '103.123.46.155', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2025-12-04 19:24:33.480405+00', '2025-12-04 19:24:33.480405+00'),
('df1d6777-4fa0-47f6-ab84-077259989920', 'DEP4391FE2AFE585CB59C3E', 'f5c92b21-fe15-463c-bfef-65fedad2ad9f', '100000', '0', '200000', 'IDR', '48604997-4746-47d8-862e-f49efd23bdd7', NULL, NULL, NULL, 'SUCCESS', NULL, NULL, 'ID', '103.123.68.26', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2025-12-04 19:24:33.480405+00', '2025-12-04 19:24:33.480405+00'),
('7ed5cac8-b83a-4cb9-858b-4840ecee49db', 'DEP6B5954526C3FFF7525C9', 'fdf2b185-2e2d-4bfb-8e0c-07eac735fc69', '200000', '0', '100000', 'IDR', '48604997-4746-47d8-862e-f49efd23bdd7', NULL, NULL, NULL, 'SUCCESS', NULL, NULL, 'ID', '103.123.184.238', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2025-12-04 19:24:33.480405+00', '2025-12-04 19:24:33.480405+00'),
('6f7740e3-9231-4f61-875f-177ac80c577f', 'DEP40A6B4682F9949DC84AD', '6f672cb5-12ae-4003-8e56-a9ac013f04d5', '200000', '0', '200000', 'IDR', '48604997-4746-47d8-862e-f49efd23bdd7', NULL, NULL, NULL, 'PENDING', NULL, NULL, 'ID', '103.123.136.220', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2025-12-04 19:24:33.480405+00', '2025-12-04 19:24:33.480405+00'),
('344257f3-73b9-493d-b4bb-3b4811727eb6', 'DEP1F11A2AA521F182F6B1C', 'f5c92b21-fe15-463c-bfef-65fedad2ad9f', '1000000', '0', '200000', 'IDR', 'd5606ef6-ee54-46df-9739-64220fdc1b28', NULL, NULL, NULL, 'SUCCESS', NULL, NULL, 'ID', '103.123.239.128', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2025-12-04 19:24:33.480405+00', '2025-12-04 19:24:33.480405+00'),
('15511e51-c6bf-4022-9013-9542b8cdb859', 'DEP1815B804C1F60F5C6F6D', 'fdf2b185-2e2d-4bfb-8e0c-07eac735fc69', '1000000', '0', '500000', 'IDR', 'd5606ef6-ee54-46df-9739-64220fdc1b28', NULL, NULL, NULL, 'SUCCESS', NULL, NULL, 'ID', '103.123.163.22', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2025-12-04 19:24:33.480405+00', '2025-12-04 19:24:33.480405+00'),
('a7d24567-c463-462f-8e10-8b3d40564680', 'DEPD64E93589D5884F7240D', '6f672cb5-12ae-4003-8e56-a9ac013f04d5', '1000000', '0', '200000', 'IDR', 'd5606ef6-ee54-46df-9739-64220fdc1b28', NULL, NULL, NULL, 'PENDING', NULL, NULL, 'ID', '103.123.54.6', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2025-12-04 19:24:33.480405+00', '2025-12-04 19:24:33.480405+00');