-- Name: refunds; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.refunds (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    transaction_id uuid,
    deposit_id uuid,
    invoice_number character varying(50) NOT NULL,
    amount bigint NOT NULL,
    currency public.currency_code NOT NULL,
    refund_to character varying(50) NOT NULL,
    status character varying(50) DEFAULT 'PROCESSING'::character varying,
    reason text,
    processed_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    completed_at timestamp with time zone,
    CONSTRAINT refund_has_reference CHECK ((((transaction_id IS NOT NULL) AND (deposit_id IS NULL)) OR ((transaction_id IS NULL) AND (deposit_id IS NOT NULL))))
);


ALTER TABLE public.refunds OWNER TO gate;

--

-- Name: refunds refunds_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.refunds
    ADD CONSTRAINT refunds_pkey PRIMARY KEY (id);


--


-- SEED DATA --


