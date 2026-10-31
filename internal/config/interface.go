package config

type Config struct {
	Primary  Primary
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type Primary struct {
	Env      string // development | production
	LogLevel string // debug | info | warn | error
}

type ServerConfig struct {
	Port               string
	ReadTimeout        int // seconds
	WriteTimeout       int // seconds; 0 disables it, which token streaming will need
	IdleTimeout        int // seconds
	ShutdownTimeout    int // seconds
	CORSAllowedOrigins []string
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxConns        int
	MinConns        int
	ConnMaxLifetime int // seconds
	ConnMaxIdleTime int // seconds
}

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}
