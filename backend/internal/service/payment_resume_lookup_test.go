//go:build unit

package service

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

type paymentResumeLookupProvider struct {
	queryCount int
}

type paymentResumeReturnProvider struct {
	queryCount   int
	verifyCount  int
	verifiedRaw  string
	notification *payment.PaymentNotification
}

func (p *paymentResumeLookupProvider) Name() string { return "resume-lookup-provider" }

func (p *paymentResumeLookupProvider) ProviderKey() string { return payment.TypeAlipay }

func (p *paymentResumeLookupProvider) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeAlipay}
}

func (p *paymentResumeLookupProvider) CreatePayment(context.Context, payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	panic("unexpected call")
}

func (p *paymentResumeLookupProvider) QueryOrder(context.Context, string) (*payment.QueryOrderResponse, error) {
	p.queryCount++
	return &payment.QueryOrderResponse{Status: payment.ProviderStatusPending}, nil
}

func (p *paymentResumeLookupProvider) VerifyNotification(context.Context, string, map[string]string) (*payment.PaymentNotification, error) {
	panic("unexpected call")
}

func (p *paymentResumeLookupProvider) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	panic("unexpected call")
}

func (p *paymentResumeReturnProvider) Name() string { return "resume-return-provider" }

func (p *paymentResumeReturnProvider) ProviderKey() string { return payment.TypeEasyPay }

func (p *paymentResumeReturnProvider) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeAlipay}
}

func (p *paymentResumeReturnProvider) CreatePayment(context.Context, payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	panic("unexpected call")
}

func (p *paymentResumeReturnProvider) QueryOrder(context.Context, string) (*payment.QueryOrderResponse, error) {
	p.queryCount++
	return &payment.QueryOrderResponse{Status: payment.ProviderStatusPending}, nil
}

func (p *paymentResumeReturnProvider) VerifyNotification(_ context.Context, rawBody string, _ map[string]string) (*payment.PaymentNotification, error) {
	p.verifyCount++
	p.verifiedRaw = rawBody
	return p.notification, nil
}

func (p *paymentResumeReturnProvider) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	panic("unexpected call")
}

func TestGetPublicOrderByResumeTokenReturnsMatchingOrder(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("resume@example.com").
		SetPasswordHash("hash").
		SetUsername("resume-user").
		Save(ctx)
	require.NoError(t, err)

	instanceID := "12"
	providerKey := payment.TypeEasyPay
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("RESUME-ORDER").
		SetOutTradeNo("sub2_resume_lookup").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-1").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		SetProviderInstanceID(instanceID).
		SetProviderKey(providerKey).
		Save(ctx)
	require.NoError(t, err)

	resumeSvc := NewPaymentResumeService([]byte("0123456789abcdef0123456789abcdef"))
	token, err := resumeSvc.CreateToken(ResumeTokenClaims{
		OrderID:            order.ID,
		UserID:             user.ID,
		ProviderInstanceID: instanceID,
		ProviderKey:        providerKey,
		PaymentType:        payment.TypeAlipay,
		CanonicalReturnURL: "https://app.example.com/payment/result",
	})
	require.NoError(t, err)

	svc := &PaymentService{
		entClient:     client,
		resumeService: resumeSvc,
	}

	got, err := svc.GetPublicOrderByResumeToken(ctx, token)
	require.NoError(t, err)
	require.Equal(t, order.ID, got.ID)
}

func TestGetPublicOrderByResumeTokenRejectsSnapshotMismatch(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("resume-mismatch@example.com").
		SetPasswordHash("hash").
		SetUsername("resume-mismatch-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("RESUME-MISMATCH").
		SetOutTradeNo("sub2_resume_lookup_mismatch").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-2").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		SetProviderInstanceID("12").
		SetProviderKey(payment.TypeEasyPay).
		Save(ctx)
	require.NoError(t, err)

	resumeSvc := NewPaymentResumeService([]byte("0123456789abcdef0123456789abcdef"))
	token, err := resumeSvc.CreateToken(ResumeTokenClaims{
		OrderID:            order.ID,
		UserID:             user.ID,
		ProviderInstanceID: "99",
		ProviderKey:        payment.TypeEasyPay,
		PaymentType:        payment.TypeAlipay,
		CanonicalReturnURL: "https://app.example.com/payment/result",
	})
	require.NoError(t, err)

	svc := &PaymentService{
		entClient:     client,
		resumeService: resumeSvc,
	}

	_, err = svc.GetPublicOrderByResumeToken(ctx, token)
	require.Error(t, err)
	require.Equal(t, "INVALID_RESUME_TOKEN", infraerrors.Reason(err))
}

func TestGetPublicOrderByResumeTokenUsesSnapshotAuthorityWhenColumnsDiffer(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("resume-snapshot-authority@example.com").
		SetPasswordHash("hash").
		SetUsername("resume-snapshot-authority-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("RESUME-SNAPSHOT-AUTHORITY").
		SetOutTradeNo("sub2_resume_snapshot_authority").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-snapshot-authority").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		SetProviderInstanceID("legacy-column-instance").
		SetProviderKey(payment.TypeAlipay).
		SetProviderSnapshot(map[string]any{
			"schema_version":       2,
			"provider_instance_id": "snapshot-instance",
			"provider_key":         payment.TypeEasyPay,
		}).
		Save(ctx)
	require.NoError(t, err)

	resumeSvc := NewPaymentResumeService([]byte("0123456789abcdef0123456789abcdef"))
	token, err := resumeSvc.CreateToken(ResumeTokenClaims{
		OrderID:            order.ID,
		UserID:             user.ID,
		ProviderInstanceID: "snapshot-instance",
		ProviderKey:        payment.TypeEasyPay,
		PaymentType:        payment.TypeAlipay,
		CanonicalReturnURL: "https://app.example.com/payment/result",
	})
	require.NoError(t, err)

	svc := &PaymentService{
		entClient:     client,
		resumeService: resumeSvc,
	}

	got, err := svc.GetPublicOrderByResumeToken(ctx, token)
	require.NoError(t, err)
	require.Equal(t, order.ID, got.ID)
}

func TestGetPublicOrderByResumeTokenChecksUpstreamForPendingOrder(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("resume-refresh@example.com").
		SetPasswordHash("hash").
		SetUsername("resume-refresh-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("RESUME-PENDING").
		SetOutTradeNo("sub2_resume_lookup_pending").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-pending").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	resumeSvc := NewPaymentResumeService([]byte("0123456789abcdef0123456789abcdef"))
	token, err := resumeSvc.CreateToken(ResumeTokenClaims{
		OrderID:            order.ID,
		UserID:             user.ID,
		PaymentType:        payment.TypeAlipay,
		CanonicalReturnURL: "https://app.example.com/payment/result",
	})
	require.NoError(t, err)

	registry := payment.NewRegistry()
	provider := &paymentResumeLookupProvider{}
	registry.Register(provider)

	svc := &PaymentService{
		entClient:       client,
		registry:        registry,
		resumeService:   resumeSvc,
		providersLoaded: true,
	}

	got, err := svc.GetPublicOrderByResumeToken(ctx, token)
	require.NoError(t, err)
	require.Equal(t, order.ID, got.ID)
	require.Equal(t, 1, provider.queryCount)
}

func TestGetPublicOrderByResumeTokenAppliesSignedProviderReturn(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("resume-return@example.com").
		SetPasswordHash("hash").
		SetUsername("resume-return-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(10).
		SetPayAmount(10).
		SetFeeRate(0).
		SetRechargeCode("RESUME-RETURN").
		SetOutTradeNo("sub2_resume_return_paid").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusCancelled).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	resumeSvc := NewPaymentResumeService([]byte("0123456789abcdef0123456789abcdef"))
	token, err := resumeSvc.CreateToken(ResumeTokenClaims{
		OrderID:            order.ID,
		UserID:             user.ID,
		PaymentType:        payment.TypeAlipay,
		CanonicalReturnURL: "https://app.example.com/payment/result",
	})
	require.NoError(t, err)

	userRepo := &mockUserRepo{
		getByIDUser: &User{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Balance:  0,
		},
	}
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		require.Equal(t, user.ID, id)
		userRepo.getByIDUser.Balance += amount
		return nil
	}
	redeemRepo := &paymentOrderLifecycleRedeemRepo{
		codesByCode: map[string]*RedeemCode{
			order.RechargeCode: {
				ID:     1,
				Code:   order.RechargeCode,
				Type:   RedeemTypeBalance,
				Value:  order.Amount,
				Status: StatusUnused,
			},
		},
	}
	redeemService := NewRedeemService(
		redeemRepo,
		userRepo,
		nil,
		nil,
		nil,
		client,
		nil,
		nil,
	)

	provider := &paymentResumeReturnProvider{
		notification: &payment.PaymentNotification{
			TradeNo: "upstream-return-1",
			OrderID: order.OutTradeNo,
			Amount:  10,
			Status:  payment.NotificationStatusSuccess,
			Metadata: map[string]string{
				"pid": "pid-1",
			},
		},
	}
	registry := payment.NewRegistry()
	registry.Register(provider)
	svc := &PaymentService{
		entClient:       client,
		registry:        registry,
		redeemService:   redeemService,
		userRepo:        userRepo,
		resumeService:   resumeSvc,
		providersLoaded: true,
	}

	returnQuery := url.Values{}
	returnQuery.Set("status", "success")
	returnQuery.Set("order_id", "999")
	returnQuery.Set("resume_token", token)
	returnQuery.Set("pid", "pid-1")
	returnQuery.Set("trade_no", "upstream-return-1")
	returnQuery.Set("out_trade_no", order.OutTradeNo)
	returnQuery.Set("type", payment.TypeAlipay)
	returnQuery.Set("name", "Sub2API 10.00 CNY")
	returnQuery.Set("money", "10.00")
	returnQuery.Set("trade_status", "TRADE_SUCCESS")
	returnQuery.Set("sign", "signed-return")
	returnQuery.Set("sign_type", "MD5")

	got, err := svc.GetPublicOrderByResumeTokenWithProviderReturn(ctx, token, returnQuery.Encode())
	require.NoError(t, err)
	require.Equal(t, order.ID, got.ID)
	require.Equal(t, OrderStatusCompleted, got.Status)
	require.Equal(t, "upstream-return-1", got.PaymentTradeNo)
	require.Equal(t, 1, provider.verifyCount)
	require.Equal(t, 0, provider.queryCount)
	require.NotContains(t, provider.verifiedRaw, "status=success")
	require.NotContains(t, provider.verifiedRaw, "order_id=")
	require.NotContains(t, provider.verifiedRaw, "resume_token=")
	require.Equal(t, 10.0, userRepo.getByIDUser.Balance)
}

func TestVerifyOrderPublicDoesNotCheckUpstreamForPendingOrder(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("public-verify@example.com").
		SetPasswordHash("hash").
		SetUsername("public-verify-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("PUBLIC-VERIFY").
		SetOutTradeNo("sub2_public_verify_pending").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-public-verify").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	registry := payment.NewRegistry()
	provider := &paymentResumeLookupProvider{}
	registry.Register(provider)

	svc := &PaymentService{
		entClient:       client,
		registry:        registry,
		providersLoaded: true,
	}

	got, err := svc.VerifyOrderPublic(ctx, order.OutTradeNo)
	require.NoError(t, err)
	require.Equal(t, order.ID, got.ID)
	require.Equal(t, 0, provider.queryCount)
}

func TestVerifyOrderPublicRejectsBlankOutTradeNo(t *testing.T) {
	svc := &PaymentService{
		entClient: newPaymentConfigServiceTestClient(t),
	}

	_, err := svc.VerifyOrderPublic(context.Background(), "   ")
	require.Error(t, err)
	require.Equal(t, "INVALID_OUT_TRADE_NO", infraerrors.Reason(err))
}
