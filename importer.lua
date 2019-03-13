--[[
Copyright 2019 Denis Bernard <db047h@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
--]]

factorio_path = arg[1]
package.path = factorio_path .. "/data/core/lualib/?.lua;" .. factorio_path .. "/data/base/?.lua;" .. package.path

-- define a few things needed to import data
defines = {
    difficulty_settings = {
        recipe_difficulty = {
            normal = 0
        },
        technology_difficulty = {
            normal = 0
        }
    },
    direction = {
        north = 0,
        east = 2,
        south = 4,
        west = 6
    }
}
data = {
    raw = {
        ["gui-style"] = {default = {}}
    }
}
function data:extend(a)
    for i, v in ipairs(a) do
        if data.raw[v.type] == nil then
            data.raw[v.type] = {}
        end
        data.raw[v.type][v.name] = v
    end
end

require("data")

-- hard-coded pseudo factories and recipes

data.raw["recipe"]["coal"] = {
    type = "recipe",
    name = "coal",
    energy_required = 1,
    category = "basic-solid",
    results = {{"coal", 1}},
}
data.raw["recipe"]["copper-ore"] = {
    type = "recipe",
    name = "copper-ore",
    energy_required = 1,
    category = "basic-solid",
    results = {{"copper-ore", 1}},
}
data.raw["recipe"]["iron-ore"] = {
    type = "recipe",
    name = "iron-ore",
    energy_required = 1,
    category = "basic-solid",
    results = {{"iron-ore", 1}},
}
data.raw["recipe"]["stone"] = {
    type = "recipe",
    name = "stone",
    energy_required = 1,
    category = "basic-solid",
    results = {{"stone", 1}},
}

-- TODO: pseudo-items for research have weighted amounts for military and production science packs
-- based on guesstimates for end-game infinit research.
-- These certainly need to be refined base on data extracted from the game tech tree. Also see
-- https://wiki.factorio.com/Technologies
data.raw["assembling-machine"]["lab"] = {
    type = "lab",
    name = "lab",
    crafting_speed = data.raw["lab"]["lab"]["researching_speed"],
    crafting_categories = { "research" },
}
data.raw["recipe"]["automation-science"] = {
    type = "recipe",
    name = "automation-science",
    energy_required = 30,
    category = "research",
    ingredients = {
        {"automation-science-pack", 1},
    },
    results = {{"automation-science", 1}},
}
data.raw["recipe"]["logistic-science"] = {
    type = "recipe",
    name = "logistic-science",
    energy_required = 30,
    category = "research",
    ingredients = {
        {"automation-science-pack", 1},
        {"logistic-science-pack", 1},
    },
    results = {{"logistic-science", 1}},
}
data.raw["recipe"]["military-science"] = {
    type = "recipe",
    name = "military-science",
    energy_required = 30,
    category = "research",
    ingredients = {
        {"automation-science-pack", 1},
        {"logistic-science-pack", 1},
        {"military-science-pack", 0.8},
    },
    results = {{"military-science", 1}},
}
data.raw["recipe"]["chemical-science"] = {
    type = "recipe",
    name = "chemical-science",
    energy_required = 30,
    category = "research",
    ingredients = {
        {"automation-science-pack", 1},
        {"logistic-science-pack", 1},
        {"military-science-pack", 0.8},
        {"chemical-science-pack", 1},
    },
    results = {{"chemical-science", 1}},
}
data.raw["recipe"]["production-science"] = {
    type = "recipe",
    name = "production-science",
    energy_required = 30,
    category = "research",
    ingredients = {
        {"automation-science-pack", 1},
        {"logistic-science-pack", 1},
        {"military-science-pack", 0.8},
        {"chemical-science-pack", 1},
        {"production-science-pack", 0.4},
    },
    results = {{"production-science", 1}},
}
data.raw["recipe"]["utility-science"] = {
    type = "recipe",
    name = "utility-science",
    energy_required = 30,
    category = "research",
    ingredients = {
        {"automation-science-pack", 1},
        {"logistic-science-pack", 1},
        {"military-science-pack", 0.8},
        {"chemical-science-pack", 1},
        {"production-science-pack", 0.4},
        {"utility-science-pack", 1},
    },
    results = {{"utility-science", 1}},
}
data.raw["recipe"]["space-science"] = {
    type = "recipe",
    name = "space-science",
    energy_required = 30,
    category = "research",
    ingredients = {
        {"automation-science-pack", 1},
        {"logistic-science-pack", 1},
        {"military-science-pack", 0.8},
        {"chemical-science-pack", 1},
        {"production-science-pack", 0.4},
        {"utility-science-pack", 1},
        {"space-science-pack", 1},
    },
    results = {{"space-science", 1}},
}

-- hack for space science packs
-- TODO: find a better way for this. See https://wiki.factorio.com/Rocket_silo
if data.raw["recipe"]["rocket-part"]["result"] ~= nil then
    r =  data.raw["recipe"]["rocket-part"]
    r.results = {{name = r.result, amount = r.result_count or 1}}
    data.raw["recipe"]["rocket-part"]["result"] = nil
end
table.insert(data.raw["recipe"]["rocket-part"]["results"], 
    {"space-science-pack", 10})



-- remove escape-pod-assembler

data.raw["assembling-machine"]["escape-pod-assembler"] = nil

-- remove recipes that generate the same products
-- TODO: handle this and return an optimized prodution plan

data.raw["recipe"]["basic-oil-processing"] = nil
data.raw["recipe"]["coal-liquefaction"] = nil
data.raw["recipe"]["heavy-oil-cracking"] = nil
data.raw["recipe"]["light-oil-cracking"] = nil
data.raw["recipe"]["solid-fuel-from-petroleum-gas"] = nil
data.raw["recipe"]["solid-fuel-from-heavy-oil"] = nil
data.raw["recipe"]["kovarex-enrichment-process"] = nil
data.raw["recipe"]["nuclear-fuel-reprocessing"] = nil

-- import done. GO!

function writeFactories(types)
    io.write("var factories = []Factory{\n")
    for type, props in pairs(types) do
        for mn, m in pairs(data.raw[type]) do 
            io.write('\t{ID: "', mn, '\", Speed: ', m[props.speed], ', Categories: []string{')
            first = true
            for i, cat in ipairs(m[props.categories]) do
                if not first then 
                    io.write(', ')
                else
                    first = false
                end
                io.write('"', cat, '"')
            end
            io.write('}},\n')
        end
    end
    io.write('}\n\n')
end

io.output("data.go")

-- header

io.write([[
// generated by factorio importer.lua; DO NOT EDIT

package main

]])

-- factories 

writeFactories({
        ["mining-drill"] = {speed = "mining_speed", categories = "resource_categories"},
        ["furnace"] = {speed = "crafting_speed", categories = "crafting_categories"},
        ["assembling-machine"] = {speed = "crafting_speed", categories = "crafting_categories"},
        ["rocket-silo"] = {speed = "crafting_speed", categories = "crafting_categories"},
    })

-- for rn, r in pairs(data.raw["resource"]) do 
--     print(rn)
--     -- for k, v in pairs(r.flags) do 
--     --     print("\t", k, v)
--     -- end
--     for k, v in pairs(r.minable) do 
--         if k == "results" then 
--             print ("\tresults")
--             for res, resv in pairs(v[1]) do
--                 print("\t\t", res, resv)
--             end
--         else
--             print("\t", k, v)
--         end
--     end     
-- end

io.write('var recipes = []Recipe{\n')

for name, r in pairs(data.raw["recipe"]) do
    if r.normal then
        -- overlay normal cost parameters over root recipe
        for k, v in pairs(r.normal) do
            r[k] = v
        end
    end
    if r.result ~= nil then
        if r.results ~= nil then
            print("result conflict for " .. name)
        end
        r.results = { {r.result, r.result_count or 1} }
    end
    io.write('\t{\n\t\tID: "', name, '",\n')
    io.write('\t\tTime: ', r.energy_required or 0.5, ',\n')
    io.write('\t\tCategory: "', r.category or "basic-crafting", '",\n')
    
    if r.ingredients ~= nil then
        io.write('\t\tIngredients: []Item{\n')
        for i, ig in ipairs(r.ingredients) do
            io.write('\t\t\t{Name: "', ig.name or ig[1],
                '", Amount: ', ig.amount or ig[2], '},\n')
        end
        io.write('\t\t},\n')
    end

    io.write('\t\tResults: []Item{\n')
    for i, res in ipairs(r.results) do
        io.write('\t\t\t{Name: "', res.name or res[1],
            '", Amount: ', (res.amount or res[2]) * (res.probability or 1), '},\n')
    end
    io.write('\t\t},\n')
    
    io.write('\t},\n')
end

io.write('}\n')

io.close()
