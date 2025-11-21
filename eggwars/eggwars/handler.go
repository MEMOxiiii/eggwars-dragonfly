package eggwars

import (
        "github.com/df-mc/dragonfly/server/block/cube"
        "github.com/df-mc/dragonfly/server/item"
        "github.com/df-mc/dragonfly/server/player"
        "github.com/df-mc/dragonfly/server/world"
)

type PlayerHandler struct {
        player.NopHandler
        gm *GameManager
        p  *player.Player
}

func NewPlayerHandler(gm *GameManager, p *player.Player) *PlayerHandler {
        return &PlayerHandler{
                gm: gm,
                p:  p,
        }
}

func (h *PlayerHandler) HandleQuit(p *player.Player) {
        pd := h.gm.GetPlayerDataTyped(p.Name())
        if pd != nil && pd.Arena != nil {
                pd.Arena.RemovePlayer(p)
        }
}

func (h *PlayerHandler) HandleDeath(p *player.Player, src world.DamageSource, keepInv *bool) {
        pd := h.gm.GetPlayerDataTyped(p.Name())
        if pd != nil && pd.Arena != nil {
                pd.Arena.HandlePlayerDeath(p)
                *keepInv = true
        }
}


func (h *PlayerHandler) HandleBlockBreak(ctx *player.Context, pos cube.Pos, drops *[]item.Stack, xp *int) {
        pd := h.gm.GetPlayerDataTyped(h.p.Name())
        if pd != nil && pd.Arena != nil {
                if !pd.Arena.CanBreakBlock(h.p, pos) {
                        ctx.Cancel()
                }
        }
}

func (h *PlayerHandler) HandleBlockPlace(ctx *player.Context, pos cube.Pos, b world.Block) {
        pd := h.gm.GetPlayerDataTyped(h.p.Name())
        if pd != nil && pd.Arena != nil {
                if pd.Arena.IsPlaying() {
                        pd.Arena.TrackPlacedBlock(pos)
                }
        }
}

func (h *PlayerHandler) HandleItemUse(ctx *player.Context) {
        pd := h.gm.GetPlayerDataTyped(h.p.Name())
        if pd != nil && pd.Arena != nil {
                held, _ := h.p.HeldItems()
                if _, ok := held.Item().(item.Paper); ok {
                        ctx.Cancel()
                        pd.Arena.OpenShop(h.p)
                }
        }
}
