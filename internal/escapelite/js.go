package escapelite

import "encoding/json"

// JSLiteral formats v as a JavaScript literal.
func JSLiteral(v any) (string, error) {
	data, err := json.Marshal(v)
	return string(data), err
}
