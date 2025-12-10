-- Name: payment_channel_gateways; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.payment_channel_gateways (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    channel_id uuid NOT NULL,
    gateway_id uuid NOT NULL,
    payment_type public.payment_type NOT NULL,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.payment_channel_gateways OWNER TO gate;

--

-- Name: payment_channel_gateways payment_channel_gateways_channel_id_payment_type_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.payment_channel_gateways
    ADD CONSTRAINT payment_channel_gateways_channel_id_payment_type_key UNIQUE (channel_id, payment_type);


--

-- Name: payment_channel_gateways payment_channel_gateways_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.payment_channel_gateways
    ADD CONSTRAINT payment_channel_gateways_pkey PRIMARY KEY (id);


--


-- SEED DATA --


INSERT INTO public.payment_channel_gateways (id, channel_id, gateway_id, payment_type, is_active, created_at, updated_at) VALUES
('79616e0b-3b07-4846-b9fb-27036673deb2', 'c708c6d0-844d-47ac-8475-fb2f3b81b196', '04bc68be-f58c-475b-947f-1d4301e99fbf', 'purchase', 't', '2025-12-04 19:24:30.017436+00', '2025-12-04 19:24:30.017436+00'),
('61335d85-4bec-4ea4-bada-b48c608e8637', '0d28ee79-9767-4fbf-ad0b-732fb42a9e05', '47f1ccd5-df1c-4848-bf32-67324c60f18b', 'purchase', 't', '2025-12-04 19:24:30.017436+00', '2025-12-04 19:24:30.017436+00'),
('69c017bc-3cfd-4592-98b5-ff7340b18d14', 'a45b214c-cd62-4be9-862c-83c034ddf4f2', '47f1ccd5-df1c-4848-bf32-67324c60f18b', 'purchase', 't', '2025-12-04 19:24:30.017436+00', '2025-12-04 19:24:30.017436+00'),
('26ca3397-6a09-4f96-aa0d-b2ad74458eaa', '48604997-4746-47d8-862e-f49efd23bdd7', '4a7124cf-a073-4cee-adbe-afb047f6a314', 'purchase', 't', '2025-12-04 19:24:30.017436+00', '2025-12-04 19:24:30.017436+00'),
('63db3fc4-478e-424e-8281-4f8b18ae70b2', '48604997-4746-47d8-862e-f49efd23bdd7', '4a7124cf-a073-4cee-adbe-afb047f6a314', 'deposit', 't', '2025-12-04 19:24:30.017436+00', '2025-12-04 19:24:30.017436+00'),
('8aeaca44-8e4f-4271-a948-3815717412b1', 'd5606ef6-ee54-46df-9739-64220fdc1b28', '84ff7eef-e31a-4a13-ad84-8eae3bb1ec43', 'purchase', 't', '2025-12-04 19:24:30.017436+00', '2025-12-04 19:24:30.017436+00'),
('0d6d86b8-1fb9-4c03-85e2-25b1e599579a', 'd5606ef6-ee54-46df-9739-64220fdc1b28', '84ff7eef-e31a-4a13-ad84-8eae3bb1ec43', 'deposit', 't', '2025-12-04 19:24:30.017436+00', '2025-12-04 19:24:30.017436+00'),
('8219b93b-a299-41ea-a2ee-5436860a2739', 'cb1707a4-285f-4260-9d4b-5908be50d087', '4e8cbd84-4eb2-403a-a365-3cce2f9b5dbc', 'purchase', 't', '2025-12-04 19:24:30.017436+00', '2025-12-04 19:24:30.017436+00'),
('18aa3fe5-d615-4e7d-9c2c-4e06983c85f9', 'cb1707a4-285f-4260-9d4b-5908be50d087', '4e8cbd84-4eb2-403a-a365-3cce2f9b5dbc', 'deposit', 't', '2025-12-04 19:24:30.017436+00', '2025-12-04 19:24:30.017436+00'),
('a7f60964-d62c-4322-ac7f-18d1b6534843', 'a7251199-b644-4a34-be74-298819a2841e', '4e8cbd84-4eb2-403a-a365-3cce2f9b5dbc', 'purchase', 't', '2025-12-04 19:24:30.017436+00', '2025-12-04 19:24:30.017436+00'),
('131f9821-ad89-4cd2-8841-e174581b1ac2', 'a7251199-b644-4a34-be74-298819a2841e', '4e8cbd84-4eb2-403a-a365-3cce2f9b5dbc', 'deposit', 't', '2025-12-04 19:24:30.017436+00', '2025-12-04 19:24:30.017436+00'),
('2de60641-61d7-42c3-97aa-42f8ad93fe34', '2162d8a6-906b-4c03-a2eb-c905f1228879', '04bc68be-f58c-475b-947f-1d4301e99fbf', 'purchase', 't', '2025-12-04 19:24:30.017436+00', '2025-12-09 09:45:44.228178+00'),
('0da330a3-87a4-426f-a5c2-85f726c9e0b3', '2162d8a6-906b-4c03-a2eb-c905f1228879', '04bc68be-f58c-475b-947f-1d4301e99fbf', 'deposit', 't', '2025-12-04 19:24:30.017436+00', '2025-12-09 09:45:44.228178+00');