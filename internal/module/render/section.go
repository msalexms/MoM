package render

import (
	"fmt"
	"strings"

	"github.com/ams/mom/internal/theme"
)

// Section renders a complete module section applying the current variant's
// framing around the provided content lines.
//
// Parameters:
//   - title:   section header text (e.g. "Timers", "System")
//   - module:  module name for theme color lookup
//   - compact: one-liner summary for compact variant (empty = skip compact)
//   - lines:   content lines (already formatted with colors, no indent or frame)
//
// Section centralizes ALL variant-specific framing logic so that modules only
// need to produce their content lines and call Section(). Adding a new variant
// only requires modifying this function.
func (r *Renderer) Section(title, module, compact string, lines []string) string {
	th := r.Opts.Theme

	switch r.Opts.Variant {
	case VariantCompact:
		return r.sectionCompact(title, module, compact)
	case VariantBoxed:
		return r.sectionBoxed(title, lines)
	case VariantCards:
		return r.sectionCards(title, lines)
	case VariantPowerline:
		return r.sectionPowerline(title, module, th, lines)
	case VariantMinimal:
		return r.sectionMinimal(title, module, th, lines)
	default:
		return r.sectionDefault(title, module, lines)
	}
}

// SectionCustom is like Section but allows a custom header (e.g. "Containers (docker)")
// instead of deriving the color from the module name.
func (r *Renderer) SectionCustom(title, headerColor, compact string, lines []string) string {
	th := r.Opts.Theme

	switch r.Opts.Variant {
	case VariantCompact:
		var sb strings.Builder
		sb.WriteString(r.HeaderColor(title, headerColor))
		if compact != "" {
			sb.WriteString("\n    " + th.Dim(compact))
		}
		return sb.String()
	case VariantBoxed:
		return r.sectionBoxed(title, lines)
	case VariantCards:
		return r.sectionCards(title, lines)
	case VariantPowerline:
		var sb strings.Builder
		sb.WriteString(r.HeaderColor(title, headerColor))
		sb.WriteString("\n\n")
		for _, l := range lines {
			sb.WriteString(fmt.Sprintf("    %s %s\n", th.Color("▌", th.Palette.Accent), l))
		}
		return strings.TrimRight(sb.String(), "\n")
	case VariantMinimal:
		var sb strings.Builder
		sb.WriteString("  " + th.Bold(th.Color(title, headerColor)))
		sb.WriteString("\n")
		for _, l := range lines {
			sb.WriteString("    " + l + "\n")
		}
		return strings.TrimRight(sb.String(), "\n")
	default:
		var sb strings.Builder
		sb.WriteString(r.HeaderColor(title, headerColor))
		sb.WriteString("\n\n")
		for _, l := range lines {
			sb.WriteString("    " + l + "\n")
		}
		return strings.TrimRight(sb.String(), "\n")
	}
}

func (r *Renderer) sectionCompact(title, module, compact string) string {
	th := r.Opts.Theme
	var sb strings.Builder
	sb.WriteString(r.Header(title, module))
	if compact != "" {
		sb.WriteString("\n    " + th.Dim(compact))
	}
	return sb.String()
}

func (r *Renderer) sectionBoxed(title string, lines []string) string {
	content := strings.Join(lines, "\n")
	return Indent(r.Box(content, title), "  ")
}

func (r *Renderer) sectionCards(title string, lines []string) string {
	// Cards content has 2-space indent inside the card
	var padded []string
	for _, l := range lines {
		padded = append(padded, "  "+l)
	}
	content := strings.Join(padded, "\n")
	return Indent(r.Card(content, title), "  ")
}

func (r *Renderer) sectionPowerline(title, module string, th *theme.Theme, lines []string) string {
	var sb strings.Builder
	sb.WriteString(r.Header(title, module))
	sb.WriteString("\n\n")
	for _, l := range lines {
		sb.WriteString(fmt.Sprintf("    %s %s\n", th.Color("▌", th.Palette.Accent), l))
	}
	return strings.TrimRight(sb.String(), "\n")
}

func (r *Renderer) sectionMinimal(title, module string, th *theme.Theme, lines []string) string {
	color := th.SectionColor(module)
	var sb strings.Builder
	sb.WriteString("  " + th.Bold(th.Color(title, color)))
	sb.WriteString("\n")
	for _, l := range lines {
		sb.WriteString("    " + l + "\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

func (r *Renderer) sectionDefault(title, module string, lines []string) string {
	var sb strings.Builder
	sb.WriteString(r.Header(title, module))
	sb.WriteString("\n\n")
	for _, l := range lines {
		sb.WriteString("    " + l + "\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}
