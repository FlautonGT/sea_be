-- Name: settings; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.settings (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    category character varying(50) NOT NULL,
    key character varying(100) NOT NULL,
    value jsonb NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.settings OWNER TO gate;

--

-- Name: settings settings_category_key_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.settings
    ADD CONSTRAINT settings_category_key_key UNIQUE (category, key);


--

-- Name: settings settings_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.settings
    ADD CONSTRAINT settings_pkey PRIMARY KEY (id);


--

-- Name: settings update_settings_updated_at; Type: TRIGGER; Schema: public; Owner: seaply
--

CREATE TRIGGER update_settings_updated_at BEFORE UPDATE ON public.settings FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--


-- SEED DATA --


INSERT INTO public.settings (id, category, key, value, description, created_at, updated_at) VALUES
('05027e12-b0e5-4e4b-b1b7-3b427eb5e51a', 'general', 'siteName', '"Gate.co.id"', 'Site name', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00'),
('a1c006fd-48e3-42fc-b569-74c0a4b7741f', 'general', 'siteDescription', '"Top Up Game & Voucher Digital Terpercaya"', 'Site description', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00'),
('ebf18ac1-956f-4432-9ced-ddb9e962f041', 'general', 'maintenanceMode', 'false', 'Enable maintenance mode', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00'),
('6920b4e5-ffa3-4e88-bb8d-dceba82ac9dc', 'general', 'maintenanceMessage', 'null', 'Maintenance message', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00'),
('6108019a-a8a1-4177-8e09-6ce75a872b1c', 'transaction', 'orderExpiry', '3600', 'Order expiry time in seconds', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00'),
('ea08902c-0ee6-4bf1-a0d4-85bab6f225c9', 'transaction', 'autoRefundOnFail', 'true', 'Auto refund on failed transaction', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00'),
('4d96a24c-b469-454d-8e49-530d95893ceb', 'transaction', 'maxRetryAttempts', '3', 'Max retry attempts for provider', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00'),
('c29e8ca8-2f82-4ea8-89ea-c5e6aa645d13', 'notification', 'emailEnabled', 'true', 'Enable email notifications', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00'),
('f8484074-2671-4d11-b478-2719883cde93', 'notification', 'whatsappEnabled', 'true', 'Enable WhatsApp notifications', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00'),
('a919852a-e396-4003-b5d0-fa5d7ca6f1bd', 'notification', 'telegramEnabled', 'false', 'Enable Telegram notifications', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00'),
('80a3fa12-6652-49ab-b35b-919d82f54011', 'security', 'maxLoginAttempts', '5', 'Max login attempts before lockout', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00'),
('100b14d5-d5e0-4a26-99d2-918673824b4e', 'security', 'lockoutDuration', '900', 'Lockout duration in seconds', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00'),
('dce5ab3d-cac1-4e8a-8e7a-7870a3bda24b', 'security', 'sessionTimeout', '3600', 'Session timeout in seconds', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00'),
('cfcba510-c3eb-46bd-aee3-42c5dfb9993e', 'security', 'mfaRequired', 'true', 'Require MFA for admin', '2025-12-04 19:24:29.958357+00', '2025-12-04 19:24:29.958357+00');