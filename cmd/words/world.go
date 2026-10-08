package main

import (
	"strings"

	"github.com/espresso20/ageforge/config"
)

// world.go holds the game facts the reader needs to say when a player first
// meets a line: the order of the ages, which age a building, tech or badge
// belongs to, and what things are called. It reads them from package config,
// so they are the game's own.

type world struct {
	ageKeys  []string
	ageNames []string
	ageIdx   map[string]int // age key -> place in the order, from 1
	eraAge   map[string]int // era key -> its first age's place
	eraName  map[string]string
	keyAge   map[string]int    // building, tech, milestone and other keys -> the age they arrive in
	name     map[string]string // any key -> what the game calls it
	wonder   map[string]bool   // the building keys that are wonders
	keys     map[string]bool   // every bare key of the game's data
}

func newWorld() *world {
	w := &world{
		ageIdx: map[string]int{}, eraAge: map[string]int{}, eraName: map[string]string{},
		keyAge: map[string]int{}, name: map[string]string{}, wonder: map[string]bool{}, keys: map[string]bool{},
	}
	for i, a := range config.Ages() {
		w.ageKeys = append(w.ageKeys, a.Key)
		w.ageNames = append(w.ageNames, a.Name)
		w.ageIdx[a.Key] = i + 1
		w.name[a.Key] = a.Name
	}
	for _, e := range config.Epochs() {
		w.eraName[e.Key] = e.Name
		w.name[e.Key] = e.Name
		if len(e.Ages) > 0 {
			w.eraAge[e.Key] = w.ageIdx[e.Ages[0]]
		}
		// A catastrophe ends its era: a player meets it in the era's last age.
		if len(e.Ages) > 0 {
			w.keyAge["catastrophe:"+e.Key] = w.ageIdx[e.Ages[len(e.Ages)-1]]
		}
	}
	put := func(key, name string, age int) {
		w.keys[key[strings.LastIndexByte(key, ':')+1:]] = true
		if name != "" {
			w.name[key] = name
		}
		if age > 0 {
			w.keyAge[key] = age
		}
	}
	for _, b := range config.BaseBuildings() {
		put(b.Key, b.Name, w.ageIdx[b.RequiredAge])
		w.wonder[b.Key] = b.Category == "wonder"
	}
	for _, t := range config.Technologies() {
		put("tech:"+t.Key, t.Name, w.ageIdx[t.Age])
	}
	for _, r := range config.BaseResources() {
		put("res:"+r.Key, r.Name, w.ageIdx[r.Age])
	}
	for _, x := range config.Milestones() {
		put("milestone:"+x.Key, x.Name, w.ageIdx[x.MinAge])
	}
	for _, x := range config.BaseFactions() {
		put("civ:"+x.Key, x.Name, w.ageIdx[x.MinAge])
	}
	for _, x := range config.Harbingers() {
		put("harbinger:"+x.Key, x.Name, w.ageIdx[x.Age])
	}
	for _, d := range config.WorkerDomains() {
		w.keys[d] = true
	}
	for _, l := range config.TechLanes() {
		w.keys[l.Key] = true
	}
	for _, x := range config.FeatureLocks() {
		if t, ok := config.TechByKey()[x.Tech]; ok {
			put("feature:"+x.Key, x.Name, w.ageIdx[t.Age])
		}
	}
	return w
}

// ageOf resolves a key of any kind to the place of the age a player first
// meets it in, or -1.
func (w *world) ageOf(key string) int {
	if key == "" {
		return -1
	}
	if i, ok := w.ageIdx[key]; ok {
		return i
	}
	if i, ok := w.eraAge[key]; ok {
		return i
	}
	if i, ok := w.keyAge[key]; ok {
		return i
	}
	// A badge key names its subject after a dot: "age.stone_age".
	if i := strings.LastIndexByte(key, '.'); i >= 0 {
		return w.ageOf(key[i+1:])
	}
	return -1
}

// called returns what the game calls a key, or the key in plain words.
func (w *world) called(key string) string {
	if n, ok := w.name[key]; ok {
		return n
	}
	for _, p := range []string{"tech:", "res:", "milestone:", "civ:", "harbinger:"} {
		if n, ok := w.name[p+key]; ok {
			return n
		}
	}
	return strings.ReplaceAll(key, "_", " ")
}

// agesText names a span of ages for a where sentence: "the Stone Age", or
// "the Primitive Age to the Classical Age".
func (w *world) agesText(from, to int) string {
	if from < 1 || to < from || to > len(w.ageNames) {
		return "every age"
	}
	if from == to {
		return "the " + w.ageNames[from-1]
	}
	return "the " + w.ageNames[from-1] + " to the " + w.ageNames[to-1]
}
