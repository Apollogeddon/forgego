// Package yamlx edits YAML documents as node trees, so the comments and key order of a
// file someone else wrote survive forgego's changes.
package yamlx

import (
	"bytes"
	"fmt"

	"go.yaml.in/yaml/v3"
)

// Parse returns the root mapping of a YAML document; an empty document is an empty mapping.
func Parse(content string) (*yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
		return nil, err
	}
	if len(doc.Content) == 0 {
		return Map(), nil
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected a mapping at the top level")
	}
	return root, nil
}

// Encode renders a root mapping with two-space indentation.
func Encode(root *yaml.Node) (string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(root); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Map returns an empty block mapping.
func Map() *yaml.Node { return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"} }

// Str returns a string scalar.
func Str(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
}

// Seq returns a block sequence of string scalars.
func Seq(values ...string) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, v := range values {
		seq.Content = append(seq.Content, Str(v))
	}
	return seq
}

// Get returns the value for key in a mapping, or nil.
func Get(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// Set replaces the value for key, or appends the key when it's missing.
func Set(m *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = value
			return
		}
	}
	m.Content = append(m.Content, Str(key), value)
}

// SetIfAbsent sets key unless it already exists; force replaces it anyway. It reports
// whether the mapping changed.
func SetIfAbsent(m *yaml.Node, key string, value *yaml.Node, force bool) bool {
	if Get(m, key) != nil && !force {
		return false
	}
	Set(m, key, value)
	return true
}

// Child returns the mapping under key, creating it when missing.
func Child(m *yaml.Node, key string) *yaml.Node {
	if child := Get(m, key); child != nil && child.Kind == yaml.MappingNode {
		return child
	}
	child := Map()
	Set(m, key, child)
	return child
}

// Merge deep-merges overlay into base: mappings merge key by key, sequences gain the
// overlay's items they don't already hold, and any other overlay value wins.
func Merge(base, overlay *yaml.Node) *yaml.Node {
	switch {
	case base == nil:
		return overlay
	case overlay == nil:
		return base
	case base.Kind == yaml.MappingNode && overlay.Kind == yaml.MappingNode:
		for i := 0; i+1 < len(overlay.Content); i += 2 {
			key := overlay.Content[i].Value
			Set(base, key, Merge(Get(base, key), overlay.Content[i+1]))
		}
		return base
	case base.Kind == yaml.SequenceNode && overlay.Kind == yaml.SequenceNode:
		for _, item := range overlay.Content {
			if !containsScalar(base, item) {
				base.Content = append(base.Content, item)
			}
		}
		return base
	default:
		return overlay
	}
}

func containsScalar(seq, item *yaml.Node) bool {
	if item.Kind != yaml.ScalarNode {
		return false
	}
	for _, existing := range seq.Content {
		if existing.Kind == yaml.ScalarNode && existing.Value == item.Value {
			return true
		}
	}
	return false
}
