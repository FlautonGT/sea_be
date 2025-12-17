-- Name: reviews; Type: TABLE; Schema: public; Owner: gate
--

CREATE TABLE public.reviews (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    invoice_number character varying(50) NOT NULL,
    transaction_id uuid NOT NULL,
    user_id uuid,
    rating integer NOT NULL,
    comment text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT reviews_rating_check CHECK ((rating >= 1 AND rating <= 5))
);


ALTER TABLE public.reviews OWNER TO gate;

--
-- Name: reviews reviews_invoice_number_key; Type: CONSTRAINT; Schema: public; Owner: gate
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_invoice_number_key UNIQUE (invoice_number);


--
-- Name: reviews reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: gate
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_pkey PRIMARY KEY (id);


--
-- Name: reviews reviews_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gate
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id) ON DELETE CASCADE;


--
-- Name: reviews reviews_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gate
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: idx_reviews_invoice; Type: INDEX; Schema: public; Owner: gate
--

CREATE INDEX idx_reviews_invoice ON public.reviews USING btree (invoice_number);


--
-- Name: idx_reviews_transaction; Type: INDEX; Schema: public; Owner: gate
--

CREATE INDEX idx_reviews_transaction ON public.reviews USING btree (transaction_id);


--
-- Name: idx_reviews_user; Type: INDEX; Schema: public; Owner: gate
--

CREATE INDEX idx_reviews_user ON public.reviews USING btree (user_id);


--
-- Name: idx_reviews_created_at; Type: INDEX; Schema: public; Owner: gate
--

CREATE INDEX idx_reviews_created_at ON public.reviews USING btree (created_at DESC);


--
-- Name: reviews update_reviews_updated_at; Type: TRIGGER; Schema: public; Owner: gate
--

CREATE TRIGGER update_reviews_updated_at BEFORE UPDATE ON public.reviews FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

