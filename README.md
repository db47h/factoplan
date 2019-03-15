# Factoplan

A simple command-line factory planner for [Factorio].

Given a list of "wanted" items and the rates at which one wants to get them, it produces a list of needed factories (mining drills, assembly machines, etc.) and what they should produce.

## Installation

Prerequisites: [LUA] and [Go].

The list of production recipes and items is not provided with factoplan. This is mostly because I do not want to have to keep this repo in sync with latest Factorio experimental versions. However, the provided LUA importer script will extract the necessary data from your Factorio game files. Once done, you'll just need to recompile factoplan.

In short:

```bash
go get github.com/db47h/factoplan

#if you have no GOPATH set this will be $HOME/go/src/github.com/db47h/factoplan
cd $GOPATH/src/github.com/db47h/factoplan

# replace the second arg with the actual path to your Factorio installation
lua ./importer.lua "$HOME/.local/share/Steam/steamapps/common/Factorio"

go install
```

And you're all set. Note that you might need to do this every time you update Factorio.

## Usage

`factoplan -list` will give you a list of known items.

For example, we want to produce every minute:

- 2.4 solar panels per minute
- 2 accumulators
- 20 piercing rounds magazines
- 4 cannon-shells

Just run:

    factoplan solar-panel:2.4 accumulator:2.0 piercing-rounds-magazine:10 cannon-shell:4

This will produce:

```txt
#  Factory               Item                     items/m items/s
-- --------------------- -----------------------  ------- -------
1  assembling-machine-3  accumulator              2.0     0.03
1  assembling-machine-3  cannon-shell             4.0     0.07
1  assembling-machine-3  copper-cable             108.0   1.80
1  assembling-machine-3  electronic-circuit       36.0    0.60
1  assembling-machine-3  firearm-magazine         10.0    0.17
1  assembling-machine-3  piercing-rounds-magazine 10.0    0.17
1  assembling-machine-3  solar-panel              2.4     0.04
1  chemical-plant        battery                  10.0    0.17
1  chemical-plant        explosives               4.0     0.07
1  chemical-plant        plastic-bar              8.0     0.13
1  chemical-plant        sulfur                   22.0    0.37
1  chemical-plant        sulfuric-acid            200.0   3.33
4  electric-furnace      copper-plate             126.0   2.10
7  electric-furnace      iron-plate               244.0   4.07
4  electric-furnace      steel-plate              30.0    0.50
1  electric-mining-drill coal                     6.0     0.10
5  electric-mining-drill copper-ore               126.0   2.10
9  electric-mining-drill iron-ore                 244.0   4.07
1  oil-refinery          petroleum-gas            410.0   6.83
```

The first column is the needed number of factories needed to produce the item mentioned in the third column at the given rate. Note that the rates in the right hand-side columns are the item rates required to achieve the set goals; they do not necessarily match the given number of factories (which are rounded up). To illustrate this, consider our example where in order to produce the requested rates of solar panels, accumulators and amo, we need to produce 4.07 iron ore per second. However, the 9 mining drills announced are capable of producing 4.5 items per second. This is not a bug, just a missing feature where the displayed items rates should be what will be produced for real.

The item rates can still give a good estimate of the required belt speeds.

## Limitations

For any given production, the selected type of factory (or assembly machine) will be the one that requires the lesser amount of factories to do the job. Energy consumption is not taken into consideration. i.e. for very small production units, you may see one assembly machine 1, never more (since a single assembly machine 3 can replace a pair of 1's), occasionally a pair of assembly machine 2, but that's it.

The current algorithm cannot handle multiple recipes producing the same item (like solid fuel from petroleum gas and from light oil). As a result, the following recipes are removed during the import:

- basic-oil-processing
- coal-liquefaction
- heavy-oil-cracking
- light-oil-cracking
- solid-fuel-from-petroleum-gas
- solid-fuel-from-heavy-oil
- kovarex-enrichment-process
- nuclear-fuel-reprocessing

Facotplan still gives rough production needs for uranium and oil as well as their derivates, but these are not accurate: giving an optimal setup would require a much more complex algorithm. Also oil production can easily be controlled using simple circuit network logic (the difficult part is not to produce enough, but to prevent production stalls due to overflow).

The only recipe considered to produce petrleum gas is advanced-oil-processing and solid fuel from light oil.

## Pseudo items

In order to help with production planning for research, a few pseudo items have been added:

- automation-science
- logistic-science
- military-science
- production-science
- utility-science
- space-science

They are agregates that represent one research point for a tech that requires 30 seconds of research. Each item above is an aggregate of itself added to the previous one. i.e. military-science is automation + logistic + military (that last one is weighted at 0.8, see below).

You can see them as a quick way of checking how many research labs are needed (plus what's needed to keep them busy) for that many research points / minute. For example here is the output of, `factoplan logistic-science:12`:

```txt
#  Factory               Item                    items/m items/s
-- --------------------- ----------------------- ------- -------
1  assembling-machine-3  automation-science-pack 12.0    0.20
1  assembling-machine-3  copper-cable            36.0    0.60
1  assembling-machine-3  electronic-circuit      12.0    0.20
1  assembling-machine-3  inserter                12.0    0.20
1  assembling-machine-3  iron-gear-wheel         30.0    0.50
1  assembling-machine-3  logistic-science-pack   12.0    0.20
1  assembling-machine-3  transport-belt          12.0    0.20
1  electric-furnace      copper-plate            30.0    0.50
3  electric-furnace      iron-plate              90.0    1.50
1  electric-mining-drill copper-ore              30.0    0.50
3  electric-mining-drill iron-ore                90.0    1.50
6  lab                   logistic-science        12.0    0.20
```

For technologies with a 30 seconds cycle, this setup will keep the 6 labs busy. For technologies with a lower cycle, this will not be optimal (i.e. some labs will be idle), but technologies with low cycles are generally cheap. For technologies with a 60 second cycle, this same setup could keep 12 labs busy. So just build more labs!

Again, this is only a helper. The actual "ingredients" for the "sepace-science" item are:

- automation-science-pack: 1
- logistic-science-pack: 1
- military-science-pack: 0.8
- chemical-science-pack: 1
- production-science-pack: 0.4
- utility-science-pack: 1
- space-science-pack: 1

The non-round values for military and production science packs are based on guesstimates for end-game infinite research (where not all techs need these). These certainly need to be refined base on data extracted from the game tech tree. Also see https://wiki.factorio.com/Technologies

The "space-science-pack" is also a pseudo-item added as a by-product of producing rocket parts in the rocket silo. The current value is 10 space-science packs for 1 rocket-part, but this is wrong and needs to be corrected, as well as the whole rocket-part recipe that needs to account for the 40 something seconds where a silo stops production to launch a rocket.

## TODO

- A graphwiz export with item rates would be nice.

[Factorio]: https://www.factorio.com/
[Go]: https://golang.org/
[LUA]: https://www.lua.org/
