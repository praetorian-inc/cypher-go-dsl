package cypher

import "errors"

// Reduction is the reduce() list function:
// reduce(accumulator = initialValue, variable IN list | expression).
type Reduction struct {
	ExpressionContainer
	accumulator    SymbolicName
	initialValue   Expression
	variable       SymbolicName
	listExpression Expression
	expression     Expression
	key            string
	err            error
	notNil         bool
}

func Reduce(accumulator SymbolicName, initialValue Expression, variable SymbolicName, listExpression Expression, expression Expression) Reduction {
	for _, e := range []Expression{initialValue, listExpression, expression} {
		if e == nil || !e.isNotNil() {
			return Reduction{err: errors.New("reduce: initial value, list and expression are required")}
		}
		if e.GetError() != nil {
			return Reduction{err: e.GetError()}
		}
	}
	for _, n := range []SymbolicName{accumulator, variable} {
		if !n.isNotNil() {
			return Reduction{err: errors.New("reduce: accumulator and variable names are required")}
		}
		if n.GetError() != nil {
			return Reduction{err: n.GetError()}
		}
	}
	r := Reduction{
		accumulator:    accumulator,
		initialValue:   initialValue,
		variable:       variable,
		listExpression: listExpression,
		expression:     expression,
		notNil:         true,
	}
	r.key = getAddress(&r)
	r.ExpressionContainer = ExpressionWrap(r)
	return r
}

func (r Reduction) GetError() error {
	return r.err
}

func (r Reduction) accept(visitor *CypherRenderer) {
	visitor.enter(r)
	r.accumulator.accept(visitor)
	ASSIGMENT.accept(visitor)
	r.initialValue.accept(visitor)
	visitor.append(", ")
	r.variable.accept(visitor)
	IN.accept(visitor)
	r.listExpression.accept(visitor)
	PIPE.accept(visitor)
	r.expression.accept(visitor)
	visitor.leave(r)
}

func (r Reduction) enter(renderer *CypherRenderer) {
	renderer.append("reduce(")
}

func (r Reduction) leave(renderer *CypherRenderer) {
	renderer.append(")")
}

func (r Reduction) getKey() string {
	return r.key
}

func (r Reduction) isNotNil() bool {
	return r.notNil
}

func (r Reduction) GetExpressionType() ExpressionType {
	return "Reduction"
}
