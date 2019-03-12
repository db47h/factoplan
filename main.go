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
	Factories   []*Factory
}

// A Factory wraps a factorio factory
type Factory struct {
	ID    string
	Speed float64 // items per second
}

// Ingredient represents an ingredient in a crafting recipe.
//
type Ingredient struct {
	Quantity float64
	Recipe   *Recipe
}

var factories = make(map[string]*Factory)
var recipes = make(map[string]*Recipe)

func loadFactories() error {
	type jsonStation struct {
		Speed float64
	}

	var fs map[string]jsonStation

	data, err := ioutil.ReadFile("factories.json")
	if err != nil {
		return errors.Wrap(err, "failed to read factories.json")
	}
	err = json.Unmarshal(data, &fs)
	if err != nil {
		return errors.Wrap(err, "failed to read json data from factories.json")
	}
	for id, s := range fs {
		factories[id] = &Factory{id, s.Speed}
	}
	return nil
}

func loadRecipes() error {
	type jsonRecipe struct {
		Time        float64
		Product     map[string]float64
		Ingredients map[string]float64
		Factories   []string
	}

	var rs []jsonRecipe

	data, err := ioutil.ReadFile("recipes.json")
	if err != nil {
		return errors.Wrap(err, "failed to read recipes.json")
	}
	err = json.Unmarshal(data, &rs)
	if err != nil {
		return errors.Wrap(err, "failed to read json data from recipes.json")
	}

	for _, r := range rs {
		if r.Time <= 0 {
			return errors.Errorf("recipe for items %v: invalid negative or 0 time", r.Product)
		}
		if len(r.Factories) == 0 {
			return errors.Errorf("recipe for items %v: no factories", r.Product)
		}
		fs := make([]*Factory, 0, len(r.Factories))
		for _, fid := range r.Factories {
			f := factories[fid]
			if f == nil {
				return errors.Errorf("recipe for items %v: unknown factory %s", r.Product, fid)
			}
			fs = append(fs, f)
		}
		for id, q := range r.Product {
			recipes[id] = &Recipe{
				ID:        id,
				Time:      r.Time / q,
				Factories: fs,
			}
		}
	}

	// fill in ingredients
	for _, r := range rs {
		for id, pq := range r.Product {
			is := make([]Ingredient, 0, len(r.Ingredients))
			for iid, iq := range r.Ingredients {
				ir := recipes[iid]
				if ir == nil {
					return errors.Errorf("recipe for item %s: unknown ingredient %s", id, iid)
				}
				is = append(is, Ingredient{Quantity: iq / pq, Recipe: ir})
			}
			recipes[id].Ingredients = is
		}
	}

	return nil
}

type Production struct {
	r   *Recipe
	ips float64
	f   *Factory
	fc  float64
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

func NewProduction(name string, ips float64) (ProdList, error) {
	pl := make(ProdList)

	if r := recipes[name]; r != nil {
		pl.add(r, ips)
	} else {
		return nil, errors.Errorf("unknown item %s", name)
	}

	// update production list with appropriate factories
	for _, p := range pl {
		for _, f := range p.r.Factories {
			// pick the slowest compatible factory as long as the number of factories does not change in order to save energy
			c := math.Ceil(p.ips * p.r.Time / f.Speed)
			if p.f == nil || c < p.fc || (c == p.fc && f.Speed < p.f.Speed) {
				p.f = f
				p.fc = c
			}
		}
		if p.f == nil {
			return nil, errors.Errorf("no factory available to produce item %s", p.r.ID)
		}
	}

	return pl, nil
}

func main() {
	var (
		err error

		list = flag.Bool("list", false, "lists known items")
		name = flag.String("i", "", "item `name`")
		ips  = flag.Float64("r", 1.0, "production rate in `items per second`")
	)

	flag.Parse()

	if err = loadFactories(); err != nil {
		panic(err)
	}
	if err = loadRecipes(); err != nil {
		panic(err)
	}

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

	if *ips <= 0.0 {
		fmt.Fprintf(os.Stderr, "error: items per second must be a non-zero positive number")
	}

	*name = strings.Replace(strings.ToLower(*name), " ", "_", -1)

	pl, err := NewProduction(*name, *ips)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var spl []*Production
	for _, p := range pl {
		spl = append(spl, p)
	}
	sort.Slice(spl, func(i, j int) bool {
		return spl[i].f.ID < spl[j].f.ID || (spl[i].f.ID == spl[j].f.ID && spl[i].r.ID < spl[j].r.ID)
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "#\tFactory\tItem\titems/s\titems/m\n")
	fmt.Fprintf(w, "--\t---------------------\t-----------------------\t-------\t-------\n")

	for _, p := range spl {
		fmt.Fprintf(w, "%.0f\t%s\t%s\t%.2f\t%.0f\n", p.fc, p.f.ID, p.r.ID, p.ips, p.ips*60.0)
	}
	w.Flush()
}
