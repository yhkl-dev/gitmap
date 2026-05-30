package main

import "testing"

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		query  string
		target string
		want   bool
	}{
		{"", "anything", true},
		{"a", "abc", true},
		{"abc", "aXbYcZ", true},
		{"gmp", "gitmap", true},
		{"GMP", "GitMap", true},
		{"xyz", "abc", false},
		{"你好", "你好世界", true},
		{"好世", "你好世界", true},
		{"世好", "你好世界", false}, // wrong order
		{"abc", "a", false},         // query longer than target
		{"aa", "aba", true},         // repeated char, each matches once
		{"aab", "abacab", true},
	}

	for _, tt := range tests {
		got := fuzzyMatch(tt.query, tt.target)
		if got != tt.want {
			t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", tt.query, tt.target, got, tt.want)
		}
	}
}

func TestVisualRange(t *testing.T) {
	tests := []struct {
		name        string
		cursor      int
		visualStart int
		wantStart   int
		wantEnd     int
	}{
		{"cursor before start", 2, 5, 2, 5},
		{"start before cursor", 5, 2, 2, 5},
		{"same position", 3, 3, 3, 3},
		{"top selection", 0, 10, 0, 10},
		{"bottom selection", 10, 0, 0, 10},
	}

	for _, tt := range tests {
		m := model{cursor: tt.cursor, visualStart: tt.visualStart, visualMode: true}
		start, end := m.visualRange()
		if start != tt.wantStart || end != tt.wantEnd {
			t.Errorf("%s: visualRange() = (%d, %d), want (%d, %d)",
				tt.name, start, end, tt.wantStart, tt.wantEnd)
		}
	}
}
