-- Name: promos; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.promos (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code character varying(50) NOT NULL,
    title character varying(200) NOT NULL,
    description text,
    note text,
    max_usage integer,
    max_daily_usage integer,
    max_usage_per_id integer DEFAULT 1,
    max_usage_per_device integer DEFAULT 1,
    max_usage_per_ip integer DEFAULT 1,
    min_amount bigint DEFAULT 0,
    max_promo_amount bigint,
    promo_flat bigint DEFAULT 0,
    promo_percentage numeric(5,2) DEFAULT 0,
    days_available text[],
    start_at timestamp with time zone,
    expired_at timestamp with time zone,
    is_active boolean DEFAULT true,
    total_usage integer DEFAULT 0,
    total_discount_given bigint DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.promos OWNER TO gate;

--

-- Name: promos promos_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.promos
    ADD CONSTRAINT promos_code_key UNIQUE (code);


--

-- Name: promos promos_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.promos
    ADD CONSTRAINT promos_pkey PRIMARY KEY (id);


--

-- Name: idx_promos_code; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_promos_code ON public.promos USING btree (code);


--

-- Name: idx_promos_expired_at; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_promos_expired_at ON public.promos USING btree (expired_at);


--

-- Name: idx_promos_is_active; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_promos_is_active ON public.promos USING btree (is_active);


--

-- Name: promos update_promos_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_promos_updated_at BEFORE UPDATE ON public.promos FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.promos (id, code, title, description, note, max_usage, max_daily_usage, max_usage_per_id, max_usage_per_device, max_usage_per_ip, min_amount, max_promo_amount, promo_flat, promo_percentage, days_available, start_at, expired_at, is_active, total_usage, total_discount_given, created_at, updated_at) VALUES
('0136662d-c470-4dba-bfd4-e4f26f40f3d7', 'NEWYEAR25', 'Promo Tahun Baru 2025', 'Diskon 10% untuk semua transaksi', 'Berlaku untuk semua produk', '1000', '100', '1', '1', '1', '50000', '50000', '0', '10.00', NULL, '2025-12-04 19:24:30.028758+00', '2026-01-03 19:24:30.028758+00', 't', '0', '0', '2025-12-04 19:24:30.028758+00', '2025-12-04 19:24:30.028758+00'),
('939fff99-942e-4c04-80f6-d79c8e574dd6', 'MLBB10K', 'Diskon MLBB 10K', 'Potongan Rp10.000 untuk Mobile Legends', 'Khusus produk Mobile Legends', '500', '50', '2', '1', '1', '80000', '10000', '10000', '0.00', NULL, '2025-12-04 19:24:30.028758+00', '2025-12-18 19:24:30.028758+00', 't', '0', '0', '2025-12-04 19:24:30.028758+00', '2025-12-04 19:24:30.028758+00'),
('e3e488e3-83e5-4fef-9703-f78108969401', 'WEEKEND20', 'Weekend Special', 'Diskon 20% untuk transaksi di akhir pekan', 'Berlaku Sabtu-Minggu', '2000', '200', '2', '1', '1', '50000', '30000', '0', '20.00', NULL, '2025-12-04 19:24:33.489468+00', '2026-01-03 19:24:33.489468+00', 't', '0', '0', '2025-12-04 19:24:33.489468+00', '2025-12-04 19:24:33.489468+00'),
('4e0f4be6-73a6-4b81-bfe1-c039dcacd085', 'FLAT5K', 'Flat 5K Off', 'Potongan langsung Rp5.000', 'Minimal transaksi Rp25.000', '5000', '500', '3', '1', '1', '25000', '5000', '5000', '0.00', NULL, '2025-12-04 19:24:33.489468+00', '2026-02-02 19:24:33.489468+00', 't', '0', '0', '2025-12-04 19:24:33.489468+00', '2025-12-04 19:24:33.489468+00'),
('b7a452b5-e5bf-454e-88f9-a2b366b70c2f', 'VIP25', 'VIP Member 25%', 'Diskon khusus member VIP', 'Khusus member Prestige/Royal', '1000', '100', '5', '1', '1', '100000', '100000', '0', '25.00', NULL, '2025-12-04 19:24:33.489468+00', '2026-03-04 19:24:33.489468+00', 't', '0', '0', '2025-12-04 19:24:33.489468+00', '2025-12-04 19:24:33.489468+00'),
('7b546b04-c6d5-4e70-8d73-98e205cf8738', 'DANA1212', 'Dana X Seaply 1212', 'Dana', '', '1', '0', '1', '1', '1', '50000', '5000', '5000', '0.00', '{MON,TUE,WED,FRI,SUN,SAT,THU}', '2025-12-08 08:54:00+00', '2025-12-16 08:54:00+00', 't', '0', '0', '2025-12-09 08:55:10.985156+00', '2025-12-09 08:55:10.985156+00');