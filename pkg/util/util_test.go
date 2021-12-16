package util

import (
	"testing"
	"time"
)

func TestSplitFindContains(t *testing.T) {
	tests := []struct {
		str, target, sep string
		match, expected  bool
	}{
		{"餐厅", "/", ",", false, false},
		{"餐厅", "/", ",", true, false},
		{"餐厅", "餐厅", ",", false, false},
		{"餐厅", "餐厅", ",", true, true},
	}

	for _, test := range tests {
		if output := SplitFindContains(test.str, test.target, test.sep, test.match); output != test.expected {
			t.Errorf("Output %v not equal to expected %v", output, test.expected)
		}
	}
}

func TestSplitFindTimeIntervalPositiveCases(t *testing.T) {
	tests := []struct {
		timeStr         string
		targetTime      time.Time
		match, expected bool
	}{
		{"11:00-13:00", time.Date(2021, 12, 15, 12, 34, 05, 0, time.UTC), false, false},
		{"11:00-13:00", time.Date(2021, 12, 15, 12, 34, 05, 0, time.UTC), true, true},
		{"11:00-13:00", time.Date(2021, 12, 15, 13, 34, 05, 0, time.UTC), false, false},
		{"11:00-13:00", time.Date(2021, 12, 15, 13, 34, 05, 0, time.UTC), true, false},

		{"23:00-01:00", time.Date(2021, 12, 15, 23, 55, 05, 0, time.UTC), true, true},
		{"23:00-01:00", time.Date(2021, 12, 15, 0, 34, 05, 0, time.UTC), true, true},
		{"23:00-01:00:05", time.Date(2021, 12, 15, 23, 55, 05, 0, time.UTC), true, true},
		{"23:00:05-01:00:00", time.Date(2021, 12, 15, 0, 34, 05, 0, time.UTC), true, true},
	}

	for _, test := range tests {
		if output, err := SplitFindTimeInterval(test.timeStr, test.targetTime, test.match); err != nil || output != test.expected {
			if err != nil {
				t.Error(err.Error())
			}
			t.Errorf("Output %v not equal to expected %v", output, test.expected)
		}
	}
}

func TestSplitFindTimeIntervalNegativeCases(t *testing.T) {
	tests := []struct {
		timeStr    string
		targetTime time.Time
		match      bool
	}{
		{"23:00-24:00", time.Date(2021, 12, 15, 12, 34, 05, 0, time.UTC), false},
		{"24:00-13:00", time.Date(2021, 12, 15, 12, 34, 05, 0, time.UTC), true},
		{"13:00-14:60", time.Date(2021, 12, 15, 12, 34, 05, 0, time.UTC), true},
		{"13:00-14:00:60", time.Date(2021, 12, 15, 12, 34, 05, 0, time.UTC), true},
		{"abc-def", time.Date(2021, 12, 15, 12, 34, 05, 0, time.UTC), false},
	}

	for _, test := range tests {
		if _, err := SplitFindTimeInterval(test.timeStr, test.targetTime, test.match); err == nil {
			t.Errorf("Function should throw an error. Input time string: `%s`", test.timeStr)
		}
	}
}
