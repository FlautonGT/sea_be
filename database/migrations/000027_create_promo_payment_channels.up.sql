-- Name: promo_payment_channels; Type: TABLE; Schema: public; Owner: seaply
--

CREATE TABLE public.promo_payment_channels (
    promo_id uuid NOT NULL,
    channel_id uuid NOT NULL
);


ALTER TABLE public.promo_payment_channels OWNER TO gate;

--

-- Name: promo_payment_channels promo_payment_channels_pkey; Type: CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.promo_payment_channels
    ADD CONSTRAINT promo_payment_channels_pkey PRIMARY KEY (promo_id, channel_id);


--


-- SEED DATA --


INSERT INTO public.promo_payment_channels (promo_id, channel_id) VALUES
('7b546b04-c6d5-4e70-8d73-98e205cf8738', 'c708c6d0-844d-47ac-8475-fb2f3b81b196');