package bloc_client

import (
	"strconv"
)

type Opts []*Opt

type Opt struct {
	Key         string    `json:"key"`
	Description string    `json:"description"`
	ValueType   ValueType `json:"value_type"`
	IsArray     bool      `json:"is_array"`
}

func (opt *Opt) String() string {
	return opt.Key + opt.Description + string(opt.ValueType) + strconv.FormatBool(opt.IsArray)
}
