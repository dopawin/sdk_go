package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	dopawin "github.com/dopawin/sdk_go"
)

type stats struct {
	start     time.Time
	bets      int
	wins      int
	losses    int
	errors    int
	profit    float64
	streak    int
	maxStreak int
	last      *dopawin.DiceBetResponse
	lastErr   error
}

func main() {
	var (
		apiKey     = flag.String("api-key", os.Getenv("DOPAWIN_API_KEY"), "Dopa API key, or DOPAWIN_API_KEY")
		baseURL    = flag.String("base-url", dopawin.DefaultBaseURL, "node API base URL")
		coin       = flag.String("coin", "PLAY", "coin id")
		baseAmount = flag.Float64("base", 0.1, "base wager")
		target     = flag.Int("target", 500000, "dice target, 1-999999")
		rollUnder  = flag.Bool("under", true, "roll under target; false rolls over target")
		maxLosses  = flag.Int("max-losses", 8, "stop after this many consecutive losses")
		delay      = flag.Duration("delay", 0, "delay between bets")
		limit      = flag.Int("limit", 0, "number of bets to run; 0 means until stopped")
		live       = flag.Bool("live", false, "place real bets; default is dry-run simulator")
		quiet      = flag.Bool("quiet", false, "print less per-bet output")
	)
	flag.Parse()

	if *target < dopawin.DiceMinTarget || *target > dopawin.DiceMaxTarget {
		fatalf("target must be between %d and %d", dopawin.DiceMinTarget, dopawin.DiceMaxTarget)
	}
	if *baseAmount <= 0 {
		fatalf("base must be positive")
	}
	if *live && strings.TrimSpace(*apiKey) == "" {
		fatalf("--api-key or DOPAWIN_API_KEY is required with --live")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := dopawin.New(
		dopawin.WithBaseURL(*baseURL),
		dopawin.WithAPIKey(*apiKey),
		dopawin.WithUserAgent("dopawin-martingale/0.1.0"),
	)
	sim := dopawin.NewDiceSimulator(0)
	st := stats{start: time.Now()}
	wager := *baseAmount

	printHeader(*live, *baseURL, *coin, *baseAmount, *target, *rollUnder)
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		if *limit > 0 && st.bets >= *limit {
			break
		}
		select {
		case <-ctx.Done():
			fmt.Println()
			printSummary(st)
			return
		default:
		}

		req := dopawin.DiceBetRequest{
			CoinID:    *coin,
			Amount:    formatAmount(wager),
			Target:    *target,
			RollUnder: *rollUnder,
		}

		start := time.Now()
		var resp *dopawin.DiceBetResponse
		var err error
		if *live {
			resp, err = client.BetDiceWithRetry(ctx, req, dopawin.DefaultRetryConfig())
		} else {
			resp = sim.Bet(req)
		}
		latency := time.Since(start)
		applyResult(&st, resp, err)

		if err == nil && resp != nil {
			if resp.Won {
				wager = *baseAmount
			} else {
				wager *= 2
			}
			if st.streak >= *maxLosses {
				fmt.Printf("\nstop: hit max loss streak %d\n", *maxLosses)
				break
			}
		}

		if !*quiet {
			printBetLine(st, req.Amount, latency)
		}
		printStatus(st)

		if *delay > 0 {
			select {
			case <-ctx.Done():
				fmt.Println()
				printSummary(st)
				return
			case <-time.After(*delay):
			}
		} else {
			select {
			case <-ctx.Done():
				fmt.Println()
				printSummary(st)
				return
			case <-ticker.C:
			default:
			}
		}
	}
	fmt.Println()
	printSummary(st)
}

func applyResult(st *stats, resp *dopawin.DiceBetResponse, err error) {
	st.bets++
	st.last = resp
	st.lastErr = err
	if err != nil {
		st.errors++
		return
	}
	if resp.Won {
		st.wins++
		st.streak = 0
	} else {
		st.losses++
		st.streak++
		if st.streak > st.maxStreak {
			st.maxStreak = st.streak
		}
	}
	st.profit += parseProfit(resp.Profit)
}

func printHeader(live bool, baseURL, coin string, base float64, target int, under bool) {
	mode := "DRY-RUN"
	if live {
		mode = "LIVE"
	}
	dir := "under"
	if !under {
		dir = "over"
	}
	fmt.Printf("dopawin martingale %s | %s | %s base %.8f | %s %d | ctrl+c to stop\n", mode, baseURL, coin, base, dir, target)
}

func printBetLine(st stats, wager string, latency time.Duration) {
	if st.lastErr != nil {
		fmt.Printf("\n#%-5d wager=%s error=%v", st.bets, wager, st.lastErr)
		return
	}
	win := "LOSS"
	if st.last.Won {
		win = "WIN "
	}
	id := st.last.NodeBetID
	if st.last.BetID != 0 {
		id = fmt.Sprintf("%d", st.last.BetID)
	}
	fmt.Printf("\n#%-5d %-4s id=%s roll=%06d wager=%s profit=%s latency=%s",
		st.bets, win, id, st.last.Roll, wager, st.last.Profit, latency.Round(time.Millisecond))
}

func printStatus(st stats) {
	elapsed := time.Since(st.start).Seconds()
	if elapsed <= 0 {
		elapsed = 1
	}
	bps := float64(st.bets) / elapsed
	winRate := 0.0
	if st.bets > 0 {
		winRate = float64(st.wins) / float64(st.bets) * 100
	}
	fmt.Printf(" | %.1f bets/s | W/L %d/%d %.1f%% | streak %d max %d | net %.8f",
		bps, st.wins, st.losses, winRate, st.streak, st.maxStreak, st.profit)
}

func printSummary(st stats) {
	fmt.Printf("summary: bets=%d wins=%d losses=%d errors=%d max_loss_streak=%d net=%.8f avg_bets_sec=%.2f\n",
		st.bets, st.wins, st.losses, st.errors, st.maxStreak, st.profit, float64(st.bets)/math.Max(time.Since(st.start).Seconds(), 1))
}

func formatAmount(v float64) string {
	return strconv.FormatFloat(v, 'f', 8, 64)
}

func parseProfit(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(2)
}
