package conf

import (
	"fmt"
)

type Server struct {
	Port int `protobuf:"bytes,1,opt,name=port,proto3" json:"port,omitempty"`
}

func (r *Server) GetPort() int {
	if r == nil {
		fmt.Println("Error nil config server")
		return 0
	}
	return r.Port
}

type Config struct {
	Server *Server `protobuf:"bytes,1,opt,name=server,proto3" json:"server,omitempty"`
}

func (r *Config) GetServer() *Server {
	if r == nil {
		fmt.Println("Error nil config")
		return nil
	}
	return r.Server
}

// func NewConfig() (*Config, error) {
// 	slog.Info("New config load")
// 	conf, err := loadConfig()
// 	if err != nil {
// 		slog.Error(err.Error())
// 		return nil, err
// 	}
// 	slog.Info("Load config finish")
// 	return conf, err
// }
//
// func loadConfig() (*Config, error) {
// 	err := godotenv.Load()
// 	if err != nil {
// 		fmt.Errorf("Error load .env config: %s", err.Error())
// 	}
// 	viper.SetConfigName("config")
// 	viper.SetConfigType("yaml")
// 	viper.AddConfigPath("../../config")
//
// 	// enable automatic environment variables substitution in the config file
// 	viper.AutomaticEnv()
//
// 	// read config
// 	if err := viper.ReadInConfig(); err != nil {
// 		return nil, fmt.Errorf("Error reading config file: %v", err)
// 	}
//
// 	fmt.Println("Read config after the viper load")
//
// 	// manual map environment variable
// 	appPort := os.Getenv("APP_PORT")
// 	if len(appPort) <= 0 {
// 		viper.Set("server.port", appPort)
// 	}
// 	portStr := viper.GetString("server.port")
// 	port, err := strconv.Atoi(portStr)
// 	if err != nil {
// 		return nil, fmt.Errorf("Invalid port number: %s", portStr)
// 	}
//
// 	config := &Config{
// 		server: Server{
// 			port: port,
// 		},
// 	}
// 	fmt.Println(config)
// 	return config, nil
// }
