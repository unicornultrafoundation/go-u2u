package monitoring

// DefaultConfig is the default config for monitorings used in U2U.
type Config struct {
	Port             int `toml:",omitempty"`
}

// DefaultConfig is the default config for monitoring used in U2U.
var DefaultConfig = Config{
	Port: 19090,
}
