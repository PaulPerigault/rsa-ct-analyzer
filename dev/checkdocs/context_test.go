package main

import "testing"

func TestCheckContext(t *testing.T) {
	cases := []struct {
		name   string
		readme string
		want   int
	}{
		{"valid", "## Overview\nx\n## Context\n- Period: 2025-10 to 2025-12\n## Architecture\n", 0},
		{"missing section", "## Overview\nx\n## Architecture\n", 1},
		{"wrong place", "## Overview\nx\n## Architecture\n## Context\n2025-10 to 2025-12\n", 1},
		{"no period", "## Overview\n## Context\nsome text\n## Architecture\n", 1},
		{"bad month", "## Overview\n## Context\n2025-13 to 2025-12\n", 1},
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
