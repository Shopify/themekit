package decor

import (
	"fmt"
	"testing"
)

func TestCounterKiB(t *testing.T) {
	cases := map[string]struct {
		value    int64
		verb     string
		expected string
	}{
		"verb %f":   {12345678, "%f", "11.773756MiB"},
		"verb %.0f": {12345678, "%.0f", "12MiB"},
		"verb %.1f": {12345678, "%.1f", "11.8MiB"},
		"verb %.2f": {12345678, "%.2f", "11.77MiB"},
		"verb %.3f": {12345678, "%.3f", "11.774MiB"},

		"verb % f":   {12345678, "% f", "11.773756 MiB"},
		"verb % .0f": {12345678, "% .0f", "12 MiB"},
		"verb % .1f": {12345678, "% .1f", "11.8 MiB"},
		"verb % .2f": {12345678, "% .2f", "11.77 MiB"},
		"verb % .3f": {12345678, "% .3f", "11.774 MiB"},

		"verb %8.f":  {12345678, "%8.f", "   12MiB"},
		"verb %8.0f": {12345678, "%8.0f", "   12MiB"},
		"verb %8.1f": {12345678, "%8.1f", " 11.8MiB"},
		"verb %8.2f": {12345678, "%8.2f", "11.77MiB"},
		"verb %8.3f": {12345678, "%8.3f", "11.774MiB"},

		"verb % 8.f":  {12345678, "% 8.f", "  12 MiB"},
		"verb % 8.0f": {12345678, "% 8.0f", "  12 MiB"},
		"verb % 8.1f": {12345678, "% 8.1f", "11.8 MiB"},

		"verb %-8.f":  {12345678, "%-8.f", "12MiB   "},
		"verb %-8.0f": {12345678, "%-8.0f", "12MiB   "},
		"verb %-8.1f": {12345678, "%-8.1f", "11.8MiB "},
		"verb %-8.2f": {12345678, "%8.2f", "11.77MiB"},
		"verb %-8.3f": {12345678, "%8.3f", "11.774MiB"},

		"verb % -8.f":  {12345678, "% -8.f", "12 MiB  "},
		"verb % -8.0f": {12345678, "% -8.0f", "12 MiB  "},
		"verb % -8.1f": {12345678, "% -8.1f", "11.8 MiB"},

		"1000 %f":           {1000, "%f", "1000b"},
		"1000 %d":           {1000, "%d", "1000b"},
		"1000 %s":           {1000, "%s", "1000b"},
		"1024 %f":           {1024, "%f", "1.000000KiB"},
		"1024 %d":           {1024, "%d", "1KiB"},
		"1024 %.1f":         {1024, "%.1f", "1.0KiB"},
		"1024 %s":           {1024, "%s", "1.0KiB"},
		"3*MiB+140KiB %f":   {3*MiB + 140*KiB, "%f", "3.136719MiB"},
		"3*MiB+140KiB %d":   {3*MiB + 140*KiB, "%d", "3MiB"},
		"3*MiB+140KiB %.1f": {3*MiB + 140*KiB, "%.1f", "3.1MiB"},
		"3*MiB+140KiB %s":   {3*MiB + 140*KiB, "%s", "3.1MiB"},
		"2*GiB %f":          {2 * GiB, "%f", "2.000000GiB"},
		"2*GiB %d":          {2 * GiB, "%d", "2GiB"},
		"2*GiB %.1f":        {2 * GiB, "%.1f", "2.0GiB"},
		"2*GiB %s":          {2 * GiB, "%s", "2.0GiB"},
		"4*TiB %f":          {4 * TiB, "%f", "4.000000TiB"},
		"4*TiB %d":          {4 * TiB, "%d", "4TiB"},
		"4*TiB %.1f":        {4 * TiB, "%.1f", "4.0TiB"},
		"4*TiB %s":          {4 * TiB, "%s", "4.0TiB"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := fmt.Sprintf(tc.verb, CounterKiB(tc.value))
			if got != tc.expected {
				t.Fatalf("expected: %q, got: %q\n", tc.expected, got)
			}
		})
	}
}

func TestCounterKB(t *testing.T) {
	cases := map[string]struct {
		value    int64
		verb     string
		expected string
	}{
		"verb %f":   {12345678, "%f", "12.345678MB"},
		"verb %.0f": {12345678, "%.0f", "12MB"},
		"verb %.1f": {12345678, "%.1f", "12.3MB"},
		"verb %.2f": {12345678, "%.2f", "12.35MB"},
		"verb %.3f": {12345678, "%.3f", "12.346MB"},

		"verb % f":   {12345678, "% f", "12.345678 MB"},
		"verb % .0f": {12345678, "% .0f", "12 MB"},
		"verb % .1f": {12345678, "% .1f", "12.3 MB"},
		"verb % .2f": {12345678, "% .2f", "12.35 MB"},
		"verb % .3f": {12345678, "% .3f", "12.346 MB"},

		"verb %8.f":  {12345678, "%8.f", "    12MB"},
		"verb %8.0f": {12345678, "%8.0f", "    12MB"},
		"verb %8.1f": {12345678, "%8.1f", "  12.3MB"},
		"verb %8.2f": {12345678, "%8.2f", " 12.35MB"},
		"verb %8.3f": {12345678, "%8.3f", "12.346MB"},

		"verb % 8.f":  {12345678, "% 8.f", "   12 MB"},
		"verb % 8.0f": {12345678, "% 8.0f", "   12 MB"},
		"verb % 8.1f": {12345678, "% 8.1f", " 12.3 MB"},

		"verb %-8.f":  {12345678, "%-8.f", "12MB    "},
		"verb %-8.0f": {12345678, "%-8.0f", "12MB    "},
		"verb %-8.1f": {12345678, "%-8.1f", "12.3MB  "},
		"verb %-8.2f": {12345678, "%8.2f", " 12.35MB"},
		"verb %-8.3f": {12345678, "%8.3f", "12.346MB"},

		"verb % -8.f":  {12345678, "% -8.f", "12 MB   "},
		"verb % -8.0f": {12345678, "% -8.0f", "12 MB   "},
		"verb % -8.1f": {12345678, "% -8.1f", "12.3 MB "},

		"1000 %f":          {1000, "%f", "1.000000kB"},
		"1000 %d":          {1000, "%d", "1kB"},
		"1000 %s":          {1000, "%s", "1.0kB"},
		"1024 %f":          {1024, "%f", "1.024000kB"},
		"1024 %d":          {1024, "%d", "1kB"},
		"1024 %.1f":        {1024, "%.1f", "1.0kB"},
		"1024 %s":          {1024, "%s", "1.0kB"},
		"3*MB+140*KB %f":   {3*MB + 140*KB, "%f", "3.140000MB"},
		"3*MB+140*KB %d":   {3*MB + 140*KB, "%d", "3MB"},
		"3*MB+140*KB %.1f": {3*MB + 140*KB, "%.1f", "3.1MB"},
		"3*MB+140*KB %s":   {3*MB + 140*KB, "%s", "3.1MB"},
		"2*GB %f":          {2 * GB, "%f", "2.000000GB"},
		"2*GB %d":          {2 * GB, "%d", "2GB"},
		"2*GB %.1f":        {2 * GB, "%.1f", "2.0GB"},
		"2*GB %s":          {2 * GB, "%s", "2.0GB"},
		"4*TB %f":          {4 * TB, "%f", "4.000000TB"},
		"4*TB %d":          {4 * TB, "%d", "4TB"},
		"4*TB %.1f":        {4 * TB, "%.1f", "4.0TB"},
		"4*TB %s":          {4 * TB, "%s", "4.0TB"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := fmt.Sprintf(tc.verb, CounterKB(tc.value))
			if got != tc.expected {
				t.Fatalf("expected: %q, got: %q\n", tc.expected, got)
			}
		})
	}
}
