package resumeparsergraph

import (
	"encoding/json"
	"strings"
)

func boolFromMeta(meta map[string]any, key string) bool {
	if meta == nil {
		return false
	}
	val, ok := meta[key]
	if !ok {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1", "yes", "y", "是":
			return true
		default:
			return false
		}
	case float64:
		return v != 0
	case float32:
		return v != 0
	case int:
		return v != 0
	case int64:
		return v != 0
	case json.Number:
		return jsonNumberToBool(v)
	default:
		return false
	}
}

func jsonNumberToBool(n json.Number) bool {
	if strings.EqualFold(string(n), "true") || strings.EqualFold(string(n), "false") {
		return strings.EqualFold(string(n), "true")
	}

	f64, err := n.Float64()
	if err == nil {
		return f64 != 0
	}

	i64, err := n.Int64()
	if err == nil {
		return i64 != 0
	}

	switch strings.ToLower(strings.TrimSpace(n.String())) {
	case "true", "1", "yes", "y", "是":
		return true
	default:
		return false
	}
}
