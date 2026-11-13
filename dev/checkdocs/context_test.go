package main

import "testing"

func TestCheckContext(t *testing.T) {
	cases := []struct {
		name   string
		readme string
		want   int
	}{
		{"valid period", "## Overview\nx\n## Context\n- Period: 2025-10 to 2025-12\n## Architecture\n", 0},
		{"year only", "## Overview\n## Context\nWritten in 2024.\n## Architecture\n", 0},
		{"month only", "## Overview\n## Context\nStarted 2025-03.\n## Architecture\n", 0},
		{"year to month", "## Overview\n## Context\n2024 to 2026-04\n## Architecture\n", 0},
		{"month to year", "## Overview\n## Context\n2024-05 to 2026\n## Architecture\n", 0},
		{"year to year", "## Overview\n## Context\n2024 to 2024\n## Architecture\n", 0},
		{"year inside a month of the same year", "## Overview\n## Context\n2024 to 2024-03\n## Architecture\n", 0},
		{"ends before it starts", "## Overview\n## Context\n2026-04 to 2024\n## Architecture\n", 1},
		{"ends before it starts by month", "## Overview\n## Context\n2025-10 to 2025-03\n## Architecture\n", 1},
		{"missing section", "## Overview\nx\n## Architecture\n", 1},
		{"wrong place", "## Overview\nx\n## Architecture\n## Context\n2025-10 to 2025-12\n", 1},
		{"no date", "## Overview\n## Context\nsome text\n## Architecture\n", 1},
		{"number that is not a year", "## Overview\n## Context\nversion 1234 and 99\n## Architecture\n", 1},
		{"bad month and no other date", "## Overview\n## Context\n2025-13\n", 1},
		{"no overview", "## Context\n2025-10 to 2025-12\n", 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := checkContext(c.readme); len(got) != c.want {
				t.Fatalf("got %d problems %v, want %d", len(got), got, c.want)
			}
		})
	}
}
