package monitoring

// DefaultConfig is the default config for monitorings used in U2U.
type Config struct {
	HTTP            	   string `toml:",omitempty"`
	Port 				   int `toml:",omitempty"`
}

// DefaultConfig is the default config for monitorings used in U2U.
var DefaultConfig = Config{
	HTTP:             	    "127.0.0.1",
	Port: 					19090,
}
