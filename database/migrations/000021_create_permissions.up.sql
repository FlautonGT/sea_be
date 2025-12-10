-- Name: permissions; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.permissions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code character varying(50) NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    category character varying(50) NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.permissions OWNER TO gate;

--

-- Name: permissions permissions_code_key; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_code_key UNIQUE (code);


--

-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--


-- SEED DATA --


INSERT INTO public.permissions (id, code, name, description, category, created_at) VALUES
('b4d6ae45-7db7-4613-8cef-71ccf5fa5d25', 'admin:read', 'View Admins', 'Can view admin list and details', 'Admin', '2025-12-04 19:24:29.724254+00'),
('3cdd8cfd-6294-4305-b321-b366afbe5cb3', 'admin:create', 'Create Admin', 'Can create new admin accounts', 'Admin', '2025-12-04 19:24:29.724254+00'),
('4bfeb815-9e9d-4a15-af0f-029eeab627d7', 'admin:update', 'Update Admin', 'Can update admin accounts', 'Admin', '2025-12-04 19:24:29.724254+00'),
('6ab87828-6f83-4daf-9bb7-795f7fb23422', 'admin:delete', 'Delete Admin', 'Can delete admin accounts', 'Admin', '2025-12-04 19:24:29.724254+00'),
('e7d3b7c9-5be1-48a5-a2c5-b9a4c34ad771', 'role:manage', 'Manage Roles', 'Can manage roles and permissions', 'Admin', '2025-12-04 19:24:29.724254+00'),
('7261c165-6017-41b5-9669-cac11cd4dd06', 'provider:read', 'View Providers', 'Can view provider list and details', 'Provider', '2025-12-04 19:24:29.724254+00'),
('adcba813-3af1-4d4d-9d76-38531e054029', 'provider:create', 'Create Provider', 'Can create new providers', 'Provider', '2025-12-04 19:24:29.724254+00'),
('72e82176-15c8-4c2f-be2c-938e150e417a', 'provider:update', 'Update Provider', 'Can update providers', 'Provider', '2025-12-04 19:24:29.724254+00'),
('2896b9b2-7b97-4840-8bca-a49b9a0e0d4b', 'provider:delete', 'Delete Provider', 'Can delete providers', 'Provider', '2025-12-04 19:24:29.724254+00'),
('d1355dd2-fbb7-43fb-a564-40726c081543', 'gateway:read', 'View Gateways', 'Can view payment gateways', 'Gateway', '2025-12-04 19:24:29.724254+00'),
('782be53a-f597-4b50-8b17-188905ffa60a', 'gateway:create', 'Create Gateway', 'Can create payment gateways', 'Gateway', '2025-12-04 19:24:29.724254+00'),
('086c834e-504e-4af1-8583-8ff00e5e82e7', 'gateway:update', 'Update Gateway', 'Can update payment gateways', 'Gateway', '2025-12-04 19:24:29.724254+00'),
('fa67a554-3420-440d-b98b-a026e0644dbd', 'gateway:delete', 'Delete Gateway', 'Can delete payment gateways', 'Gateway', '2025-12-04 19:24:29.724254+00'),
('4bbdfa4b-d861-4a57-a6b8-713579d26f49', 'product:read', 'View Products', 'Can view products', 'Product', '2025-12-04 19:24:29.724254+00'),
('118f2a6f-9ae8-4f1b-932b-71be199cd877', 'product:create', 'Create Product', 'Can create products', 'Product', '2025-12-04 19:24:29.724254+00'),
('6357654e-b269-4582-b761-c1afd7ca1726', 'product:update', 'Update Product', 'Can update products', 'Product', '2025-12-04 19:24:29.724254+00'),
('f1d91bc4-328a-4182-a106-6500ad911405', 'product:delete', 'Delete Product', 'Can delete products', 'Product', '2025-12-04 19:24:29.724254+00'),
('c238850a-ecb6-4a07-9ea2-18c56e06fec0', 'sku:read', 'View SKUs', 'Can view SKUs', 'SKU', '2025-12-04 19:24:29.724254+00'),
('2ffcd3eb-3f56-4656-8403-939d08fede58', 'sku:create', 'Create SKU', 'Can create SKUs', 'SKU', '2025-12-04 19:24:29.724254+00'),
('3e73a255-ebb6-468a-8764-c19c7be289fd', 'sku:update', 'Update SKU', 'Can update SKUs', 'SKU', '2025-12-04 19:24:29.724254+00'),
('47c43e3a-a9f5-4fd8-b0c6-5a0ed9ef64ee', 'sku:delete', 'Delete SKU', 'Can delete SKUs', 'SKU', '2025-12-04 19:24:29.724254+00'),
('c04699c3-8551-4860-b460-e98a926e98ab', 'sku:sync', 'Sync SKU', 'Can sync SKUs from provider', 'SKU', '2025-12-04 19:24:29.724254+00'),
('b63b76e2-1fae-4d89-bf45-eef09b37592f', 'transaction:read', 'View Transactions', 'Can view transactions', 'Transaction', '2025-12-04 19:24:29.724254+00'),
('3663ba95-4fd5-426f-b6c3-7afddfffa996', 'transaction:update', 'Update Transaction', 'Can update transaction status', 'Transaction', '2025-12-04 19:24:29.724254+00'),
('076c58df-7b61-4825-bdbf-20d702a135e0', 'transaction:refund', 'Refund Transaction', 'Can process refunds', 'Transaction', '2025-12-04 19:24:29.724254+00'),
('be5c3ccc-5520-4fb1-a61f-fee000490f97', 'transaction:manual', 'Manual Process', 'Can manually process transactions', 'Transaction', '2025-12-04 19:24:29.724254+00'),
('ba771aa7-40a7-4410-a8af-485a4e91eb8f', 'user:read', 'View Users', 'Can view users', 'User', '2025-12-04 19:24:29.724254+00'),
('69327111-788e-4e65-be81-e18fc297a353', 'user:update', 'Update User', 'Can update user accounts', 'User', '2025-12-04 19:24:29.724254+00'),
('97cfadbf-fefd-4f8c-8d1c-da2110e3dbbd', 'user:suspend', 'Suspend User', 'Can suspend user accounts', 'User', '2025-12-04 19:24:29.724254+00'),
('c9c460ca-6bab-4958-aa05-7624f31a3bfb', 'user:balance', 'Adjust Balance', 'Can adjust user balance', 'User', '2025-12-04 19:24:29.724254+00'),
('4f5602a7-52bd-4596-ad79-b816de8ec1a5', 'promo:read', 'View Promos', 'Can view promos', 'Promo', '2025-12-04 19:24:29.724254+00'),
('bcc39899-b8d0-486b-8e42-27c058d08a53', 'promo:create', 'Create Promo', 'Can create promos', 'Promo', '2025-12-04 19:24:29.724254+00'),
('92ea62fc-9202-4ce6-bfc3-e796257365dc', 'promo:update', 'Update Promo', 'Can update promos', 'Promo', '2025-12-04 19:24:29.724254+00'),
('34ce7a3e-2f48-45c1-8522-d15ea1224f5d', 'promo:delete', 'Delete Promo', 'Can delete promos', 'Promo', '2025-12-04 19:24:29.724254+00'),
('6b0694ab-d0de-46d4-b412-855f8bdff402', 'content:read', 'View Content', 'Can view content', 'Content', '2025-12-04 19:24:29.724254+00'),
('e6c7041d-16a2-4a35-8b12-698e13a83432', 'content:banner', 'Manage Banners', 'Can manage banners', 'Content', '2025-12-04 19:24:29.724254+00'),
('045cef44-46e2-4fb1-87d1-0de5c6438f15', 'content:popup', 'Manage Popups', 'Can manage popups', 'Content', '2025-12-04 19:24:29.724254+00'),
('8e5f8eaf-c0ab-48c1-948d-660bf25f2a06', 'report:read', 'View Reports', 'Can view reports', 'Report', '2025-12-04 19:24:29.724254+00'),
('3bb671ae-842b-4385-963b-6e95868be71d', 'report:export', 'Export Reports', 'Can export reports', 'Report', '2025-12-04 19:24:29.724254+00'),
('08d882ab-bd28-4e16-967a-2e7f12d7b3e4', 'audit:read', 'View Audit Logs', 'Can view audit logs', 'Audit', '2025-12-04 19:24:29.724254+00'),
('857bec5b-0993-4d16-96b3-b4d58e7e64cd', 'setting:read', 'View Settings', 'Can view settings', 'Setting', '2025-12-04 19:24:29.724254+00'),
('9123cc1b-d7ad-4dbd-82ec-f08c29be5ad6', 'setting:update', 'Update Settings', 'Can update settings', 'Setting', '2025-12-04 19:24:29.724254+00');