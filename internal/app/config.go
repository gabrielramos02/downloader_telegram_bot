package app

import "fmt"

type Config struct {
	BotToken string
	Env      string
	GPURL    string
	GPToken  string
}

func validateEnvVars(vars map[string]string) error {
	for _, v := range requiredEnvVars {
		if vars[v] == "" {
			return fmt.Errorf("required env variable %s not set", v)
		}
	}
	return nil
}

func LoadConfig(vars map[string]string) (Config, error) {
	if err := validateEnvVars(vars); err != nil {
		return Config{}, err
	}
	return Config{
		BotToken: vars["BOT_TOKEN"],
		Env:      vars["ENV"],
		GPURL:    vars["GP_URL"],
		GPToken:  vars["GP_TOKEN"],
	}, nil
}

var requiredEnvVars = []string{
	"BOT_TOKEN",
	"ENV",
	"GP_URL",
	"GP_TOKEN",
}
