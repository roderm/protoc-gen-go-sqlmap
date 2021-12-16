package pg

import (
	"fmt"
	"regexp"
	"strings"
)

func joinN(n int, paramBase *int, sep string) string {
	arr := make([]string, n)
	for i := range arr {
		*paramBase++
		arr[i] = fmt.Sprintf("$%d", *paramBase)
	}
	return strings.Join(arr, sep)
}

func escapeColName(in string) string {
	m := regexp.MustCompile("/[A-z|0-9|_]*/").FindAllString(in, len(in)+1)
	return strings.Join(m, "")
}
