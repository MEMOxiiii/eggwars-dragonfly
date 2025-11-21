package server

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
)

type EventListener struct{}

func (e *EventListener) HandlePlayerJoin(p *player.Player) {
	p.Message("Welcome to EggWars! Prepare for battle.")
	// Additional logic for player join can be added here.
}

func (e *EventListener) HandlePlayerQuit(p *player.Player) {
	// Logic for handling player quit.
	p.Message("Goodbye! Thanks for playing EggWars.")
}

func (e *EventListener) HandleBlockBreak(ctx *event.Context[any], p *player.Player) {
	// Prevent block breaking in the lobby.
	ctx.Cancel()
	p.Message("You cannot break blocks here.")
}

func RegisterEvents() {
	// Register event handlers here.
}
