package team

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

type Color string

const (
	Red    Color = "§c"
	Blue   Color = "§9"
	Green  Color = "§a"
	Yellow Color = "§e"
)

type Team struct {
	Name       string
	Color      Color
	ColorName  string
	Spawn      mgl64.Vec3
	EggPos     cube.Pos
	EggAlive   bool
	Players    []string
	Generator  mgl64.Vec3
}

func NewTeam(name string, color Color, colorName string, spawn mgl64.Vec3, eggPos cube.Pos, generator mgl64.Vec3) *Team {
	return &Team{
		Name:      name,
		Color:     color,
		ColorName: colorName,
		Spawn:     spawn,
		EggPos:    eggPos,
		EggAlive:  true,
		Players:   make([]string, 0),
		Generator: generator,
	}
}

func (t *Team) AddPlayer(playerName string) {
	t.Players = append(t.Players, playerName)
}

func (t *Team) RemovePlayer(playerName string) {
	for i, name := range t.Players {
		if name == playerName {
			t.Players = append(t.Players[:i], t.Players[i+1:]...)
			return
		}
	}
}

func (t *Team) HasPlayer(playerName string) bool {
	for _, name := range t.Players {
		if name == playerName {
			return true
		}
	}
	return false
}

func (t *Team) PlayerCount() int {
	return len(t.Players)
}

func (t *Team) IsAlive() bool {
	return t.EggAlive || len(t.Players) > 0
}

func (t *Team) BreakEgg() {
	t.EggAlive = false
}
