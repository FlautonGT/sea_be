-- Name: payment_channels; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.payment_channels (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code character varying(50) NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    image character varying(500),
    category_id uuid,
    fee_type public.fee_type DEFAULT 'PERCENTAGE'::public.fee_type,
    fee_amount bigint DEFAULT 0,
    fee_percentage numeric(5,2) DEFAULT 0,
    min_fee bigint DEFAULT 0,
    max_fee bigint DEFAULT 0,
    min_amount bigint DEFAULT 0,
    max_amount bigint DEFAULT 0,
    supported_types public.payment_type[] DEFAULT ARRAY['purchase'::public.payment_type, 'deposit'::public.payment_type],
    instruction text,
    is_active boolean DEFAULT true,
    is_featured boolean DEFAULT false,
    sort_order integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    gateway_code character varying(50) DEFAULT ''::character varying
);


ALTER TABLE public.payment_channels OWNER TO gate;

--

-- Name: payment_channels payment_channels_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.payment_channels
    ADD CONSTRAINT payment_channels_code_key UNIQUE (code);


--

-- Name: payment_channels payment_channels_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.payment_channels
    ADD CONSTRAINT payment_channels_pkey PRIMARY KEY (id);


--

-- Name: payment_channels update_payment_channels_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_payment_channels_updated_at BEFORE UPDATE ON public.payment_channels FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.payment_channels (id, code, name, description, image, category_id, fee_type, fee_amount, fee_percentage, min_fee, max_fee, min_amount, max_amount, supported_types, instruction, is_active, is_featured, sort_order, created_at, updated_at, gateway_code) VALUES
('2162d8a6-906b-4c03-a2eb-c905f1228879', 'QRIS', 'QRIS', 'Bayar dengan scan QRIS', 'https://gate.nos.jkt-1.neo.id/payment/073ba2eb-6678-4f3e-8c95-60e6f5ccbcfc.png', NULL, 'PERCENTAGE', '0', '1.00', '0', '0', '1000', '10000000', '{purchase,deposit}', 'Scan QR code menggunakan aplikasi e-wallet favoritmu', 't', 't', '1', '2025-12-04 19:24:30.009002+00', '2025-12-09 09:45:44.228178+00', 'QRIS'),
('a45b214c-cd62-4be9-862c-83c034ddf4f2', 'SHOPEEPAY', 'ShopeePay', 'Bayar dengan ShopeePay', 'https://gate.nos.jkt-1.neo.id/payment/90ed2c3e-748d-4c30-930b-86e8095620ad.webp', '25b0365c-8eba-4daa-b8a7-36cf9ccf3c66', 'PERCENTAGE', '0', '1.50', '0', '0', '998', '10000000', '{purchase}', 'Kamu akan diarahkan ke aplikasi Shopee untuk menyelesaikan pembayaran', 't', 'f', '4', '2025-12-04 19:24:30.009002+00', '2025-12-09 09:19:02.83773+00', 'SHOPEEPAY'),
('d5606ef6-ee54-46df-9739-64220fdc1b28', 'BRI_VA', 'BRI Virtual Account', 'Transfer via BRI', 'https://gate.nos.jkt-1.neo.id/payment/917902fb-29d1-4eb0-b807-034b48dc8086.png', '695ccf5c-dcc1-414b-be4c-ccc1338357bd', 'FIXED', '4000', '0.00', '0', '0', '10000', '100000000', '{purchase,deposit}', 'Transfer ke nomor Virtual Account yang akan diberikan', 't', 'f', '2', '2025-12-04 19:24:30.009002+00', '2025-12-09 11:19:49.50656+00', '002'),
('48604997-4746-47d8-862e-f49efd23bdd7', 'BCA_VA', 'BCA Virtual Account', 'Transfer via BCA', 'https://gate.nos.jkt-1.neo.id/payment/c55e5dd0-1aa5-4248-93d3-78fb4ad70a0c.png', '695ccf5c-dcc1-414b-be4c-ccc1338357bd', 'FIXED', '4000', '0.00', '0', '0', '10000', '100000000', '{purchase,deposit}', 'Transfer ke nomor Virtual Account yang akan diberikan', 't', 'f', '1', '2025-12-04 19:24:30.009002+00', '2025-12-09 11:22:49.376492+00', '014'),
('9457bd7d-216c-4ae4-a654-02272dec87d8', 'BALANCE', 'Saldo Gate', 'Bayar dengan saldo Gate', 'https://gate.nos.jkt-1.neo.id/payment/d7581e04-7cdd-414d-9ef8-907191bdb37f.png', NULL, 'FIXED', '0', '0.00', '0', '0', '1000', '100000000', '{purchase}', 'Pembayaran langsung menggunakan saldo Gate kamu', 't', 't', '0', '2025-12-04 19:24:30.009002+00', '2025-12-08 13:04:35.357204+00', ''),
('c708c6d0-844d-47ac-8475-fb2f3b81b196', 'DANA', 'DANA', 'Bayar dengan DANA', 'https://gate.nos.jkt-1.neo.id/payment/174a7df2-ac2d-4481-aac3-0c34721050b8.webp', '25b0365c-8eba-4daa-b8a7-36cf9ccf3c66', 'PERCENTAGE', '0', '3.00', '0', '0', '1000', '10000000', '{purchase,deposit}', 'Kamu akan diarahkan ke aplikasi DANA untuk menyelesaikan pembayaran', 't', 'f', '2', '2025-12-04 19:24:30.009002+00', '2025-12-09 11:23:25.086224+00', 'DANA'),
('0d28ee79-9767-4fbf-ad0b-732fb42a9e05', 'GOPAY', 'GoPay', 'Bayar dengan GoPay', 'https://gate.nos.jkt-1.neo.id/payment/c66014e0-3aa9-4a6e-b0ce-9e016d970ca9.png', '25b0365c-8eba-4daa-b8a7-36cf9ccf3c66', 'PERCENTAGE', '0', '3.00', '0', '0', '10000', '10000000', '{purchase}', 'Kamu akan diarahkan ke aplikasi Gojek untuk menyelesaikan pembayaran', 't', 'f', '3', '2025-12-04 19:24:30.009002+00', '2025-12-09 11:23:34.991014+00', 'GOPAY'),
('6b754f8e-c124-4f00-9776-36d9c73fde9e', 'ALFAMART', 'Alfamart', 'Bayar menggunakan Retail Alfamart', 'https://gate.nos.jkt-1.neo.id/payment/95e4ce44-db8c-4338-a8a9-3c1349dc04fa.png', 'a2e9f28d-d545-4d77-a29c-e28f9908914e', 'FIXED', '6000', '0.00', '0', '0', '10000', '5000000', '{purchase}', 'p', 't', 't', '1', '2025-12-08 19:04:27.679511+00', '2025-12-09 09:19:02.923162+00', 'ALFAMART'),
('f72f575c-4cd8-42fa-b1a3-6bf78bfd6a0c', 'INDOMARET', 'Indomaret', 'Bayar menggunakan Retail Indomaret', 'https://gate.nos.jkt-1.neo.id/payment/d577f0b5-75b3-4b34-ad69-c5c0e0630ff6.png', 'a2e9f28d-d545-4d77-a29c-e28f9908914e', 'FIXED', '6000', '0.00', '0', '0', '10000', '5000000', '{purchase}', '', 't', 'f', '2', '2025-12-08 19:05:08.691338+00', '2025-12-09 09:19:04.00439+00', 'INDOMARET'),
('a7251199-b644-4a34-be74-298819a2841e', 'PERMATA_VA', 'Permata Virtual Account', 'Transfer via Permata', 'https://gate.nos.jkt-1.neo.id/payment/70eba21d-644c-426a-b591-3fda41a1e22c.svg', '695ccf5c-dcc1-414b-be4c-ccc1338357bd', 'FIXED', '4000', '0.00', '0', '0', '10000', '100000000', '{purchase,deposit}', 'Transfer ke nomor Virtual Account yang akan diberikan', 't', 'f', '4', '2025-12-04 19:24:30.009002+00', '2025-12-09 11:25:09.563524+00', '013'),
('cb1707a4-285f-4260-9d4b-5908be50d087', 'MANDIRI_VA', 'Mandiri Virtual Account', 'Transfer via Mandiri', 'https://gate.nos.jkt-1.neo.id/payment/631c1fda-f613-4b6b-8312-880d96bff1a9.png', '695ccf5c-dcc1-414b-be4c-ccc1338357bd', 'FIXED', '4000', '0.00', '0', '0', '10000', '100000000', '{purchase,deposit}', 'Transfer ke nomor Virtual Account yang akan diberikan', 't', 'f', '3', '2025-12-04 19:24:30.009002+00', '2025-12-09 11:25:19.801298+00', '008');