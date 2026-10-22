package lib

import (
	"github.com/ant-src/ant/value"
)

func init() {
	register("Regex", func() value.Value {
		return namespace("Regex",
			constant("Pattern", value.RegexPatternTag),
		)
	})
}
