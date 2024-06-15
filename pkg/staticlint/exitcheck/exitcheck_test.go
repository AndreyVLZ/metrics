package exitcheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestMyAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, NewAnalyzer(), "./...")
}
