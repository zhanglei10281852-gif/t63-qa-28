package httpapi

import (
	"fmt"
	"net/http"
	"testing"
)

func TestDailyReviewIncludesEveryScheduledShift(t *testing.T) {
	h := newAPIHarness(t)
	h.createRouteAndShift(t, "604")
	for index := 0; index < 100; index++ {
		h.createRouteAndShift(t, fmt.Sprintf("%03d", 700+index))
	}
	report := h.request(t, http.MethodGet, "/api/v1/reconciliation?service_date=2026-08-18", nil, http.StatusOK)
	if report["shift_count"].(float64) != 101 {
		t.Fatalf("shift count=%v", report["shift_count"])
	}
}
