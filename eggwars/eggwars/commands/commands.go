package commands

import (
        "github.com/df-mc/dragonfly/server/cmd"
        "github.com/df-mc/dragonfly/server/player"
        "github.com/df-mc/dragonfly/server/world"
)

type GameManager interface {
        GetArena(name string) interface{}
        GetArenas() map[string]interface{}
        GetPlayerData(name string) interface{}
        GetStatsManager() interface{}
        Log() interface{}
        JoinArena(p *player.Player, arenaName string) bool
        LeaveArena(p *player.Player) bool
        ListArenas(p *player.Player)
        ShowStats(p *player.Player)
}

var globalGameManager GameManager

func RegisterCommands(gm GameManager) {
        globalGameManager = gm
        
        cmd.Register(cmd.New("join", "Join an EggWars arena", []string{}, JoinArenaCommand{}))
        cmd.Register(cmd.New("leave", "Leave current arena", []string{}, LeaveArenaCommand{}))
        cmd.Register(cmd.New("arenas", "List all arenas", []string{}, ListArenasCommand{}))
        cmd.Register(cmd.New("ewstats", "Show your statistics", []string{}, StatsShowCommand{}))
}

type EggWarsCommand struct {
        Sub cmd.SubCommand `cmd:"sub"`
}

func (e EggWarsCommand) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
        p, ok := src.(*player.Player)
        if !ok {
                o.Error("This command can only be used by players")
                return
        }
        
        p.Message("<yellow>EggWars Commands:</yellow>")
        p.Message("<white>/eggwars join <arena> - Join an arena</white>")
        p.Message("<white>/eggwars leave - Leave current arena</white>")
        p.Message("<white>/eggwars list - List all arenas</white>")
        p.Message("<white>/eggwars stats - View your statistics</white>")
}

type JoinArenaCommand struct {
        Arena string `cmd:"arena"`
}

func (j JoinArenaCommand) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
        p, ok := src.(*player.Player)
        if !ok {
                o.Error("Only players can use this command")
                return
        }
        
        if globalGameManager != nil {
                globalGameManager.JoinArena(p, j.Arena)
        }
}

type LeaveArenaCommand struct{}

func (l LeaveArenaCommand) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
        p, ok := src.(*player.Player)
        if !ok {
                o.Error("Only players can use this command")
                return
        }
        
        if globalGameManager != nil {
                globalGameManager.LeaveArena(p)
        }
}

type ListArenasCommand struct{}

func (l ListArenasCommand) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
        p, ok := src.(*player.Player)
        if !ok {
                o.Error("Only players can use this command")
                return
        }
        
        if globalGameManager != nil {
                globalGameManager.ListArenas(p)
        }
}

type StatsShowCommand struct{}

func (s StatsShowCommand) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
        p, ok := src.(*player.Player)
        if !ok {
                o.Error("Only players can use this command")
                return
        }
        
        if globalGameManager != nil {
                globalGameManager.ShowStats(p)
        }
}
