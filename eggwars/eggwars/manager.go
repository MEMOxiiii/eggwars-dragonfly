package eggwars

import (
	"fmt"
	"sync"

	"github.com/eggwars-dragonfly/eggwars/eggwars/arena"
	"github.com/eggwars-dragonfly/eggwars/eggwars/commands"
	"github.com/eggwars-dragonfly/eggwars/eggwars/config"
	"github.com/eggwars-dragonfly/eggwars/eggwars/stats"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/sirupsen/logrus"
)

type GameManager struct {
	log     *logrus.Logger
	server  *server.Server
	arenas  map[string]*arena.Arena
	players map[string]*arena.PlayerData
	stats   *stats.StatsManager
	config  *config.Config
	mu      sync.RWMutex
}

func NewGameManager(log *logrus.Logger, srv *server.Server) *GameManager {
	cfg := config.LoadConfig(log)

	gm := &GameManager{
		log:     log,
		server:  srv,
		arenas:  make(map[string]*arena.Arena),
		players: make(map[string]*arena.PlayerData),
		stats:   stats.NewStatsManager(log),
		config:  cfg,
	}

	commands.RegisterCommands(gm)

	return gm
}

func (gm *GameManager) LoadArenas() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Use the server's overworld for arenas.
	w := gm.server.World()

	for name, arenaCfg := range gm.config.Arenas {
		a := arena.NewArena(name, arenaCfg, gm.log, w)
		gm.arenas[name] = a
		gm.log.Infof("Loaded arena: %s", name)
	}

	gm.log.Infof("Loaded %d arenas", len(gm.arenas))
}

func (gm *GameManager) HandlePlayer(p *player.Player) {
	gm.mu.Lock()
	pd := &arena.PlayerData{
		Player: p,
	}
	gm.players[p.Name()] = pd
	gm.mu.Unlock()

	handler := NewPlayerHandler(gm, p)
	p.Handle(handler)
}

func (gm *GameManager) GetArena(name string) interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.arenas[name]
}

func (gm *GameManager) GetArenas() map[string]interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	arenas := make(map[string]interface{})
	for k, v := range gm.arenas {
		arenas[k] = v
	}
	return arenas
}

func (gm *GameManager) GetPlayerData(name string) interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.players[name]
}

func (gm *GameManager) GetStatsManager() interface{} {
	return gm.stats
}

func (gm *GameManager) Log() interface{} {
	return gm.log
}

func (gm *GameManager) GetArenaTyped(name string) *arena.Arena {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.arenas[name]
}

func (gm *GameManager) GetPlayerDataTyped(name string) *arena.PlayerData {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.players[name]
}

func (gm *GameManager) JoinArena(p *player.Player, arenaName string) bool {
	a := gm.GetArenaTyped(arenaName)
	if a == nil {
		p.Message(fmt.Sprintf("§c✗ Arena '%s' not found!", arenaName))
		p.Message("§eUse /arenas to see available arenas.")
		return false
	}

	pd := gm.GetPlayerDataTyped(p.Name())
	if pd == nil {
		p.Message("§c✗ Error: Player data not found! Please reconnect.")
		return false
	}

	if pd.Arena != nil {
		p.Message(fmt.Sprintf("§c✗ You are already in arena '%s'! Use /leave first.", pd.Arena.Name))
		return false
	}

	if a.AddPlayer(p, pd) {
		return true
	}

	p.Message(fmt.Sprintf("§c✗ Could not join arena '%s'.", arenaName))

	switch a.State {
	case arena.Playing:
		p.Message("§eReason: Game is already in progress.")
	case arena.Ending:
		p.Message("§eReason: Game is ending.")
	default:
		if len(a.Players) >= a.Config.MaxPlayers {
			p.Message(fmt.Sprintf("§eReason: Arena is full (%d/%d players).", len(a.Players), a.Config.MaxPlayers))
		} else {
			p.Message("§eReason: Unknown error.")
		}
	}

	return false
}

func (gm *GameManager) LeaveArena(p *player.Player) bool {
	pd := gm.GetPlayerDataTyped(p.Name())
	if pd == nil || pd.Arena == nil {
		p.Message("§c✗ You are not in an arena!")
		return false
	}

	arenaName := pd.Arena.Name
	pd.Arena.RemovePlayer(p)

	gm.mu.Lock()
	pd.Arena = nil
	pd.Team = nil
	pd.IsAlive = false
	gm.mu.Unlock()

	p.Message(fmt.Sprintf("§a✓ You left arena '%s'.", arenaName))
	return true
}

func (gm *GameManager) ListArenas(p *player.Player) {
	p.Message("§6╔═══════════════════════════════╗")
	p.Message("§6║    Available EggWars Arenas   ║")
	p.Message("§6╚═══════════════════════════════╝")
	p.Message("")

	gm.mu.RLock()
	arenas := make([]*arena.Arena, 0, len(gm.arenas))
	for _, a := range gm.arenas {
		arenas = append(arenas, a)
	}
	gm.mu.RUnlock()

	if len(arenas) == 0 {
		p.Message("§eNo arenas configured yet.")
		p.Message("")
		return
	}

	for _, a := range arenas {
		stateMsg := ""
		stateColor := "§7"

		switch a.State {
		case arena.Waiting:
			stateMsg = "Waiting"
			stateColor = "§a"
		case arena.Starting:
			stateMsg = "Starting"
			stateColor = "§e"
		case arena.Playing:
			stateMsg = "Playing"
			stateColor = "§c"
		case arena.Ending:
			stateMsg = "Ending"
			stateColor = "§8"
		}

		players := len(a.Players)
		min := a.Config.MinPlayers
		max := a.Config.MaxPlayers

		p.Message(fmt.Sprintf("§f  %s §f%s%s", a.Name, stateColor, stateMsg))
		p.Message(fmt.Sprintf("    §7Players: %d/%d (min: %d)", players, max, min))
		p.Message(fmt.Sprintf("    §7Command: /join %s", a.Name))
		p.Message("")
	}

	pd := gm.GetPlayerDataTyped(p.Name())
	if pd != nil && pd.Arena != nil {
		p.Message(fmt.Sprintf("§eℹ You are currently in: %s", pd.Arena.Name))
		p.Message("")
	}
}

func (gm *GameManager) ShowStats(p *player.Player) {
	stats := gm.stats.GetStats(p.Name())

	p.Message("§6╔═══════════════════════════════╗")
	p.Message(fmt.Sprintf("§6║  Statistics for %s", p.Name()))
	p.Message("§6╚═══════════════════════════════╝")
	p.Message("")
	p.Message(fmt.Sprintf("§f  Kills:   §e%d", stats.Kills))
	p.Message(fmt.Sprintf("§f  Deaths:  §e%d", stats.Deaths))
	p.Message(fmt.Sprintf("§f  Wins:    §a%d", stats.Wins))
	p.Message(fmt.Sprintf("§f  Losses:  §c%d", stats.Losses))
	p.Message(fmt.Sprintf("§f  Games:   §8%d", stats.Games))
	p.Message("")

	kd := 0.0
	if stats.Deaths > 0 {
		kd = float64(stats.Kills) / float64(stats.Deaths)
	} else if stats.Kills > 0 {
		kd = float64(stats.Kills)
	}

	kdColor := "§8"
	if kd >= 2.0 {
		kdColor = "§a"
	} else if kd >= 1.0 {
		kdColor = "§e"
	} else if kd > 0 {
		kdColor = "§c"
	}

	p.Message(fmt.Sprintf("§f  K/D Ratio: %s%.2f", kdColor, kd))
	p.Message("")
}
