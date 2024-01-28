package exitcheck

import (
	"go/ast"
	"go/token"
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
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				if x.Name.Name == "main" {
					if isContainExit, poses := isFuncNodeContainExit(node); isContainExit {
						for _, pos := range poses {
							pass.Reportf(pos, "Must not contain os.Exit expression")
						}
					}
				}
			}

			return true
		})
	}

	return nil, nil
}

func isFuncNodeContainExit(node ast.Node) (bool, []token.Pos) {
	exitExpressionPos := make([]token.Pos, 0)

	ast.Inspect(node, func(node ast.Node) bool {
		if x, ok := node.(*ast.CallExpr); ok {
			if isExitExpression(x) {
				exitExpressionPos = append(exitExpressionPos, x.Pos())
			}
		}

		return true
	})

	return len(exitExpressionPos) > 0, exitExpressionPos
}

func isExitExpression(c *ast.CallExpr) bool {
	if s, ok := c.Fun.(*ast.SelectorExpr); ok {
		if x, ok := s.X.(*ast.Ident); ok {
			return x.Name == "os" && s.Sel.Name == "Exit"
		}
	}

	return false
}
