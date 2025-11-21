package config

import (
        "github.com/df-mc/dragonfly/server/block/cube"
        "github.com/go-gl/mathgl/mgl64"
        "github.com/pelletier/go-toml/v2"
        "github.com/sirupsen/logrus"
        "os"
)

type Config struct {
        Arenas map[string]*ArenaConfig `toml:"arenas"`
        Shop   *ShopConfig             `toml:"shop"`
}

type ArenaConfig struct {
        World       string                  `toml:"world"`
        MinPlayers  int                     `toml:"min_players"`
        MaxPlayers  int                     `toml:"max_players"`
        LobbySpawn  mgl64.Vec3              `toml:"lobby_spawn"`
        Teams       map[string]*TeamConfig  `toml:"teams"`
}

type TeamConfig struct {
        Spawn     mgl64.Vec3 `toml:"spawn"`
        Egg       cube.Pos   `toml:"egg"`
        Generator mgl64.Vec3 `toml:"generator"`
}

type ShopConfig struct {
        Items map[string]*ShopItem `toml:"items"`
}

type ShopItem struct {
        Name     string            `toml:"name"`
        Price    int               `toml:"price"`
        Currency string            `toml:"currency"`
        Item     string            `toml:"item"`
        Amount   int               `toml:"amount"`
}

func LoadConfig(log *logrus.Logger) *Config {
        if _, err := os.Stat("arenas.toml"); os.IsNotExist(err) {
                cfg := createDefaultConfig()
                saveConfig(cfg, "arenas.toml", log)
                return cfg
        }
        
        data, err := os.ReadFile("arenas.toml")
        if err != nil {
                log.Fatalf("Failed to read config: %v", err)
        }
        
        var cfg Config
        if err := toml.Unmarshal(data, &cfg); err != nil {
                log.Fatalf("Failed to decode config: %v", err)
        }
        
        return &cfg
}

func createDefaultConfig() *Config {
        return &Config{
                Arenas: map[string]*ArenaConfig{
                        "default": {
                                World:      "world",
                                MinPlayers: 2,
                                MaxPlayers: 8,
                                LobbySpawn: mgl64.Vec3{0, 100, 0},
                                Teams: map[string]*TeamConfig{
                                        "red": {
                                                Spawn:     mgl64.Vec3{50, 100, 0},
                                                Egg:       cube.Pos{45, 101, 0},
                                                Generator: mgl64.Vec3{48, 100, 0},
                                        },
                                        "blue": {
                                                Spawn:     mgl64.Vec3{-50, 100, 0},
                                                Egg:       cube.Pos{-45, 101, 0},
                                                Generator: mgl64.Vec3{-48, 100, 0},
                                        },
                                        "green": {
                                                Spawn:     mgl64.Vec3{0, 100, 50},
                                                Egg:       cube.Pos{0, 101, 45},
                                                Generator: mgl64.Vec3{0, 100, 48},
                                        },
                                        "yellow": {
                                                Spawn:     mgl64.Vec3{0, 100, -50},
                                                Egg:       cube.Pos{0, 101, -45},
                                                Generator: mgl64.Vec3{0, 100, -48},
                                        },
                                },
                        },
                        "islands": {
                                World:      "world",
                                MinPlayers: 2,
                                MaxPlayers: 8,
                                LobbySpawn: mgl64.Vec3{0, 150, 0},
                                Teams: map[string]*TeamConfig{
                                        "red": {
                                                Spawn:     mgl64.Vec3{100, 150, 100},
                                                Egg:       cube.Pos{95, 151, 100},
                                                Generator: mgl64.Vec3{98, 150, 100},
                                        },
                                        "blue": {
                                                Spawn:     mgl64.Vec3{-100, 150, -100},
                                                Egg:       cube.Pos{-95, 151, -100},
                                                Generator: mgl64.Vec3{-98, 150, -100},
                                        },
                                        "green": {
                                                Spawn:     mgl64.Vec3{100, 150, -100},
                                                Egg:       cube.Pos{95, 151, -100},
                                                Generator: mgl64.Vec3{98, 150, -100},
                                        },
                                        "yellow": {
                                                Spawn:     mgl64.Vec3{-100, 150, 100},
                                                Egg:       cube.Pos{-95, 151, 100},
                                                Generator: mgl64.Vec3{-98, 150, 100},
                                        },
                                },
                        },
                },
                Shop: &ShopConfig{
                        Items: map[string]*ShopItem{
                                "sword_wood": {
                                        Name:     "Wooden Sword",
                                        Price:    10,
                                        Currency: "iron",
                                        Item:     "minecraft:wooden_sword",
                                        Amount:   1,
                                },
                                "blocks": {
                                        Name:     "Blocks (16x)",
                                        Price:    5,
                                        Currency: "iron",
                                        Item:     "minecraft:wool",
                                        Amount:   16,
                                },
                        },
                },
        }
}

func saveConfig(cfg *Config, path string, log *logrus.Logger) {
        data, err := toml.Marshal(cfg)
        if err != nil {
                log.Fatalf("Failed to marshal config: %v", err)
        }
        
        if err := os.WriteFile(path, data, 0644); err != nil {
                log.Fatalf("Failed to write config: %v", err)
        }
        
        log.Infof("Created default config at %s", path)
}
