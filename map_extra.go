package frozen

import (
	"encoding/json"

	"github.com/arr-ai/frozen/errors"
)

// MarshalJSON implements json.Marshaler.
func (m Map) MarshalJSON() ([]byte, error) {
	proxy := map[string]interface{}{}
	for i := m.Range(); i.Next(); {
		if s, ok := i.Key().(string); ok {
			proxy[s] = i.Value()
		} else {
			return m.marshalJSONArray()
		}
	}
	data, err := json.Marshal(proxy)
	return data, errors.Wrap(err, 0)
}

// Ensure that Map implements json.Marshaler.
var _ json.Marshaler = Map{}

func (m Map) marshalJSONArray() ([]byte, error) {
	proxy := make([]interface{}, 0, m.Count())
	for i := m.Range(); i.Next(); {
		proxy = append(proxy, []interface{}{i.Key(), i.Value()})
	}
	data, err := json.Marshal(proxy)
	return data, errors.Wrap(err, 0)
}
