package payment

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	dana "github.com/dana-id/dana-go"
	danaconfig "github.com/dana-id/dana-go/config"
	payment_gateway "github.com/dana-id/dana-go/payment_gateway/v1"
)

// DANAGatewayConfig encapsulates runtime configuration for the DANA SDK integration.
type DANAGatewayConfig struct {
	PartnerID      string
	ClientSecret   string
	MerchantID     string
	ShopId         string
	ChannelID      string
	Origin         string
	Environment    string
	PrivateKey     string
	PrivateKeyPath string
	CallbackURL    string
	ReturnURL      string
	DefaultMCC     string
	Debug          bool
}

// DANAGateway implements the payment.Gateway interface using the official DANA SNAP SDK.
type DANAGateway struct {
	client *dana.APIClient
	cfg    DANAGatewayConfig
}

// NewDANAGateway instantiates a DANA gateway using the official SDK.
func NewDANAGateway(cfg DANAGatewayConfig) (*DANAGateway, error) {
	if cfg.PartnerID == "" {
		return nil, fmt.Errorf("dana partner id is required")
	}
	if cfg.MerchantID == "" {
		return nil, fmt.Errorf("dana merchant id is required")
	}

	apiCfg := danaconfig.NewConfiguration()
	apiCfg.Debug = cfg.Debug
	apiCfg.APIKey = &danaconfig.APIKey{
		DANA_ENV:         strings.ToUpper(defaultString(cfg.Environment, "SANDBOX")),
		X_PARTNER_ID:     cfg.PartnerID,
		CHANNEL_ID:       cfg.ChannelID,
		ORIGIN:           cfg.Origin,
		PRIVATE_KEY:      cfg.PrivateKey,
		PRIVATE_KEY_PATH: cfg.PrivateKeyPath,
		CLIENT_SECRET:    cfg.ClientSecret,
	}

	return &DANAGateway{
		client: dana.NewAPIClient(apiCfg),
		cfg: DANAGatewayConfig{
			PartnerID:      cfg.PartnerID,
			ClientSecret:   cfg.ClientSecret,
			MerchantID:     cfg.MerchantID,
			ShopId:         cfg.ShopId,
			ChannelID:      cfg.ChannelID,
			Origin:         cfg.Origin,
			Environment:    strings.ToUpper(defaultString(cfg.Environment, "SANDBOX")),
			PrivateKey:     cfg.PrivateKey,
			PrivateKeyPath: cfg.PrivateKeyPath,
			CallbackURL:    cfg.CallbackURL,
			ReturnURL:      cfg.ReturnURL,
			DefaultMCC:     defaultString(cfg.DefaultMCC, "6012"),
			Debug:          cfg.Debug,
		},
	}, nil
}

func defaultString(val, fallback string) string {
	if strings.TrimSpace(val) != "" {
		return val
	}
	return fallback
}

// GetName identifies the gateway.
func (d *DANAGateway) GetName() string {
	return "DANA_DIRECT"
}

// GetSupportedChannels returns supported payment channels.
func (d *DANAGateway) GetSupportedChannels() []string {
	return []string{"DANA", "QRIS"}
}

// CreatePayment dispatches payment creation based on the requested channel.
func (d *DANAGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	switch strings.ToUpper(req.Channel) {
	case "QRIS":
		return d.createQRISPayment(ctx, req)
	case "DANA":
		fallthrough
	default:
		return d.createRedirectPayment(ctx, req)
	}
}

func (d *DANAGateway) createRedirectPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	money := d.buildMoney(req.Amount, req.Currency)
	urlParams := d.buildURLParamsHosted(req)

	// Use BALANCE pay option for DANA wallet payment (similar to QRIS but with BALANCE)
	payDetail := payment_gateway.NewPayOptionDetail(
		string(payment_gateway.PAYMETHOD_BALANCE_),
		"BALANCE", // PayOption BALANCE (no constant in SDK)
		*money,
	)

	apiReq := payment_gateway.NewCreateOrderByApiRequest([]payment_gateway.PayOptionDetail{}, req.RefID, d.cfg.MerchantID, *money, urlParams)
	apiReq.SetPayOptionDetails([]payment_gateway.PayOptionDetail{*payDetail})
	apiReq.SetValidUpTo(d.formatWIB(d.expiryTime(req.ExpiryDuration)))

	// Use ShopId from config, fallback to MerchantID if not set
	externalStoreId := d.cfg.ShopId
	if externalStoreId == "" {
		externalStoreId = d.cfg.MerchantID
	}
	apiReq.SetExternalStoreId(externalStoreId)

	additional := d.buildAPIAdditionalInfo(req)
	apiReq.SetAdditionalInfo(additional)

	request := payment_gateway.CreateOrderByApiRequestAsCreateOrderRequest(apiReq)
	resp, httpResp, err := d.client.PaymentGatewayAPI.CreateOrder(ctx).CreateOrderRequest(request).Execute()
	if err != nil {
		return nil, err
	}

	// Debug logging
	if d.cfg.Debug {
		fmt.Printf("[DANA BALANCE] Response Code: %s, Message: %s\n", resp.GetResponseCode(), resp.GetResponseMessage())
		fmt.Printf("[DANA BALANCE] Reference No: %s\n", resp.GetReferenceNo())
		fmt.Printf("[DANA BALANCE] Web Redirect URL: %s\n", resp.GetWebRedirectUrl())
		if httpResp != nil {
			fmt.Printf("[DANA BALANCE] HTTP Status: %s\n", httpResp.Status)
		}
	}

	if err := d.ensureSuccess(resp.GetResponseCode(), resp.GetResponseMessage()); err != nil {
		return nil, err
	}

	gatewayRef := resp.GetReferenceNo()
	if gatewayRef == "" {
		gatewayRef = req.RefID
	}

	return &PaymentResponse{
		RefID:        req.RefID,
		GatewayRefID: gatewayRef,
		Channel:      "DANA",
		Amount:       req.Amount,
		Fee:          0,
		TotalAmount:  req.Amount,
		Currency:     money.GetCurrency(),
		Status:       PaymentStatusPending,
		PaymentURL:   resp.GetWebRedirectUrl(),
		ExpiresAt:    d.expiryTime(req.ExpiryDuration),
		CreatedAt:    time.Now(),
		Instructions: []string{
			"1. Tekan tombol 'Bayar dengan DANA' untuk membuka halaman pembayaran.",
			"2. Login ke akun DANA Anda.",
			"3. Pastikan detail transaksi sesuai, lalu konfirmasi.",
			"4. Masukkan PIN DANA untuk menyelesaikan pembayaran.",
		},
	}, nil
}

func (d *DANAGateway) createQRISPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	money := d.buildMoney(req.Amount, req.Currency)
	urlParams := d.buildURLParamsHosted(req)

	// Use QRIS pay option for hosted checkout
	payDetail := payment_gateway.NewPayOptionDetail(
		string(payment_gateway.PAYMETHOD_NETWORK_PAY_),
		string(payment_gateway.PAYOPTION_NETWORK_PAY_PG_QRIS_),
		*money,
	)

	apiReq := payment_gateway.NewCreateOrderByApiRequest([]payment_gateway.PayOptionDetail{}, req.RefID, d.cfg.MerchantID, *money, urlParams)
	apiReq.SetPayOptionDetails([]payment_gateway.PayOptionDetail{*payDetail})
	apiReq.SetValidUpTo(d.formatWIB(d.expiryTime(req.ExpiryDuration)))
	// Use ShopId from config, fallback to MerchantID if not set
	externalStoreId := d.cfg.ShopId
	if externalStoreId == "" {
		externalStoreId = d.cfg.MerchantID
	}
	apiReq.SetExternalStoreId(externalStoreId)

	additional := d.buildAPIAdditionalInfo(req)
	apiReq.SetAdditionalInfo(additional)

	request := payment_gateway.CreateOrderByApiRequestAsCreateOrderRequest(apiReq)
	resp, httpResp, err := d.client.PaymentGatewayAPI.CreateOrder(ctx).CreateOrderRequest(request).Execute()
	if err != nil {
		return nil, err
	}

	// Debug logging
	if d.cfg.Debug {
		fmt.Printf("[DANA QRIS] Response Code: %s, Message: %s\n", resp.GetResponseCode(), resp.GetResponseMessage())
		fmt.Printf("[DANA QRIS] Reference No: %s\n", resp.GetReferenceNo())
		fmt.Printf("[DANA QRIS] Web Redirect URL: %s\n", resp.GetWebRedirectUrl())
		if info := resp.GetAdditionalInfo(); info != (payment_gateway.CreateOrderResponseAdditionalInfo{}) {
			fmt.Printf("[DANA QRIS] PaymentCode: %s\n", info.GetPaymentCode())
		}
		if httpResp != nil {
			fmt.Printf("[DANA QRIS] HTTP Status: %s\n", httpResp.Status)
		}
	}

	if err := d.ensureSuccess(resp.GetResponseCode(), resp.GetResponseMessage()); err != nil {
		return nil, err
	}

	// Get payment code (QRIS string) from additionalInfo
	paymentCode := ""
	if info := resp.GetAdditionalInfo(); info != (payment_gateway.CreateOrderResponseAdditionalInfo{}) {
		paymentCode = info.GetPaymentCode()
	}

	gatewayRef := resp.GetReferenceNo()
	if gatewayRef == "" {
		gatewayRef = req.RefID
	}

	// If paymentCode is empty, DANA might have returned redirect URL instead
	// This happens when merchant is not configured for hosted checkout
	if paymentCode == "" {
		webRedirectUrl := resp.GetWebRedirectUrl()
		if webRedirectUrl != "" {
			// Fall back to redirect-based payment
			fmt.Printf("[DANA QRIS] Warning: No paymentCode returned, falling back to redirect URL\n")
			return &PaymentResponse{
				RefID:        req.RefID,
				GatewayRefID: gatewayRef,
				Channel:      "QRIS",
				Amount:       req.Amount,
				Fee:          0,
				TotalAmount:  req.Amount,
				Currency:     money.GetCurrency(),
				Status:       PaymentStatusPending,
				PaymentURL:   webRedirectUrl,
				ExpiresAt:    d.expiryTime(req.ExpiryDuration),
				CreatedAt:    time.Now(),
				Instructions: []string{
					"1. Klik tombol 'Bayar dengan QRIS' untuk membuka halaman pembayaran.",
					"2. Scan QR code yang ditampilkan menggunakan aplikasi e-wallet.",
					"3. Konfirmasi pembayaran di aplikasi.",
				},
			}, nil
		}
		return nil, fmt.Errorf("dana did not return payment code or redirect URL for QRIS")
	}

	return &PaymentResponse{
		RefID:        req.RefID,
		GatewayRefID: gatewayRef,
		Channel:      "QRIS",
		Amount:       req.Amount,
		Fee:          0,
		TotalAmount:  req.Amount,
		Currency:     money.GetCurrency(),
		Status:       PaymentStatusPending,
		PaymentCode:  paymentCode,
		QRCode:       paymentCode,
		QRCodeURL:    d.qrCodeURL(paymentCode),
		ExpiresAt:    d.expiryTime(req.ExpiryDuration),
		CreatedAt:    time.Now(),
		Instructions: []string{
			"1. Buka aplikasi mobile banking atau e-wallet yang mendukung QRIS.",
			"2. Pilih menu Scan QR / QRIS.",
			"3. Scan kode yang ditampilkan dan konfirmasi pembayaran.",
		},
	}, nil
}

// CheckStatus queries DANA for the latest transaction status.
func (d *DANAGateway) CheckStatus(ctx context.Context, paymentID string) (*PaymentStatus, error) {
	request := payment_gateway.NewQueryPaymentRequest("54", d.cfg.MerchantID)
	request.SetOriginalPartnerReferenceNo(paymentID)
	request.SetOriginalReferenceNo(paymentID)

	resp, _, err := d.client.PaymentGatewayAPI.QueryPayment(ctx).QueryPaymentRequest(*request).Execute()
	if err != nil {
		return nil, err
	}
	if err := d.ensureSuccess(resp.GetResponseCode(), resp.GetResponseMessage()); err != nil {
		return nil, err
	}

	status := d.mapLatestStatus(resp.GetLatestTransactionStatus())
	amount := d.moneyToFloat(resp.GetAmount())
	paidAt := d.parseWIB(resp.GetPaidTime())

	return &PaymentStatus{
		RefID:        resp.GetOriginalPartnerReferenceNo(),
		GatewayRefID: resp.GetOriginalReferenceNo(),
		Status:       status,
		Amount:       amount,
		Fee:          d.moneyToFloat(resp.GetFeeAmount()),
		PaidAt:       paidAt,
		UpdatedAt:    time.Now(),
	}, nil
}

// HealthCheck performs a lightweight connectivity check.
// Note: DANA API doesn't have a root health endpoint, so we just check if the server is reachable.
// A 404 response is acceptable as it means the server is up, just no root endpoint.
func (d *DANAGateway) HealthCheck(ctx context.Context) error {
	serverURL := d.client.GetConfig().Servers[0].URL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		return err
	}

	resp, err := d.client.GetConfig().HTTPClient.Do(req)
	if err != nil {
		// Connection error - server is unreachable
		return fmt.Errorf("dana server unreachable: %w", err)
	}
	defer resp.Body.Close()

	// Accept any response from the server (including 404)
	// as it indicates the server is up and responding
	// Only 5xx errors indicate server-side issues
	if resp.StatusCode >= 500 {
		return fmt.Errorf("dana server error: status %d", resp.StatusCode)
	}

	return nil
}

// Helper methods -------------------------------------------------------------

func (d *DANAGateway) buildMoney(amount float64, currency string) *payment_gateway.Money {
	curr := defaultString(currency, "IDR")
	return payment_gateway.NewMoney(fmt.Sprintf("%.2f", amount), curr)
}

func (d *DANAGateway) buildRedirectAdditionalInfo(req *PaymentRequest) payment_gateway.CreateOrderByRedirectAdditionalInfo {
	env := payment_gateway.NewEnvInfo(
		string(payment_gateway.SOURCEPLATFORM_IPG_),
		string(payment_gateway.TERMINALTYPE_WEB_),
	)
	order := payment_gateway.NewOrderRedirectObject(d.orderTitle(req))
	order.SetScenario("API")

	info := payment_gateway.NewCreateOrderByRedirectAdditionalInfo(d.cfg.DefaultMCC, *env)
	info.SetOrder(*order)
	return *info
}

func (d *DANAGateway) buildAPIAdditionalInfo(req *PaymentRequest) payment_gateway.CreateOrderByApiAdditionalInfo {
	// Use SYSTEM terminal type for hosted/API checkout
	env := payment_gateway.NewEnvInfo(
		string(payment_gateway.SOURCEPLATFORM_IPG_),
		string(payment_gateway.TERMINALTYPE_SYSTEM_),
	)
	// Set orderTerminalType to WEB
	env.SetOrderTerminalType(string(payment_gateway.TERMINALTYPE_WEB_))

	// Create order with scenario "API" for hosted checkout (returns QRIS string)
	order := payment_gateway.NewOrderApiObject(d.orderTitle(req))
	order.SetScenario("API")

	// Add buyer info if available
	if req.CustomerPhone != "" || req.CustomerName != "" {
		buyerID := req.CustomerPhone
		if buyerID == "" {
			buyerID = req.RefID
		}
		buyer := payment_gateway.NewBuyer()
		buyer.SetExternalUserId("BUYER-" + buyerID)
		buyer.SetExternalUserType("BUYER-" + req.CustomerPhone)
		order.SetBuyer(*buyer)
	}

	info := payment_gateway.NewCreateOrderByApiAdditionalInfo(d.cfg.DefaultMCC, *env)
	info.SetOrder(*order)
	return *info
}

func (d *DANAGateway) buildURLParams(req *PaymentRequest) []payment_gateway.UrlParam {
	var params []payment_gateway.UrlParam

	notify := defaultString(req.CallbackURL, d.cfg.CallbackURL)
	if notify != "" {
		params = append(params, *payment_gateway.NewUrlParam(notify, string(payment_gateway.TYPE_NOTIFICATION_), "false"))
	}

	ret := defaultString(req.SuccessURL, d.cfg.ReturnURL)
	if ret != "" {
		params = append(params, *payment_gateway.NewUrlParam(ret, string(payment_gateway.TYPE_PAY_RETURN_), "false"))
	}

	if len(params) == 0 {
		params = append(params, *payment_gateway.NewUrlParam("https://"+defaultString(d.cfg.Origin, "seaply.co"), string(payment_gateway.TYPE_NOTIFICATION_), "false"))
	}
	return params
}

// buildURLParamsHosted creates URL params for hosted/API checkout with isDeeplink="Y"
func (d *DANAGateway) buildURLParamsHosted(req *PaymentRequest) []payment_gateway.UrlParam {
	var params []payment_gateway.UrlParam

	// Return URL first (PAY_RETURN)
	ret := defaultString(req.SuccessURL, d.cfg.ReturnURL)
	if ret != "" {
		params = append(params, *payment_gateway.NewUrlParam(ret, string(payment_gateway.TYPE_PAY_RETURN_), "Y"))
	}

	// Notification URL (callback)
	notify := defaultString(req.CallbackURL, d.cfg.CallbackURL)
	if notify != "" {
		params = append(params, *payment_gateway.NewUrlParam(notify, string(payment_gateway.TYPE_NOTIFICATION_), "Y"))
	}

	return params
}

func (d *DANAGateway) orderTitle(req *PaymentRequest) string {
	if strings.TrimSpace(req.Description) != "" {
		return req.Description
	}
	return fmt.Sprintf("Order %s", req.RefID)
}

func (d *DANAGateway) expiryTime(duration time.Duration) time.Time {
	if duration <= 0 {
		return time.Now().Add(15 * time.Minute)
	}
	return time.Now().Add(duration)
}

func (d *DANAGateway) formatWIB(t time.Time) string {
	wib := time.FixedZone("WIB", 7*3600)
	return t.In(wib).Format("2006-01-02T15:04:05-07:00")
}

func (d *DANAGateway) parseWIB(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse("2006-01-02T15:04:05-07:00", value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func (d *DANAGateway) ensureSuccess(code, message string) error {
	// DANA V2 API success codes:
	// - 2005400: Success code seen in CreateOrder response
	// - Successful: Message text variations
	if code == "2005400" || strings.EqualFold(message, "Successful") {
		return nil
	}
	return fmt.Errorf("dana error %s: %s", code, message)
}

func (d *DANAGateway) mapLatestStatus(code string) string {
	switch code {
	case "00":
		return PaymentStatusPaid
	case "01", "02":
		return PaymentStatusPending
	case "05":
		return PaymentStatusFailed
	default:
		return PaymentStatusPending
	}
}

func (d *DANAGateway) moneyToFloat(m payment_gateway.Money) float64 {
	if (m == payment_gateway.Money{}) {
		return 0
	}
	val, err := strconv.ParseFloat(m.GetValue(), 64)
	if err != nil {
		return 0
	}
	return val
}

func (d *DANAGateway) qrCodeURL(data string) string {
	if data == "" {
		return ""
	}
	return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s", url.QueryEscape(data))
}
