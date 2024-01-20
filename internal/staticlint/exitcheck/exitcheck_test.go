package exitcheck

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func TestExitCheckAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), ExitCheckAnalyzer, "./...")
}
