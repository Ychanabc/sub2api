package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

func TestBEPUSDTCreatePaymentUsesJSONAPIAndSignature(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/order/create-transaction" {
			t.Fatalf("path = %q, want create-transaction", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Fatalf("content-type = %q, want application/json", ct)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["signature"] != bepusdtSignAny(payload, "token-secret") {
			t.Fatalf("invalid signature in payload: %#v", payload)
		}
		_, _ = w.Write([]byte(`{"status_code":200,"message":"success","data":{"trade_id":"be_1","order_id":"order_usdt_1","payment_url":"https://pay.example/usdt","token":"TR7N"}}`))
	}))
	defer srv.Close()

	prov, err := NewBEPUSDT("inst_usdt", map[string]string{
		"apiToken":  "token-secret",
		"apiBase":   srv.URL,
		"notifyUrl": "https://merchant.example/notify",
		"returnUrl": "https://merchant.example/return",
		"tradeType": "usdt.trc20",
		"fiat":      "CNY",
	})
	if err != nil {
		t.Fatalf("NewBEPUSDT: %v", err)
	}

	resp, err := prov.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID:     "order_usdt_1",
		Amount:      "12.34",
		PaymentType: payment.TypeUSDT,
		Subject:     "USDT top up",
		ClientIP:    "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}

	if resp.TradeNo != "be_1" || resp.PayURL != "https://pay.example/usdt" {
		t.Fatalf("unexpected create response: %#v", resp)
	}
	if payload["amount"] != float64(12.34) {
		t.Fatalf("amount should be numeric 12.34, got %#v", payload["amount"])
	}
}

func TestBEPUSDTVerifyNotification(t *testing.T) {
	prov, err := NewBEPUSDT("inst_usdt", map[string]string{
		"apiToken":  "token-secret",
		"apiBase":   "https://pay.example",
		"notifyUrl": "https://merchant.example/notify",
		"returnUrl": "https://merchant.example/return",
	})
	if err != nil {
		t.Fatalf("NewBEPUSDT: %v", err)
	}
	payload := map[string]any{
		"trade_id":             "be_1",
		"order_id":             "order_usdt_1",
		"amount":               28.88,
		"actual_amount":        4.25,
		"token":                "TR7N",
		"block_transaction_id": "tx_1",
		"status":               2,
	}
	payload["signature"] = bepusdtSignAny(payload, "token-secret")
	body, _ := json.Marshal(payload)

	notify, err := prov.VerifyNotification(context.Background(), string(body), nil)
	if err != nil {
		t.Fatalf("VerifyNotification: %v", err)
	}
	if notify.OrderID != "order_usdt_1" || notify.TradeNo != "be_1" || notify.Status != payment.ProviderStatusSuccess {
		t.Fatalf("unexpected notification: %#v", notify)
	}
}

func TestEasyPaySupportedTypesExcludesUSDT(t *testing.T) {
	prov := &EasyPay{}
	for _, typ := range prov.SupportedTypes() {
		if typ == payment.TypeUSDT {
			t.Fatalf("EasyPay SupportedTypes() should not include %q, got %v", payment.TypeUSDT, prov.SupportedTypes())
		}
	}
}

func TestBEPUSDTSupportedTypesIncludesUSDT(t *testing.T) {
	prov := &BEPUSDT{}
	for _, typ := range prov.SupportedTypes() {
		if typ == payment.TypeUSDT {
			return
		}
	}
	t.Fatalf("BEPUSDT SupportedTypes() should include %q, got %v", payment.TypeUSDT, prov.SupportedTypes())
}
