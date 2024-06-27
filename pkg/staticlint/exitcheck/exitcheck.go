// exitcheck проверяет вызов функции Exit из пакета os.
package exitcheck

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

func NewAnalyzer() *analysis.Analyzer {
	osExitAnalyzer := analysis.Analyzer{
		Name: "osExit",
		Doc:  "проверяет наличие вызова os.Exit",
		Run:  run,
	}

	return &osExitAnalyzer
}

func run(pass *analysis.Pass) (interface{}, error) {
	var pos token.Pos

	fnCheck := func() bool {
		if pos > 0 {
			pass.Reportf(pos, "вызов os.Exit")

			return true
		}

		return false
	}

	for _, file := range pass.Files {
		if fnCheck() {
			return nil, nil
		}

		ast.Inspect(file, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			selexpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := selexpr.X.(*ast.Ident)
			if !ok || ident.Name != "os" {
				return true
			}

			if selexpr.Sel.Name == "Exit" {
				pos = selexpr.Pos()

				return false
			}

			return true
		})
	}

	if fnCheck() {
		return nil, nil
	}

	return nil, nil
}
