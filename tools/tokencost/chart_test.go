package main

import (
	"slices"
	"strings"
	"testing"
)

func TestChart(t *testing.T) {
	rows := []row{
		{SessionID: "s1", Issue: 0, Model: "claude-opus-5-5", Start: at(t, "2026-10-05T10:00:00Z"), End: at(t, "2026-10-05T11:00:00Z"), Cost: 5000},
		{SessionID: "s1", Issue: 11, Model: "claude-opus-5-5", Start: at(t, "2026-10-09T08:00:00Z"), End: at(t, "2026-10-09T09:00:00Z"), Cost: 12345},
		{SessionID: "s1", Issue: 11, Model: overheadModel, Start: at(t, "2026-10-09T08:00:00Z"), End: at(t, "2026-10-09T09:00:00Z"), Cost: 1000},
		{SessionID: "s2", Issue: 2, Model: "claude-opus-5-5", Start: at(t, "2026-10-25T23:00:00Z"), End: at(t, "2026-10-26T01:30:00Z"), Cost: 20000},
	}
	got := chart(rows)
	want := "Dados até 2026-10-26.\n\n" +
		"## Custo por issue\n\n" +
		"```mermaid\nxychart-beta\n" +
		"    title \"Custo por issue (US$)\"\n" +
		"    x-axis [\"sem issue\", \"#2\", \"#11\"]\n" +
		"    y-axis \"US$\"\n" +
		"    bar [0.5000, 2.0000, 1.3345]\n" +
		"```\n\n" +
		"## Custo acumulado por semana\n\n" +
		"```mermaid\nxychart-beta\n" +
		"    title \"Custo acumulado por semana (US$)\"\n" +
		"    x-axis [\"2026-10-05\", \"2026-10-12\", \"2026-10-19\"]\n" +
		"    y-axis \"US$\"\n" +
		"    line [1.8345, 1.8345, 3.8345]\n" +
		"```\n"
	if !strings.HasSuffix(got, want) {
		t.Errorf("chart:\n%s\nwant suffix:\n%s", got, want)
	}
	reversed := slices.Clone(rows)
	slices.Reverse(reversed)
	if again := chart(reversed); again != got {
		t.Errorf("second chart differs:\n%s\nwant:\n%s", again, got)
	}
}

func TestChartWithoutRows(t *testing.T) {
	if got := chart(nil); !strings.HasSuffix(got, "Sem dados.\n") {
		t.Errorf("chart:\n%s\nwant a no data note", got)
	}
}

func TestWeekStart(t *testing.T) {
	tests := []struct{ in, want string }{
		{"2026-10-05T00:00:00Z", "2026-10-05"},      // Monday
		{"2026-10-11T23:59:59Z", "2026-10-05"},      // Sunday
		{"2026-10-12T02:00:00+03:00", "2026-10-05"}, // Sunday 23:00 in UTC
		{"2026-10-12T00:00:00Z", "2026-10-12"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := weekStart(at(t, tt.in)).Format("2006-01-02"); got != tt.want {
				t.Errorf("weekStart(%s) = %s, want %s", tt.in, got, tt.want)
			}
		})
	}
}
