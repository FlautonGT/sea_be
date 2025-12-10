-- Name: transaction_logs; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.transaction_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    transaction_id uuid NOT NULL,
    status character varying(50) NOT NULL,
    message text,
    data jsonb,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.transaction_logs OWNER TO gate;

--

-- Name: transaction_logs transaction_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.transaction_logs
    ADD CONSTRAINT transaction_logs_pkey PRIMARY KEY (id);


--


-- SEED DATA --


INSERT INTO public.transaction_logs (id, transaction_id, status, message, data, created_at) VALUES
('e49ea366-8b4b-4713-9d6d-25125337ae52', '2b25a38d-13c5-4c35-9403-d1f509985189', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-07 19:39:41.676698+00'),
('c3c8a340-dd9c-46d4-b02f-7d382ccb167c', '9836bda7-a778-4aae-842b-d889da026ff7', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-07 19:40:07.229071+00'),
('2c034ceb-45f4-42c3-9963-513050e7424d', '63ae7a25-bdef-4764-9b3d-a152fe994202', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-07 19:48:50.244371+00'),
('e17afec4-9bbb-4004-8505-1cc87da31c82', 'fe484bec-c8f1-4308-8a1b-b1c27b46a65e', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-07 19:59:30.760491+00'),
('451f278c-071b-4e61-8bcd-a9ce3a404082', '40b24c1c-d595-4eb7-bee6-b487b00f5b03', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-07 20:00:04.419543+00'),
('3ecd8423-0f89-4bcf-9c8d-6a2315b1e927', 'ad0f42f9-6e82-4034-a64e-98107716887c', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-07 20:00:34.23521+00'),
('fa331289-cb44-4665-af93-042a8eb5a3a4', 'b0949824-eb28-4137-8b16-70e8b930cb82', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-07 20:02:51.45086+00'),
('aeec342e-e18c-4b92-a424-0bae95738c30', '3ab26fb0-52df-4c80-81d9-aea0a66fffd6', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-07 20:31:39.687526+00'),
('29539962-a908-4e5e-ac94-90c701b3ad38', '89623e76-ebaf-4a97-af99-e689cf6fcf71', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-07 21:08:07.715024+00'),
('af1b95eb-7c51-4eff-8cf4-de47aae9822d', '8b2f0308-2b5c-4e29-a3a2-95ee0af6f874', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 14:02:11.941761+00'),
('254fdd84-6aa0-4fc4-abdc-04ddbad95d77', '2e9b5366-44f0-4508-9cf8-a142fc8a26ee', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 14:56:48.412352+00'),
('92903177-1e61-438a-92ca-7e93e356f4d1', '22c23389-47c8-4bc9-a794-ae65f49ecf8a', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 15:02:42.964918+00'),
('4bf19dbf-e0ed-448f-bd03-b9af4ab4ab68', 'b9206656-b6f0-45d4-b999-b9841f22e273', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 16:04:02.05676+00'),
('039d4590-df12-432e-9290-fc29dc878ffa', '0e043da8-ee12-49c4-84cc-08a3088d3257', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 16:04:40.147059+00'),
('d7107b41-61cd-41e5-9708-65da9a98ca5d', '96545aad-9521-4cbe-9c96-856a098982d5', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 16:14:47.768406+00'),
('296d3d31-0527-49ec-8829-5bdf455bc2c9', '86746e53-14c2-4972-a6bb-5c6d5e95917c', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 16:16:31.103134+00'),
('0e1d2f88-627b-486e-803e-7039c6d289fd', 'b6010a9e-99f8-4d3e-a527-77787f64a8ff', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 16:19:18.619498+00'),
('28357c32-4727-45a0-a444-bc03f4f02704', '706f614a-aefe-4bfb-8624-9f60cbc5e9bc', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 16:24:20.953462+00'),
('9fa65f98-90f6-4cc0-81d0-6f0fb166523f', 'aeeaa0e9-c038-46f6-957f-600fbd2fccf2', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 16:30:14.685918+00'),
('45dac793-bcee-4769-95cc-f4be2faf9728', 'db5ad32b-03ff-48c8-ae6e-2cfe290ae54e', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 16:36:23.119434+00'),
('38e48cc4-8415-4bd7-95e8-80569e050ee8', '9a2ae2b7-d4ee-4ae8-8a03-99e558e04fbd', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 16:51:57.777686+00'),
('1140d404-5f01-42d4-bad5-06dd7f69ad63', '32b4913a-c833-4a61-85dd-8175c9d9dedd', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 16:59:04.836708+00'),
('5fa6b96e-d392-44b4-8fc7-caec7d57f726', '67d0e98d-ca54-4e10-b1fd-3c7739b3614f', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 17:01:43.465085+00'),
('20bcdf02-451d-4d67-8197-bf38f53f09bb', 'b13d0368-c7d3-4f9d-854c-6ae5fbcae084', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 17:42:26.106999+00'),
('e8ead963-0848-465d-9a22-6014ec562e0c', '954cf6ee-7d1d-4893-8e51-f4d00c2a0837', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 17:46:01.179161+00'),
('a1a91b10-bf45-409b-8117-7e471bfb1567', 'd348b87a-b21e-4fd6-b333-473d5e42d8aa', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 18:35:38.896483+00'),
('1f2db8b8-8c90-4167-8d18-71f29079b533', '936fc54c-ff33-4920-98fa-3632fe877091', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 19:15:10.229084+00'),
('8ff8814f-54cc-45fe-83b4-0a081c33c42c', 'fbe71972-067a-4e49-920b-3ed32d06439d', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-08 19:16:02.128924+00'),
('9c070c57-9720-4e91-be0b-787261809c94', '76d0707d-17f3-4965-b4b7-8a6259f9552c', 'PENDING', 'Order created, waiting for payment', NULL, '2025-12-09 15:31:38.814181+00');