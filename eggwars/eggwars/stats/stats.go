package stats

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

type PlayerStats struct {
	Name   string `json:"name"`
	Kills  int    `json:"kills"`
	Deaths int    `json:"deaths"`
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
	Games  int    `json:"games"`
}

type StatsManager struct {
	stats map[string]*PlayerStats
	log   *logrus.Logger
	mu    sync.RWMutex
}

func NewStatsManager(log *logrus.Logger) *StatsManager {
	sm := &StatsManager{
		stats: make(map[string]*PlayerStats),
		log:   log,
	}
	
	sm.load()
	return sm
}

func (sm *StatsManager) GetStats(playerName string) *PlayerStats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	if stats, ok := sm.stats[playerName]; ok {
		return stats
	}
	
	return &PlayerStats{
		Name: playerName,
	}
}

func (sm *StatsManager) AddKill(playerName string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if _, ok := sm.stats[playerName]; !ok {
		sm.stats[playerName] = &PlayerStats{Name: playerName}
	}
	
	sm.stats[playerName].Kills++
	sm.save()
}

func (sm *StatsManager) AddDeath(playerName string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if _, ok := sm.stats[playerName]; !ok {
		sm.stats[playerName] = &PlayerStats{Name: playerName}
	}
	
	sm.stats[playerName].Deaths++
	sm.save()
}

func (sm *StatsManager) AddWin(playerName string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if _, ok := sm.stats[playerName]; !ok {
		sm.stats[playerName] = &PlayerStats{Name: playerName}
	}
	
	sm.stats[playerName].Wins++
	sm.stats[playerName].Games++
	sm.save()
}

func (sm *StatsManager) AddLoss(playerName string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if _, ok := sm.stats[playerName]; !ok {
		sm.stats[playerName] = &PlayerStats{Name: playerName}
	}
	
	sm.stats[playerName].Losses++
	sm.stats[playerName].Games++
	sm.save()
}

func (sm *StatsManager) load() {
	if _, err := os.Stat("stats.json"); os.IsNotExist(err) {
		return
	}
	
	data, err := os.ReadFile("stats.json")
	if err != nil {
		sm.log.Errorf("Failed to load stats: %v", err)
		return
	}
	
	if err := json.Unmarshal(data, &sm.stats); err != nil {
		sm.log.Errorf("Failed to unmarshal stats: %v", err)
	}
}

func (sm *StatsManager) save() {
	data, err := json.MarshalIndent(sm.stats, "", "  ")
	if err != nil {
		sm.log.Errorf("Failed to marshal stats: %v", err)
		return
	}
	
	if err := os.WriteFile("stats.json", data, 0644); err != nil {
		sm.log.Errorf("Failed to save stats: %v", err)
	}
}
