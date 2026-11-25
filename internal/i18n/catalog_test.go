package i18n

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// translatedArgs lists the calls whose string-literal arguments are shown
// to the user and therefore need a Spanish entry, with the argument
// positions to check.
var translatedArgs = map[string][]int{
	"T": {0}, "Tf": {0}, // i18n.T / tui.T
	"Section": {0}, "Header": {0}, "Field": {0}, // render translates these
	"stringSetting": {1, 2}, "enumSetting": {1, 2}, "boolSetting": {1, 2},
	"intSetting": {1, 2}, "listSetting": {1, 2},
}

// sameInSpanish are user-facing strings that need no translation.
var sameInSpanish = map[string]bool{
	"GPU": true, "Git": true, "Fail2ban": true, "Firewall": false, "MOTD": true, "kernel": true, "shell": true,
	"total": true, "local": true, "vram": true, "temp": true, "cpu 1m": true, "cpu 5m": true, "cpu 15m": true,
	"ram": true, "swap": true, "Timers": true, "a/d": true, "%d: %s": true, "no": true, "original": true,
	"mom: error %s: %v\n": true, "Shift+↑↓": false,
}

// TestSpanishCatalogIsComplete parses the source tree and fails for every
// user-facing literal that has no Spanish translation, so a new string
// cannot silently stay in English in the Spanish interface.
func TestSpanishCatalogIsComplete(t *testing.T) {
	root := filepath.Join("..", "..")
	fset := token.NewFileSet()
	missing := map[string]string{}

	check := func(lit ast.Expr, pos token.Pos) {
		bl, ok := lit.(*ast.BasicLit)
		if !ok || bl.Kind != token.STRING {
			return
		}
		s, err := strconv.Unquote(bl.Value)
		if err != nil || strings.TrimSpace(s) == "" || sameInSpanish[s] {
			return
		}
		if _, ok := es[s]; !ok {
			missing[s] = fset.Position(pos).String()
		}
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CallExpr:
				name := ""
				switch fn := n.Fun.(type) {
				case *ast.Ident:
					name = fn.Name
				case *ast.SelectorExpr:
					name = fn.Sel.Name
				}
				for _, i := range translatedArgs[name] {
					if i < len(n.Args) {
						check(n.Args[i], n.Pos())
					}
				}
			case *ast.FuncDecl:
				// Module Title() and Description() return display text.
				if n.Recv != nil && (n.Name.Name == "Title" || n.Name.Name == "Description") && n.Body != nil {
					for _, st := range n.Body.List {
						if r, ok := st.(*ast.ReturnStmt); ok && len(r.Results) == 1 {
							check(r.Results[0], r.Pos())
						}
					}
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for s, pos := range missing {
		t.Errorf("%s: no Spanish translation for %q", pos, s)
	}
}

func TestTranslate(t *testing.T) {
	t.Cleanup(func() { Set(EN) })
	if T("Modules") != "Modules" {
		t.Error("English must return the source text")
	}
	Set(ES)
	if T("Modules") != "Módulos" {
		t.Errorf("T(Modules) = %q", T("Modules"))
	}
	if T("not in the catalog") != "not in the catalog" {
		t.Error("unknown text must fall back to the source")
	}
	if Parse("fr") != Default || Parse("es") != ES {
		t.Error("Parse")
	}
}
