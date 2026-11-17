package graphBetaServicePrincipalSynchronizationJobSchema

// This file has no dependencies outside the standard library so the parser can be
// tested without compiling the Graph SDK.

import (
	"fmt"
	"strings"
)

// Source node types used by Graph for attributeMappingSource.type.
const (
	sourceTypeAttribute = "Attribute"
	sourceTypeConstant  = "Constant"
	sourceTypeFunction  = "Function"
)

// exprNode is a parsed provisioning expression, for example
// `Switch([IsSoftDeleted], , "False", "1", "True", "0")`.
type exprNode struct {
	Type string
	// Name is the attribute name, the unescaped constant value or the function name.
	Name string
	// Args holds the function arguments in order. A nil entry is an empty argument,
	// such as the omitted default value in `Switch([x], , "a", "b")`.
	Args []*exprNode
}

// functionParameter mirrors attributeMappingParameterSchema returned by
// GET .../synchronization/jobs/{jobId}/schema/functions.
type functionParameter struct {
	Name                     string `json:"name"`
	Required                 bool   `json:"required"`
	AllowMultipleOccurrences bool   `json:"allowMultipleOccurrences"`
}

// functionSignature mirrors attributeMappingFunctionSchema.
type functionSignature struct {
	ID         string              `json:"id"`
	Parameters []functionParameter `json:"parameters"`
}

// parseExpression parses an Entra provisioning expression. The grammar is:
//
//	expr     = attribute | constant | function
//	attribute = "[" name "]"
//	constant = '"' chars '"'      (a double quote inside is written \")
//	function = name "(" [arg {"," arg}] ")"
//	arg      = expr | (empty)
func parseExpression(input string) (*exprNode, error) {
	p := &exprParser{src: input}
	p.skipSpace()
	if p.eof() {
		return nil, fmt.Errorf("expression is empty")
	}
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	p.skipSpace()
	if !p.eof() {
		return nil, p.errorf("unexpected %q after the end of the expression", p.src[p.pos:])
	}
	return node, nil
}

type exprParser struct {
	src string
	pos int
}

func (p *exprParser) eof() bool { return p.pos >= len(p.src) }

func (p *exprParser) peek() byte { return p.src[p.pos] }

func (p *exprParser) skipSpace() {
	for !p.eof() {
		switch p.peek() {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func (p *exprParser) errorf(format string, args ...any) error {
	return fmt.Errorf("at character %d: %s", p.pos+1, fmt.Sprintf(format, args...))
}

func (p *exprParser) parseExpr() (*exprNode, error) {
	if p.eof() {
		return nil, p.errorf("expected an attribute, constant or function")
	}
	switch c := p.peek(); {
	case c == '[':
		return p.parseAttribute()
	case c == '"':
		return p.parseConstant()
	case isIdentStart(c):
		return p.parseFunction()
	default:
		return nil, p.errorf("unexpected %q; expected [attribute], \"constant\" or Function(...)", string(c))
	}
}

func (p *exprParser) parseAttribute() (*exprNode, error) {
	start := p.pos
	p.pos++ // [
	end := strings.IndexByte(p.src[p.pos:], ']')
	if end < 0 {
		p.pos = start
		return nil, p.errorf("attribute reference is missing its closing ]")
	}
	name := p.src[p.pos : p.pos+end]
	p.pos += end + 1
	if strings.TrimSpace(name) == "" {
		p.pos = start
		return nil, p.errorf("attribute reference [] has no name")
	}
	return &exprNode{Type: sourceTypeAttribute, Name: name}, nil
}

func (p *exprParser) parseConstant() (*exprNode, error) {
	start := p.pos
	p.pos++ // opening quote
	var b strings.Builder
	for !p.eof() {
		c := p.peek()
		switch {
		case c == '\\' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '"':
			b.WriteByte('"')
			p.pos += 2
		case c == '"':
			p.pos++
			return &exprNode{Type: sourceTypeConstant, Name: b.String()}, nil
		default:
			b.WriteByte(c)
			p.pos++
		}
	}
	p.pos = start
	return nil, p.errorf("string constant is missing its closing double quote")
}

func (p *exprParser) parseFunction() (*exprNode, error) {
	start := p.pos
	for !p.eof() && isIdentPart(p.peek()) {
		p.pos++
	}
	name := p.src[start:p.pos]
	p.skipSpace()
	if p.eof() || p.peek() != '(' {
		return nil, p.errorf("expected ( after function name %q", name)
	}
	p.pos++ // (
	node := &exprNode{Type: sourceTypeFunction, Name: name}
	p.skipSpace()
	if !p.eof() && p.peek() == ')' {
		p.pos++
		return node, nil
	}
	for {
		p.skipSpace()
		if p.eof() {
			return nil, p.errorf("function %s is missing its closing )", name)
		}
		var arg *exprNode
		if c := p.peek(); c != ',' && c != ')' {
			var err error
			if arg, err = p.parseExpr(); err != nil {
				return nil, err
			}
			p.skipSpace()
		}
		node.Args = append(node.Args, arg)
		if p.eof() {
			return nil, p.errorf("function %s is missing its closing )", name)
		}
		switch p.peek() {
		case ',':
			p.pos++
		case ')':
			p.pos++
			return node, nil
		default:
			return nil, p.errorf("expected , or ) in the arguments of %s, found %q", name, string(p.peek()))
		}
	}
}

func isIdentStart(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c == '_'
}

func isIdentPart(c byte) bool {
	return isIdentStart(c) || c >= '0' && c <= '9'
}

// canonical returns the expression in the form Graph stores and returns it:
// arguments separated by ", ", an empty argument written as nothing (", ,").
func (n *exprNode) canonical() string {
	switch n.Type {
	case sourceTypeAttribute:
		return "[" + n.Name + "]"
	case sourceTypeConstant:
		return `"` + strings.ReplaceAll(n.Name, `"`, `\"`) + `"`
	default:
		args := make([]string, len(n.Args))
		for i, a := range n.Args {
			if a != nil {
				args[i] = a.canonical()
			}
		}
		return n.Name + "(" + strings.Join(args, ", ") + ")"
	}
}

// canonicalExpression parses an expression and returns its canonical form.
func canonicalExpression(expression string) (string, error) {
	node, err := parseExpression(expression)
	if err != nil {
		return "", err
	}
	return node.canonical(), nil
}

// toSource builds the attributeMappingSource JSON object Graph expects, in the same
// shape Graph returns from GET .../schema. Function argument keys come from the
// function's parameter list: arguments are matched to parameters by position, and a
// parameter that allows multiple occurrences takes every remaining argument. Empty
// arguments are left out, as Graph does.
func (n *exprNode) toSource(signatures map[string]functionSignature) (map[string]any, error) {
	source := map[string]any{
		"expression": n.canonical(),
		"name":       n.Name,
		"parameters": []any{},
		"type":       n.Type,
	}
	if n.Type != sourceTypeFunction {
		return source, nil
	}
	sig, ok := signatures[n.Name]
	if !ok {
		return nil, fmt.Errorf("unknown function %q%s", n.Name, suggestFunction(n.Name, signatures))
	}
	keys, err := argumentKeys(sig, n.Args)
	if err != nil {
		return nil, err
	}
	params := make([]any, 0, len(n.Args))
	for i, arg := range n.Args {
		if arg == nil {
			continue
		}
		value, err := arg.toSource(signatures)
		if err != nil {
			return nil, err
		}
		params = append(params, map[string]any{"key": keys[i], "value": value})
	}
	source["parameters"] = params
	return source, nil
}

func argumentKeys(sig functionSignature, args []*exprNode) ([]string, error) {
	keys := make([]string, len(args))
	multi := -1
	for i, p := range sig.Parameters {
		if p.AllowMultipleOccurrences {
			multi = i
			break
		}
	}
	for i := range args {
		switch {
		case multi >= 0 && i >= multi:
			keys[i] = sig.Parameters[multi].Name
		case i < len(sig.Parameters):
			keys[i] = sig.Parameters[i].Name
		default:
			return nil, fmt.Errorf("function %s takes at most %d arguments, got %d", sig.ID, len(sig.Parameters), len(args))
		}
	}
	for i, p := range sig.Parameters {
		if !p.Required {
			continue
		}
		present := false
		for j, arg := range args {
			if arg != nil && (j == i || p.AllowMultipleOccurrences && j >= i) {
				present = true
				break
			}
		}
		if !present {
			return nil, fmt.Errorf("function %s is missing its required argument %q (argument %d)", sig.ID, p.Name, i+1)
		}
	}
	return keys, nil
}

func suggestFunction(name string, signatures map[string]functionSignature) string {
	for id := range signatures {
		if strings.EqualFold(id, name) {
			return fmt.Sprintf("; function names are case-sensitive, did you mean %q?", id)
		}
	}
	return ""
}
