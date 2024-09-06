package utils2

import "github.com/hokaccha/go-prettyjson"

func PrettyJson(v interface{}) string {
	s, _ := prettyjson.Marshal(v)
	return string(s)
}
