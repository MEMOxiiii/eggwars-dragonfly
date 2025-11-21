package main

import (
	"os"

	"github.com/eggwars-dragonfly/eggwars/eggwars"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/pelletier/go-toml/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.Info("Starting Dragonfly EggWars Server...")

	srv := readConfig(log)

	chat.Global.Subscribe(chat.StdoutSubscriber{})

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

func readConfig(log *logrus.Logger) *server.Server {
	c := server.DefaultConfig()

	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, _ := toml.Marshal(c)
		if err := os.WriteFile("config.toml", data, 0644); err != nil {
			log.Fatalf("Failed to create config: %v", err)
		}
	}

	data, err := os.ReadFile("config.toml")
	if err != nil {
		log.Warnf("Could not read config: %v, using defaults", err)
	} else {
		if err := toml.Unmarshal(data, &c); err != nil {
			log.Warnf("Could not decode config: %v, using defaults", err)
		}
	}

	srv := server.New()
	return srv
}
