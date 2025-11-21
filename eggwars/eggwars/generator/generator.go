package generator

import (
	"time"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type ResourceType string

const (
	Iron    ResourceType = "iron"
	Gold    ResourceType = "gold"
	Diamond ResourceType = "diamond"
)

type Generator struct {
	Position     mgl64.Vec3
	ResourceType ResourceType
	Interval     time.Duration
	World        *world.World
	running      bool
	stopChan     chan bool
}

func NewGenerator(pos mgl64.Vec3, resType ResourceType, w *world.World) *Generator {
	interval := 2 * time.Second

	switch resType {
	case Gold:
		interval = 5 * time.Second
	case Diamond:
		interval = 10 * time.Second
	}

	return &Generator{
		Position:     pos,
		ResourceType: resType,
		Interval:     interval,
		World:        w,
		stopChan:     make(chan bool),
	}
}

func (g *Generator) Start() {
	if g.running {
		return
	}

	g.running = true
	go g.generate()
}

func (g *Generator) Stop() {
	if !g.running {
		return
	}

	g.running = false
	g.stopChan <- true
}

func (g *Generator) generate() {
	ticker := time.NewTicker(g.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.spawnResource()
		case <-g.stopChan:
			return
		}
	}
}

func (g *Generator) spawnResource() {
	var itemStack item.Stack

	switch g.ResourceType {
	case Iron:
		itemStack = item.NewStack(item.IronIngot{}, 1)
	case Gold:
		itemStack = item.NewStack(item.GoldIngot{}, 1)
	case Diamond:
		itemStack = item.NewStack(item.Diamond{}, 1)
	}

	if g.World != nil && itemStack.Count() > 0 {
		e := g.World.EntityRegistry().Config().Item(world.EntitySpawnOpts{Position: g.Position}, itemStack)
		<-g.World.Exec(func(tx *world.Tx) {
			tx.AddEntity(e)
		})
	}
}
