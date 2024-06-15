// exitcheck проверяет вызов функции Exit из пакета os.
// примеры в testdata.
// Проблема:
// наличие цикломатической сложность [15] у ф-ии inspect.
package exitcheck

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

const (
	importNameBit  uint8 = 1 << iota // 1 import "os"
	funcNameBit                      // 10 func main()
	callPackageBit                   // 100 os
	callMethodBit                    // 1000 Exit
)

const (
	importNameOS    = "\"os\""
	packageNameMain = "main"
	funcNameMain    = "main"
	pkgNameOS       = "os"
	methodNameExit  = "Exit"
)

func NewAnalyzer() *analysis.Analyzer {
	osExitAnalyzer := analysis.Analyzer{
		Name: "osExit",
		Doc:  "проверяет наличие вызова os.Exit",
		Run:  run,
	}

	return &osExitAnalyzer
}

func inspect(astFile *ast.File) (token.Pos, bool) {
	var (
		pos    token.Pos
		resBit uint8
	)

	localPackageName := pkgNameOS

	ast.Inspect(astFile, func(astNode ast.Node) bool {
		if resBit == (importNameBit | funcNameBit | callMethodBit | callPackageBit) {
			return true
		}

		switch xNode := astNode.(type) {
		case *ast.ImportSpec:
			if xNode.Path.Value == importNameOS {
				if xNode.Name != nil {
					localPackageName = xNode.Name.Name
				}

				resBit |= importNameBit
			}
		case *ast.FuncDecl:
			if resBit&importNameBit != 0 && xNode.Name.String() == funcNameMain {
				resBit |= funcNameBit
			}
		case *ast.SelectorExpr:
			if resBit&(importNameBit|funcNameBit|(resBit&^callPackageBit)) != 0 && xNode.Sel.Name == methodNameExit {
				resBit |= callMethodBit
			}
		case *ast.Ident:
			if resBit&(importNameBit|funcNameBit|callMethodBit) != 0 && xNode.Name == localPackageName {
				pos = xNode.Pos()
				resBit |= callPackageBit

				return true
			}

			resBit &^= (callMethodBit | callPackageBit)
		}

		return true
	})

	if resBit == (importNameBit | funcNameBit | callMethodBit | callPackageBit) {
		return pos, true
	}

	return 0, false
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() != packageNameMain { // если пакет не main, то пропускаем файл
			continue
		}

		pos, isFind := inspect(file)
		if isFind {
			pass.Reportf(pos, "прямой вызов os.Exit в функции main")
		}
	}

	return nil, nil
}
