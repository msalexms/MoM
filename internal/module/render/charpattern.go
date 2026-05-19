package render

// charPattern returns the 5x5 ASCII pixel pattern for a single character.
// Supports A-Z, 0-9, and a few punctuation marks. Anything else returns an
// empty 5x5 block (used as a space).
//
// Each row is exactly 5 cells wide so callers can concatenate them
// horizontally without alignment fiddling.
func charPattern(ch rune) []string {
	switch ch {
	case 'A':
		return []string{"  █  ", " █ █ ", "█████", "█   █", "█   █"}
	case 'B':
		return []string{"████ ", "█   █", "████ ", "█   █", "████ "}
	case 'C':
		return []string{" ████", "█    ", "█    ", "█    ", " ████"}
	case 'D':
		return []string{"████ ", "█   █", "█   █", "█   █", "████ "}
	case 'E':
		return []string{"█████", "█    ", "███  ", "█    ", "█████"}
	case 'F':
		return []string{"█████", "█    ", "███  ", "█    ", "█    "}
	case 'G':
		return []string{" ████", "█    ", "█  ██", "█   █", " ████"}
	case 'H':
		return []string{"█   █", "█   █", "█████", "█   █", "█   █"}
	case 'I':
		return []string{"█████", "  █  ", "  █  ", "  █  ", "█████"}
	case 'J':
		return []string{"█████", "   █ ", "   █ ", "█  █ ", " ██  "}
	case 'K':
		return []string{"█   █", "█  █ ", "███  ", "█  █ ", "█   █"}
	case 'L':
		return []string{"█    ", "█    ", "█    ", "█    ", "█████"}
	case 'M':
		return []string{"█   █", "██ ██", "█ █ █", "█   █", "█   █"}
	case 'N':
		return []string{"█   █", "██  █", "█ █ █", "█  ██", "█   █"}
	case 'O':
		return []string{" ███ ", "█   █", "█   █", "█   █", " ███ "}
	case 'P':
		return []string{"████ ", "█   █", "████ ", "█    ", "█    "}
	case 'Q':
		return []string{" ███ ", "█   █", "█ █ █", "█  █ ", " ██ █"}
	case 'R':
		return []string{"████ ", "█   █", "████ ", "█  █ ", "█   █"}
	case 'S':
		return []string{" ████", "█    ", " ███ ", "    █", "████ "}
	case 'T':
		return []string{"█████", "  █  ", "  █  ", "  █  ", "  █  "}
	case 'U':
		return []string{"█   █", "█   █", "█   █", "█   █", " ███ "}
	case 'V':
		return []string{"█   █", "█   █", "█   █", " █ █ ", "  █  "}
	case 'W':
		return []string{"█   █", "█   █", "█ █ █", "██ ██", "█   █"}
	case 'X':
		return []string{"█   █", " █ █ ", "  █  ", " █ █ ", "█   █"}
	case 'Y':
		return []string{"█   █", " █ █ ", "  █  ", "  █  ", "  █  "}
	case 'Z':
		return []string{"█████", "   █ ", "  █  ", " █   ", "█████"}
	case '0':
		return []string{" ███ ", "█  ██", "█ █ █", "██  █", " ███ "}
	case '1':
		return []string{"  █  ", " ██  ", "  █  ", "  █  ", "█████"}
	case '2':
		return []string{" ███ ", "█   █", "  ██ ", " █   ", "█████"}
	case '3':
		return []string{" ███ ", "   █ ", "  ██ ", "   █ ", " ███ "}
	case '4':
		return []string{"█   █", "█   █", "█████", "    █", "    █"}
	case '5':
		return []string{"█████", "█    ", "████ ", "    █", "████ "}
	case '6':
		return []string{" ████", "█    ", "████ ", "█   █", " ████"}
	case '7':
		return []string{"█████", "   █ ", "  █  ", " █   ", " █   "}
	case '8':
		return []string{" ████", "█   █", " ███ ", "█   █", " ████"}
	case '9':
		return []string{" ████", "█   █", " ████", "    █", " ███ "}
	case '!':
		return []string{"  █  ", "  █  ", "  █  ", "     ", "  █  "}
	case '?':
		return []string{" ████", "    █", "  ██ ", "     ", "  █  "}
	case '.':
		return []string{"     ", "     ", "     ", "     ", "  █  "}
	case '-':
		return []string{"     ", "     ", "█████", "     ", "     "}
	case '_':
		return []string{"     ", "     ", "     ", "     ", "█████"}
	default:
		return []string{"     ", "     ", "     ", "     ", "     "}
	}
}
