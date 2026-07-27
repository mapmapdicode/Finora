package service

import (
	"context"
	"testing"

	"wealthos-backend/internal/domain"
	"wealthos-backend/internal/storage"
)

type mockCache struct {
	store map[string]*domain.UserSettings
}

func (m *mockCache) GetUserSettings(ctx context.Context, userID string) (*domain.UserSettings, error) {
	if st, ok := m.store[userID]; ok {
		return st, nil
	}
	return nil, nil
}

func (m *mockCache) SetUserSettings(ctx context.Context, userID string, settings *domain.UserSettings) error {
	m.store[userID] = settings
	return nil
}

func (m *mockCache) InvalidateUserSettings(ctx context.Context, userID string) error {
	delete(m.store, userID)
	return nil
}

func TestUserSettingsRedisCacheAndFallback(t *testing.T) {
	store := storage.NewInMemoryStore()
	cache := &mockCache{store: make(map[string]*domain.UserSettings)}
	svc := NewWealthService(store, cache)

	ctx := context.Background()
	userID := domain.ID("user-123")

	// 1. Initial get - should return default 'full' and populate cache
	st1, err := svc.GetUserSettings(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st1.AmountDisplayMode != domain.AmountDisplayModeFull {
		t.Errorf("expected default mode 'full', got '%s'", st1.AmountDisplayMode)
	}

	// Verify cache was populated
	if cached, ok := cache.store[string(userID)]; !ok || cached.AmountDisplayMode != domain.AmountDisplayModeFull {
		t.Errorf("expected cache to be populated with 'full', got %v", cached)
	}

	// 2. Update settings to 'compact'
	updated, err := svc.UpdateUserSettings(ctx, userID, domain.AmountDisplayModeCompact)
	if err != nil {
		t.Fatalf("update error: %v", err)
	}
	if updated.AmountDisplayMode != domain.AmountDisplayModeCompact {
		t.Errorf("expected updated mode 'compact', got '%s'", updated.AmountDisplayMode)
	}

	// Verify cache was updated to 'compact'
	if cached, ok := cache.store[string(userID)]; !ok || cached.AmountDisplayMode != domain.AmountDisplayModeCompact {
		t.Errorf("expected cache to be updated to 'compact', got %v", cached)
	}

	// 3. GetUserSettings should read directly from cache
	st2, err := svc.GetUserSettings(ctx, userID)
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if st2.AmountDisplayMode != domain.AmountDisplayModeCompact {
		t.Errorf("expected cached mode 'compact', got '%s'", st2.AmountDisplayMode)
	}
}
