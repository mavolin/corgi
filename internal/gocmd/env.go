package gocmd

func (c *Cmd) Env(env string) string {
	val, _ := c.command("env", env)
	return string(val)
}

func (c *Cmd) Env_GOMODCACHE() string {
	c.goModCacheOnce.Do(func() {
		c.goModCache = c.Env("GOMODCACHE")
	})
	return c.goModCache
}
