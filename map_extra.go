package frozen

import (
	"encoding/json"

	"github.com/arr-ai/frozen/pkg/errors"
)

func marshalJSON(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	return data, errors.Wrap(err, 0)
}

// MarshalJSON implements json.Marshaler.
func (m Map[K, V]) MarshalJSON() ([]byte, error) {
	switch m := any(m).(type) {
	case Map[string, V]:
		proxy := map[string]any{}
		for i := m.Range(); i.Next(); {
			proxy[i.Key()] = i.Value()
		}
		return marshalJSON(proxy)
	}
	return m.marshalJSONArray()
}

// Ensure that Map implements json.Marshaler.
var _ json.Marshaler = Map[any, any]{}

func (m Map[K, V]) marshalJSONArray() ([]byte, error) {
	proxy := make([]any, 0, m.Count())
	for i := m.Range(); i.Next(); {
		proxy = append(proxy, []any{i.Key(), i.Value()})
	}
	data, err := json.Marshal(proxy)
	return data, errors.Wrap(err, 0)
}
