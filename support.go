package cypher

import (
	"strconv"
	"sync/atomic"

	"github.com/rs/xid"
)

var SYMBOLIC_GENERATOR = xid.New()

var keyCounter uint64

func ExpressionsToVisitables(expressions []Expression) []Visitable {
	visitables := make([]Visitable, len(expressions))
	for i := range expressions {
		visitables[i] = expressions[i]
	}
	return visitables
}

func PatternElementsToVisitables(patterns []PatternElement) []Visitable {
	visitables := make([]Visitable, len(patterns))
	for i := range patterns {
		visitables[i] = patterns[i]
	}
	return visitables
}

func getAddress(a interface{}) string {
	return strconv.FormatUint(atomic.AddUint64(&keyCounter, 1), 10)
}
