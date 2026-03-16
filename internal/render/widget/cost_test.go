package widget

import (
	"strings"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

func TestCostWidget_ZeroCostReturnsEmpty(t *testing.T) {
	ctx := &model.RenderContext{SessionCostUSD: 0}
	cfg := defaultCfg()

	if got := Cost(ctx, cfg); got != "" {
		t.Errorf("Cost with zero: expected empty string, got %q", got)
	}
}

func TestCostWidget_SubDollar(t *testing.T) {
	ctx := &model.RenderContext{SessionCostUSD: 0.42}
	cfg := defaultCfg()

	got := Cost(ctx, cfg)
	if !strings.Contains(got, "$0.42") {
		t.Errorf("Cost sub-dollar: expected '$0.42' in output, got %q", got)
	}
}

func TestCostWidget_MultiDollar(t *testing.T) {
	ctx := &model.RenderContext{SessionCostUSD: 1.23}
	cfg := defaultCfg()

	got := Cost(ctx, cfg)
	if !strings.Contains(got, "$1.23") {
		t.Errorf("Cost multi-dollar: expected '$1.23' in output, got %q", got)
	}
}

func TestCostWidget_LargerAmount(t *testing.T) {
	ctx := &model.RenderContext{SessionCostUSD: 12.50}
	cfg := defaultCfg()

	got := Cost(ctx, cfg)
	if !strings.Contains(got, "$12.50") {
		t.Errorf("Cost larger: expected '$12.50' in output, got %q", got)
	}
}

func TestCostWidget_NilCostEquivalent(t *testing.T) {
	// SessionCostUSD defaults to zero (its zero value) when cost data is unavailable.
	// This mirrors the nil-cost case: gather.go only sets SessionCostUSD when
	// StdinData.Cost is non-nil, so a missing cost object leaves it at 0.
	ctx := &model.RenderContext{}
	cfg := defaultCfg()

	if got := Cost(ctx, cfg); got != "" {
		t.Errorf("Cost with default (unavailable) context: expected empty string, got %q", got)
	}
}

func TestCostWidget_RegisteredInRegistry(t *testing.T) {
	if _, ok := Registry["cost"]; !ok {
		t.Error("Registry missing 'cost' widget")
	}
}
