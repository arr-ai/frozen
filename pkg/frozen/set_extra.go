package frozen

import "encoding/json"

func (s Set) MarshalJSON() ([]byte, error) {
	proxy := make([]interface{}, 0, s.Count())
	for i := s.Range(); i.Next(); {
		proxy = append(proxy, i.Value())
	}
	return json.Marshal(proxy)
}
