package analyzer

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

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
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return
		}
		pkgName := ident.Name
		methodName := sel.Sel.Name
		isSlog := pkgName == "slog" || pkgName == "log"
		isZap := pkgName == "zap" || pkgName == "logger"
		if !isSlog && !isZap {
			return
		}
		if methodName != "Info" && methodName != "Error" && methodName != "Debug" && methodName != "Warn" {
			return
		}
		if len(call.Args) == 0 {
			return
		}
		msgArg := call.Args[0]
		checkRules(pass, msgArg)
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
	checkSensitiveData(pass, expr)

}

func checkIsLowerCase(pass *analysis.Pass, lit *ast.BasicLit, msg string) {
	firstRune := []rune(msg)[0]
	if unicode.IsUpper(firstRune) {
		pass.Reportf(lit.Pos(), "the log message must begin with a lowercase letter")
	}
}

func checkIsEnglish(pass *analysis.Pass, lit *ast.BasicLit, msg string) {
	for _, r := range msg {
		if r > unicode.MaxASCII && unicode.IsLetter(r) {
			pass.Reportf(lit.Pos(), "the log message must be in English only")
			return
		}
	}

}

func checkIsEmoji(pass *analysis.Pass, lit *ast.BasicLit, msg string) {
	for _, r := range msg {
		if unicode.IsSymbol(r) || r == '!' || r == '?' || r == '.' {
			pass.Reportf(lit.Pos(), "the log message must not contain special characters or emojis")
			return
		}
	}
}

func checkSensitiveData(pass *analysis.Pass, expr ast.Expr) {
	// Список триггеров
	keywords := []string{"password", "token", "api_key", "secret"}

	var inspect func(e ast.Expr)
	inspect = func(e ast.Expr) {
		switch v := e.(type) {
		case *ast.BasicLit:
			if v.Kind == token.STRING {
				val, _ := strconv.Unquote(v.Value)
				s := strings.ToLower(val)
				for _, kw := range keywords {
					if strings.Contains(s, kw) {
						pass.Reportf(v.Pos(), "log message contains sensitive data: %s", kw)
					}
				}
			}
		case *ast.Ident:
			name := strings.ToLower(v.Name)
			for _, kw := range keywords {
				if strings.Contains(name, kw) {
					pass.Reportf(v.Pos(), "attempt to log sensitive variable: %s", kw)
				}
			}
		case *ast.BinaryExpr:
			if v.Op == token.ADD {
				inspect(v.X)
				inspect(v.Y)
			}
		}
	}
	inspect(expr)
}
