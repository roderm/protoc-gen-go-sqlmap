package pg

type UpdateSQL struct {
}
type UpdateConfig func(c *UpdateSQL)

func Fields(v map[string]interface{}) UpdateConfig {
	return func(c *UpdateSQL) {}
}

func Returns(v map[string]interface{}) UpdateConfig {
	return func(c *UpdateSQL) {}
}
