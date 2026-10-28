package cypher

import "testing"

func assertRendered(t *testing.T, renderer *CypherRenderer, want string, st Statement, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	got, err := renderer.Render(st)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got != want {
		t.Fatalf("want %q\ngot  %q", want, got)
	}
}

func TestReduce(t *testing.T) {
	n := ANode("Asset").NamedByString("n")
	acc, x := ASymbolic("acc"), ASymbolic("x")
	body := GenericCaseCreate1().
		When(ExpressionWrap(x).In(acc).Get()).Then(acc).
		ElseDefault(ExpressionWrap(acc).Add(x).Get())
	red := Reduce(acc, Coalesce(n.Property("tags"), ListOf()), x, AParam("p1"), body)
	st, err := Match(n).Set(n.Property("tags").To(red)).Build()
	assertRendered(t, NewRenderer(), "MATCH (n:`Asset`) SET n.tags = reduce(acc = coalesce(n.tags, []), x IN $p1 | CASE WHEN x IN acc THEN acc ELSE (acc + x) END)", st, err)
}

func TestRendererEscapingOptions(t *testing.T) {
	a := ANode("Asset").NamedByString("a")
	b := ANode("Weird Label").NamedByString("b")
	r := a.RelationshipTo(b, "HAS").NamedByString("r")
	st, err := Match(r).ReturningByNamed(a).Build()
	assertRendered(t, NewRenderer(), "MATCH (a:`Asset`)-[r:`HAS`]->(b:`Weird Label`) RETURN a", st, err)
	assertRendered(t, plainRenderer(), "MATCH (a:Asset)-[r:HAS]->(b:`Weird Label`) RETURN a", st, err)
}

func TestProcedureNamespace(t *testing.T) {
	st, err := ACall("db", "index", "fulltext", "queryNodes").WithArgs(LiteralOf("idx"), AParam("p1")).
		Yield(AliasedExpressionCreate(ASymbolic("node"), "n0")).ReturningByString("n0").Build()
	assertRendered(t, NewRenderer(), "CALL db.index.fulltext.queryNodes('idx', $p1) YIELD node AS n0 RETURN n0", st, err)
}
