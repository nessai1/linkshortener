package exitcheck

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// ExitCheckAnalyzer - анализатор проверки отсутсвия прямых вызовов os.Exit
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check direct call of os.Exit command",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.CallExpr:
				if isExitExpression(x) {
					pass.Reportf(x.Pos(), "Must not contain os.Exit expression")
				}
			}

			return true
		})
	}

	return nil, nil
}

func isExitExpression(c *ast.CallExpr) bool {
	if s, ok := c.Fun.(*ast.SelectorExpr); ok {
		if x, ok := s.X.(*ast.Ident); ok {
			return x.Name == "os" && s.Sel.Name == "Exit"
		}
	}

	return false
}
