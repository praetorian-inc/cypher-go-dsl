package cypher

import (
	"github.com/rs/xid"
	"golang.org/x/xerrors"
	"strings"
	"unicode"
)

type CypherRenderer struct {
	visitableToAliased      map[string]AliasedExpression
	separatorOnLevel        map[int]string
	resolvedSymbolicNames   map[SymbolicName]string
	visitedNamed            map[string]int
	currentVisitedElements  []string
	currentAliasedElements  []AliasedExpression
	currentLevel            int
	skipNodeContent         bool
	builder                 strings.Builder
	guid                    xid.ID
	escapeLabels            bool
	escapeRelationshipTypes bool
}

type RendererOption func(*CypherRenderer)

// WithLabelEscaping controls whether node labels are always backticked (the
// default) or only when they are not plain identifiers.
func WithLabelEscaping(always bool) RendererOption {
	return func(r *CypherRenderer) { r.escapeLabels = always }
}

// WithRelationshipTypeEscaping controls whether relationship types are always
// backticked (the default) or only when they are not plain identifiers.
func WithRelationshipTypeEscaping(always bool) RendererOption {
	return func(r *CypherRenderer) { r.escapeRelationshipTypes = always }
}

func (renderer *CypherRenderer) escapeLabel(name string) string {
	if renderer.escapeLabels {
		return escapeName(name)
	}
	return EscapeIfNecessary(name)
}

func (renderer *CypherRenderer) escapeRelationshipType(name string) string {
	if renderer.escapeRelationshipTypes {
		return escapeName(name)
	}
	return EscapeIfNecessary(name)
}

func (renderer *CypherRenderer) append(content string) *CypherRenderer {
	renderer.builder.WriteString(content)
	return renderer
}

func NewRenderer(opts ...RendererOption) *CypherRenderer {
	r := &CypherRenderer{
		visitableToAliased:      make(map[string]AliasedExpression),
		separatorOnLevel:        make(map[int]string),
		resolvedSymbolicNames:   make(map[SymbolicName]string),
		visitedNamed:            make(map[string]int),
		currentVisitedElements:  make([]string, 0),
		currentAliasedElements:  make([]AliasedExpression, 0),
		skipNodeContent:         false,
		builder:                 strings.Builder{},
		guid:                    xid.New(),
		escapeLabels:            true,
		escapeRelationshipTypes: true,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (renderer CypherRenderer) Render(statement Statement) (string, error) {
	if statement == nil {
		return "", xerrors.New("can not render nil statement")
	}
	if statement.GetError() != nil {
		return "", statement.GetError()
	}
	statement.accept(&renderer)
	return strings.TrimSpace(renderer.builder.String()), nil
}

func (renderer *CypherRenderer) enableSeparator(level int, on bool) {
	if on {
		renderer.separatorOnLevel[level] = ""
	} else {
		delete(renderer.separatorOnLevel, level)
	}
}

func (renderer *CypherRenderer) separatorOnCurrentLevel() (string, bool) {
	value, isExist := renderer.separatorOnLevel[renderer.currentLevel]
	return value, isExist
}

func (renderer *CypherRenderer) resetSeparatorOnCurrentLevel() {
	renderer.separatorOnLevel[renderer.currentLevel] = ""
}

func (renderer *CypherRenderer) resolve(name SymbolicName) string {
	if _, isExist := renderer.resolvedSymbolicNames[name]; !isExist {
		value := name.value
		if len(value) > 0 {
			return EscapeIfNecessary(value)
		}
		return renderer.guid.String()
	}
	return renderer.resolvedSymbolicNames[name]
}

func (renderer *CypherRenderer) enter(visitable Visitable) {
	used := renderer.getAliasedIfSeen(visitable)
	if renderer.PreEnter(&used) {
		renderer.currentVisitedElements = push(renderer.currentVisitedElements, used.getKey())
		used.enter(renderer)
	}
}

func (renderer *CypherRenderer) leave(visitable Visitable) {
	used := renderer.getAliasedIfSeen(visitable)
	current := peek(renderer.currentVisitedElements)
	if current == (used).getKey() {
		used.leave(renderer)
		renderer.postLeave(used)
		renderer.currentVisitedElements = pop(renderer.currentVisitedElements)
	}
	aliasedExpression, isAliased := (visitable).(AliasedExpression)
	if isAliased {
		renderer.visitableToAliased[aliasedExpression.delegate.getKey()] = aliasedExpression
	}
}

func (renderer *CypherRenderer) PreEnter(visitable *Visitable) bool {
	lastAliased := peekAliased(renderer.currentAliasedElements)
	visited := false
	if lastAliased != nil {
		_, visited = renderer.visitableToAliased[lastAliased.delegate.getKey()]
	}
	if renderer.skipNodeContent || visited {
		return false
	}
	aliasedExpression, isAliased := (*visitable).(AliasedExpression)
	if isAliased {
		renderer.currentAliasedElements = pushAliased(renderer.currentAliasedElements, aliasedExpression)
	}
	renderer.currentLevel = renderer.currentLevel + 1
	nextLevel := renderer.currentLevel + 1
	if _, isSub := (*visitable).(SubVisitable); isSub {
		renderer.enableSeparator(nextLevel, true)
	}
	if separator, isSeparator := renderer.separatorOnCurrentLevel(); isSeparator {
		renderer.append(separator)
		renderer.resetSeparatorOnCurrentLevel()
	}
	return !renderer.skipNodeContent
}

func (renderer *CypherRenderer) postLeave(visitable Visitable) {

	if _, isSeparator := renderer.separatorOnCurrentLevel(); isSeparator {
		renderer.separatorOnLevel[renderer.currentLevel] = ", "
	}
	if _, isSub := visitable.(SubVisitable); isSub {
		renderer.enableSeparator(renderer.currentLevel+1, false)
	}
	if peekAliased(renderer.currentAliasedElements) != nil &&
		peekAliased(renderer.currentAliasedElements).getKey() == visitable.getKey() {
		renderer.currentAliasedElements = popAliased(renderer.currentAliasedElements)
	}
	renderer.currentLevel = renderer.currentLevel - 1
}

func (renderer *CypherRenderer) getAliasedIfSeen(visitable Visitable) Visitable {
	if expression, isExist := renderer.visitableToAliased[visitable.getKey()]; isExist {
		return expression
	}
	return visitable
}

const RelTypSeparator = "|"
const RelTypeStart = ":"
const NodeLabelStart = ":"

func push(queue []string, element string) []string {
	queue = append(queue, element)
	return queue
}

func pop(queue []string) []string {
	if len(queue) > 0 {
		return queue[:len(queue)-1]
	}
	return queue
}

func peek(queue []string) string {
	if len(queue) > 0 {
		return queue[len(queue)-1]
	}
	return ""
}

func pushAliased(queue []AliasedExpression, element AliasedExpression) []AliasedExpression {
	queue = append(queue, element)
	return queue
}

func popAliased(queue []AliasedExpression) []AliasedExpression {
	if len(queue) > 0 {
		return queue[:len(queue)-1]
	}
	return queue
}

func peekAliased(queue []AliasedExpression) *AliasedExpression {
	if len(queue) > 0 {
		return &queue[len(queue)-1]
	}
	return nil
}

func escapeName(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func EscapeIfNecessary(name string) string {
	if len(strings.TrimSpace(name)) == 0 || isIdentifier(name) {
		return name
	}
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func isIdentifier(name string) bool {
	for i, c := range name {
		if !unicode.IsLetter(c) && c != '_' && (i == 0 || !unicode.IsDigit(c)) {
			return false
		}
	}
	return name != ""

}
