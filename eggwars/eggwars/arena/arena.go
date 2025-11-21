package arena

import (
	"fmt"
	"sync"
	"time"

	"github.com/eggwars-dragonfly/eggwars/eggwars/config"
	"github.com/eggwars-dragonfly/eggwars/eggwars/generator"
	"github.com/eggwars-dragonfly/eggwars/eggwars/team"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sirupsen/logrus"
)

type GameState int

const (
	Waiting GameState = iota
	Starting
	Playing
	Ending
)

// shopMenu is a MenuSubmittable used for the in-game shop.
type shopMenu struct {
	a      *Arena
	Iron   form.Button
	Sword  form.Button
	Shield form.Button
	Close  form.Button
}

func (m shopMenu) Submit(submitter form.Submitter, pressed form.Button, tx *world.Tx) {
	p, ok := submitter.(*player.Player)
	if !ok {
		return
	}

	if pressed == m.Close {
		p.Message("§aShop closed!")
		return
	}

	pd := m.a.GetPlayerData(p.Name())
	if pd == nil {
		p.Message("§c✗ Error processing purchase")
		return
	}

	switch {
	case pressed == m.Iron:
		if iron, ok := pd.Resources["iron"]; ok && iron >= 10 {
			pd.Resources["iron"] -= 10
			p.Message("§a✓ Purchased Iron Helmet!")
		} else {
			p.Message("§c✗ Not enough Iron! Need: 10, Have: " + fmt.Sprint(pd.Resources["iron"]))
		}
	case pressed == m.Sword:
		if diamond, ok := pd.Resources["diamond"]; ok && diamond >= 5 {
			pd.Resources["diamond"] -= 5
			p.Message("§a✓ Purchased Diamond Sword!")
		} else {
			p.Message("§c✗ Not enough Diamond! Need: 5, Have: " + fmt.Sprint(pd.Resources["diamond"]))
		}
	case pressed == m.Shield:
		gold := pd.Resources["gold"]
		iron := pd.Resources["iron"]
		if gold >= 5 && iron >= 10 {
			pd.Resources["gold"] -= 5
			pd.Resources["iron"] -= 10
			p.Message("§a✓ Purchased Shield!")
		} else {
			p.Message(fmt.Sprintf("§c✗ Not enough resources! Need: 5 Gold + 10 Iron, Have: %d Gold + %d Iron", gold, iron))
		}
	}
}

type Arena struct {
	Name         string
	Config       *config.ArenaConfig
	State        GameState
	Teams        map[team.Color]*team.Team
	Players      map[string]*PlayerData
	PlacedBlocks map[cube.Pos]bool
	Generators   map[team.Color]*generator.Generator
	World        *world.World
	log          *logrus.Logger
	mu           sync.RWMutex
	startTimer   *time.Timer
}

type PlayerData struct {
	Player    *player.Player
	Team      *team.Team
	Arena     *Arena
	Kills     int
	Deaths    int
	IsAlive   bool
	Resources map[string]int
}

func NewArena(name string, cfg *config.ArenaConfig, log *logrus.Logger, w *world.World) *Arena {
	a := &Arena{
		Name:         name,
		Config:       cfg,
		State:        Waiting,
		Teams:        make(map[team.Color]*team.Team),
		Players:      make(map[string]*PlayerData),
		PlacedBlocks: make(map[cube.Pos]bool),
		Generators:   make(map[team.Color]*generator.Generator),
		World:        w,
		log:          log,
	}

	a.initTeams()
	a.initGenerators()
	return a
}

func (a *Arena) initTeams() {
	if red, ok := a.Config.Teams["red"]; ok {
		a.Teams[team.Red] = team.NewTeam("Red", team.Red, "red", red.Spawn, red.Egg, red.Generator)
	}
	if blue, ok := a.Config.Teams["blue"]; ok {
		a.Teams[team.Blue] = team.NewTeam("Blue", team.Blue, "blue", blue.Spawn, blue.Egg, blue.Generator)
	}
	if green, ok := a.Config.Teams["green"]; ok {
		a.Teams[team.Green] = team.NewTeam("Green", team.Green, "green", green.Spawn, green.Egg, green.Generator)
	}
	if yellow, ok := a.Config.Teams["yellow"]; ok {
		a.Teams[team.Yellow] = team.NewTeam("Yellow", team.Yellow, "yellow", yellow.Spawn, yellow.Egg, yellow.Generator)
	}
}

func (a *Arena) initGenerators() {
	if a.World == nil {
		return
	}

	for color, t := range a.Teams {
		a.Generators[color] = generator.NewGenerator(t.Generator, generator.Iron, a.World)
	}
}

func (a *Arena) AddPlayer(p *player.Player, externalPd *PlayerData) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.State != Waiting && a.State != Starting {
		return false
	}

	if len(a.Players) >= a.Config.MaxPlayers {
		return false
	}

	assignedTeam := a.getSmallestTeam()
	if assignedTeam == nil {
		return false
	}

	if externalPd != nil {
		externalPd.Team = assignedTeam
		externalPd.Arena = a
		externalPd.IsAlive = true
		externalPd.Kills = 0
		externalPd.Deaths = 0
		if externalPd.Resources == nil {
			externalPd.Resources = make(map[string]int)
		}
		a.Players[p.Name()] = externalPd
	} else {
		pd := &PlayerData{
			Player:    p,
			Team:      assignedTeam,
			Arena:     a,
			IsAlive:   true,
			Resources: make(map[string]int),
		}
		a.Players[p.Name()] = pd
	}

	assignedTeam.AddPlayer(p.Name())

	p.Teleport(a.Config.LobbySpawn)
	p.Message(fmt.Sprintf("<green>✓ Joined arena '%s' as %steam %s!</green>", a.Name, assignedTeam.Color, assignedTeam.ColorName))
	a.broadcast(fmt.Sprintf("%s%s<white> joined the game! (%d/%d)</white>",
		assignedTeam.Color, p.Name(), len(a.Players), a.Config.MaxPlayers))

	if len(a.Players) >= a.Config.MinPlayers && a.State == Waiting {
		a.startCountdown()
	}

	return true
}

func (a *Arena) RemovePlayer(p *player.Player) {
	a.mu.Lock()
	defer a.mu.Unlock()

	pd, ok := a.Players[p.Name()]
	if !ok {
		return
	}

	if pd.Team != nil {
		pd.Team.RemovePlayer(p.Name())
	}

	delete(a.Players, p.Name())

	a.broadcast(fmt.Sprintf("<yellow>%s left the game!</yellow>", p.Name()))

	if a.State == Playing {
		a.checkWinCondition()
	} else if len(a.Players) < a.Config.MinPlayers && a.State == Starting {
		a.cancelCountdown()
	}
}

func (a *Arena) getSmallestTeam() *team.Team {
	var smallest *team.Team
	minPlayers := 999

	for _, t := range a.Teams {
		if t.PlayerCount() < minPlayers {
			minPlayers = t.PlayerCount()
			smallest = t
		}
	}

	return smallest
}

func (a *Arena) startCountdown() {
	a.State = Starting
	a.broadcast("<green>Game starting in 10 seconds!</green>")

	a.startTimer = time.AfterFunc(10*time.Second, func() {
		a.startGame()
	})
}

func (a *Arena) cancelCountdown() {
	if a.startTimer != nil {
		a.startTimer.Stop()
	}
	a.State = Waiting
	a.broadcast("<red>Not enough players! Countdown cancelled.</red>")
}

func (a *Arena) startGame() {
	a.mu.Lock()
	a.State = Playing
	a.mu.Unlock()

	a.broadcast("<gold>===== GAME STARTED! =====</gold>")
	a.broadcast("<yellow>Protect your egg and destroy others!</yellow>")

	for _, pd := range a.Players {
		pd.IsAlive = true
		if pd.Team != nil {
			pd.Player.Teleport(pd.Team.Spawn)
			pd.Player.Message(fmt.Sprintf("%sYou are in team %s!", pd.Team.Color, pd.Team.ColorName))
		}
		if pd.Resources == nil {
			pd.Resources = make(map[string]int)
		}
		pd.Resources["iron"] = 0
		pd.Resources["gold"] = 0
		pd.Resources["diamond"] = 0
	}

	for color, gen := range a.Generators {
		if t, ok := a.Teams[color]; ok && t.EggAlive {
			gen.Start()
			a.log.Infof("Generator started for team %s at %v", t.Name, t.Generator)
		}
	}
}

func (a *Arena) HandlePlayerDeath(p *player.Player) {
	a.mu.Lock()
	defer a.mu.Unlock()

	pd, ok := a.Players[p.Name()]
	if !ok {
		return
	}

	pd.Deaths++

	if pd.Team != nil && pd.Team.EggAlive {
		time.AfterFunc(3*time.Second, func() {
			pd.IsAlive = true
			p.Teleport(pd.Team.Spawn)
			p.Message("<green>You respawned!</green>")
		})
	} else {
		pd.IsAlive = false
		teamMsg := ""
		if pd.Team != nil {
			teamMsg = string(pd.Team.Color)
		}
		a.broadcast(fmt.Sprintf("%s%s<white> was eliminated!</white>", teamMsg, p.Name()))
		a.checkWinCondition()
	}
}

func (a *Arena) checkWinCondition() {
	aliveTeams := 0
	var winningTeam *team.Team

	for _, t := range a.Teams {
		if t.IsAlive() {
			aliveTeams++
			winningTeam = t
		}
	}

	if aliveTeams <= 1 && a.State == Playing {
		a.endGame(winningTeam)
	}
}

func (a *Arena) endGame(winningTeam *team.Team) {
	a.State = Ending

	for _, gen := range a.Generators {
		gen.Stop()
	}

	if winningTeam != nil {
		a.broadcast("<gold>===== GAME OVER! =====</gold>")
		a.broadcast(fmt.Sprintf("%sTeam %s wins!", winningTeam.Color, winningTeam.ColorName))
	} else {
		a.broadcast("<gold>Game ended with no winners!</gold>")
	}

	time.AfterFunc(10*time.Second, func() {
		a.reset()
	})
}

func (a *Arena) reset() {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, gen := range a.Generators {
		gen.Stop()
	}

	for _, pd := range a.Players {
		pd.Player.Teleport(a.Config.LobbySpawn)
	}

	a.Players = make(map[string]*PlayerData)
	a.PlacedBlocks = make(map[cube.Pos]bool)
	a.State = Waiting
	a.initTeams()
	a.initGenerators()
}

func (a *Arena) broadcast(message string) {
	for _, pd := range a.Players {
		pd.Player.Message(message)
	}
}

func (a *Arena) CanBreakBlock(p *player.Player, pos cube.Pos) bool {
	if a.PlacedBlocks[pos] {
		return true
	}

	for _, t := range a.Teams {
		if t.EggPos == pos && t.EggAlive {
			pd := a.Players[p.Name()]
			if pd != nil && pd.Team != nil && pd.Team.Color != t.Color {
				t.BreakEgg()
				a.broadcast(fmt.Sprintf("<red>%sTeam %s's egg was destroyed!</red>", t.Color, t.ColorName))
				return true
			}
			return false
		}
	}

	return false
}

func (a *Arena) TrackPlacedBlock(pos cube.Pos) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.PlacedBlocks[pos] = true
}

func (a *Arena) IsPlaying() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.State == Playing
}

func (a *Arena) OpenShop(p *player.Player) {
	pd := a.GetPlayerData(p.Name())
	if pd == nil {
		p.Message("§c✗ Error opening shop")
		return
	}

	resources := ""
	if pd.Resources != nil {
		if iron, ok := pd.Resources["iron"]; ok {
			resources += fmt.Sprintf("§6Iron: §f%d  ", iron)
		}
		if gold, ok := pd.Resources["gold"]; ok {
			resources += fmt.Sprintf("§eGold: §f%d  ", gold)
		}
		if diamond, ok := pd.Resources["diamond"]; ok {
			resources += fmt.Sprintf("§bDiamond: §f%d", diamond)
		}
	}

	menu := form.NewMenu(shopMenu{
		a: a,
		Iron:   form.NewButton("§cIron Helmet\n§7Cost: 10 Iron", ""),
		Sword:  form.NewButton("§f⚔ Diamond Sword\n§7Cost: 5 Diamond", ""),
		Shield: form.NewButton("§9Shield\n§7Cost: 5 Gold + 10 Iron", ""),
		Close:  form.NewButton("§8Close", ""),
	}, "§6EggWars Shop")

	menu = menu.WithBody(fmt.Sprintf("Your Resources:\n%s\n\nSelect an item:", resources))

	p.SendForm(menu)
}

func (a *Arena) GetPlayerData(name string) *PlayerData {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Players[name]
}
