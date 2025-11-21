package server

import (
	"github.com/df-mc/dragonfly/server/player"
)

type Scoreboard struct {
	Title string
	Lines []string
}

func (s *Scoreboard) Update(p *player.Player) {
	// Logic to update the scoreboard for the player.
	p.Message("Scoreboard updated: " + s.Title)
}

func NewGameScoreboard() *Scoreboard {
	return &Scoreboard{
		Title: "EggWars",
		Lines: []string{
			"Players: 0",
			"Time Left: 10:00",
		},
	}
}
