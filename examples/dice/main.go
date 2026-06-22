package main

import (
	"context"
	"fmt"
	"os"
	"time"

	dopawin "github.com/dopawin/sdk_go"
)

func main() {
	apiKey := os.Getenv("DOPAWIN_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "set DOPAWIN_API_KEY first")
		os.Exit(2)
	}

	client := dopawin.New(dopawin.WithAPIKey(apiKey))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bet, err := client.BetDice(ctx, dopawin.DiceBetRequest{
		CoinID:    "PLAY",
		Amount:    "0.1",
		Target:    500000,
		RollUnder: true,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("bet=%s won=%t roll=%d profit=%s balance=%s\n",
		bet.NodeBetID, bet.Won, bet.Roll, bet.Profit, bet.NewBalance)
}
