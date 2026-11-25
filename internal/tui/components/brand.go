package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// thorPixels is the AMSTech Thor mark (the head with the >_ prompt) on a
// 12×12 grid: '#' head, 'p' prompt, '.' empty. Two pixel rows are packed
// into one text row with half blocks, so it renders 12 cells × 6 rows.
var thorPixels = []string{
	"..#......#..",
	"..##....##..",
	".####..####.",
	"############",
	"############",
	"###p########",
	"####p#######",
	"#####p######",
	"####p#######",
	"###p##pppp##",
	"############",
	".##########.",
}

// ThorLogo renders the brand mark with half-block characters.
func ThorLogo() string {
	color := func(px byte) string {
		switch px {
		case '#':
			return ColorThor
		case 'p':
			return ColorInk
		}
		return ""
	}

	var rows []string
	for y := 0; y+1 < len(thorPixels); y += 2 {
		var sb strings.Builder
		for x := range len(thorPixels[y]) {
			top, bottom := color(thorPixels[y][x]), color(thorPixels[y+1][x])
			sb.WriteString(halfBlock(top, bottom))
		}
		rows = append(rows, sb.String())
	}
	return strings.Join(rows, "\n")
}

func halfBlock(top, bottom string) string {
	st := lipgloss.NewStyle()
	switch {
	case top == "" && bottom == "":
		return " "
	case top == bottom:
		return st.Foreground(lipgloss.Color(top)).Render("█")
	case top == "":
		return st.Foreground(lipgloss.Color(bottom)).Render("▄")
	case bottom == "":
		return st.Foreground(lipgloss.Color(top)).Render("▀")
	default:
		return st.Foreground(lipgloss.Color(top)).Background(lipgloss.Color(bottom)).Render("▀")
	}
}
