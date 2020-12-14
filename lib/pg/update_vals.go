package pg

import (
	"fmt"
	"strings"
)

type UpdateSQL struct {
	ValueMap map[string]interface{}
}

func (c *UpdateSQL) String(base *int) string {
	mRows := []string{}
	for col := range c.ValueMap {
		mRows = append(mRows, fmt.Sprintf("%s = $%d", col, base))
	}
	return strings.Join(mRows, ", ")
}
func (c *UpdateSQL) Values(base *int) []interface{} {
	res := []interface{}{}
	for _, val := range c.ValueMap {
		res = append(res, val)
	}
	return res
}

type UpdateConfig func(c *UpdateSQL)

func Fields(v map[string]interface{}) UpdateConfig {
	return func(c *UpdateSQL) {}
}

func Returns(v map[string]interface{}) UpdateConfig {
	return func(c *UpdateSQL) {}
}
