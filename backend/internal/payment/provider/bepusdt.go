package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

const (
	bepusdtHTTPTimeout     = 10 * time.Second
	bepusdtMaxResponseSize = 1 << 20
	bepusdtStatusPaid      = 2
	bepusdtStatusSuccess   = 200
	bepusdtLegacySuccess   = 1
	bepusdtCreatePath      = "/api/v1/order/create-transaction"
	bepusdtQueryPath       = "/api/v1/order/order-status"
	bepusdtDefaultTrade    = "usdt.trc20"
)

// BEPUSDT implements the BEpusdt JSON API as an independent USDT provider.
type BEPUSDT struct {
	instanceID string
	config     map[string]string
	httpClient *http.Client
}

func NewBEPUSDT(instanceID string, config map[string]string) (*BEPUSDT, error) {
	for _, k := range []string{"apiBase", "apiToken", "notifyUrl", "returnUrl"} {
		if strings.TrimSpace(config[k]) == "" {
			return nil, fmt.Errorf("usdt config missing required key: %s", k)
		}
	}
	cfg := make(map[string]string, len(config))
	for k, v := range config {
		cfg[k] = v
	}
	cfg["apiBase"] = normalizeBEPUSDTBase(cfg["apiBase"])
	return &BEPUSDT{
		instanceID: instanceID,
		config:     cfg,
		httpClient: &http.Client{Timeout: bepusdtHTTPTimeout},
	}, nil
}

func (b *BEPUSDT) Name() string        { return "BEpusdt" }
func (b *BEPUSDT) ProviderKey() string { return payment.TypeUSDT }
func (b *BEPUSDT) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeUSDT}
}

func (b *BEPUSDT) MerchantIdentityMetadata() map[string]string {
	if b == nil {
		return nil
	}
	if apiBase := strings.TrimSpace(b.config["apiBase"]); apiBase != "" {
		return map[string]string{"apiBase": apiBase}
	}
	return nil
}

func (b *BEPUSDT) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	notifyURL, returnURL := b.resolveURLs(req)
	params := map[string]any{
		"order_id":     req.OrderID,
		"amount":       bepusdtAmount(req.Amount),
		"notify_url":   notifyURL,
		"redirect_url": returnURL,
		"trade_type":   firstNonEmpty(b.config["tradeType"], bepusdtDefaultTrade),
		"fiat":         firstNonEmpty(b.config["fiat"], payment.DefaultPaymentCurrency),
		"name":         req.Subject,
	}
	if address := strings.TrimSpace(b.config["address"]); address != "" {
		params["address"] = address
	}
	if timeout := strings.TrimSpace(b.config["timeout"]); timeout != "" {
		if seconds, err := strconv.Atoi(timeout); err == nil && seconds > 0 {
			params["timeout"] = seconds
		}
	}
	if rate := strings.TrimSpace(b.config["rate"]); rate != "" {
		params["rate"] = rate
	}
	params["signature"] = bepusdtSignAny(params, b.config["apiToken"])

	body, err := b.postJSON(ctx, b.endpoint(firstNonEmpty(b.config["createPath"], bepusdtCreatePath)), params)
	if err != nil {
		return nil, fmt.Errorf("bepusdt create: %w", err)
	}
	var resp struct {
		StatusCode int    `json:"status_code"`
		Code       int    `json:"code"`
		Message    string `json:"message"`
		Msg        string `json:"msg"`
		Data       struct {
			TradeID    string `json:"trade_id"`
			OrderID    string `json:"order_id"`
			PaymentURL string `json:"payment_url"`
			Token      string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("bepusdt parse: %w", err)
	}
	if !bepusdtResponseOK(resp.StatusCode, resp.Code) {
		msg := firstNonEmpty(resp.Message, resp.Msg, summarizeEasyPayResponse(body))
		return nil, fmt.Errorf("bepusdt error: %s", msg)
	}
	return &payment.CreatePaymentResponse{
		TradeNo: firstNonEmpty(resp.Data.TradeID, resp.Data.OrderID, req.OrderID),
		PayURL:  resp.Data.PaymentURL,
		QRCode:  firstNonEmpty(resp.Data.PaymentURL, resp.Data.Token),
	}, nil
}

func (b *BEPUSDT) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	params := map[string]any{"trade_id": tradeNo}
	params["signature"] = bepusdtSignAny(params, b.config["apiToken"])
	body, err := b.postJSON(ctx, b.endpoint(firstNonEmpty(b.config["queryPath"], bepusdtQueryPath)), params)
	if err != nil {
		return nil, fmt.Errorf("bepusdt query: %w", err)
	}
	var resp struct {
		StatusCode int `json:"status_code"`
		Code       int `json:"code"`
		Data       struct {
			TradeID      string `json:"trade_id"`
			OrderID      string `json:"order_id"`
			Amount       string `json:"amount"`
			ActualAmount string `json:"actual_amount"`
			Status       int    `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("bepusdt parse query: %w", err)
	}
	status := payment.ProviderStatusPending
	if resp.Data.Status == bepusdtStatusPaid {
		status = payment.ProviderStatusPaid
	}
	amount, _ := strconv.ParseFloat(firstNonEmpty(resp.Data.Amount, resp.Data.ActualAmount), 64)
	return &payment.QueryOrderResponse{
		TradeNo:  firstNonEmpty(resp.Data.TradeID, resp.Data.OrderID, tradeNo),
		Status:   status,
		Amount:   amount,
		Metadata: b.MerchantIdentityMetadata(),
	}, nil
}

func (b *BEPUSDT) VerifyNotification(_ context.Context, rawBody string, _ map[string]string) (*payment.PaymentNotification, error) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(rawBody), &payload); err != nil {
		return nil, fmt.Errorf("parse bepusdt notify: %w", err)
	}
	sign := fmt.Sprint(payload["signature"])
	if strings.TrimSpace(sign) == "" {
		return nil, fmt.Errorf("missing signature")
	}
	if !hmac.Equal([]byte(bepusdtSignAny(payload, b.config["apiToken"])), []byte(sign)) {
		return nil, fmt.Errorf("invalid signature")
	}
	statusCode, _ := strconv.Atoi(bepusdtString(payload["status"]))
	status := payment.ProviderStatusFailed
	if statusCode == bepusdtStatusPaid {
		status = payment.ProviderStatusSuccess
	}
	amount, _ := strconv.ParseFloat(firstNonEmpty(bepusdtString(payload["amount"]), bepusdtString(payload["actual_amount"])), 64)
	orderID := bepusdtString(payload["order_id"])
	return &payment.PaymentNotification{
		TradeNo:  firstNonEmpty(bepusdtString(payload["trade_id"]), orderID),
		OrderID:  orderID,
		Amount:   amount,
		Status:   status,
		RawData:  rawBody,
		Metadata: b.MerchantIdentityMetadata(),
	}, nil
}

func (b *BEPUSDT) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, fmt.Errorf("bepusdt refund is not supported")
}

func (b *BEPUSDT) resolveURLs(req payment.CreatePaymentRequest) (string, string) {
	notifyURL := req.NotifyURL
	if notifyURL == "" {
		notifyURL = b.config["notifyUrl"]
	}
	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = b.config["returnUrl"]
	}
	return notifyURL, returnURL
}

func (b *BEPUSDT) endpoint(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimRight(b.config["apiBase"], "/") + path
}

func (b *BEPUSDT) postJSON(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
	payload, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := b.httpClient
	if client == nil {
		client = &http.Client{Timeout: bepusdtHTTPTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, bepusdtMaxResponseSize))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, summarizeEasyPayResponse(body))
	}
	return body, nil
}

func bepusdtResponseOK(statusCode, legacyCode int) bool {
	return statusCode == bepusdtStatusSuccess || (statusCode == 0 && legacyCode == bepusdtLegacySuccess)
}

func bepusdtAmount(raw string) any {
	d, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return raw
	}
	return d
}

func bepusdtSignAny(params map[string]any, token string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "signature" || bepusdtEmpty(v) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			_ = buf.WriteByte('&')
		}
		_, _ = buf.WriteString(k + "=" + bepusdtString(params[k]))
	}
	_, _ = buf.WriteString(token)
	hash := md5.Sum([]byte(buf.String()))
	return hex.EncodeToString(hash[:])
}

func bepusdtEmpty(v any) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s) == ""
	}
	return false
}

func bepusdtString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case json.Number:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}

func normalizeBEPUSDTBase(apiBase string) string {
	base := strings.TrimSpace(apiBase)
	if base == "" {
		return ""
	}
	if parsed, err := url.Parse(base); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		parsed.RawQuery = ""
		parsed.Fragment = ""
		parsed.RawPath = ""
		return strings.TrimRight(parsed.String(), "/")
	}
	return strings.TrimRight(base, "/")
}
