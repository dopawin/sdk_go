package dopawin

import (
	"context"
	"net/http"
)

const (
	DiceMinTarget = 1
	DiceMaxTarget = 999999
)

type DiceBetRequest struct {
	CoinID    string  `json:"coin_id,omitempty"`
	Currency  string  `json:"currency,omitempty"`
	Amount    string  `json:"amount,omitempty"`
	Wager     string  `json:"wager,omitempty"`
	Target    int     `json:"target"`
	RollUnder bool    `json:"roll_under"`
	EdgeBPS   int     `json:"edge_bps,omitempty"`
	HouseEdge float64 `json:"houseEdge,omitempty"`
}

type DiceBetResponse struct {
	NodeBetID     string        `json:"nodeBetId"`
	BetID         int64         `json:"betId,omitempty"`
	Roll          int           `json:"roll"`
	Target        int           `json:"target"`
	Won           bool          `json:"won"`
	Profit        string        `json:"profit"`
	Multiplier    string        `json:"multiplier"`
	NewBalance    string        `json:"newBalance"`
	ClientSeed    string        `json:"clientSeed"`
	Nonce         uint64        `json:"nonce"`
	BankrollStats *BankrollStat `json:"bankrollStats,omitempty"`
}

type BankrollStat struct {
	Currency      string `json:"currency"`
	Balance       string `json:"balance"`
	TotalShares   string `json:"totalShares"`
	SharePrice    string `json:"sharePrice"`
	InvestorCount int    `json:"investorCount"`
	MaxBet        string `json:"maxBet"`
	TotalProfit   string `json:"totalProfit"`
	ATH           string `json:"ath"`
	TotalBets     int    `json:"totalBets"`
	TotalWagered  string `json:"totalWagered"`
	TotalXP       string `json:"totalXp"`
}

func (c *Client) BetDice(ctx context.Context, req DiceBetRequest) (*DiceBetResponse, error) {
	var out DiceBetResponse
	if err := c.do(ctx, http.MethodPost, "/api/bet/dice", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
