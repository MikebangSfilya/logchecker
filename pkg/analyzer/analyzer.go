package analyzer

import (
	"go/ast"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var sensitiveRegex = regexp.MustCompile(`(?i)(password|token|api_key|secret)`)

var Analyzer = &analysis.Analyzer{
	Name:     "logchecker",
	Doc:      "chek the logs",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		obj := pass.TypesInfo.ObjectOf(sel.Sel)
		if obj == nil || obj.Pkg() == nil {
			return
		}
		pkgPath := obj.Pkg().Path()
		isSlog := pkgPath == "log/slog" || pkgPath == "log"
		isZap := pkgPath == "go.uber.org/zap"
		if !isSlog && !isZap {
			return
		}
		var msgIndex int
		switch sel.Sel.Name {
		case "Info", "Error", "Debug", "Warn", "Fatal":
			msgIndex = 0
		case "InfoContext", "ErrorContext", "DebugContext", "WarnContext":
			msgIndex = 1
		default:
			return
		}

		if len(call.Args) <= msgIndex {
			return
		}

		checkRules(pass, call.Args[msgIndex])
		checkSensitiveDataArgs(pass, call.Args)
	})

	return nil, nil
}

func checkRules(pass *analysis.Pass, expr ast.Expr) {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		strVal, err := strconv.Unquote(lit.Value)
		if err == nil && len(strVal) > 0 {
			checkIsLowerCase(pass, lit, strVal)
			checkIsEnglish(pass, lit, strVal)
			checkIsEmoji(pass, lit, strVal)
		}
	}

}

func checkIsLowerCase(pass *analysis.Pass, lit *ast.BasicLit, msg string) {
	if !isLowerCase(msg) {
		pass.Reportf(lit.Pos(), "the log message must begin with a lowercase letter")
	}
}

func checkIsEnglish(pass *analysis.Pass, lit *ast.BasicLit, msg string) {
	if !isEnglish(msg) {
		pass.Reportf(lit.Pos(), "the log message must be in English only")
	}
}

func checkIsEmoji(pass *analysis.Pass, lit *ast.BasicLit, msg string) {
	if hasSpecialCharsOrEmoji(msg) {
		pass.Reportf(lit.Pos(), "the log message must not contain special characters or emojis")
	}
}

func checkSensitiveDataArgs(pass *analysis.Pass, args []ast.Expr) {
	for _, arg := range args {
		ast.Inspect(arg, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.BasicLit:
				if v.Kind == token.STRING {
					val, err := strconv.Unquote(v.Value)
					if err == nil {
						if match := sensitiveRegex.FindString(val); match != "" {
							pass.Reportf(v.Pos(), "log message contains sensitive data: %s", strings.ToLower(match))
						}
					}
				}
			case *ast.Ident:
				if match := sensitiveRegex.FindString(v.Name); match != "" {
					pass.Reportf(v.Pos(), "attempt to log sensitive variable: %s", strings.ToLower(match))
				}
			}
			return true
		})
	}
}

func isLowerCase(msg string) bool {
	for _, r := range msg {
		if unicode.IsLetter(r) {
			return !unicode.IsUpper(r)
		}
	}
	return true
}

func isEnglish(msg string) bool {
	for _, r := range msg {
		if r > unicode.MaxASCII && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func hasSpecialCharsOrEmoji(msg string) bool {
	if msg == "" {
		return false
	}

	runes := []rune(msg)
	for i, r := range runes {
		if unicode.IsSymbol(r) || r == '!' || r == '?' {
			return true
		}
		if r == '.' {
			if i == len(runes)-1 {
				return true
			}
			if i+1 < len(runes) && runes[i+1] == '.' {
				return true
			}
		}
	}
	return false
}
