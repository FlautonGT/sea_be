package admin

import "net/http"

// Admin Management Handlers
func HandleGetAdmins(deps *Dependencies) http.HandlerFunc {
	return HandleGetAdminsImpl(deps)
}

func HandleGetAdmin(deps *Dependencies) http.HandlerFunc {
	return HandleGetAdminImpl(deps)
}

func HandleCreateAdmin(deps *Dependencies) http.HandlerFunc {
	return HandleCreateAdminImpl(deps)
}

func HandleUpdateAdmin(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateAdminImpl(deps)
}

func HandleDeleteAdmin(deps *Dependencies) http.HandlerFunc {
	return HandleDeleteAdminImpl(deps)
}

func HandleGetRoles(deps *Dependencies) http.HandlerFunc {
	return HandleGetRolesImpl(deps)
}

func HandleUpdateRolePermissions(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateRolePermissionsImpl(deps)
}

// Provider Handlers
func HandleGetProviders(deps *Dependencies) http.HandlerFunc {
	return HandleGetProvidersImpl(deps)
}

func HandleGetProvider(deps *Dependencies) http.HandlerFunc {
	return HandleGetProviderImpl(deps)
}

func HandleCreateProvider(deps *Dependencies) http.HandlerFunc {
	return HandleCreateProviderImpl(deps)
}

func HandleUpdateProvider(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateProviderImpl(deps)
}

func HandleDeleteProvider(deps *Dependencies) http.HandlerFunc {
	return HandleDeleteProviderImpl(deps)
}

func HandleTestProvider(deps *Dependencies) http.HandlerFunc {
	return HandleTestProviderImpl(deps)
}

func HandleSyncProvider(deps *Dependencies) http.HandlerFunc {
	return HandleSyncProviderImpl(deps)
}

// Payment Channel Handlers
func HandleAdminGetPaymentChannels(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetPaymentChannelsImplAdmin(deps)
}

func HandleGetChannelAssignments(deps *Dependencies) http.HandlerFunc {
	return HandleGetChannelAssignmentsImpl(deps)
}

func HandleGetPaymentChannel(deps *Dependencies) http.HandlerFunc {
	return HandleGetPaymentChannelImpl(deps)
}

func HandleCreatePaymentChannel(deps *Dependencies) http.HandlerFunc {
	return HandleCreatePaymentChannelImpl(deps)
}

func HandleUpdatePaymentChannel(deps *Dependencies) http.HandlerFunc {
	return HandleUpdatePaymentChannelImpl(deps)
}

func HandleUpdateChannelAssignment(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateChannelAssignmentImpl(deps)
}

func HandleDeletePaymentChannel(deps *Dependencies) http.HandlerFunc {
	return HandleDeletePaymentChannelImpl(deps)
}

func HandleGetPaymentChannelCategories(deps *Dependencies) http.HandlerFunc {
	return HandleGetPaymentChannelCategoriesImpl(deps)
}

func HandleCreatePaymentChannelCategory(deps *Dependencies) http.HandlerFunc {
	return HandleCreatePaymentChannelCategoryImpl(deps)
}

func HandleUpdatePaymentChannelCategory(deps *Dependencies) http.HandlerFunc {
	return HandleUpdatePaymentChannelCategoryImpl(deps)
}

func HandleDeletePaymentChannelCategory(deps *Dependencies) http.HandlerFunc {
	return HandleDeletePaymentChannelCategoryImpl(deps)
}

// Product Handlers
func HandleAdminGetProducts(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetProductsImpl(deps)
}

func HandleAdminGetProduct(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetProductImpl(deps)
}

func HandleCreateProduct(deps *Dependencies) http.HandlerFunc {
	return HandleCreateProductImpl(deps)
}

func HandleUpdateProduct(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateProductImpl(deps)
}

func HandleDeleteProduct(deps *Dependencies) http.HandlerFunc {
	return HandleDeleteProductImpl(deps)
}

func HandleAdminGetFields(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetFieldsImpl(deps)
}

func HandleUpdateFields(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateFieldsImpl(deps)
}

// Category Handlers
func HandleAdminGetCategories(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetCategoriesImpl(deps)
}

func HandleCreateCategory(deps *Dependencies) http.HandlerFunc {
	return HandleCreateCategoryImpl(deps)
}

func HandleUpdateCategory(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateCategoryImpl(deps)
}

func HandleDeleteCategory(deps *Dependencies) http.HandlerFunc {
	return HandleDeleteCategoryImpl(deps)
}

// Section Handlers
func HandleAdminGetSections(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetSectionsImpl(deps)
}

func HandleCreateSection(deps *Dependencies) http.HandlerFunc {
	return HandleCreateSectionImpl(deps)
}

func HandleUpdateSection(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateSectionImpl(deps)
}

func HandleDeleteSection(deps *Dependencies) http.HandlerFunc {
	return HandleDeleteSectionImpl(deps)
}

func HandleAssignSectionProducts(deps *Dependencies) http.HandlerFunc {
	return HandleAssignSectionProductsImpl(deps)
}

// SKU Handlers
func HandleAdminGetSKUs(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetSKUsImpl(deps)
}

func HandleAdminGetSKU(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetSKUImpl(deps)
}

func HandleCreateSKU(deps *Dependencies) http.HandlerFunc {
	return HandleCreateSKUImpl(deps)
}

func HandleUpdateSKU(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateSKUImpl(deps)
}

func HandleDeleteSKU(deps *Dependencies) http.HandlerFunc {
	return HandleDeleteSKUImpl(deps)
}

func HandleBulkUpdatePrice(deps *Dependencies) http.HandlerFunc {
	return HandleBulkUpdatePriceImpl(deps)
}

func HandleSyncSKUs(deps *Dependencies) http.HandlerFunc {
	return HandleSyncSKUsImpl(deps)
}

func HandleAdminGetSKUImages(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetSKUImagesImpl(deps)
}

// Transaction Admin Handlers
func HandleAdminGetTransactions(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetTransactionsImpl(deps)
}

func HandleAdminGetTransaction(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetTransactionImpl(deps)
}

func HandleUpdateTransactionStatus(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateTransactionStatusImpl(deps)
}

func HandleRefundTransaction(deps *Dependencies) http.HandlerFunc {
	return HandleRefundTransactionImpl(deps)
}

func HandleRetryTransaction(deps *Dependencies) http.HandlerFunc {
	return HandleRetryTransactionImpl(deps)
}

func HandleManualProcess(deps *Dependencies) http.HandlerFunc {
	return HandleManualProcessImpl(deps)
}

// User Admin Handlers
func HandleAdminGetUsers(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetUsersImpl(deps)
}

func HandleAdminGetUser(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetUserImpl(deps)
}

func HandleUpdateUserStatus(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateUserStatusImpl(deps)
}

func HandleAdjustBalance(deps *Dependencies) http.HandlerFunc {
	return HandleAdjustBalanceImpl(deps)
}

func HandleUserTransactions(deps *Dependencies) http.HandlerFunc {
	return HandleUserTransactionsImpl(deps)
}

func HandleUserMutations(deps *Dependencies) http.HandlerFunc {
	return HandleUserMutationsImpl(deps)
}

// Promo Handlers
func HandleAdminGetPromos(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetPromosImpl(deps)
}

func HandleAdminGetPromo(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetPromoImpl(deps)
}

func HandleCreatePromo(deps *Dependencies) http.HandlerFunc {
	return HandleCreatePromoImpl(deps)
}

func HandleUpdatePromo(deps *Dependencies) http.HandlerFunc {
	return HandleUpdatePromoImpl(deps)
}

func HandleDeletePromo(deps *Dependencies) http.HandlerFunc {
	return HandleDeletePromoImpl(deps)
}

func HandleGetPromoStats(deps *Dependencies) http.HandlerFunc {
	return HandleGetPromoStatsImpl(deps)
}

// Banner/Popup Handlers
func HandleAdminGetBanners(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetBannersImpl(deps)
}

func HandleCreateBanner(deps *Dependencies) http.HandlerFunc {
	return HandleCreateBannerImpl(deps)
}

func HandleUpdateBanner(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateBannerImpl(deps)
}

func HandleDeleteBanner(deps *Dependencies) http.HandlerFunc {
	return HandleDeleteBannerImpl(deps)
}

func HandleAdminGetPopups(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetPopupsImpl(deps)
}

func HandleCreatePopup(deps *Dependencies) http.HandlerFunc {
	return HandleCreatePopupImpl(deps)
}

func HandleUpdatePopup(deps *Dependencies) http.HandlerFunc {
	return HandleUpdatePopupImpl(deps)
}

// Deposit Admin Handlers
func HandleAdminGetDeposits(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetDepositsImpl(deps)
}

func HandleAdminGetDeposit(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetDepositImpl(deps)
}

func HandleConfirmDeposit(deps *Dependencies) http.HandlerFunc {
	return HandleConfirmDepositImpl(deps)
}

func HandleCancelDeposit(deps *Dependencies) http.HandlerFunc {
	return HandleCancelDepositImpl(deps)
}

func HandleRefundDeposit(deps *Dependencies) http.HandlerFunc {
	return HandleRefundDepositImpl(deps)
}

// Invoice Handlers
func HandleAdminGetInvoices(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetInvoicesImpl(deps)
}

func HandleSearchInvoice(deps *Dependencies) http.HandlerFunc {
	return HandleSearchInvoiceImpl(deps)
}

func HandleSendInvoiceEmail(deps *Dependencies) http.HandlerFunc {
	return HandleSendInvoiceEmailImpl(deps)
}

// Report Handlers
func HandleGetDashboard(deps *Dependencies) http.HandlerFunc {
	return HandleGetDashboardImpl(deps)
}

func HandleGetRevenueReport(deps *Dependencies) http.HandlerFunc {
	return HandleGetRevenueReportImpl(deps)
}

func HandleGetTransactionReport(deps *Dependencies) http.HandlerFunc {
	return HandleGetTransactionReportImpl(deps)
}

func HandleGetProductReport(deps *Dependencies) http.HandlerFunc {
	return HandleGetProductReportImpl(deps)
}

func HandleGetProviderReport(deps *Dependencies) http.HandlerFunc {
	return HandleGetProviderReportImpl(deps)
}

func HandleExportReport(deps *Dependencies) http.HandlerFunc {
	return HandleExportReportImpl(deps)
}

func HandleGetExportStatus(deps *Dependencies) http.HandlerFunc {
	return HandleGetExportStatusImpl(deps)
}

// Audit Log Handlers
func HandleGetAuditLogs(deps *Dependencies) http.HandlerFunc {
	return HandleGetAuditLogsImpl(deps)
}

// Settings Handlers
func HandleGetSettings(deps *Dependencies) http.HandlerFunc {
	return HandleGetSettingsImpl(deps)
}

func HandleUpdateSettings(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateSettingsImpl(deps)
}

func HandleGetContactSettings(deps *Dependencies) http.HandlerFunc {
	return HandleGetContactSettingsImpl(deps)
}

func HandleUpdateContactSettings(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateContactSettingsImpl(deps)
}

// Region Handlers
func HandleAdminGetRegions(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetRegionsImpl(deps)
}

func HandleCreateRegion(deps *Dependencies) http.HandlerFunc {
	return HandleCreateRegionImpl(deps)
}

func HandleUpdateRegion(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateRegionImpl(deps)
}

func HandleDeleteRegion(deps *Dependencies) http.HandlerFunc {
	return HandleDeleteRegionImpl(deps)
}

// Language Handlers
func HandleAdminGetLanguages(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetLanguagesImpl(deps)
}

func HandleCreateLanguage(deps *Dependencies) http.HandlerFunc {
	return HandleCreateLanguageImpl(deps)
}

func HandleUpdateLanguage(deps *Dependencies) http.HandlerFunc {
	return HandleUpdateLanguageImpl(deps)
}

func HandleDeleteLanguage(deps *Dependencies) http.HandlerFunc {
	return HandleDeleteLanguageImpl(deps)
}

// Admin Auth Handlers
func HandleAdminLogin(deps *Dependencies) http.HandlerFunc {
	return HandleAdminLoginImpl(deps)
}

func HandleAdminVerifyMFA(deps *Dependencies) http.HandlerFunc {
	return HandleAdminVerifyMFAImpl(deps)
}

func HandleAdminRefreshToken(deps *Dependencies) http.HandlerFunc {
	return HandleAdminRefreshTokenImpl(deps)
}

func HandleAdminLogout(deps *Dependencies) http.HandlerFunc {
	return HandleAdminLogoutImpl(deps)
}
