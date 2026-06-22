# Dopa Go SDK

Small Go SDK for Dopa betting APIs.

```bash
go get github.com/dopawin/sdk_go
```

## Dice Bet

```go
client := dopawin.New(dopawin.WithAPIKey(os.Getenv("DOPAWIN_API_KEY")))

bet, err := client.BetDice(ctx, dopawin.DiceBetRequest{
    CoinID:    "PLAY",
    Amount:    "0.1",
    Target:    500000,
    RollUnder: true,
})
```

`profit` is net user profit: positive on win, negative on loss.

## Martingale Bot

The included bot dry-runs by default and does not spend funds unless `--live` is set.

Dry-run:

```bash
go run ./cmd/dopawin-martingale --limit 25
```

Live:

```bash
go run ./cmd/dopawin-martingale \
  --live \
  --api-key "$DOPAWIN_API_KEY" \
  --coin PLAY \
  --base 0.1 \
  --target 500000 \
  --max-losses 8
```

Flags:

- `--api-key`: Dopa API key. Can also use `DOPAWIN_API_KEY`.
- `--base-url`: node URL. Defaults to `https://eu-1.dopa.win`.
- `--coin`: coin id. Defaults to `PLAY`.
- `--base`: base wager. Defaults to `0.1`.
- `--target`: dice target. Defaults to `500000`.
- `--under`: roll under target. Defaults to `true`.
- `--max-losses`: stop after this many consecutive losses.
- `--limit`: stop after N bets. `0` means until interrupted.
- `--delay`: optional delay between bets.
- `--quiet`: reduce per-bet output.

## Notes

Use a node URL for betting, currently:

```text
https://eu-1.dopa.win
```

The SDK authenticates with `X-API-Key`.
