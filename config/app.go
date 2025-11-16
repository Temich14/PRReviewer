package config

import "os"

type AppConfig struct {
	ServerCfg *ServerConfig
	DBCfg     *DBConfig
}

func MustLoadConfig() *AppConfig {
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	dbConnString := os.Getenv("DB_CONNECTION_STRING")
	if dbConnString == "" {
		dbConnString = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	serverCfg := ServerConfig{
		Port: serverPort,
	}

	db := DBConfig{
		ConnectionString: dbConnString,
	}

	return &AppConfig{
		ServerCfg: &serverCfg,
		DBCfg:     &db,
	}
}
