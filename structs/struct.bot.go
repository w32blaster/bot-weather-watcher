package structs

// Opts command line arguments
type Opts struct {
	Port           int    `env:"PORT" envDefault:"8444"`
	Host           string `env:"HOST" envDefault:"localhost"`
	IsDebug        bool   `env:"IS_DEBUG"`
	BotToken       string `env:"BOT_TOKEN,required"`
	MetofficeAppID string `env:"METOFFICE_APP_ID"`
}
