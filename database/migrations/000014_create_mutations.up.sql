-- Name: mutations; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.mutations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    invoice_number character varying(50),
    reference_type character varying(50) NOT NULL,
    reference_id uuid,
    description text NOT NULL,
    mutation_type public.mutation_type NOT NULL,
    amount bigint NOT NULL,
    balance_before bigint NOT NULL,
    balance_after bigint NOT NULL,
    currency public.currency_code NOT NULL,
    admin_id uuid,
    admin_note text,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.mutations OWNER TO gate;

--

-- Name: mutations mutations_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.mutations
    ADD CONSTRAINT mutations_pkey PRIMARY KEY (id);


--

-- Name: idx_mutations_created_at; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_mutations_created_at ON public.mutations USING btree (created_at DESC);


--

-- Name: idx_mutations_type; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_mutations_type ON public.mutations USING btree (mutation_type);


--

-- Name: idx_mutations_user; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_mutations_user ON public.mutations USING btree (user_id);


--


-- SEED DATA --


INSERT INTO public.mutations (id, user_id, invoice_number, reference_type, reference_id, description, mutation_type, amount, balance_before, balance_after, currency, admin_id, admin_note, created_at) VALUES
('5f262922-3368-4c79-96f1-3f18d848b703', 'f5c92b21-fe15-463c-bfef-65fedad2ad9f', NULL, 'DEPOSIT', NULL, 'Deposit via QRIS', 'CREDIT', '100000', '400000', '500000', 'IDR', NULL, NULL, '2025-12-04 19:24:33.485665+00');