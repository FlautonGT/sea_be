-- Name: product_fields; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.product_fields (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    product_id uuid NOT NULL,
    name character varying(100) NOT NULL,
    key character varying(50) NOT NULL,
    field_type public.field_type NOT NULL,
    label character varying(200) NOT NULL,
    placeholder character varying(200),
    hint text,
    pattern character varying(255),
    is_required boolean DEFAULT true,
    min_length integer,
    max_length integer,
    options jsonb,
    sort_order integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.product_fields OWNER TO gate;

--

-- Name: product_fields product_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.product_fields
    ADD CONSTRAINT product_fields_pkey PRIMARY KEY (id);


--

-- Name: product_fields product_fields_product_id_key_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.product_fields
    ADD CONSTRAINT product_fields_product_id_key_key UNIQUE (product_id, key);


--


-- SEED DATA --


INSERT INTO public.product_fields (id, product_id, name, key, field_type, label, placeholder, hint, pattern, is_required, min_length, max_length, options, sort_order, created_at, updated_at) VALUES
('5938fd50-5ee5-42b5-9c68-cf8057118e63', 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', 'User ID', 'userId', 'number', 'User ID', 'Contoh: 123456789', 'Buka profile game untuk melihat User ID', '^[0-9]{6,12}$', 't', NULL, NULL, NULL, '1', '2025-12-04 19:24:29.971306+00', '2025-12-04 19:24:29.971306+00'),
('2461f571-4684-4a4d-a237-00895f3b00ca', 'd3fd0451-8ad0-4dd2-99f2-96e451b49a6d', 'Server ID', 'serverId', 'number', 'Server ID', 'Contoh: 1234', 'Terletak di sebelah User ID dalam kurung', '^[0-9]{4,5}$', 't', NULL, NULL, NULL, '2', '2025-12-04 19:24:29.971306+00', '2025-12-04 19:24:29.971306+00'),
('a5fc9e9d-4de1-4e1e-b0bd-4df5f6f13c29', 'e16042d2-0a9c-4e23-9ae0-c488a0900d88', 'User ID', 'userId', 'number', 'ID Free Fire', 'Contoh: 123456789', 'Buka profile untuk melihat ID', NULL, 't', NULL, NULL, NULL, '1', '2025-12-04 19:24:29.974628+00', '2025-12-04 19:24:29.974628+00'),
('36882214-81f4-4235-803a-4423e24dda79', 'e2552085-cffd-4a34-ab76-dca9e834fd16', 'User ID', 'userId', 'number', 'ID PUBG Mobile', 'Contoh: 5123456789', 'Buka profile untuk melihat ID', NULL, 't', NULL, NULL, NULL, '1', '2025-12-04 19:24:29.977774+00', '2025-12-04 19:24:29.977774+00'),
('7bc2c744-f392-41c7-b1ac-3c5ba3c326eb', 'f7d62818-0275-44f0-af91-d329b3be08a2', 'UID', 'uid', 'number', 'UID Genshin', 'Contoh: 812345678', 'Buka menu Paimon untuk melihat UID', NULL, 't', NULL, NULL, NULL, '1', '2025-12-04 19:24:29.980069+00', '2025-12-04 19:24:29.980069+00'),
('1d006f4e-81bc-457d-9911-c1c96c7bcb12', 'f7d62818-0275-44f0-af91-d329b3be08a2', 'Server', 'server', 'select', 'Server', '', 'Pilih server yang sesuai', NULL, 't', NULL, NULL, NULL, '2', '2025-12-04 19:24:29.980069+00', '2025-12-04 19:24:29.980069+00');