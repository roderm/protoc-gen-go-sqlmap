package pg

import (
	"fmt"
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
