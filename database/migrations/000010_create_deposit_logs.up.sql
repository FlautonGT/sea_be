-- Name: deposit_logs; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.deposit_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    deposit_id uuid NOT NULL,
    status character varying(50) NOT NULL,
    message text,
    data jsonb,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.deposit_logs OWNER TO gate;

--

-- Name: deposit_logs deposit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.deposit_logs
    ADD CONSTRAINT deposit_logs_pkey PRIMARY KEY (id);


--


-- SEED DATA --


