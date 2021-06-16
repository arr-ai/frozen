package frozen

import (
	"encoding/json"
)

// MarshalJSON implements json.Marshaler.
func (m StringMap) MarshalJSON() ([]byte, error) {
	proxy := map[string]interface{}{}
	for i := m.Range(); i.Next(); {
		proxy[i.Key()] = i.Value()
	}
	data, err := json.Marshal(proxy)
	return data, errorsWrap(err, 0)
}

// Ensure that StringMap implements json.Marshaler.
var _ json.Marshaler = StringMap{}
