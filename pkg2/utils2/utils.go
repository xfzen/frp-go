package utils2

import (
	"github.com/hokaccha/go-prettyjson"
	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

func PrettyJson(v interface{}) string {
	s, _ := prettyjson.Marshal(v)
	return string(s)
}

func CopyStruct(to interface{}, from interface{}) {
	err := copier.Copy(to, from)
	if err != nil {
		logx.Errorf("CopyStruct err: %v", err)
	}
}
