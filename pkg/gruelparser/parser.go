package gruelparser

import (
	"fmt"
	"strconv"
	"strings"
)

// An AST node
type GruelAstNode struct {
	// The token for this node or the operator for non-atomic nodes
	Value string
	// Types (the type for non-atomic nodes is TypeParenthesis)
	Type TokenType
	// The parameters for non-atomic nodes (nil for atomic ones)
	Parameters []GruelAstNode
}

// Parses a lisp-like expression into an AST tree
//
// The values still need further validation though.
func Parse(expr string) (GruelAstNode, error) {
	r := NewTokenReader(expr)
	branch := make([]GruelAstNode, 0, 16)
	var current GruelAstNode
	for {
		token, tokenType, err := r.NextToken()
		if err != nil {
			return current, err
		}
		current.Type = tokenType
		if tokenType == TypeParenthesis {
			if token == "(" {
				operator, operatorType, err := r.NextToken()
				if err != nil {
					return current, err
				}
				if operatorType != TypeSymbol {
					return current, fmt.Errorf("expecting symbolic operator")
				}
				current.Value = operator
				branch = append(branch, current)
			} else if token == ")" {
				var i int
				for i = len(branch) - 1; i >= 0 && branch[i].Type != TypeParenthesis ||
					branch[i].Parameters != nil; i-- {
				}
				if i < 0 {
					return current, fmt.Errorf("unexpected parenthesis")
				}
				branch[i].Parameters = make([]GruelAstNode, len(branch)-i-1)
				copy(branch[i].Parameters, branch[i+1:])
				branch = branch[0 : i+1]
				if i == 0 {
					return branch[0], nil
				}
			} else {
				return current, fmt.Errorf("open parenthesis")
			}
		} else {
			current.Value = token
			current.Parameters = nil
			branch = append(branch, current)
			if len(branch) == 1 {
				return branch[0], nil
			}
		}
		current = GruelAstNode{}
	}
}

// Implements fmt.Stringer
func (node *GruelAstNode) String() string {
	switch node.Type {
	case TypeParenthesis:
		sb := strings.Builder{}
		sb.WriteByte('(')
		sb.WriteString(node.Value)
		for _, v := range node.Parameters {
			sb.WriteString(" ")
			sb.WriteString(v.String())
		}
		sb.WriteByte(')')
		return sb.String()
	case TypeString:
		return strconv.Quote(node.Value)
	case TypeBool:
		return "#" + node.Value
	default:
		return node.Value
	}
}
