package outboundgroup

import (
	"testing"

	C "github.com/metacubex/mihomo/constant"
)

func TestParseFallbackModeDefaultsToOrdered(t *testing.T) {
	for _, value := range []string{"", "ordered", "RELIABLE"} {
		mode, err := parseFallbackMode(value)
		if err != nil {
			t.Fatalf("parseFallbackMode(%q) error = %v", value, err)
		}
		if value == "RELIABLE" && mode != fallbackModeReliable {
			t.Fatalf("parseFallbackMode(%q) = %q, want reliable", value, mode)
		}
		if value != "RELIABLE" && mode != fallbackModeOrdered {
			t.Fatalf("parseFallbackMode(%q) = %q, want ordered", value, mode)
		}
	}
}

func TestParseFallbackModeRejectsUnknownValue(t *testing.T) {
	if _, err := parseFallbackMode("unknown"); err == nil {
		t.Fatal("parseFallbackMode(unknown) succeeded, want error")
	}
}

func TestReliabilityScoreRequiresMinimumSamples(t *testing.T) {
	score, ready := reliabilityScore([]C.DelayHistory{
		{Delay: 42},
		{Delay: 0},
	})

	if ready {
		t.Fatal("score with too few samples must not be ready")
	}
	if score != 0 {
		t.Fatalf("score with too few samples = %v, want 0", score)
	}
}

func TestReliabilityScoreCountsSuccessfulChecks(t *testing.T) {
	score, ready := reliabilityScore([]C.DelayHistory{
		{Delay: 42},
		{Delay: 0},
		{Delay: 21},
		{Delay: 18},
	})

	if !ready {
		t.Fatal("score with enough samples must be ready")
	}
	if score != 0.75 {
		t.Fatalf("score = %v, want 0.75", score)
	}
}

func TestReliabilityScoreUsesLatestTenSamples(t *testing.T) {
	history := make([]C.DelayHistory, 0, 12)
	for i := 0; i < 2; i++ {
		history = append(history, C.DelayHistory{})
	}
	for i := 0; i < 8; i++ {
		history = append(history, C.DelayHistory{Delay: 100})
	}
	for i := 0; i < 2; i++ {
		history = append(history, C.DelayHistory{})
	}

	score, ready := reliabilityScore(history)
	if !ready {
		t.Fatal("score with enough samples must be ready")
	}
	if score != 0.8 {
		t.Fatalf("score = %v, want 0.8", score)
	}
}

func TestChooseReliableCandidatePreservesCurrentWithinHysteresis(t *testing.T) {
	candidates := []reliableCandidate{
		{name: "current", alive: true, ready: true, score: 0.8},
		{name: "better", alive: true, ready: true, score: 0.85},
	}

	if got := chooseReliableCandidate(candidates, "current"); got != "current" {
		t.Fatalf("candidate = %q, want current", got)
	}

	candidates[1].score = 0.91
	if got := chooseReliableCandidate(candidates, "current"); got != "better" {
		t.Fatalf("candidate = %q, want better", got)
	}
}

func TestChooseReliableCandidateLeavesUnreadyCurrentForReadyCandidate(t *testing.T) {
	candidates := []reliableCandidate{
		{name: "current", alive: true},
		{name: "ready", alive: true, ready: true, score: 0.6},
	}

	if got := chooseReliableCandidate(candidates, "current"); got != "ready" {
		t.Fatalf("candidate = %q, want ready", got)
	}
}

func TestChooseReliableCandidateFallsBackToOrderWithoutReadyScores(t *testing.T) {
	candidates := []reliableCandidate{
		{name: "first", alive: true},
		{name: "second", alive: true},
	}

	if got := chooseReliableCandidate(candidates, ""); got != "first" {
		t.Fatalf("candidate = %q, want first", got)
	}
}
