package cypher

import "testing"

func plainRenderer() *CypherRenderer {
	return NewRenderer(WithLabelEscaping(false), WithRelationshipTypeEscaping(false))
}

func TestRawLabel(t *testing.T) {
	n := NodeCreateWithLabels(RawLabel("Asset|Seed")).NamedByString("n0")
	st, err := Match(n).ReturningByNamed(n).Build()
	assertRendered(t, NewRenderer(), "MATCH (n0:Asset|Seed) RETURN n0", st, err)
	assertRendered(t, plainRenderer(), "MATCH (n0:Asset|Seed) RETURN n0", st, err)

	mixed := NodeCreateWithLabels(NodeLabelCreate("Asset"), RawLabel("Seed|Risk")).NamedByString("n0")
	st, err = Match(mixed).ReturningByNamed(mixed).Build()
	assertRendered(t, NewRenderer(), "MATCH (n0:`Asset`:Seed|Risk) RETURN n0", st, err)
}

func TestRelationshipTypeSeparator(t *testing.T) {
	a := ANode("Asset").NamedByString("n0")
	b := ANode("Risk").NamedByString("n1")
	r := a.RelationshipTo(b, "HAS", "SCANNED").NamedByString("r0")
	st, err := Match(r).ReturningByNamed(a).Build()
	assertRendered(t, plainRenderer(), "MATCH (n0:Asset)-[r0:HAS|SCANNED]->(n1:Risk) RETURN n0", st, err)

	st, err = Match(r.WithTypeSeparator(":")).ReturningByNamed(a).Build()
	assertRendered(t, plainRenderer(), "MATCH (n0:Asset)-[r0:HAS:SCANNED]->(n1:Risk) RETURN n0", st, err)

	types := RelationshipTypesCreateWithSeparator(":", "HAS", "SCANNED")
	st, err = Match(RelationshipCreate3(a, RelationshipDetailsCreate2(LTR(), ASymbolic("r0"), types), b)).ReturningByNamed(a).Build()
	assertRendered(t, plainRenderer(), "MATCH (n0:Asset)-[r0:HAS:SCANNED]->(n1:Risk) RETURN n0", st, err)
}

func TestUsingIndexHint(t *testing.T) {
	n := ANode("Asset").NamedByString("n0")
	st, err := Match(n).UsingIndex("n0", "Asset", "username", "created").
		WhereConditionContainer(ExpressionWrap(n.Property("created")).IsNotNull()).
		ReturningByNamed(n).Build()
	assertRendered(t, plainRenderer(), "MATCH (n0:Asset) USING INDEX n0:Asset(username, created) WHERE n0.created IS NOT NULL RETURN n0", st, err)

	st, err = Match(n).UsingIndex("n0", "Asset", "username").UsingIndex("n0", "Asset", "key").ReturningByNamed(n).Build()
	assertRendered(t, plainRenderer(), "MATCH (n0:Asset) USING INDEX n0:Asset(username) USING INDEX n0:Asset(key) RETURN n0", st, err)

	m := ANode("Risk").NamedByString("n1")
	st, err = Match(n).OptionalMatch(m).UsingIndex("n1", "Risk", "key").ReturningByNamed(n).Build()
	assertRendered(t, plainRenderer(), "MATCH (n0:Asset) OPTIONAL MATCH (n1:Risk) USING INDEX n1:Risk(key) RETURN n0", st, err)

	if _, err := Match(n).UsingIndex("n0", "Asset").ReturningByNamed(n).Build(); err == nil {
		t.Fatal("expected an error for a hint without properties")
	}
}

func TestShortestK(t *testing.T) {
	a := ANode("Asset").NamedByString("n0")
	b := ANode("Risk").NamedByString("n1")
	path := APath("t0").ShortestK(2).DefinedBy(a.RelationshipTo(b, "HAS").NamedByString("r0").Unbounded())
	st, err := Match(path).ReturningByString("t0").Build()
	assertRendered(t, plainRenderer(), "MATCH t0 = SHORTEST 2 (n0:Asset)-[r0:HAS*]->(n1:Risk) RETURN t0", st, err)

	st, err = Match(a).OptionalMatch(path).ReturningByString("t0").Build()
	assertRendered(t, plainRenderer(), "MATCH (n0:Asset) OPTIONAL MATCH t0 = SHORTEST 2 (n0)-[r0:HAS*]->(n1:Risk) RETURN t0", st, err)

	plain := APath("t0").DefinedBy(a.RelationshipTo(b, "HAS").NamedByString("r0"))
	st, err = Match(plain).ReturningByString("t0").Build()
	assertRendered(t, plainRenderer(), "MATCH t0 = (n0:Asset)-[r0:HAS]->(n1:Risk) RETURN t0", st, err)

	if _, err := Match(APath("t0").ShortestK(0).DefinedBy(a.RelationshipTo(b, "HAS"))).ReturningByString("t0").Build(); err == nil {
		t.Fatal("expected an error for SHORTEST 0")
	}
}
