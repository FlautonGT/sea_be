-- Name: admins; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.admins (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(200) NOT NULL,
    email character varying(255) NOT NULL,
    phone_number character varying(20),
    password_hash character varying(255) NOT NULL,
    role_id uuid NOT NULL,
    status public.user_status DEFAULT 'ACTIVE'::public.user_status,
    mfa_enabled boolean DEFAULT false,
    mfa_secret character varying(255),
    mfa_backup_codes text[],
    last_login_at timestamp with time zone,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.admins OWNER TO gate;

--

-- Name: admins admins_email_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.admins
    ADD CONSTRAINT admins_email_key UNIQUE (email);


--

-- Name: admins admins_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.admins
    ADD CONSTRAINT admins_pkey PRIMARY KEY (id);


--

-- Name: idx_admins_email; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_admins_email ON public.admins USING btree (email);


--

-- Name: idx_admins_role; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_admins_role ON public.admins USING btree (role_id);


--

-- Name: idx_admins_status; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_admins_status ON public.admins USING btree (status);


--

-- Name: admins update_admins_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_admins_updated_at BEFORE UPDATE ON public.admins FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.admins (id, name, email, phone_number, password_hash, role_id, status, mfa_enabled, mfa_secret, mfa_backup_codes, last_login_at, created_by, created_at, updated_at) VALUES
('7351b34c-8e5a-4439-909a-a12fcb54e85a', 'Admin Gate', 'admin@gate.co.id', NULL, '$2a$12$3hLryJJCSRDwq9UZ67VvE.vSvwSBFFtlWOq4C.4bFQ2Ez8s85jEaO', '1fe234c5-dc9e-49e0-9cf6-72f9f61fd594', 'ACTIVE', 'f', NULL, NULL, NULL, NULL, '2025-12-04 19:24:30.566674+00', '2025-12-04 19:24:30.566674+00'),
('5705b094-63b7-434e-8a0d-49d9b1a93e84', 'Finance Gate', 'finance@gate.co.id', NULL, '$2a$12$HYFYJnnh31WY/i4sWIIRJ.XcXhCEKgny/jJLglk99ZOPseaEqewoS', 'b82f4665-465f-499d-bf40-2728e0bdd8f2', 'ACTIVE', 'f', NULL, NULL, NULL, NULL, '2025-12-04 19:24:30.743186+00', '2025-12-04 19:24:30.743186+00'),
('19478bfa-994c-42dc-af7a-058717d43d4f', 'CS Lead Gate', 'cslead@gate.co.id', NULL, '$2a$12$dQkeOCTgwBssRwNHFI2J4OZnLbEASC2Z10WInOgVxIasN85w2bz7C', '017ecc3f-9b41-4293-9d70-5f48e2ee3144', 'ACTIVE', 'f', NULL, NULL, NULL, NULL, '2025-12-04 19:24:30.919256+00', '2025-12-04 19:24:30.919256+00'),
('7c0df3dd-9826-4396-80e6-01312d70372a', 'CS Gate', 'cs@gate.co.id', NULL, '$2a$12$SUBro98KwpJhAWN9lxB9v.wScbQMB82UzJYJdIJrzrJJ.zgMi245.', '24839bab-1760-45e5-9ff7-b557d0fc6796', 'ACTIVE', 'f', NULL, NULL, NULL, NULL, '2025-12-04 19:24:31.094702+00', '2025-12-04 19:24:31.094702+00'),
('a7466525-cf91-4890-b38a-9ff0d48d4b87', 'Super Admin', 'superadmin@gate.co.id', NULL, '$2a$12$CK.zZ4h5K4T3EOWuwKbaZOC8.c7Xlyp5To0hK17lEnDiBS7N3fZlu', '9de4b9f5-d986-4dee-8e73-16e82f2fbe2e', 'ACTIVE', 'f', NULL, NULL, '2025-12-09 20:02:37.038159+00', NULL, '2025-12-04 19:24:29.752689+00', '2025-12-09 20:02:37.038159+00');