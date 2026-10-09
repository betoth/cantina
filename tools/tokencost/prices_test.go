package main

import "testing"

func TestPriceLookup(t *testing.T) {
	opus := price{Input: 4, Output: 20, CacheWrite5m: 5, CacheWrite1h: 8, CacheRead: 0.2}
	sonnet := price{Input: 2, Output: 10, CacheWrite5m: 2.5, CacheWrite1h: 4, CacheRead: 0.1}
	table := priceTable{Models: map[string]price{
		"claude-opus-5-5":   opus,
		"claude-sonnet-5":   {Input: 2},
		"claude-sonnet-5-5": sonnet,
		"claude-haiku-5-5":  {},
	}}
	tests := []struct {
		model  string
		want   price
		wantOK bool
	}{
		{"claude-opus-5-5", opus, true},
		{"claude-sonnet-5-5-20260101", sonnet, true},
		{"claude-haiku-5-5", price{}, false},
		{"claude-fable-5-1", price{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			got, ok := table.lookup(tt.model)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("lookup(%q) = %+v, %v; want %+v, %v", tt.model, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestCost(t *testing.T) {
	opus := price{Input: 4, Output: 20, CacheWrite5m: 5, CacheWrite1h: 8, CacheRead: 0.2}
	tests := []struct {
		name string
		u    usage
		want int64
	}{
		// 1000×4 + 2000×20 + 10000×5 + 20000×8 + 500000×0.2 = 354000 → US$ 0.354
		{"all categories", usage{Input: 1000, Output: 2000, CacheWrite5m: 10000, CacheWrite1h: 20000, CacheRead: 500000}, 3540},
		// 1h writes cost more than 5m writes: 100000×8 = US$ 0.8 against 100000×5 = US$ 0.5
		{"1h cache write", usage{CacheWrite1h: 100000}, 8000},
		{"5m cache write", usage{CacheWrite5m: 100000}, 5000},
		// 123 output tokens × 20 = US$ 0.00246, rounded to 0.0025
		{"rounds to 4 places", usage{Output: 123}, 25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cost(tt.u, opus); got != tt.want {
				t.Errorf("cost = %d, want %d", got, tt.want)
			}
		})
	}
}
