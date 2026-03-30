package dashboard

import (
	"strconv"
)

func decStr(f *float64) string {
	if f == nil {
		return "0"
	}
	return strconv.FormatFloat(*f, 'f', -1, 64)
}
