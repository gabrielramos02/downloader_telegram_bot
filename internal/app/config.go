package app

import "fmt"

type Config struct {
	BotToken   string
	Env        string
	QBURL      string
	QBUsername string
	QBPassword string
	GPURL      string
	GPToken    string
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
		BotToken:   vars["BOT_TOKEN"],
		Env:        vars["ENV"],
		QBURL:      vars["QB_URL"],
		QBUsername: vars["QB_USERNAME"],
		QBPassword: vars["QB_PASSWORD"],
		GPURL:      vars["GP_URL"],
		GPToken:    vars["GP_TOKEN"],
	}, nil
}

var requiredEnvVars = []string{
	"BOT_TOKEN",
	"ENV",
	"QB_URL",
	"QB_USERNAME",
	"QB_PASSWORD",
	"GP_URL",
	"GP_TOKEN",
}
