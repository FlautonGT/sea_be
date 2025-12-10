-- Name: admin_sessions admin_sessions_admin_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.admin_sessions
    ADD CONSTRAINT admin_sessions_admin_id_fkey FOREIGN KEY (admin_id) REFERENCES public.admins(id) ON DELETE CASCADE;


--

-- Name: admins admins_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.admins
    ADD CONSTRAINT admins_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.admins(id);


--

-- Name: admins admins_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.admins
    ADD CONSTRAINT admins_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id);


--

-- Name: audit_logs audit_logs_admin_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_admin_id_fkey FOREIGN KEY (admin_id) REFERENCES public.admins(id);


--

-- Name: banner_regions banner_regions_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.banner_regions
    ADD CONSTRAINT banner_regions_banner_id_fkey FOREIGN KEY (banner_id) REFERENCES public.banners(id) ON DELETE CASCADE;


--

-- Name: category_regions category_regions_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.category_regions
    ADD CONSTRAINT category_regions_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE CASCADE;


--

-- Name: deposit_logs deposit_logs_deposit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.deposit_logs
    ADD CONSTRAINT deposit_logs_deposit_id_fkey FOREIGN KEY (deposit_id) REFERENCES public.deposits(id) ON DELETE CASCADE;


--

-- Name: deposits deposits_cancelled_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.deposits
    ADD CONSTRAINT deposits_cancelled_by_fkey FOREIGN KEY (cancelled_by) REFERENCES public.admins(id);


--

-- Name: deposits deposits_confirmed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.deposits
    ADD CONSTRAINT deposits_confirmed_by_fkey FOREIGN KEY (confirmed_by) REFERENCES public.admins(id);


--

-- Name: deposits deposits_payment_channel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.deposits
    ADD CONSTRAINT deposits_payment_channel_id_fkey FOREIGN KEY (payment_channel_id) REFERENCES public.payment_channels(id);


--

-- Name: deposits deposits_payment_gateway_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.deposits
    ADD CONSTRAINT deposits_payment_gateway_id_fkey FOREIGN KEY (payment_gateway_id) REFERENCES public.payment_gateways(id);


--

-- Name: deposits deposits_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.deposits
    ADD CONSTRAINT deposits_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--

-- Name: email_verifications email_verifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.email_verifications
    ADD CONSTRAINT email_verifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--

-- Name: mutations mutations_admin_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.mutations
    ADD CONSTRAINT mutations_admin_id_fkey FOREIGN KEY (admin_id) REFERENCES public.admins(id);


--

-- Name: mutations mutations_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.mutations
    ADD CONSTRAINT mutations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--

-- Name: password_resets password_resets_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.password_resets
    ADD CONSTRAINT password_resets_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--

-- Name: payment_channel_gateways payment_channel_gateways_channel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.payment_channel_gateways
    ADD CONSTRAINT payment_channel_gateways_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES public.payment_channels(id) ON DELETE CASCADE;


--

-- Name: payment_channel_gateways payment_channel_gateways_gateway_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.payment_channel_gateways
    ADD CONSTRAINT payment_channel_gateways_gateway_id_fkey FOREIGN KEY (gateway_id) REFERENCES public.payment_gateways(id);


--

-- Name: payment_channel_regions payment_channel_regions_channel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.payment_channel_regions
    ADD CONSTRAINT payment_channel_regions_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES public.payment_channels(id) ON DELETE CASCADE;


--

-- Name: payment_channels payment_channels_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.payment_channels
    ADD CONSTRAINT payment_channels_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.payment_channel_categories(id);


--

-- Name: product_fields product_fields_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.product_fields
    ADD CONSTRAINT product_fields_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--

-- Name: product_regions product_regions_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.product_regions
    ADD CONSTRAINT product_regions_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--

-- Name: product_sections product_sections_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.product_sections
    ADD CONSTRAINT product_sections_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--

-- Name: product_sections product_sections_section_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.product_sections
    ADD CONSTRAINT product_sections_section_id_fkey FOREIGN KEY (section_id) REFERENCES public.sections(id) ON DELETE CASCADE;


--

-- Name: products products_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id);


--

-- Name: promo_payment_channels promo_payment_channels_channel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.promo_payment_channels
    ADD CONSTRAINT promo_payment_channels_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES public.payment_channels(id) ON DELETE CASCADE;


--

-- Name: promo_payment_channels promo_payment_channels_promo_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.promo_payment_channels
    ADD CONSTRAINT promo_payment_channels_promo_id_fkey FOREIGN KEY (promo_id) REFERENCES public.promos(id) ON DELETE CASCADE;


--

-- Name: promo_products promo_products_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.promo_products
    ADD CONSTRAINT promo_products_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--

-- Name: promo_products promo_products_promo_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.promo_products
    ADD CONSTRAINT promo_products_promo_id_fkey FOREIGN KEY (promo_id) REFERENCES public.promos(id) ON DELETE CASCADE;


--

-- Name: promo_regions promo_regions_promo_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.promo_regions
    ADD CONSTRAINT promo_regions_promo_id_fkey FOREIGN KEY (promo_id) REFERENCES public.promos(id) ON DELETE CASCADE;


--

-- Name: promo_usages promo_usages_promo_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.promo_usages
    ADD CONSTRAINT promo_usages_promo_id_fkey FOREIGN KEY (promo_id) REFERENCES public.promos(id);


--

-- Name: promo_usages promo_usages_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.promo_usages
    ADD CONSTRAINT promo_usages_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--

-- Name: refunds refunds_deposit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.refunds
    ADD CONSTRAINT refunds_deposit_id_fkey FOREIGN KEY (deposit_id) REFERENCES public.deposits(id);


--

-- Name: refunds refunds_processed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.refunds
    ADD CONSTRAINT refunds_processed_by_fkey FOREIGN KEY (processed_by) REFERENCES public.admins(id);


--

-- Name: refunds refunds_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.refunds
    ADD CONSTRAINT refunds_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id);


--

-- Name: role_permissions role_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES public.permissions(id) ON DELETE CASCADE;


--

-- Name: role_permissions role_permissions_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--

-- Name: sku_pricing sku_pricing_sku_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.sku_pricing
    ADD CONSTRAINT sku_pricing_sku_id_fkey FOREIGN KEY (sku_id) REFERENCES public.skus(id) ON DELETE CASCADE;


--

-- Name: skus skus_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.skus
    ADD CONSTRAINT skus_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id);


--

-- Name: skus skus_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.skus
    ADD CONSTRAINT skus_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.providers(id);


--

-- Name: skus skus_section_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.skus
    ADD CONSTRAINT skus_section_id_fkey FOREIGN KEY (section_id) REFERENCES public.sections(id);


--

-- Name: transaction_logs transaction_logs_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.transaction_logs
    ADD CONSTRAINT transaction_logs_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id) ON DELETE CASCADE;


--

-- Name: transactions transactions_payment_channel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_payment_channel_id_fkey FOREIGN KEY (payment_channel_id) REFERENCES public.payment_channels(id);


--

-- Name: transactions transactions_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id);


--

-- Name: transactions transactions_promo_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_promo_id_fkey FOREIGN KEY (promo_id) REFERENCES public.promos(id);


--

-- Name: transactions transactions_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.providers(id);


--

-- Name: transactions transactions_sku_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_sku_id_fkey FOREIGN KEY (sku_id) REFERENCES public.skus(id);


--

-- Name: transactions transactions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--

-- Name: user_sessions user_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: seaply
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict c3QP85HytuRcK9VXINTPw6aS0oD3Nq9I2uA9HUC4QCwqO7cVPLyTSmd2pjIpcbh