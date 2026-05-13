package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

func (s *PaymentService) GetPublicOrderByResumeToken(ctx context.Context, token string) (*dbent.PaymentOrder, error) {
	return s.GetPublicOrderByResumeTokenWithProviderReturn(ctx, token, "")
}

func (s *PaymentService) GetPublicOrderByResumeTokenWithProviderReturn(ctx context.Context, token string, providerReturnQuery string) (*dbent.PaymentOrder, error) {
	claims, err := s.paymentResume().ParseToken(strings.TrimSpace(token))
	if err != nil {
		return nil, err
	}

	order, err := s.entClient.PaymentOrder.Get(ctx, claims.OrderID)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, infraerrors.NotFound("NOT_FOUND", "order not found")
		}
		return nil, fmt.Errorf("get order by resume token: %w", err)
	}
	if claims.UserID > 0 && order.UserID != claims.UserID {
		return nil, invalidResumeTokenMatchError()
	}
	snapshot := psOrderProviderSnapshot(order)
	orderProviderInstanceID := strings.TrimSpace(psStringValue(order.ProviderInstanceID))
	orderProviderKey := strings.TrimSpace(psStringValue(order.ProviderKey))
	if snapshot != nil {
		if snapshot.ProviderInstanceID != "" {
			orderProviderInstanceID = snapshot.ProviderInstanceID
		}
		if snapshot.ProviderKey != "" {
			orderProviderKey = snapshot.ProviderKey
		}
	}
	if claims.ProviderInstanceID != "" && orderProviderInstanceID != claims.ProviderInstanceID {
		return nil, invalidResumeTokenMatchError()
	}
	if claims.ProviderKey != "" && !strings.EqualFold(orderProviderKey, claims.ProviderKey) {
		return nil, invalidResumeTokenMatchError()
	}
	if claims.PaymentType != "" && NormalizeVisibleMethod(order.PaymentType) != NormalizeVisibleMethod(claims.PaymentType) {
		return nil, invalidResumeTokenMatchError()
	}
	if paymentResumeLookupShouldReconcile(order.Status) && s.applyProviderReturnForOrder(ctx, order, providerReturnQuery) {
		order, err = s.entClient.PaymentOrder.Get(ctx, order.ID)
		if err != nil {
			return nil, fmt.Errorf("reload order after provider return: %w", err)
		}
	}
	if paymentResumeLookupShouldReconcile(order.Status) {
		result := s.checkPaid(ctx, order)
		if result == checkPaidResultAlreadyPaid {
			order, err = s.entClient.PaymentOrder.Get(ctx, order.ID)
			if err != nil {
				return nil, fmt.Errorf("reload order by resume token: %w", err)
			}
		}
	}

	return order, nil
}

func paymentResumeLookupShouldReconcile(status string) bool {
	switch status {
	case OrderStatusPending, OrderStatusExpired, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

func (s *PaymentService) applyProviderReturnForOrder(ctx context.Context, order *dbent.PaymentOrder, rawQuery string) bool {
	rawQuery = strings.TrimSpace(rawQuery)
	if order == nil || rawQuery == "" {
		return false
	}
	sanitizedQuery, err := sanitizeEasyPayReturnQuery(rawQuery)
	if err != nil {
		return false
	}
	prov, err := s.getOrderProvider(ctx, order)
	if err != nil {
		slog.Warn("payment return recovery provider lookup failed", "orderID", order.ID, "error", err)
		return false
	}
	if prov == nil || !strings.EqualFold(strings.TrimSpace(prov.ProviderKey()), payment.TypeEasyPay) {
		return false
	}
	notification, err := prov.VerifyNotification(ctx, sanitizedQuery, nil)
	if err != nil {
		slog.Warn("payment return recovery verify failed", "orderID", order.ID, "error", err)
		return false
	}
	if notification == nil || notification.Status != payment.NotificationStatusSuccess {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(notification.OrderID), strings.TrimSpace(order.OutTradeNo)) {
		slog.Warn("payment return recovery order mismatch",
			"orderID", order.ID,
			"returnOutTradeNo", notification.OrderID,
		)
		return false
	}
	if err := s.HandlePaymentNotification(ctx, notification, prov.ProviderKey()); err != nil {
		slog.Error("payment return recovery fulfillment failed", "orderID", order.ID, "error", err)
		return false
	}
	return true
}

func sanitizeEasyPayReturnQuery(rawQuery string) (string, error) {
	rawQuery = strings.TrimPrefix(strings.TrimSpace(rawQuery), "?")
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", fmt.Errorf("parse easypay return query: %w", err)
	}
	if strings.TrimSpace(values.Get("sign")) == "" ||
		strings.TrimSpace(values.Get("trade_status")) == "" ||
		strings.TrimSpace(values.Get("out_trade_no")) == "" {
		return "", fmt.Errorf("missing easypay return fields")
	}

	sanitized := url.Values{}
	for key, entries := range values {
		if _, skip := easyPayReturnAppQueryKeys[strings.ToLower(strings.TrimSpace(key))]; skip {
			continue
		}
		for _, entry := range entries {
			if strings.TrimSpace(entry) == "" {
				continue
			}
			sanitized.Add(key, entry)
		}
	}
	if strings.TrimSpace(sanitized.Get("sign")) == "" ||
		strings.TrimSpace(sanitized.Get("trade_status")) == "" ||
		strings.TrimSpace(sanitized.Get("out_trade_no")) == "" {
		return "", fmt.Errorf("missing sanitized easypay return fields")
	}
	return sanitized.Encode(), nil
}

var easyPayReturnAppQueryKeys = map[string]struct{}{
	"order_id":            {},
	"resume_token":        {},
	"status":              {},
	"wechat_resume_token": {},
}

func invalidResumeTokenMatchError() error {
	return infraerrors.BadRequest("INVALID_RESUME_TOKEN", "resume token does not match the payment order")
}

func (s *PaymentService) ParseWeChatPaymentResumeToken(token string) (*WeChatPaymentResumeClaims, error) {
	return s.paymentResume().ParseWeChatPaymentResumeToken(strings.TrimSpace(token))
}
