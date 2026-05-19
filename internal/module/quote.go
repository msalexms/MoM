package module

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/ams/mom/internal/module/render"
)

// QuoteModule displays a random quote about technology or programming.
type QuoteModule struct{}

func (m *QuoteModule) Name() string           { return "quote" }
func (m *QuoteModule) Title() string          { return "Quote" }
func (m *QuoteModule) Description() string    { return "Random programming/tech quote" }
func (m *QuoteModule) Dependencies() []string { return nil }
func (m *QuoteModule) Available() bool        { return true }
func (m *QuoteModule) DefaultEnabled() bool   { return false }

func (m *QuoteModule) Variants() []render.Variant {
	return []render.Variant{render.VariantDefault, render.VariantMinimal, render.VariantBoxed, render.VariantPowerline, render.VariantCards}
}
func (m *QuoteModule) DefaultVariant() render.Variant { return render.VariantDefault }
func (m *QuoteModule) Settings() []SettingDef         { return nil }

func (m *QuoteModule) Generate(ctx context.Context) (string, error) {
	return m.GenerateThemed(ctx, render.DefaultOptions())
}

func (m *QuoteModule) GenerateThemed(ctx context.Context, opts render.Options) (string, error) {
	q := quotes[rand.IntN(len(quotes))]
	r := render.New(opts)
	th := r.Theme()

	var sb strings.Builder

	switch r.Variant() {
	case render.VariantBoxed:
		var content strings.Builder
		wrapped := wordWrap(q.Text, 38)
		for _, line := range wrapped {
			content.WriteString(th.Italic(line) + "\n")
		}
		content.WriteString("\n" + th.Dim("— "+q.Author))
		sb.WriteString(render.Indent(r.Box(content.String(), "Quote"), "  "))

	case render.VariantPowerline:
		sb.WriteString(r.Header("Quote", "quote"))
		sb.WriteString("\n\n")
		wrapped := wordWrap(q.Text, 44)
		for _, line := range wrapped {
			sb.WriteString("    " + th.Color("▌", th.Palette.Secondary) + " " + th.Italic(line) + "\n")
		}
		sb.WriteString("    " + th.Color("▌", th.Palette.Subtle) + " " + th.Dim("— "+q.Author))

	case render.VariantCards:
		var content strings.Builder
		wrapped := wordWrap(q.Text, 36)
		for _, line := range wrapped {
			content.WriteString("  " + th.Italic(line) + "\n")
		}
		content.WriteString("\n  " + th.Dim("— "+q.Author))
		sb.WriteString(render.Indent(r.Card(content.String(), "Quote"), "  "))

	case render.VariantMinimal:
		sb.WriteString(fmt.Sprintf("  %s\"%s\" — %s", r.Icon("quote")+" ", q.Text, q.Author))

	default:
		sb.WriteString(r.Header("Quote", "quote"))
		sb.WriteString("\n\n")
		wrapped := wordWrap(q.Text, 42)
		for _, line := range wrapped {
			sb.WriteString("  " + th.Italic(th.Color(line, th.Palette.Foreground)) + "\n")
		}
		sb.WriteString(fmt.Sprintf("\n    %s", th.Dim("— "+q.Author)))
	}

	return sb.String(), nil
}

func wordWrap(text string, width int) []string {
	words := strings.Fields(text)
	var lines []string
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
		} else if len(current)+1+len(word) <= width {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

type quote struct {
	Text   string
	Author string
}

var quotes = []quote{
	{"Talk is cheap. Show me the code.", "Linus Torvalds"},
	{"The best way to predict the future is to invent it.", "Alan Kay"},
	{"Simplicity is the soul of efficiency.", "Austin Freeman"},
	{"First, solve the problem. Then, write the code.", "John Johnson"},
	{"Any fool can write code that a computer can understand. Good programmers write code that humans can understand.", "Martin Fowler"},
	{"Programs must be written for people to read, and only incidentally for machines to execute.", "Harold Abelson"},
	{"The most dangerous phrase in the language is: We've always done it this way.", "Grace Hopper"},
	{"Unix is simple. It just takes a genius to understand its simplicity.", "Dennis Ritchie"},
	{"Perfection is achieved not when there is nothing more to add, but when there is nothing left to take away.", "Antoine de Saint-Exupery"},
	{"Debugging is twice as hard as writing the code in the first place.", "Brian Kernighan"},
	{"It works on my machine.", "Every developer ever"},
	{"There are only two hard things in Computer Science: cache invalidation and naming things.", "Phil Karlton"},
	{"Measuring programming progress by lines of code is like measuring aircraft building progress by weight.", "Bill Gates"},
	{"Before software can be reusable it first has to be usable.", "Ralph Johnson"},
	{"The best error message is the one that never shows up.", "Thomas Fuchs"},
	{"Code is like humor. When you have to explain it, it's bad.", "Cory House"},
	{"Optimism is an occupational hazard of programming; feedback is the treatment.", "Kent Beck"},
	{"One of my most productive days was throwing away 1000 lines of code.", "Ken Thompson"},
	{"If debugging is the process of removing bugs, then programming must be the process of putting them in.", "Edsger Dijkstra"},
	{"It's not a bug; it's an undocumented feature.", "Anonymous"},
	{"Good code is its own best documentation.", "Steve McConnell"},
	{"Premature optimization is the root of all evil.", "Donald Knuth"},
	{"The best thing about a boolean is even if you are wrong, you are only off by a bit.", "Anonymous"},
	{"Truth can only be found in one place: the code.", "Robert C. Martin"},
	{"Always code as if the guy who ends up maintaining your code will be a violent psychopath who knows where you live.", "John Woods"},
	{"Documentation is a love letter that you write to your future self.", "Damian Conway"},
	{"Linux is only free if your time has no value.", "Jamie Zawinski"},
	{"There is no cloud. It's just someone else's computer.", "Anonymous"},
	{"I don't always test my code, but when I do, I do it in production.", "Anonymous"},
	{"In a world without fences and walls, who needs Gates and Windows?", "Linux proverb"},
	{"Linux: Because rebooting is for adding new hardware.", "Anonymous"},
	{"UNIX was not designed to stop its users from doing stupid things, as that would also stop them from doing clever things.", "Doug Gwyn"},
	{"Life is too short to remove USB safely.", "Anonymous"},
}
