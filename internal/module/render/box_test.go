package render

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ams/mom/internal/theme"
)

func TestBox_AlignedBorders(t *testing.T) {
	opts := Options{Theme: theme.MustGet("default"), Variant: VariantBoxed}
	r := New(opts)

	content := "host      myserver\nkernel    Linux 6.8.0\nuptime    3d 14h 22m"
	box := Indent(r.Box(content, "System"), "  ")

	lines := strings.Split(box, "\n")
	if len(lines) < 3 {
		t.Fatalf("box too short: %d lines", len(lines))
	}

	// All lines should have same visible width
	topW := visibleLen(lines[0])
	for i, l := range lines {
		w := visibleLen(l)
		if w != topW {
			t.Errorf("line %d width %d != top width %d\n  line: %q", i, w, topW, l)
		}
	}

	// All lines should start with "  " (the indent)
	for i, l := range lines {
		if !strings.HasPrefix(l, "  ") {
			t.Errorf("line %d missing indent: %q", i, l)
		}
	}

	fmt.Println(box)
}
