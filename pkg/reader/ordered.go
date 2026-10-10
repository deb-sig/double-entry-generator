package reader

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// OrderedMap is a string map that remembers declaration order. Column order
// in a template is meaningful, and yaml maps lose it.
type OrderedMap struct {
	Keys   []string
	Values map[string]string
}

func (m *OrderedMap) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected a mapping, got %s", value.Tag)
	}
	m.Keys = m.Keys[:0]
	m.Values = map[string]string{}
	for i := 0; i+1 < len(value.Content); i += 2 {
		k := value.Content[i].Value
		if _, dup := m.Values[k]; dup {
			return fmt.Errorf("duplicate column %q", k)
		}
		m.Keys = append(m.Keys, k)
		m.Values[k] = value.Content[i+1].Value
	}
	return nil
}

func (m OrderedMap) MarshalYAML() (interface{}, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for _, k := range m.Keys {
		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: k},
			&yaml.Node{Kind: yaml.ScalarNode, Value: m.Values[k]},
		)
	}
	return node, nil
}

func (m OrderedMap) Len() int { return len(m.Keys) }
