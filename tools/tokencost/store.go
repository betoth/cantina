package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"time"
)

// row is one line of the costs file: usage of one session in one issue with
// one model.
type row struct {
	SessionID string
	Issue     int
	Model     string
	Start     time.Time
	End       time.Time
	Requests  int64
	Usage     usage
	// Cost is in units of costUnit.
	Cost int64
}

type rowKey struct {
	SessionID string
	Issue     int
	Model     string
}

func (r row) key() rowKey {
	return rowKey{r.SessionID, r.Issue, r.Model}
}

var header = []string{
	"session_id", "issue", "model", "start", "end", "requests",
	"input_tokens", "output_tokens", "cache_write_5m_tokens", "cache_write_1h_tokens", "cache_read_tokens",
	"cost_usd",
}

const timeLayout = "2006-01-02T15:04:05.000Z"

// readRows reads the costs file. A missing file has no rows.
func readRows(path string) ([]row, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open costs: %w", err)
	}
	defer f.Close()
	rows, err := decodeRows(f)
	if err != nil {
		return nil, fmt.Errorf("read costs %s: %w", path, err)
	}
	return rows, nil
}

func decodeRows(r io.Reader) ([]row, error) {
	recs, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(recs) == 0 {
		return nil, nil
	}
	if !slices.Equal(recs[0], header) {
		return nil, fmt.Errorf("header %v, want %v", recs[0], header)
	}
	var rows []row
	for i, rec := range recs[1:] {
		rw, err := decodeRow(rec)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+2, err)
		}
		rows = append(rows, rw)
	}
	return rows, nil
}

func decodeRow(rec []string) (row, error) {
	if len(rec) != len(header) {
		return row{}, fmt.Errorf("want %d fields, got %d", len(header), len(rec))
	}
	var (
		rw   row
		errs []error
	)
	integer := func(s string) int64 {
		n, err := strconv.ParseInt(s, 10, 64)
		errs = append(errs, err)
		return n
	}
	timestamp := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		errs = append(errs, err)
		return t.UTC()
	}
	rw.SessionID = rec[0]
	rw.Issue = int(integer(rec[1]))
	rw.Model = rec[2]
	rw.Start = timestamp(rec[3])
	rw.End = timestamp(rec[4])
	rw.Requests = integer(rec[5])
	rw.Usage = usage{
		Input:        integer(rec[6]),
		Output:       integer(rec[7]),
		CacheWrite5m: integer(rec[8]),
		CacheWrite1h: integer(rec[9]),
		CacheRead:    integer(rec[10]),
	}
	usd, err := strconv.ParseFloat(rec[11], 64)
	errs = append(errs, err)
	rw.Cost = int64(math.Round(usd / costUnit))
	return rw, errors.Join(errs...)
}

// writeRows writes the rows sorted, replacing the file atomically.
func writeRows(path string, rows []row) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create costs dir: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".ai-costs-*.csv")
	if err != nil {
		return fmt.Errorf("create costs temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	if err := encodeRows(tmp, rows); err != nil {
		tmp.Close()
		return fmt.Errorf("write costs: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write costs: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("replace costs: %w", err)
	}
	return nil
}

func encodeRows(w io.Writer, rows []row) error {
	sortRows(rows)
	cw := csv.NewWriter(w)
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, r := range rows {
		rec := []string{
			r.SessionID,
			strconv.Itoa(r.Issue),
			r.Model,
			r.Start.UTC().Format(timeLayout),
			r.End.UTC().Format(timeLayout),
			strconv.FormatInt(r.Requests, 10),
			strconv.FormatInt(r.Usage.Input, 10),
			strconv.FormatInt(r.Usage.Output, 10),
			strconv.FormatInt(r.Usage.CacheWrite5m, 10),
			strconv.FormatInt(r.Usage.CacheWrite1h, 10),
			strconv.FormatInt(r.Usage.CacheRead, 10),
			formatCost(r.Cost),
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func formatCost(units int64) string {
	return fmt.Sprintf("%d.%04d", units/1e4, units%1e4)
}

// sortRows orders by start, then by the key, so the file diff stays stable.
func sortRows(rows []row) {
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if !a.Start.Equal(b.Start) {
			return a.Start.Before(b.Start)
		}
		if a.SessionID != b.SessionID {
			return a.SessionID < b.SessionID
		}
		if a.Issue != b.Issue {
			return a.Issue < b.Issue
		}
		return a.Model < b.Model
	})
}
