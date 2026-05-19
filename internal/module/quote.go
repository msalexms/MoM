package module

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
)

// QuoteModule displays a random quote about technology or programming.
type QuoteModule struct{}

func (m *QuoteModule) Name() string        { return "quote" }
func (m *QuoteModule) Title() string       { return "Quote of the Day" }
func (m *QuoteModule) Description() string { return "Random programming/tech quote" }
func (m *QuoteModule) Dependencies() []string { return nil }
func (m *QuoteModule) Available() bool     { return true }
func (m *QuoteModule) DefaultEnabled() bool { return false }

func (m *QuoteModule) Generate(ctx context.Context) (string, error) {
	quote := quotes[rand.IntN(len(quotes))]

	// Word wrap the quote to fit in the box
	wrapped := wordWrap(quote, 37)

	var sb strings.Builder
	sb.WriteString("┌─ Quote ──────────────────────────────┐\n")
	for _, line := range wrapped {
		sb.WriteString(fmt.Sprintf("│ %-37s │\n", line))
	}
	sb.WriteString("└───────────────────────────────────────┘")

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

var quotes = []string{
	"\"Talk is cheap. Show me the code.\" — Linus Torvalds",
	"\"The best way to predict the future is to invent it.\" — Alan Kay",
	"\"Simplicity is the soul of efficiency.\" — Austin Freeman",
	"\"First, solve the problem. Then, write the code.\" — John Johnson",
	"\"Any fool can write code that a computer can understand. Good programmers write code that humans can understand.\" — Martin Fowler",
	"\"Programs must be written for people to read, and only incidentally for machines to execute.\" — Harold Abelson",
	"\"The most dangerous phrase in the language is: We've always done it this way.\" — Grace Hopper",
	"\"Unix is simple. It just takes a genius to understand its simplicity.\" — Dennis Ritchie",
	"\"In theory, there is no difference between theory and practice. In practice, there is.\" — Yogi Berra",
	"\"Perfection is achieved not when there is nothing more to add, but when there is nothing left to take away.\" — Antoine de Saint-Exupéry",
	"\"The only way to do great work is to love what you do.\" — Steve Jobs",
	"\"Debugging is twice as hard as writing the code in the first place.\" — Brian Kernighan",
	"\"It works on my machine.\" — Every developer ever",
	"\"There are only two hard things in Computer Science: cache invalidation and naming things.\" — Phil Karlton",
	"\"The computer was born to solve problems that did not exist before.\" — Bill Gates",
	"\"Software is like entropy: it is difficult to grasp, weighs nothing, and obeys the Second Law of Thermodynamics; i.e., it always increases.\" — Norman Augustine",
	"\"Measuring programming progress by lines of code is like measuring aircraft building progress by weight.\" — Bill Gates",
	"\"Before software can be reusable it first has to be usable.\" — Ralph Johnson",
	"\"The best error message is the one that never shows up.\" — Thomas Fuchs",
	"\"Code is like humor. When you have to explain it, it's bad.\" — Cory House",
	"\"Fix the cause, not the symptom.\" — Steve Maguire",
	"\"Optimism is an occupational hazard of programming; feedback is the treatment.\" — Kent Beck",
	"\"The function of good software is to make the complex appear to be simple.\" — Grady Booch",
	"\"One of my most productive days was throwing away 1000 lines of code.\" — Ken Thompson",
	"\"If debugging is the process of removing bugs, then programming must be the process of putting them in.\" — Edsger Dijkstra",
	"\"Walking on water and developing software from a specification are easy if both are frozen.\" — Edward V Berard",
	"\"It's not a bug; it's an undocumented feature.\" — Anonymous",
	"\"The most important property of a program is whether it accomplishes the intention of its user.\" — C.A.R. Hoare",
	"\"Good code is its own best documentation.\" — Steve McConnell",
	"\"Premature optimization is the root of all evil.\" — Donald Knuth",
	"\"Computers are useless. They can only give you answers.\" — Pablo Picasso",
	"\"The best thing about a boolean is even if you are wrong, you are only off by a bit.\" — Anonymous",
	"\"Experience is the name everyone gives to their mistakes.\" — Oscar Wilde",
	"\"Java is to JavaScript what car is to carpet.\" — Chris Heilmann",
	"\"Truth can only be found in one place: the code.\" — Robert C. Martin",
	"\"Give someone a program, you frustrate them for a day; teach them how to program, you frustrate them for a lifetime.\" — David Leinweber",
	"\"Programming is the art of telling another human what one wants the computer to do.\" — Donald Knuth",
	"\"Always code as if the guy who ends up maintaining your code will be a violent psychopath who knows where you live.\" — John Woods",
	"\"Documentation is a love letter that you write to your future self.\" — Damian Conway",
	"\"Linux is only free if your time has no value.\" — Jamie Zawinski",
	"\"The Internet? Is that thing still around?\" — Homer Simpson",
	"\"There is no cloud. It's just someone else's computer.\" — Anonymous",
	"\"rm -rf / — A sysadmin's lullaby.\" — Anonymous",
	"\"I don't always test my code, but when I do, I do it in production.\" — Anonymous",
	"\"With great power comes great responsibility.\" — Uncle Ben (and every sudo user)",
	"\"A computer lets you make more mistakes faster than any other invention with the possible exceptions of handguns and Tequila.\" — Mitch Ratcliffe",
	"\"In a world without fences and walls, who needs Gates and Windows?\" — Linux proverb",
	"\"Linux: Because rebooting is for adding new hardware.\" — Anonymous",
	"\"UNIX was not designed to stop its users from doing stupid things, as that would also stop them from doing clever things.\" — Doug Gwyn",
	"\"Life is too short to remove USB safely.\" — Anonymous",
}
