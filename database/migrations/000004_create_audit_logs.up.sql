-- Name: audit_logs; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.audit_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    admin_id uuid,
    admin_name character varying(200),
    admin_email character varying(255),
    action public.audit_action NOT NULL,
    resource character varying(100) NOT NULL,
    resource_id uuid,
    description text,
    changes jsonb,
    ip_address inet,
    user_agent text,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.audit_logs OWNER TO gate;

--

-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--

-- Name: idx_audit_logs_action; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_audit_logs_action ON public.audit_logs USING btree (action);


--

-- Name: idx_audit_logs_admin; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_audit_logs_admin ON public.audit_logs USING btree (admin_id);


--

-- Name: idx_audit_logs_created_at; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_audit_logs_created_at ON public.audit_logs USING btree (created_at DESC);


--

-- Name: idx_audit_logs_resource; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_audit_logs_resource ON public.audit_logs USING btree (resource);


--


-- SEED DATA --


INSERT INTO public.audit_logs (id, admin_id, admin_name, admin_email, action, resource, resource_id, description, changes, ip_address, user_agent, created_at) VALUES
('2a92d3ca-45bc-4a3e-b132-2671218e4183', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'POPUP', '7ad33173-7ce3-4b81-995c-5196bf43fe4e', 'Updated popup for region ID', NULL, NULL, NULL, '2025-12-04 20:29:44.313379+00'),
('fc96eeb1-9849-4847-aac4-987c7d42a070', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'POPUP', 'bc6f38af-b748-4ca1-8207-3a8b1dec48d1', 'Updated popup for region MY', NULL, NULL, NULL, '2025-12-04 20:30:14.294598+00'),
('c4fae96c-0d5b-4509-92f5-e652f93233e2', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'POPUP', '7ad33173-7ce3-4b81-995c-5196bf43fe4e', 'Updated popup for region ID', NULL, NULL, NULL, '2025-12-04 20:31:20.46774+00'),
('fe15e2e2-ae43-4769-bf62-1beb2697b5e2', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'POPUP', '7ad33173-7ce3-4b81-995c-5196bf43fe4e', 'Updated popup for region ID', NULL, NULL, NULL, '2025-12-04 20:31:46.447075+00'),
('7c601052-1fa1-4fad-89e7-e1093a750c45', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'POPUP', 'bc6f38af-b748-4ca1-8207-3a8b1dec48d1', 'Updated popup for region MY', NULL, NULL, NULL, '2025-12-04 20:33:47.857842+00'),
('9993233e-4406-4b7c-aab0-0a50031033c8', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'POPUP', '24ede188-7049-4ab6-bbbb-a20981e8f843', 'Updated popup for region PH', NULL, NULL, NULL, '2025-12-04 20:34:45.319228+00'),
('d82d0c44-a02f-480a-9a86-c53eb9dab14a', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'POPUP', 'f30e88a1-8569-4242-9dbd-d9d2285ab5bd', 'Updated popup for region SG', NULL, NULL, NULL, '2025-12-04 20:35:31.332675+00'),
('fe62517a-dfad-47a1-9d33-72e908d6da4a', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'POPUP', '1fcdd188-e7d2-47a3-b43e-002a36da7ae9', 'Updated popup for region TH', NULL, NULL, NULL, '2025-12-04 20:37:03.329616+00'),
('ac124548-290e-4495-af0a-b16724d17060', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'BANNER', '85f97a4f-6748-4c05-b6d3-c5c0af2ebfb6', 'Updated banner', NULL, NULL, NULL, '2025-12-04 20:46:34.868499+00'),
('5a9900fc-f90b-4463-8032-bbba2f73e4ef', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'BANNER', '7ef7d361-e377-4e8b-9585-82b64a82937e', 'Updated banner', NULL, NULL, NULL, '2025-12-04 20:46:46.740801+00'),
('71af4ab1-bdc9-4676-a534-2bd07853b58d', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'BANNER', '576bcbdb-ec68-4461-b425-90e8a529f1f2', 'Updated banner', NULL, NULL, NULL, '2025-12-04 20:46:55.877638+00'),
('7d86687e-3c0d-4edf-9174-602941a7f7de', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'BANNER', '85f97a4f-6748-4c05-b6d3-c5c0af2ebfb6', 'Updated banner', NULL, NULL, NULL, '2025-12-04 20:47:01.042633+00'),
('d3811922-dba4-4b7b-8a4d-639e9f3f3fe6', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'BANNER', '7ef7d361-e377-4e8b-9585-82b64a82937e', 'Updated banner', NULL, NULL, NULL, '2025-12-04 20:47:16.26755+00'),
('82a08e3b-2cdc-4009-b9c4-1022f5c25bdb', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'REGION', '0c506105-70a5-4af9-a649-61632926f982', 'Updated region', NULL, NULL, NULL, '2025-12-04 21:23:21.582553+00'),
('fcc00f76-8007-4e3f-9b4b-9c38fcb24acb', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'REGION', '23ec7a7d-3212-4e43-95c1-c0b52c01d712', 'Updated region', NULL, NULL, NULL, '2025-12-04 21:23:34.57551+00'),
('eadef4e9-7e29-492c-9609-3c963b3bbfc4', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'REGION', 'abfd2347-b943-4fb3-8c24-10e8f28a8cce', 'Updated region', NULL, NULL, NULL, '2025-12-04 21:23:40.038638+00'),
('65dc065a-a59a-4a3b-a923-970e97153e54', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'REGION', 'f37ab772-2188-4822-8677-d773e1aa5c64', 'Updated region', NULL, NULL, NULL, '2025-12-04 21:23:45.302146+00'),
('98e49fbf-b372-44fa-ac60-8af521177512', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'REGION', '6cb4eb90-c32b-4600-a2e0-aefea711303e', 'Updated region', NULL, NULL, NULL, '2025-12-04 21:23:53.757225+00'),
('0c96dd21-1f10-48c5-954c-81ce0afedf80', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'LANGUAGE', '06733d0e-d494-444c-b40b-b245549dfca7', 'Updated language', NULL, NULL, NULL, '2025-12-04 21:41:38.629544+00'),
('25892d41-2573-4735-a634-c50ad90c0d2b', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'LANGUAGE', '2f4eebaf-5321-4e39-8d34-a208cbf6306c', 'Updated language', NULL, NULL, NULL, '2025-12-04 21:41:44.207015+00'),
('2598a2ff-c49c-4696-af19-a51ff36a6653', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'DELETE', 'BANNER', '576bcbdb-ec68-4461-b425-90e8a529f1f2', 'Deleted banner', NULL, NULL, NULL, '2025-12-04 21:48:16.884257+00'),
('41aee244-b51d-418b-aeb6-315c2302741a', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'CATEGORY', 'a9341837-8c47-48ed-be60-c2baee160c1c', 'Updated category', NULL, NULL, NULL, '2025-12-08 09:01:20.599328+00'),
('672974c8-da58-4197-b02f-b92273fa34d4', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'CATEGORY', 'a9341837-8c47-48ed-be60-c2baee160c1c', 'Updated category', NULL, NULL, NULL, '2025-12-08 09:01:25.252175+00'),
('b04b7b81-f695-4cb3-8a7e-8f20efa1873c', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'CATEGORY', 'a9341837-8c47-48ed-be60-c2baee160c1c', 'Updated category', NULL, NULL, NULL, '2025-12-08 09:01:32.643201+00'),
('eab0f3cc-b988-41d4-b2a1-e4cc53daa544', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'CATEGORY', 'a9341837-8c47-48ed-be60-c2baee160c1c', 'Updated category', NULL, NULL, NULL, '2025-12-08 09:02:07.857974+00'),
('1115d895-7733-493c-9598-84461c3209e6', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'DELETE', 'CATEGORY', '79ff9ade-1337-4786-929f-a5352b98c3ab', 'Deleted category', NULL, NULL, NULL, '2025-12-08 09:02:23.991218+00'),
('0d94b35a-62e0-460c-8a20-ea01689685fe', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'DELETE', 'CATEGORY', '522caeb9-f8bf-468a-bd9c-7b37f3c521f4', 'Deleted category', NULL, NULL, NULL, '2025-12-08 09:02:52.131689+00'),
('62450ec1-cd6e-4979-b676-5b448380f671', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'CATEGORY', '5b7dcf51-84fb-4900-baa8-387672e0f0d2', 'Updated category', NULL, NULL, NULL, '2025-12-08 09:03:15.366158+00'),
('726456dc-3eab-4ba5-8137-d23ed31f4fb7', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'CATEGORY', 'a9341837-8c47-48ed-be60-c2baee160c1c', 'Updated category', NULL, NULL, NULL, '2025-12-08 09:03:51.422751+00'),
('af4cab4c-79b0-4f81-a3da-59504f904b24', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'CATEGORY', 'a9341837-8c47-48ed-be60-c2baee160c1c', 'Updated category', NULL, NULL, NULL, '2025-12-08 09:04:08.510811+00'),
('23ff9343-d8cb-48a3-86e3-9c34b6e11d6b', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'CATEGORY', 'a9341837-8c47-48ed-be60-c2baee160c1c', 'Updated category', NULL, NULL, NULL, '2025-12-08 09:04:28.024796+00'),
('728a0542-fc3c-4052-8c7f-24243ac8f54f', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'CATEGORY', 'a9341837-8c47-48ed-be60-c2baee160c1c', 'Updated category', NULL, NULL, NULL, '2025-12-08 09:09:21.983702+00'),
('6b250cb3-3db8-45d2-905f-d6d93d468fd2', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'CATEGORY', '5b7dcf51-84fb-4900-baa8-387672e0f0d2', 'Updated category', NULL, NULL, NULL, '2025-12-08 09:09:29.427437+00'),
('2f18106c-5827-47bb-b126-59bdb275123b', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'CATEGORY', '0421994d-df98-4d13-a527-8276a4b6a606', 'Updated category', NULL, NULL, NULL, '2025-12-08 09:09:35.655524+00'),
('de583d83-6aae-420c-9c7f-d3b2f9a97947', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'CATEGORY', 'c3042b7e-2b19-47a1-9170-f719ea7c143d', 'Updated category', NULL, NULL, NULL, '2025-12-08 09:09:40.383588+00'),
('5b9a7e83-83f5-4b40-8a26-b536d2c148ab', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'CATEGORY', 'c3042b7e-2b19-47a1-9170-f719ea7c143d', 'Updated category', NULL, NULL, NULL, '2025-12-08 09:09:49.237387+00'),
('0defdd11-34a0-46b5-b507-fb4654b4f782', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'SECTION', '481115b1-0462-4429-8bba-265cbd183bfe', 'Updated section', NULL, NULL, NULL, '2025-12-08 09:15:18.493269+00'),
('16a2e8e2-83ee-4534-8b6d-b33c93a4b330', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'SECTION', '838caff3-a279-47dd-a95e-a658425bf1d5', 'Updated section', NULL, NULL, NULL, '2025-12-08 09:15:21.99739+00'),
('558c54ad-aa79-4854-bf67-81a75fbe7076', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'SECTION', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', 'Updated section', NULL, NULL, NULL, '2025-12-08 09:16:05.175995+00'),
('c5e51396-c6ef-4f66-b7f1-8b2e40a64340', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'SECTION', '83e12b66-a25a-4e4e-8acf-bdbfd67e57df', 'Updated section', NULL, NULL, NULL, '2025-12-08 09:18:13.000257+00'),
('1064d3b8-79fb-4764-98ca-9779ad0c3743', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'SECTION', '838caff3-a279-47dd-a95e-a658425bf1d5', 'Updated section', NULL, NULL, NULL, '2025-12-08 09:18:48.201785+00'),
('6ec6bb5e-1d31-497f-9d62-64ef7c8fa6e1', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'SECTION', '95e9d0bc-55b1-477f-8629-0b7fc3e1ee9b', 'Updated section', NULL, NULL, NULL, '2025-12-08 09:18:54.613713+00'),
('680881bd-9d1c-462f-bf1e-a14a6857390e', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'SECTION', 'cb63f4d0-2cc6-49a0-9001-885b7744fdc9', 'Updated section', NULL, NULL, NULL, '2025-12-08 09:19:00.425597+00'),
('63b57616-c318-4735-a4f6-68a6b7a76a64', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'SECTION', '481115b1-0462-4429-8bba-265cbd183bfe', 'Updated section', NULL, NULL, NULL, '2025-12-08 09:19:03.941629+00'),
('a88cb412-83ae-4291-8931-bcf89348f693', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'DELETE', 'PROMO', '34105508-fe42-4c8f-b391-320115347a36', 'Deleted promo', NULL, NULL, NULL, '2025-12-08 10:46:47.071663+00'),
('e748ce78-4f3b-487e-803d-864d168515d4', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'DELETE', 'PROMO', 'fb72c4cf-2bf3-495f-a47f-5a9f2c88a365', 'Deleted promo', NULL, NULL, NULL, '2025-12-08 10:46:58.314004+00'),
('e29b76a7-8ebb-4dd4-ba52-4c09388f6d35', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'CREATE', 'PROMO', '4c4f65b1-f13d-47e7-a160-4f0f2e73ac34', 'Created promo DANA1212', NULL, NULL, NULL, '2025-12-08 10:48:43.264597+00'),
('7ebf722c-4355-459d-9871-08a3fb4936d0', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PROMO', '4c4f65b1-f13d-47e7-a160-4f0f2e73ac34', 'Updated promo DANA1212', NULL, NULL, NULL, '2025-12-08 10:48:52.379072+00'),
('59afbaf4-95e2-4e4e-80d0-f9422ca99931', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PROMO', '4c4f65b1-f13d-47e7-a160-4f0f2e73ac34', 'Updated promo DANA1212', NULL, NULL, NULL, '2025-12-08 10:49:49.795286+00'),
('74d16794-3247-4aa4-8b84-ef4892457b2c', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PROMO', '4c4f65b1-f13d-47e7-a160-4f0f2e73ac34', 'Updated promo DANA1212', NULL, NULL, NULL, '2025-12-08 10:50:19.55266+00'),
('03ab8b0f-f8ef-4a1f-8665-dd61b57299ff', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', 'c708c6d0-844d-47ac-8475-fb2f3b81b196', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 12:30:51.380657+00'),
('c2fc6521-7c9c-4e8e-934b-c1aec955defd', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '2162d8a6-906b-4c03-a2eb-c905f1228879', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 12:31:26.288929+00'),
('150f8518-8806-4a7f-a3c6-1adb35a82d5a', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '2162d8a6-906b-4c03-a2eb-c905f1228879', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 12:31:40.465051+00'),
('e6d69a2f-157f-402a-9567-825b94e22ffc', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '2162d8a6-906b-4c03-a2eb-c905f1228879', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 12:36:11.544409+00'),
('99826c75-1289-4cb5-85ef-bde3ab020872', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '2162d8a6-906b-4c03-a2eb-c905f1228879', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 12:36:17.574086+00'),
('f4c7387b-789e-425a-abc5-a32931f2d205', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '2162d8a6-906b-4c03-a2eb-c905f1228879', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 12:36:39.826+00'),
('994c205c-4ca2-4619-b93f-d0d012045219', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '9457bd7d-216c-4ae4-a654-02272dec87d8', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 12:37:42.598002+00'),
('7499b1b2-d915-4ae2-a674-8ef48b08220f', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '2162d8a6-906b-4c03-a2eb-c905f1228879', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 12:38:25.699162+00'),
('8792e441-69d2-4b53-9bd5-3c9a7b32b763', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '2162d8a6-906b-4c03-a2eb-c905f1228879', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 12:48:52.454291+00'),
('0426df44-af3a-4ff7-9392-0860d322f639', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', 'c708c6d0-844d-47ac-8475-fb2f3b81b196', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 13:04:07.397053+00'),
('132f5356-920c-4e82-9c4c-f5d49338d477', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', 'c708c6d0-844d-47ac-8475-fb2f3b81b196', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 13:04:22.125256+00'),
('42adf197-d7dd-4f1d-861e-c292bf4ab548', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '9457bd7d-216c-4ae4-a654-02272dec87d8', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 13:04:35.357204+00'),
('a885fb60-d1fc-4742-b306-002f93ab80bd', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '0d28ee79-9767-4fbf-ad0b-732fb42a9e05', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 13:06:59.409436+00'),
('827bd85b-582a-4570-ab7e-09a167683cf6', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'DELETE', 'CATEGORY', '5b7dcf51-84fb-4900-baa8-387672e0f0d2', 'Deleted category', NULL, NULL, NULL, '2025-12-08 14:18:19.333696+00'),
('d20a6e1d-50c6-4296-926d-38681ca409c5', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'BANNER', '85f97a4f-6748-4c05-b6d3-c5c0af2ebfb6', 'Updated banner', NULL, NULL, NULL, '2025-12-08 14:51:15.096278+00'),
('1297de1c-b8fb-475a-8b9a-5b0e31b28d4f', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '48604997-4746-47d8-862e-f49efd23bdd7', 'Updated payment channel', NULL, NULL, NULL, '2025-12-08 15:37:17.443317+00'),
('8ce8119c-7b4a-4859-b0a6-b6bc6e4f4fe7', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'CREATE', 'PAYMENT_CHANNEL', '6b754f8e-c124-4f00-9776-36d9c73fde9e', 'Created payment channel', NULL, NULL, NULL, '2025-12-08 19:04:27.679511+00'),
('4f83e23c-3dd2-4e8e-a0ba-693aca1c7c15', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'CREATE', 'PAYMENT_CHANNEL', 'f72f575c-4cd8-42fa-b1a3-6bf78bfd6a0c', 'Created payment channel', NULL, NULL, NULL, '2025-12-08 19:05:08.691338+00'),
('2719f9df-c1c9-4d37-8959-352efd48918d', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'DELETE', 'PROMO', '4c4f65b1-f13d-47e7-a160-4f0f2e73ac34', 'Deleted promo', NULL, NULL, NULL, '2025-12-09 08:54:07.146768+00'),
('d572a548-630b-40e4-812b-d52f54b5f793', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'CREATE', 'PROMO', '7b546b04-c6d5-4e70-8d73-98e205cf8738', 'Created promo DANA1212', NULL, NULL, NULL, '2025-12-09 08:55:10.985156+00'),
('87837329-6d32-4f27-a879-03f33c289c89', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', 'a45b214c-cd62-4be9-862c-83c034ddf4f2', 'Updated payment channel', NULL, NULL, NULL, '2025-12-09 09:03:39.335219+00'),
('db52e43e-4db3-4df8-ad8d-507cc0986e37', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '2162d8a6-906b-4c03-a2eb-c905f1228879', 'Updated payment channel', NULL, NULL, NULL, '2025-12-09 09:45:44.228178+00'),
('3e9194f0-1164-47ef-8073-514308744322', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', 'd5606ef6-ee54-46df-9739-64220fdc1b28', 'Updated payment channel', NULL, NULL, NULL, '2025-12-09 11:19:49.50656+00'),
('acf3efe8-fc44-4a6e-819d-141a84530b14', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '48604997-4746-47d8-862e-f49efd23bdd7', 'Updated payment channel', NULL, NULL, NULL, '2025-12-09 11:22:49.376492+00'),
('ab93dc13-e536-4bb4-bd4e-55bb0a496cb8', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', 'c708c6d0-844d-47ac-8475-fb2f3b81b196', 'Updated payment channel', NULL, NULL, NULL, '2025-12-09 11:23:25.086224+00'),
('bb1c9dff-cf60-41ca-bc72-49309e1df85f', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', '0d28ee79-9767-4fbf-ad0b-732fb42a9e05', 'Updated payment channel', NULL, NULL, NULL, '2025-12-09 11:23:34.991014+00'),
('901b596e-3403-417f-b921-ca10a26b2f3e', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', 'a7251199-b644-4a34-be74-298819a2841e', 'Updated payment channel', NULL, NULL, NULL, '2025-12-09 11:25:09.563524+00'),
('1043d790-c1cf-4f3e-9048-7aa9898eafd7', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'PAYMENT_CHANNEL', 'cb1707a4-285f-4260-9d4b-5908be50d087', 'Updated payment channel', NULL, NULL, NULL, '2025-12-09 11:25:19.801298+00'),
('5d99f26c-ba60-420a-90c9-232eae14cf22', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'SECTION', '838caff3-a279-47dd-a95e-a658425bf1d5', 'Updated section', NULL, NULL, NULL, '2025-12-09 12:17:27.57918+00'),
('3fabfcc9-bc4f-4900-9aed-9c1b5d156e5b', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'UPDATE', 'SECTION', '95e9d0bc-55b1-477f-8629-0b7fc3e1ee9b', 'Updated section', NULL, NULL, NULL, '2025-12-09 12:17:32.787633+00'),
('97f44da1-7dbb-4567-ba13-538b621edbec', 'a7466525-cf91-4890-b38a-9ff0d48d4b87', NULL, NULL, 'DELETE', 'SECTION', '481115b1-0462-4429-8bba-265cbd183bfe', 'Deleted section', NULL, NULL, NULL, '2025-12-09 20:39:40.214227+00');