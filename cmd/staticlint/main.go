// Linter for
package main

import (
	"github.com/nessai1/linkshortener/internal/staticlint/exitcheck"
	"github.com/sivchari/containedctx"
	magic_numbers "github.com/tommy-muehle/go-mnd"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"honnef.co/go/tools/staticcheck"
	"strings"

	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
)

func main() {

	// Подключаем анализаторы пакета staticcheck
	var mychecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		// Подключаем все анализаторы класса SA и анализатор использования бесконечного for (S1006)
		if strings.HasPrefix(v.Analyzer.Name, "SA") || v.Analyzer.Name == "S1006" {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	mychecks = append(
		mychecks,

		// Анализатор, проверяющий отсутсвие контекстов в структурах
		containedctx.Analyzer,

		// Анализатор, проверяющий использование магических чисел
		magic_numbers.Analyzer,

		// Анализатор, проверяющий отсутсвие прямых вызовов os.Exit
		exitcheck.ExitCheckAnalyzer,

		// Анализатор, проверяющий верную сигнатуру строк с плейсхолдерами
		printf.Analyzer,

		// Анализатор, проверяющий корректность структурных тегов
		structtag.Analyzer,

		// Анализатор, проверяющий потенциально неверное затенение переменных
		shadow.Analyzer,

		// Анализатор, проверяющий бесполезное сравнение с nil
		nilfunc.Analyzer,

		// Анализатор, проверяющий замыкания переменных цикла
		loopclosure.Analyzer,

		// Анализатор, проверяющий недоступные куски кода
		unreachable.Analyzer,
	)

	multichecker.Main(
		mychecks...,
	)
}
