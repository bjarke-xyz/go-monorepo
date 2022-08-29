package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"
)

type Config struct {
	apiPort           string
	workerWebUIPort   string
	ServerReadTimeout int

	JwtJwksUrl          string
	JwtKeycloakUrl      string
	JwtKeycloakUsername string
	JwtKeycloakPassword string

	dbHost     string
	dbPort     string
	dbName     string
	dbUser     string
	dbPassword string

	redisHost     string
	redisPort     string
	redisUser     string
	redisPassword string
	redisPrefix   string

	rabbitmqHost     string
	rabbitmqPort     string
	rabbitmqUser     string
	rabbitmqPassword string
	rabbitmqVHost    string

	LogFormat string

	Migrate          string
	MigrateOnStartup bool
}

func New() *Config {
	conf := &Config{}

	flag.StringVar(&conf.apiPort, "apiport", os.Getenv("API_PORT"), "Which port for the API server to listen on")
	flag.StringVar(&conf.workerWebUIPort, "workerwebuiport", os.Getenv("WORKER_WEBUI_PORT"), "Which port for the worker web UI server to listen on")

	serverReadTimeout, err := strconv.Atoi(os.Getenv("SERVER_READ_TIMEOUT"))
	if err != nil {
		zap.S().Errorf("Could not read SERVER_READ_TIMEOUT: %v", err)
		serverReadTimeout = 60
	}
	flag.IntVar(&conf.ServerReadTimeout, "serverreadtimeout", serverReadTimeout, "Server read timeout")

	flag.StringVar(&conf.JwtJwksUrl, "jwtjwksurl", os.Getenv("JWT_JWKS_URL"), "JWT JWKS URL")
	flag.StringVar(&conf.JwtKeycloakUrl, "jwtkeycloakurl", os.Getenv("JWT_KEYCLOAK_URL"), "JWT Keycloak URL")
	flag.StringVar(&conf.JwtKeycloakUsername, "jwtkeycloakusername", os.Getenv("JWT_KEYCLOAK_USERNAME"), "Keycloak username")
	flag.StringVar(&conf.JwtKeycloakPassword, "jwtkeycloakpassword", os.Getenv("JWT_KEYCLOAK_PASSWORD"), "Keycloak password")

	flag.StringVar(&conf.dbHost, "dbhost", os.Getenv("DB_HOST"), "Database host")
	flag.StringVar(&conf.dbPort, "dbport", os.Getenv("DB_PORT"), "Database port")
	flag.StringVar(&conf.dbName, "dbname", os.Getenv("DB_NAME"), "Database name")
	flag.StringVar(&conf.dbUser, "dbuser", os.Getenv("DB_USER"), "Database user")
	flag.StringVar(&conf.dbPassword, "dbpassword", os.Getenv("DB_PASSWORD"), "Database password")

	flag.StringVar(&conf.redisHost, "redishost", os.Getenv("REDIS_HOST"), "Redis host")
	flag.StringVar(&conf.redisPort, "redisport", os.Getenv("REDIS_PORT"), "Redis port")
	flag.StringVar(&conf.redisUser, "redisuser", os.Getenv("REDIS_USER"), "Redis username")
	flag.StringVar(&conf.redisPassword, "redispassword", os.Getenv("REDIS_PASSWORD"), "Redis password")
	flag.StringVar(&conf.redisPrefix, "redisprefix", os.Getenv("REDIS_PREFIX"), "Redis key prefix")

	flag.StringVar(&conf.rabbitmqHost, "rabbitmqhost", os.Getenv("RABBITMQ_HOST"), "RabbitMQ host")
	flag.StringVar(&conf.rabbitmqPort, "rabbitmqport", os.Getenv("RABBITMQ_PORT"), "RabbitMQ port")
	flag.StringVar(&conf.rabbitmqUser, "rabbitmquser", os.Getenv("RABBITMQ_USER"), "RabbitMQ user")
	flag.StringVar(&conf.rabbitmqPassword, "rabbitmqpassword", os.Getenv("RABBITMQ_PASSWORD"), "RabbitMQ password")
	flag.StringVar(&conf.rabbitmqVHost, "rabbitmqvhost", os.Getenv("RABBITMQ_VHOST"), "RabbitMQ vHost")

	flag.StringVar(&conf.LogFormat, "logformat", os.Getenv("LOG_FORMAT"), "Log format (json or console)")

	flag.StringVar(&conf.Migrate, "migrate", "up", "Specify if we should migrate DB 'up' or 'down'")
	migrateOnStartup, err := strconv.ParseBool(os.Getenv("DB_MIGRATE_ON_STARTUP"))
	if err != nil {
		zap.S().Errorf("Could not read DB_MIGRATE_ON_STARTUP: %v", err)
		migrateOnStartup = false
	}
	flag.BoolVar(&conf.MigrateOnStartup, "migrateonstartup", migrateOnStartup, "Specify if migrations should happen on API startup")

	flag.Parse()

	return conf
}

func (c *Config) GetDBConnStr() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.dbUser, c.dbPassword, c.dbHost, c.dbPort, c.dbName,
	)
}

func (c *Config) GetRedisHost() string {
	return c.redisHost
}

func (c *Config) GetRedisPort() string {
	return c.redisPort
}

func (c *Config) GetRedisConnStr() string {
	return fmt.Sprintf("%s:%s", c.redisHost, c.redisPort)
}

func (c *Config) GetRedisUser() string {
	return c.redisUser
}

func (c *Config) GetRedisPassword() string {
	return c.redisPassword
}

func (c *Config) GetRedisClientName() string {
	return "slv4"
}

func (c *Config) GetRedisPrefix() string {
	return c.redisPrefix
}

func (c *Config) GetRabbitMqUri() string {
	return fmt.Sprintf("amqp://%v:%v@%v:%v/%v", c.rabbitmqUser, c.rabbitmqPassword, c.rabbitmqHost, c.rabbitmqPort, c.rabbitmqVHost)
}

func (c *Config) GetServerUrl() string {
	return fmt.Sprintf("0.0.0.0:%s", c.apiPort)
}

func (c *Config) GetWorkerPort() string {
	return fmt.Sprintf(":%v", c.workerWebUIPort)
}
