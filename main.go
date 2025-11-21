package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/eggwars-dragonfly/eggwars/eggwars"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player/chat"
	toml "github.com/pelletier/go-toml/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create a logrus logger for the plugin.
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.Info("Starting Dragonfly EggWars Server...")

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	conf, err := readConfig(slog.Default())
	if err != nil {
		panic(err)
	}

	srv := conf.New()
	srv.CloseOnProgramEnd()

	// Initialize EggWars manager from the external plugin package.
	eggMgr := eggwars.NewGameManager(log, srv)
	eggMgr.LoadArenas()

	srv.Listen()
	log.Info("Server is now running!")

	for p := range srv.Accept() {
		p.Message("<green>Welcome to EggWars Server!</green>")
		p.Message("<yellow>Use /eggwars list to see available arenas</yellow>")

		go eggMgr.HandlePlayer(p)
	}
}

// readConfig reads the configuration from the config.toml file, or creates the
// file if it does not yet exist.
func readConfig(log *slog.Logger) (server.Config, error) {
	c := server.DefaultConfig()
	var zero server.Config
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(c)
		if err != nil {
			return zero, fmt.Errorf("encode default config: %v", err)
		}
		if err := os.WriteFile("config.toml", data, 0644); err != nil {
			return zero, fmt.Errorf("create default config: %v", err)
		}
		return c.Config(log)
	}
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return zero, fmt.Errorf("read config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return zero, fmt.Errorf("decode config: %v", err)
	}
	return c.Config(log)
}
