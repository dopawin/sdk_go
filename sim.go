package dopawin

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type DiceSimulator struct {
	rng *rand.Rand
}

func NewDiceSimulator(seed int64) *DiceSimulator {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &DiceSimulator{rng: rand.New(rand.NewSource(seed))}
}

func (s *DiceSimulator) Bet(req DiceBetRequest) *DiceBetResponse {
	roll := s.rng.Intn(1_000_000)
	won := roll < req.Target
	if !req.RollUnder {
		won = roll > req.Target
	}

	wager := parseDecimal(req.Amount)
	if wager == 0 {
		wager = parseDecimal(req.Wager)
	}

	profit := -wager
	multiplier := 0.0
	if req.Target > 0 {
		chance := float64(req.Target) / 1_000_000
		if !req.RollUnder {
			chance = float64(999_999-req.Target) / 1_000_000
		}
		if chance > 0 {
			multiplier = 1 / chance
		}
	}
	if won {
		profit = wager * (multiplier - 1)
	}

	return &DiceBetResponse{
		NodeBetID:  "dry-run",
		Roll:       roll,
		Target:     req.Target,
		Won:        won,
		Profit:     formatDecimal(profit),
		Multiplier: fmt.Sprintf("%.3f", multiplier),
		NewBalance: "dry-run",
		ClientSeed: "dry-run",
		Nonce:      uint64(s.rng.Int63()),
		BankrollStats: &BankrollStat{
			Currency: req.CoinID,
		},
	}
}

func parseDecimal(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func formatDecimal(v float64) string {
	return strconv.FormatFloat(v, 'f', 8, 64)
}
