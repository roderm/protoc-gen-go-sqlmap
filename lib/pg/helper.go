package pg

import (
	"fmt"
	"strings"
)

func joinN(n int, param_base *int, sep string) string {
	arr := make([]string, n)
	for i, _ := range arr {
		*param_base++
		arr[i] = fmt.Sprintf("$%d", *param_base)
	}
	return strings.Join(arr, sep)
}
