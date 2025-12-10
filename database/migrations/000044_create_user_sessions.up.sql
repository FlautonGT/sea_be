-- Name: user_sessions; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.user_sessions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    refresh_token_hash character varying(255) NOT NULL,
    ip_address inet,
    user_agent text,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.user_sessions OWNER TO gate;

--

-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);


--

-- Name: idx_user_sessions_expires_at; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_user_sessions_expires_at ON public.user_sessions USING btree (expires_at);


--

-- Name: idx_user_sessions_user_id; Type: INDEX; Schema: public; Owner: seaply
--

CREATE INDEX idx_user_sessions_user_id ON public.user_sessions USING btree (user_id);


--


-- SEED DATA --


