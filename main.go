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

// Recipe wraps a factorio recipe
type Recipe struct {
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
	Recipe   *Recipe
}

var stations = make(map[string]*Station)
var recipes = make(map[string]*Recipe)

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

func readRecipes() error {
	type jsonRecipe struct {
		Batch       float64
		Time        float64
		Ingredients map[string]float64
		Stations    []string
	}

	var jr map[string]*jsonRecipe

	data, err := ioutil.ReadFile("recipes.json")
	if err != nil {
		return errors.Wrap(err, "failed to read recipes.json")
	}
	err = json.Unmarshal(data, &jr)
	if err != nil {
		return errors.Wrap(err, "failed to read json data from recipes.json")
	}

	for id, r := range jr {
		if r.Batch == 0 {
			r.Batch = 1
		}
		if r.Time <= 0 {
			return errors.Errorf("recipe for item %s: invalid negative or 0 time", id)
		}
		if len(r.Stations) == 0 {
			return errors.Errorf("recipe for item %s: no crafting stations", id)
		}
		st := make([]*Station, 0, len(r.Stations))
		for _, sid := range r.Stations {
			s := stations[sid]
			if s == nil {
				return errors.Errorf("recipe for item %s: unknown station %s", id, sid)
			}
			st = append(st, s)
		}
		recipes[id] = &Recipe{
			ID:       id,
			Time:     r.Time / r.Batch,
			Stations: st,
		}
	}

	// fill in ingredients
	for id, r := range jr {
		is := make([]Ingredient, 0, len(r.Ingredients))
		for iid, q := range r.Ingredients {
			ir := recipes[iid]
			if ir == nil {
				return errors.Errorf("recipe for item %s: unknown ingredient %s", id, iid)
			}
			is = append(is, Ingredient{Quantity: q / r.Batch, Recipe: ir})
		}
		recipes[id].Ingredients = is
	}

	return nil
}

type Production struct {
	r   *Recipe
	ips float64
	s   *Station
	sc  float64
}

type ProdList map[string]*Production

func (pl ProdList) add(r *Recipe, ips float64) {
	p := pl[r.ID]
	if p == nil {
		p = &Production{r: r}
		pl[r.ID] = p
	}
	p.ips += ips

	for _, i := range r.Ingredients {
		pl.add(i.Recipe, ips*i.Quantity)
	}
}

func NewProduction(name string, ips float64, techLevel int) (ProdList, error) {
	pl := make(ProdList)

	if r := recipes[name]; r != nil {
		pl.add(r, ips)
	} else {
		return nil, errors.Errorf("unknown item %s", name)
	}

	// update production list with appropriate stations
	for _, p := range pl {
		for _, s := range p.r.Stations {
			if s.Level > techLevel {
				continue
			}
			c := math.Ceil(p.ips * p.r.Time / s.Speed)
			if p.s == nil || c < p.sc || (s.Level < p.s.Level && c == p.sc) {
				p.s = s
				p.sc = c
			}
		}
		if p.s == nil {
			return nil, errors.Errorf("no station available to produce item %s at tech level %d", p.r.ID, techLevel)
		}
	}

	return pl, nil
}

func main() {
	var err error
	if err = readStations(); err != nil {
		panic(err)
	}
	if err = readRecipes(); err != nil {
		panic(err)
	}

	var (
		list = flag.Bool("list", false, "lists known items")
		name = flag.String("i", "", "item `name`")
		ips  = flag.Float64("r", 1.0, "production rate in `items per second`")
		l    = flag.Int("l", 3, "maximum assembly machine `level`")
	)

	flag.Parse()

	if *list {
		fmt.Println("Known recipes:")
		var el []string
		for k := range recipes {
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
		return spl[i].s.ID < spl[j].s.ID || (spl[i].s.ID == spl[j].s.ID && spl[i].r.ID < spl[j].r.ID)
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "#\tFactory\tItem\titems/s\n")
	fmt.Fprintf(w, "--\t-------------------------\t-------------------------\t-------\n")

	for _, p := range spl {
		fmt.Fprintf(w, "%.0f\t%s\t%s\t%.2f\n", p.sc, p.s.ID, p.r.ID, p.ips)
	}
	w.Flush()
}
