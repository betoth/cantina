package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
)

// price is in US$ per million tokens.
type price struct {
	Input        float64 `json:"input"`
	Output       float64 `json:"output"`
	CacheWrite5m float64 `json:"cache_write_5m"`
	CacheWrite1h float64 `json:"cache_write_1h"`
	CacheRead    float64 `json:"cache_read"`
}

func (p price) zero() bool {
	return p == price{}
}

type priceTable struct {
	Source    string           `json:"source"`
	CheckedAt string           `json:"checked_at"`
	Models    map[string]price `json:"models"`
}

func loadPrices(path string) (priceTable, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return priceTable{}, fmt.Errorf("read prices: %w", err)
	}
	var t priceTable
	if err := json.Unmarshal(b, &t); err != nil {
		return priceTable{}, fmt.Errorf("parse prices %s: %w", path, err)
	}
	return t, nil
}

// lookup finds the price of a model by exact name or, failing that, by the
// longest registered prefix. ok is false when there is no usable price.
func (t priceTable) lookup(model string) (price, bool) {
	if p, ok := t.Models[model]; ok {
		return p, !p.zero()
	}
	best := ""
	for name := range t.Models {
		if strings.HasPrefix(model, name) && len(name) > len(best) {
			best = name
		}
	}
	if best == "" {
		return price{}, false
	}
	p := t.Models[best]
	return p, !p.zero()
}

// costUnit is the stored cost precision: US$ 0.0001.
const costUnit = 1e-4

// cost returns the cost of u in units of costUnit.
func cost(u usage, p price) int64 {
	usd := (float64(u.Input)*p.Input +
		float64(u.Output)*p.Output +
		float64(u.CacheWrite5m)*p.CacheWrite5m +
		float64(u.CacheWrite1h)*p.CacheWrite1h +
		float64(u.CacheRead)*p.CacheRead) / 1e6
	return int64(math.Round(usd / costUnit))
}
