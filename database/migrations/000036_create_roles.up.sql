-- Name: roles; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.roles (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code public.admin_role NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    level integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.roles OWNER TO gate;

--

-- Name: roles roles_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_code_key UNIQUE (code);


--

-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--

-- Name: roles update_roles_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON public.roles FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.roles (id, code, name, description, level, created_at, updated_at) VALUES
('9de4b9f5-d986-4dee-8e73-16e82f2fbe2e', 'SUPERADMIN', 'Super Administrator', 'Full system access, manage admins & permissions', '1', '2025-12-04 19:24:29.727553+00', '2025-12-04 19:24:29.727553+00'),
('1fe234c5-dc9e-49e0-9cf6-72f9f61fd594', 'ADMIN', 'Administrator', 'Manage products, SKUs, promos, content', '2', '2025-12-04 19:24:29.727553+00', '2025-12-04 19:24:29.727553+00'),
('b82f4665-465f-499d-bf40-2728e0bdd8f2', 'FINANCE', 'Finance', 'View transactions, reports, manage deposits', '3', '2025-12-04 19:24:29.727553+00', '2025-12-04 19:24:29.727553+00'),
('017ecc3f-9b41-4293-9d70-5f48e2ee3144', 'CS_LEAD', 'CS Lead', 'Handle escalations, manage CS team', '4', '2025-12-04 19:24:29.727553+00', '2025-12-04 19:24:29.727553+00'),
('24839bab-1760-45e5-9ff7-b557d0fc6796', 'CS', 'Customer Service', 'View transactions, handle user issues', '5', '2025-12-04 19:24:29.727553+00', '2025-12-04 19:24:29.727553+00');