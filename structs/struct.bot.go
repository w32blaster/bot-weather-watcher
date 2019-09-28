package structs

// Opts command line arguments
type Opts struct {
	Port    int    `short:"p" long:"port" description:"The port for the bot. The default is 8444" default:"8444"`
	Host    string `short:"h" long:"host" description:"The hostname for the bot. Default is localhost" default:"localhost"`
	IsDebug bool   `short:"d" long:"debug" description:"Is it debug? Default is true. Disable it for production."`

	BotToken       string `short:"b" long:"bot-token" description:"The Bot-Token. As long as it is the sensitive data, we can't keep it in Github" required:"true"`
	MetofficeAppID string `short:"a" long:"appid" description:"AppID to use the MetOffice data point"`
}
