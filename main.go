package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
)

// Entity wraps a factorio entity
type Entity struct {
	ID          string
	Time        float64
	Ingredients []Ingredient
	Stations    []*Station
}

// A Station represents a crafting station
type Station struct {
	ID    string
	Level int
	Speed float64 // items per second
}

// Ingredient represents an ingredient in a crafting recipe.
//
type Ingredient struct {
	Quantity float64
	Entity   *Entity
}

var stations = make(map[string]*Station)
var entities = make(map[string]*Entity)

func readStations() error {
	type jsonStation struct {
		Level int
		Speed float64
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
		Batch       float64
		Time        float64
		Ingredients map[string]float64
		Stations    []string
	}

	var el map[string]*jsonEntity

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
		if e.Batch == 0 {
			e.Batch = 1
		}
		ne := &Entity{
			ID:       id,
			Time:     e.Time / e.Batch,
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
			ne.Ingredients = append(ne.Ingredients, Ingredient{Quantity: q / e.Batch, Entity: ie})
		}
	}

	return nil
}

type Production struct {
	e   *Entity
	ips float64
	s   *Station
	sc  float64
}

type ProdList map[string]*Production

func (pl ProdList) add(e *Entity, ips float64) {
	p := pl[e.ID]
	if p == nil {
		p = &Production{e: e}
		pl[e.ID] = p
	}
	p.ips += ips

	for _, i := range e.Ingredients {
		pl.add(i.Entity, ips*i.Quantity)
	}
}

func NewProduction(name string, ips float64, techLevel int) (ProdList, error) {
	pl := make(ProdList)

	if e := entities[name]; e != nil {
		pl.add(e, ips)
	} else {
		return nil, errors.Errorf("unknown entity %s", name)
	}

	// update production list with appropriate stations
	for _, p := range pl {
		for _, s := range p.e.Stations {
			if s.Level > techLevel {
				continue
			}
			c := math.Ceil(p.ips * p.e.Time / s.Speed)
			if p.s == nil || c < p.sc || (s.Level < p.s.Level && c == p.sc) {
				p.s = s
				p.sc = c
			}
		}
		if p.s == nil {
			return nil, errors.Errorf("no station available to produce entity %s at tech level %d", p.e.ID, techLevel)
		}
	}

	return pl, nil
}

func main() {
	var err error
	if err = readStations(); err != nil {
		panic(err)
	}
	if err = readEntities(); err != nil {
		panic(err)
	}

	var (
		list = flag.Bool("list", false, "lists known entities")
		name = flag.String("e", "", "entity `name`")
		ips  = flag.Float64("i", 1.0, "where `ips` is the number of items per second")
		l    = flag.Int("l", 3, "maximum assembly machine `level`")
	)

	flag.Parse()

	if *list {
		fmt.Println("Known entities:")
		var el []string
		for k := range entities {
			el = append(el, k)
		}
		sort.Strings(el)
		for _, en := range el {
			fmt.Println(en)
		}
		os.Exit(0)
	}

	if *l < 1 {
		fmt.Fprintf(os.Stderr, "error: assembly machine level must be greater or equal to 1")
	}
	if *ips <= 0.0 {
		fmt.Fprintf(os.Stderr, "error: items per second must be a non-zero positive number")
	}

	*name = strings.Replace(strings.ToLower(*name), " ", "_", -1)

	pl, err := NewProduction(*name, *ips, *l)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var spl []*Production
	for _, p := range pl {
		spl = append(spl, p)
	}
	sort.Slice(spl, func(i, j int) bool {
		return spl[i].s.ID < spl[j].s.ID || (spl[i].s.ID == spl[j].s.ID && spl[i].e.ID < spl[j].e.ID)
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "#\tProducer\tItem\titems/s\n")
	fmt.Fprintf(w, "--\t-------------------------\t-------------------------\t-------\n")

	for _, p := range spl {
		fmt.Fprintf(w, "%.0f\t%s\t%s\t%.2f\n", p.sc, p.s.ID, p.e.ID, p.ips)
	}
	w.Flush()
}
