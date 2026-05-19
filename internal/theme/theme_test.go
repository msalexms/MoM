package theme

import (
	"strings"
	"testing"
)

func TestRegistry_AllBuiltinThemesRegistered(t *testing.T) {
	expected := []string{"default", "dracula", "nord", "solarized-dark", "monochrome", "ascii"}
	for _, id := range expected {
		if _, ok := Get(id); !ok {
			t.Errorf("theme %q not registered", id)
		}
	}
}

func TestGet_UnknownReturnsDefault(t *testing.T) {
	got, ok := Get("nope")
	if ok {
		t.Error("expected ok=false for unknown theme")
	}
	if got == nil || got.ID != "default" {
		t.Errorf("expected fallback to default theme, got %#v", got)
	}
}

func TestColor_EmptyColorReturnsInputUnchanged(t *testing.T) {
	th := MustGet("default")
	if got := th.Color("hi", ""); got != "hi" {
		t.Errorf("empty color should be passthrough, got %q", got)
	}
	if got := th.Color("", th.Palette.Accent); got != "" {
		t.Errorf("empty input should be passthrough, got %q", got)
	}
}

func TestColor_AppendsResetExactlyOnce(t *testing.T) {
	th := MustGet("default")
	got := th.Color("hi", th.Palette.Accent)
	if !strings.HasSuffix(got, Reset) {
		t.Errorf("expected Reset suffix, got %q", got)
	}
	if strings.Count(got, Reset) != 1 {
		t.Errorf("expected exactly one Reset, got %q", got)
	}
}

func TestASCIITheme_HasNoEscapes(t *testing.T) {
	th := MustGet("ascii")
	if th.Palette.Accent != "" || th.Attrs.Bold != "" {
		t.Errorf("ascii theme leaked codes: %#v", th)
	}
	got := th.Color("hello", th.Palette.Accent)
	if got != "hello" {
		t.Errorf("expected raw text under ascii theme, got %q", got)
	}
	got = th.Bold("hello")
	if got != "hello" {
		t.Errorf("Bold under ascii should be passthrough, got %q", got)
	}
}

func TestPercentColor_Thresholds(t *testing.T) {
	th := MustGet("default")
	cases := []struct {
		percent float64
		want    string
	}{
		{0, th.Palette.Success},
		{50, th.Palette.Success},
		{59.9, th.Palette.Success},
		{60, th.Palette.Warning},
		{84.9, th.Palette.Warning},
		{85, th.Palette.Danger},
		{100, th.Palette.Danger},
	}
	for _, c := range cases {
		if got := th.PercentColor(c.percent); got != c.want {
			t.Errorf("PercentColor(%v) = %q, want %q", c.percent, got, c.want)
		}
	}
}

func TestSectionColor_KnownAndUnknown(t *testing.T) {
	th := MustGet("default")
	if th.SectionColor("system") != th.Palette.SectionSystem {
		t.Error("SectionColor(system) mismatch")
	}
	if th.SectionColor("does-not-exist") != th.Palette.Accent {
		t.Error("unknown section should fallback to Accent")
	}
}

func TestStatus_ClassifiesCommonStrings(t *testing.T) {
	th := MustGet("default")
	cases := []struct {
		in    string
		color string
		label string
	}{
		{"active", th.Palette.Success, "active"},
		{"running", th.Palette.Success, "active"},
		{"inactive", th.Palette.Danger, "inactive"},
		{"failed", th.Palette.Warning, "failed"},
		{"weird", th.Palette.Subtle, "weird"},
	}
	for _, c := range cases {
		gotColor, gotLabel := th.Status(c.in)
		if gotColor != c.color || gotLabel != c.label {
			t.Errorf("Status(%q) = (%q,%q), want (%q,%q)", c.in, gotColor, gotLabel, c.color, c.label)
		}
	}
}

func TestAll_SortedByID(t *testing.T) {
	all := All()
	if len(all) < 2 {
		t.Fatalf("expected multiple themes, got %d", len(all))
	}
	for i := 1; i < len(all); i++ {
		if all[i-1].ID > all[i].ID {
			t.Errorf("themes not sorted by ID: %q before %q", all[i-1].ID, all[i].ID)
		}
	}
}
