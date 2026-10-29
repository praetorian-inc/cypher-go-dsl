package cypher

import (
	"errors"
	"strings"
)

type IndexHint struct {
	alias      string
	label      string
	properties []string
	key        string
	notNil     bool
	err        error
}

func IndexHintCreate(alias string, label string, properties ...string) IndexHint {
	if alias == "" || label == "" || len(properties) == 0 {
		return IndexHint{err: errors.New("an index hint requires an alias, a label and at least one property")}
	}
	h := IndexHint{
		alias:      alias,
		label:      label,
		properties: properties,
		notNil:     true,
	}
	h.key = getAddress(&h)
	return h
}

func (h IndexHint) GetError() error {
	return h.err
}

func (h IndexHint) isNotNil() bool {
	return h.notNil
}

func (h IndexHint) getKey() string {
	return h.key
}

func (h IndexHint) accept(visitor *CypherRenderer) {
	visitor.enter(h)
	visitor.leave(h)
}

func (h IndexHint) enter(renderer *CypherRenderer) {
	renderer.append(" USING INDEX ").append(h.alias).append(":").append(h.label).
		append("(").append(strings.Join(h.properties, ", ")).append(")")
}

func (h IndexHint) leave(renderer *CypherRenderer) {
}
