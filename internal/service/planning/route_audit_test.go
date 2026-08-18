package planning

import (
	"context"
	"errors"
	"testing"
	"time"

	"sanitation-operations/internal/audit"
	"sanitation-operations/internal/clock"
	"sanitation-operations/internal/domain/workplan"
	"sanitation-operations/internal/identity"
	"sanitation-operations/internal/repository"
)

var errRouteAudit = errors.New("audit unavailable")

type routeAuditStore struct {
	repository.Store
	routes []workplan.Route
}

func (s *routeAuditStore) SaveRoute(_ context.Context, value workplan.Route) error {
	s.routes = append(s.routes, value)
	return nil
}

func (s *routeAuditStore) AppendAudit(context.Context, audit.Event) error { return errRouteAudit }

type routeAuditTx struct {
	repository.Tx
	owner   *routeAuditStore
	pending *workplan.Route
}

func (t *routeAuditTx) SaveRoute(_ context.Context, value workplan.Route) error {
	t.pending = &value
	return nil
}

func (t *routeAuditTx) AppendAudit(context.Context, audit.Event) error { return errRouteAudit }

func (s *routeAuditStore) WithTx(ctx context.Context, fn func(context.Context, repository.Tx) error) error {
	tx := &routeAuditTx{owner: s}
	if err := fn(ctx, tx); err != nil {
		return err
	}
	if tx.pending != nil {
		s.routes = append(s.routes, *tx.pending)
	}
	return nil
}

func TestRouteCreationRollsBackWhenAuditFails(t *testing.T) {
	now := time.Date(2026, 8, 18, 8, 0, 0, 0, time.UTC)
	store := &routeAuditStore{}
	service := Service{Store: store, Clock: clock.Fixed{Current: now}, IDs: &identity.Sequence{}}
	_, err := service.CreateRoute(context.Background(), CreateRouteInput{
		Code: "R-901", Name: "North", Zone: "N", RequiredCapacityKg: 100,
		ActorID: "op", RequestID: "req",
	})
	if err == nil || !errors.Is(err, errRouteAudit) {
		t.Fatalf("unexpected error %v", err)
	}
	if len(store.routes) != 0 {
		t.Fatalf("route persisted after audit failure: %+v", store.routes)
	}
}
