package outboundgroup

import (
	"fmt"
	"strings"

	C "github.com/metacubex/mihomo/constant"
)

const (
	fallbackModeOrdered  fallbackMode = "ordered"
	fallbackModeReliable fallbackMode = "reliable"

	reliableMinSamples    = 3
	reliableHistoryWindow = 10
	reliableSwitchMargin  = 0.10
)

type fallbackMode string

func parseFallbackMode(value string) (fallbackMode, error) {
	switch mode := fallbackMode(strings.ToLower(strings.TrimSpace(value))); mode {
	case "", fallbackModeOrdered:
		return fallbackModeOrdered, nil
	case fallbackModeReliable:
		return fallbackModeReliable, nil
	default:
		return "", fmt.Errorf("unsupported fallback-mode: %s", value)
	}
}

type reliableCandidate struct {
	name  string
	alive bool
	score float64
	ready bool
}

func reliabilityScore(history []C.DelayHistory) (float64, bool) {
	if len(history) > reliableHistoryWindow {
		history = history[len(history)-reliableHistoryWindow:]
	}
	if len(history) < reliableMinSamples {
		return 0, false
	}

	successes := 0
	for _, sample := range history {
		if sample.Delay > 0 {
			successes++
		}
	}
	return float64(successes) / float64(len(history)), true
}

func chooseReliableCandidate(
	candidates []reliableCandidate,
	current string,
) string {
	firstAlive := ""
	var currentCandidate *reliableCandidate
	var bestCandidate *reliableCandidate

	for _, candidate := range candidates {
		if !candidate.alive {
			continue
		}
		if firstAlive == "" {
			firstAlive = candidate.name
		}
		if candidate.name == current {
			value := candidate
			currentCandidate = &value
		}
		if !candidate.ready ||
			bestCandidate != nil && candidate.score <= bestCandidate.score {
			continue
		}
		value := candidate
		bestCandidate = &value
	}

	if firstAlive == "" {
		return candidates[0].name
	}
	if currentCandidate != nil {
		if bestCandidate == nil {
			return currentCandidate.name
		}
		if currentCandidate.ready &&
			(bestCandidate.name == currentCandidate.name ||
				bestCandidate.score <= currentCandidate.score+reliableSwitchMargin) {
			return currentCandidate.name
		}
	}
	if bestCandidate != nil {
		return bestCandidate.name
	}
	return firstAlive
}
