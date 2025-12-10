-- Name: users; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    first_name character varying(100) NOT NULL,
    last_name character varying(100),
    email character varying(255) NOT NULL,
    email_verified_at timestamp with time zone,
    phone_number character varying(20),
    password_hash character varying(255),
    profile_picture character varying(500),
    status public.user_status DEFAULT 'INACTIVE'::public.user_status,
    primary_region public.region_code DEFAULT 'ID'::public.region_code,
    current_region public.region_code DEFAULT 'ID'::public.region_code,
    balance_idr bigint DEFAULT 0,
    balance_myr bigint DEFAULT 0,
    balance_php bigint DEFAULT 0,
    balance_sgd bigint DEFAULT 0,
    balance_thb bigint DEFAULT 0,
    membership_level public.membership_level DEFAULT 'CLASSIC'::public.membership_level,
    total_spent_idr bigint DEFAULT 0,
    mfa_status public.mfa_status DEFAULT 'INACTIVE'::public.mfa_status,
    mfa_secret character varying(255),
    mfa_backup_codes text[],
    google_id character varying(255),
    last_login_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT valid_email CHECK (((email)::text ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'::text))
);


ALTER TABLE public.users OWNER TO gate;

--

-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--

-- Name: users users_google_id_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_google_id_key UNIQUE (google_id);


--

-- Name: users users_phone_number_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_phone_number_key UNIQUE (phone_number);


--

-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--

-- Name: idx_users_created_at; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_users_created_at ON public.users USING btree (created_at DESC);


--

-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--

-- Name: idx_users_google_id; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_users_google_id ON public.users USING btree (google_id);


--

-- Name: idx_users_membership; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_users_membership ON public.users USING btree (membership_level);


--

-- Name: idx_users_phone; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_users_phone ON public.users USING btree (phone_number);


--

-- Name: idx_users_primary_region; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_users_primary_region ON public.users USING btree (primary_region);


--

-- Name: idx_users_status; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_users_status ON public.users USING btree (status);


--

-- Name: users update_users_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.users (id, first_name, last_name, email, email_verified_at, phone_number, password_hash, profile_picture, status, primary_region, current_region, balance_idr, balance_myr, balance_php, balance_sgd, balance_thb, membership_level, total_spent_idr, mfa_status, mfa_secret, mfa_backup_codes, google_id, last_login_at, created_at, updated_at) VALUES
('f5c92b21-fe15-463c-bfef-65fedad2ad9f', 'John', 'Doe', 'john@example.com', '2025-12-04 19:24:30.039239+00', '081234567890', '$2a$12$b.ZlpNyPpbkEwy5.Tz1AvepKtk0PvyIZRnpMIOP66ZxNHIbTC5qb6', NULL, 'ACTIVE', 'ID', 'ID', '500000', '0', '0', '0', '0', 'CLASSIC', '0', 'INACTIVE', NULL, NULL, NULL, NULL, '2025-12-04 19:24:30.039239+00', '2025-12-04 19:24:30.039239+00'),
('fdf2b185-2e2d-4bfb-8e0c-07eac735fc69', 'Jane', 'Smith', 'jane@example.com', '2025-12-04 19:24:30.039239+00', '081234567891', '$2a$12$XGWt6D0sYdkAWiJYpxe7L./tO/wToVV7.gN5wL8Hpu1fiPjPKTure', NULL, 'ACTIVE', 'ID', 'ID', '1500000', '0', '0', '0', '0', 'PRESTIGE', '0', 'INACTIVE', NULL, NULL, NULL, NULL, '2025-12-04 19:24:30.039239+00', '2025-12-04 19:24:30.039239+00'),
('c74b4c08-18c4-483d-b1ad-98b8a30274b4', 'Admin', 'Test', 'admin@example.com', '2025-12-04 19:24:30.039239+00', '081234567892', '$2a$12$ONMbycKlMai/OKUraJbNfuZvjLE0eQXA7IVn3vr8nl5cOhNOB9f.i', NULL, 'ACTIVE', 'ID', 'ID', '0', '0', '0', '0', '0', 'CLASSIC', '0', 'INACTIVE', NULL, NULL, NULL, NULL, '2025-12-04 19:24:30.039239+00', '2025-12-04 19:24:30.039239+00'),
('b2e1f30d-7303-4a45-821a-4038ca86abc4', 'Alex', 'Johnson', 'alex@example.com', '2025-12-04 19:24:32.596825+00', '081234567893', '$2a$12$HwZV02s6IMxBuxHQuZu0FeTjHW6C./2FnRXxWqP460CGJq/0nG/e.', NULL, 'ACTIVE', 'ID', 'ID', '250000', '0', '0', '0', '0', 'CLASSIC', '0', 'INACTIVE', NULL, NULL, NULL, NULL, '2025-12-04 19:24:32.596825+00', '2025-12-04 19:24:32.596825+00'),
('9a2e5c11-e03f-49ad-90e7-de711099c9c8', 'Maria', 'Garcia', 'maria@example.com', '2025-12-04 19:24:32.596825+00', '081234567894', '$2a$12$2VmH7qSDxt.fOvtL7peh9e22SB55cmNstYxUu8PT7XBEBwJSTgNZW', NULL, 'ACTIVE', 'ID', 'ID', '750000', '0', '0', '0', '0', 'CLASSIC', '0', 'INACTIVE', NULL, NULL, NULL, NULL, '2025-12-04 19:24:32.596825+00', '2025-12-04 19:24:32.596825+00'),
('6f672cb5-12ae-4003-8e56-a9ac013f04d5', 'David', 'Chen', 'david@example.com', '2025-12-04 19:24:32.596825+00', '081234567895', '$2a$12$CYtd8i31n4VYF0hXYvU56Od7s8Ec0Xhglx5q5xyKe.pjqNF/vlZFK', NULL, 'ACTIVE', 'ID', 'ID', '2500000', '0', '0', '0', '0', 'ROYAL', '0', 'INACTIVE', NULL, NULL, NULL, NULL, '2025-12-04 19:24:32.596825+00', '2025-12-04 19:24:32.596825+00'),
('047911eb-0465-46a7-aca9-3cb1b7468570', 'Sarah', 'Kim', 'sarah@example.com', '2025-12-04 19:24:32.596825+00', '081234567896', '$2a$12$OVAilAPlzu2Gfq87B4/15.CudWw444hVmuE3BQm8PKjOnKGzQt3RW', NULL, 'SUSPENDED', 'ID', 'ID', '100000', '0', '0', '0', '0', 'CLASSIC', '0', 'INACTIVE', NULL, NULL, NULL, NULL, '2025-12-04 19:24:32.596825+00', '2025-12-04 19:24:32.596825+00'),
('fd605506-7ae8-41cb-8a9c-2da698a3f7bb', 'Michael', 'Brown', 'michael@example.com', '2025-12-04 19:24:32.596825+00', '081234567897', '$2a$12$VjMQEcgS8E0/HewZFq0WxOrx8O46iM8LyKike6zkuLNHMvqeYMgly', NULL, 'ACTIVE', 'MY', 'MY', '0', '0', '0', '0', '0', 'CLASSIC', '0', 'INACTIVE', NULL, NULL, NULL, NULL, '2025-12-04 19:24:32.596825+00', '2025-12-04 19:24:32.596825+00'),
('817d4fb7-8395-40a6-9fc9-34af8bbb1307', 'GERBANG SOLUSI', 'DIGITAL', 'gerbangsolusidigital@gmail.com', '2025-12-04 23:15:47.295421+00', NULL, NULL, 'https://lh3.googleusercontent.com/a/ACg8ocJYsmhvjbfP65xi19jF79nogm3MIsSXMlncW4QRbbD3fgrWZw=s96-c', 'ACTIVE', 'ID', 'ID', '0', '0', '0', '0', '0', 'CLASSIC', '0', 'INACTIVE', NULL, NULL, '108928377848856303040', '2025-12-05 16:18:57.465969+00', '2025-12-04 23:15:47.295421+00', '2025-12-05 16:18:57.465969+00'),
('eeb13535-5caa-4c0c-af7f-2d36a8525c7b', 'RIKO', 'BUDI SAPUTRA', 'ptgerbangteknologidigital@gmail.com', '2025-12-05 16:24:25.443126+00', '+6281234879652', '$2a$12$gP1TMjb2/D.7H4uU.90XEO9aR4cigzCs3SLuPLDHuHI/uSDGmFBfO', NULL, 'ACTIVE', 'ID', 'ID', '0', '0', '0', '0', '0', 'CLASSIC', '0', 'INACTIVE', NULL, NULL, NULL, '2025-12-05 16:25:06.473989+00', '2025-12-04 23:15:24.348245+00', '2025-12-05 16:25:06.473989+00'),
('8f3c189c-50ad-4bb6-bbb7-ffa1660506ef', 'RIKO BUDI', 'SAPUTRA', 'rikobudisaputra14@gmail.com', '2025-12-08 14:28:27.180081+00', NULL, NULL, 'https://lh3.googleusercontent.com/a/ACg8ocLp3hgL3ubV5pUNSJ8sw0WYAK6VE-eL7AZeur5RQxcMuU7tivNs=s96-c', 'ACTIVE', 'ID', 'ID', '0', '0', '0', '0', '0', 'CLASSIC', '0', 'INACTIVE', NULL, NULL, '109738127726894892781', NULL, '2025-12-08 14:28:27.180081+00', '2025-12-08 14:28:27.180081+00');