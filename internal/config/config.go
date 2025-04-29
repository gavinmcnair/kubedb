package config

type Config struct {
	DBPath string
	Token  string
}

func LoadConfig() Config {
	return Config{
		DBPath: "data/kubedb",
		Token:  "my-secret-token",
	}
}

