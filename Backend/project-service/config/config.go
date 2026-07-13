package config  // in the config file we have all our configurations like which port we will listen on which post db will run etc 

import "os"

type Config struct {
	GRPCPort     string
	PostgresHost string
	PostgresPort string
	PostgresUser string
	PostgresPass string
	PostgresDB   string
	PublicURL    string
}

func Load() Config {
	return Config{
		GRPCPort:     getEnv("PROJECT_GRPC_PORT", "50051"),
		PostgresHost: getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort: getEnv("POSTGRES_PORT", "5432"),
		PostgresUser: getEnv("POSTGRES_USER", "forge"),
		PostgresPass: getEnv("POSTGRES_PASSWORD", "forge123"),
		PostgresDB:   getEnv("POSTGRES_DB", "forge"),
		PublicURL:    getEnv("FORGE_PUBLIC_BASE_URL", "http://localhost"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}