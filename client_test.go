package dopawin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBetDiceSendsAPIKeyAndPayload(t *testing.T) {
	var gotKey string
	var gotReq DiceBetRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/bet/dice" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		gotKey = r.Header.Get("X-API-Key")
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(DiceBetResponse{
			NodeBetID:  "eu-1:1",
			Roll:       123456,
			Target:     500000,
			Won:        false,
			Profit:     "-0.1",
			NewBalance: "9.9",
		})
	}))
	defer srv.Close()

	c := New(WithBaseURL(srv.URL), WithAPIKey("test-key"))
	resp, err := c.BetDice(context.Background(), DiceBetRequest{
		CoinID:    "PLAY",
		Amount:    "0.1",
		Target:    500000,
		RollUnder: true,
	})
	if err != nil {
		t.Fatalf("BetDice error: %v", err)
	}
	if gotKey != "test-key" {
		t.Fatalf("X-API-Key = %q", gotKey)
	}
	if gotReq.CoinID != "PLAY" || gotReq.Amount != "0.1" || gotReq.Target != 500000 || !gotReq.RollUnder {
		t.Fatalf("request = %+v", gotReq)
	}
	if resp.Profit != "-0.1" || resp.NodeBetID != "eu-1:1" {
		t.Fatalf("response = %+v", resp)
	}
}

func TestBetDiceRequiresAPIKey(t *testing.T) {
	_, err := New().BetDice(context.Background(), DiceBetRequest{})
	if !errors.Is(err, ErrAPIKeyRequired) {
		t.Fatalf("err = %v, want ErrAPIKeyRequired", err)
	}
}

func TestAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid coin"}`))
	}))
	defer srv.Close()

	_, err := New(WithBaseURL(srv.URL), WithAPIKey("test")).BetDice(context.Background(), DiceBetRequest{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T %v, want APIError", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Message != "invalid coin" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
}

func TestBetDiceWithRetry(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"slow down"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(DiceBetResponse{NodeBetID: "eu-1:2", Profit: "0.1"})
	}))
	defer srv.Close()

	resp, err := New(WithBaseURL(srv.URL), WithAPIKey("test")).BetDiceWithRetry(context.Background(), DiceBetRequest{}, RetryConfig{
		MaxAttempts: 2,
		BaseDelay:   1,
		MaxDelay:    1,
	})
	if err != nil {
		t.Fatalf("BetDiceWithRetry error: %v", err)
	}
	if attempts != 2 || resp.NodeBetID != "eu-1:2" {
		t.Fatalf("attempts=%d resp=%+v", attempts, resp)
	}
}
