package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
)

// Entity wraps a factorio entity
type Entity struct {
	ID          string
	Time        float32
	Ingredients []Ingredient
	Stations    []*Station
}

// A Station represents a crafting station
type Station struct {
	ID    string
	Level int
	Speed float32 // items per second
}

// Ingredient represents an ingredient in a crafting recipe.
//
type Ingredient struct {
	Quantity int
	Entity   *Entity
}

var stations = make(map[string]*Station)
var entities = make(map[string]*Entity)

func readStations() error {
	type jsonStation struct {
		Level int
		Speed float32
	}

	var stationList map[string]jsonStation

	data, err := ioutil.ReadFile("stations.json")
	if err != nil {
		return errors.Wrap(err, "failed to read stations.json")
	}
	err = json.Unmarshal(data, &stationList)
	if err != nil {
		return errors.Wrap(err, "failed to read json data from stations.json")
	}
	for id, s := range stationList {
		stations[id] = &Station{id, s.Level, s.Speed}
	}
	return nil
}

func readEntities() error {
	type jsonEntity struct {
		Time        float32
		Ingredients map[string]int
		Stations    []string
	}

	var el map[string]jsonEntity

	data, err := ioutil.ReadFile("entities.json")
	if err != nil {
		return errors.Wrap(err, "failed to read entities.json")
	}
	err = json.Unmarshal(data, &el)
	if err != nil {
		return errors.Wrap(err, "failed to read json data from entities.json")
	}

	for id, e := range el {
		if e.Time == 0 {
			return errors.Errorf("entity %s: invalid 0 time", id)
		}
		ne := &Entity{
			ID:       id,
			Time:     e.Time,
			Stations: make([]*Station, 0, len(e.Stations)),
		}
		if len(e.Stations) == 0 {
			return errors.Errorf("entity %s: no crafting stations", id)
		}
		for _, sid := range e.Stations {
			s := stations[sid]
			if s == nil {
				return errors.Errorf("entity %s: unknown station %s", id, sid)
			}
			ne.Stations = append(ne.Stations, s)
		}
		entities[id] = ne
	}

	// fill in ingredients
	for id, e := range el {
		ne := entities[id]
		ne.Ingredients = make([]Ingredient, 0, len(e.Ingredients))
		for iid, q := range e.Ingredients {
			ie := entities[iid]
			if ie == nil {
				return errors.Errorf("entity %s: unknown ingredient %s", id, iid)
			}
			ne.Ingredients = append(ne.Ingredients, Ingredient{Quantity: q, Entity: ie})
		}

		// b, err := json.MarshalIndent(ne, "", "\t")
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Println("added entity ", id, " : ", string(b))
	}

	return nil
}

type craft struct {
	s *Station
	e *Entity
	q float32
}

type craftList map[string]*craft

func addCraft(e *Entity, s *Station, count float32, techLevel int, cl craftList) {
	c := cl[e.ID]
	if c == nil {
		c = &craft{
			s: s,
			e: e,
			q: 0,
		}
		cl[e.ID] = c
	}
	c.q += count

	cycleTime := e.Time / s.Speed

	// count / (e.Time / s.Speed)

	for _, i := range e.Ingredients {
		ie := i.Entity
		is := ie.stationForLevel(techLevel)
		itemsNeeded := count * float32(i.Quantity)
		addCraft(i.Entity, is, itemsNeeded/cycleTime*(ie.Time/is.Speed), techLevel, cl)
	}
}

func getRecipe(name string, count int, techLevel int) craftList {
	cl := make(craftList)
	e := entities[name]
	if e == nil {
		panic("unknown entity " + name)
	}
	s := e.stationForLevel(techLevel)
	addCraft(e, s, float32(count), techLevel, cl)
	return cl
}

func (e *Entity) stationForLevel(level int) *Station {
	var best *Station
	for _, s := range e.Stations {
		if s.Level > level {
			continue
		}
		if best == nil || best.Level < level {
			best = s
		}
	}
	return best
}

func main() {
	var err error
	if err = readStations(); err != nil {
		panic(err)
	}
	if err = readEntities(); err != nil {
		panic(err)
	}

	name := "automation_science_pack"
	// nname := "iron_plate"
	count := 5
	cl := getRecipe(name, count, 1)
	for _, c := range cl {
		fmt.Printf("%.2f x %s\t-> %s - %.2f items/s\n", c.q, c.s.ID, c.e.ID, c.q*c.s.Speed/c.e.Time)
	}
}
