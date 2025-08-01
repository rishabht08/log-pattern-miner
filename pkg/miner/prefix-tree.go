package miner

import (
	"strings"
	"sync"

	"github.com/rishabht08/template-miner/pkg/miner/utils"
)

type PatternTree struct {
	Root *Node
}

type Node struct {
	Children sync.Map
}

func NewPatternTree() *PatternTree {
	return &PatternTree{
		Root: &Node{
			Children: sync.Map{}, // Initialize the root node with an empty map
		},
	}
}

func (pt *PatternTree) AddOrMatch(tokens []string) ([]string, []string, string) {
	current := pt.Root
	templateIndexes := make([]int, 0)
	unMatchCount := 0

	for _i, token := range tokens {
		childAny, ok := current.Children.Load(token)
		wildNode, wildCardPresent := current.Children.Load("<*>")

		if wildCardPresent {
			templateIndexes = append(templateIndexes, _i)
			current = wildNode.(*Node)
			continue
		}

		if !ok {
			unMatchCount++
			nextKeyFound := false

			if _i < len(tokens)-1 {
				nextKey := tokens[_i+1]
				current.Children.Range(func(key, value interface{}) bool {
					internalNode := value.(*Node)
					_, exists := internalNode.Children.Load(nextKey)
					if exists {
						current.Children.LoadOrStore("<*>", internalNode)
						// avoid deleting here to keep structure safe
						nextKeyFound = true
						templateIndexes = append(templateIndexes, _i)
						current = internalNode
						return false
					}
					return true
				})
			}

			if !nextKeyFound {
				lastToken := token
				if _i == len(tokens)-1 && unMatchCount > int(0.4*float64(len(tokens))) {
					lastToken = "<*>"
					templateIndexes = append(templateIndexes, _i)
				}
				newNode := &Node{Children: sync.Map{}}
				actual, _ := current.Children.LoadOrStore(lastToken, newNode)
				childAny = actual
				current = childAny.(*Node)
			} else {
				continue // already updated current to internalNode
			}
		} else {
			current = childAny.(*Node)
		}
	}

	if unMatchCount > int(0.4*float64(len(tokens))) {
		return tokens, tokens, utils.Sha1Hex(strings.Join(tokens, "|"))
	}

	template := make([]string, len(tokens))
	copy(template, tokens)
	params := make([]string, 0)
	for _, idx := range templateIndexes {
		params = append(params, tokens[idx])
		template[idx] = "<*>"
	}

	templateID := utils.Sha1Hex(strings.Join(template, "|"))
	return template, params, templateID
}

func (pt *PatternTree) GetPattern(tokens []string) ([]string, []string, string) {
	current := pt.Root
	templateIndexes := make([]int, 0)
	unMatchCount := 0

	for _i, token := range tokens {
		childAny, ok := current.Children.Load(token)
		wildNode, wildCardPresent := current.Children.Load("<*>")

		if wildCardPresent {
			templateIndexes = append(templateIndexes, _i)
			current = wildNode.(*Node)
			continue
		}

		if !ok {
			unMatchCount++
			nextKeyFound := false

			if _i < len(tokens)-1 {
				nextKey := tokens[_i+1]
				current.Children.Range(func(key, value interface{}) bool {
					internalNode := value.(*Node)
					_, exists := internalNode.Children.Load(nextKey)
					if exists {
						// current.Children.LoadOrStore("<*>", internalNode)
						// avoid deleting here to keep structure safe
						nextKeyFound = true
						templateIndexes = append(templateIndexes, _i)
						current = internalNode
						return false
					}
					return true
				})
			}

			if !nextKeyFound {
				if _i == len(tokens)-1 && unMatchCount > int(0.4*float64(len(tokens))) {
					templateIndexes = append(templateIndexes, _i)
					continue
				}
				return tokens, tokens, utils.Sha1Hex(strings.Join(tokens, "|"))

			} else {
				continue // already updated current to internalNode
			}
		} else {
			current = childAny.(*Node)
		}
	}

	if unMatchCount > int(0.4*float64(len(tokens))) {
		return tokens, tokens, utils.Sha1Hex(strings.Join(tokens, "|"))
	}

	template := make([]string, len(tokens))
	copy(template, tokens)
	params := make([]string, 0)
	for _, idx := range templateIndexes {
		params = append(params, tokens[idx])
		template[idx] = "<*>"
	}

	templateID := utils.Sha1Hex(strings.Join(template, "|"))
	return template, params, templateID
}

func convertToSerializableNode(n *Node) *SerializableNode {
	serial := &SerializableNode{
		Children: make(map[string]*SerializableNode),
	}
	n.Children.Range(func(key, value interface{}) bool {
		k := key.(string)
		v := value.(*Node)
		serial.Children[k] = convertToSerializableNode(v)
		return true
	})
	return serial
}

func (pt *PatternTree) ToSerializable() *SerializablePatternTree {
	return &SerializablePatternTree{
		Root: convertToSerializableNode(pt.Root),
	}
}

func convertToRuntimeNode(sn *SerializableNode) *Node {
	runtime := &Node{
		Children: sync.Map{},
	}
	for key, child := range sn.Children {
		runtime.Children.Store(key, convertToRuntimeNode(child))
	}
	return runtime
}

func FromSerializableTree(spt *SerializablePatternTree) *PatternTree {
	return &PatternTree{
		Root: convertToRuntimeNode(spt.Root),
	}
}
