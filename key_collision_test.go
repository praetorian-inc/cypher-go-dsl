package cypher

import (
	"runtime"
	"testing"
)

func TestVisitableKeysAreUnique(t *testing.T) {
	seen := make(map[string]string)
	record := func(i int, kind string, v Visitable) {
		t.Helper()
		key := v.getKey()
		if prev, dup := seen[key]; dup {
			t.Fatalf("iteration %d: %s reused key %q already held by %s", i, kind, key, prev)
		}
		seen[key] = kind
	}
	for i := 0; i < 2000; i++ {
		record(i, "Operator", OperatorCreate("=", BINARY))
		record(i, "Asterisk", AsteriskCreate())
		record(i, "SymbolicName", ASymbolic("x"))
		record(i, "Literal", LiteralOf("v"))
		record(i, "Parameter", AParam("p"))
		if i%4 == 0 {
			runtime.GC()
		}
	}
}

func guardShapedStatement() (Statement, error) {
	node := AliasedExpressionCreate(ASymbolic("node"), "n0")
	n0 := node.AsName()
	where := APropertyByExpression(n0, "username").IsEqualTo(AParam("p2")).
		And(APropertyByExpression(n0, "status").StartWiths(AParam("p3")).Get()).Get()
	return ACall("db", "index", "fulltext", "queryNodes").WithArgs(LiteralOf("idx"), AParam("p1")).
		Yield(node).
		With(AnAsterisk()).
		Where(where).
		Returning(n0).
		Build()
}

func TestRenderIsDeterministicUnderGC(t *testing.T) {
	want := "CALL db.index.fulltext.queryNodes('idx', $p1) YIELD node AS n0 WITH * WHERE (n0.username = $p2 AND n0.status STARTS WITH $p3) RETURN n0"
	for i := 0; i < 1000; i++ {
		st, err := guardShapedStatement()
		if err != nil {
			t.Fatalf("iteration %d: build: %v", i, err)
		}
		got, err := plainRenderer().Render(st)
		if err != nil {
			t.Fatalf("iteration %d: render: %v", i, err)
		}
		if got != want {
			t.Fatalf("iteration %d:\nwant %q\ngot  %q", i, want, got)
		}
		if i%4 == 0 {
			runtime.GC()
		}
	}
}
