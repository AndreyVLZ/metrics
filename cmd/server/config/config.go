package config

type Config struct {
	Addr string
}

func (c *Config) SetAddr(addr string) {
	c.Addr = addr
}
