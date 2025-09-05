package iters

import (
	"cmp"
	"fmt"
	"slices"
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
)

func ExampleOrderedByKey() {
	// From a corgi training manual
	trickRewards := map[string]struct {
		Treats     int
		Difficulty string
	}{
		"Roll over on command":      {Treats: 3, Difficulty: "Medium"},
		"Sit down on command":       {Treats: 1, Difficulty: "Easy"},
		"Retrieve thrown item":      {Treats: 2, Difficulty: "Easy"},
		"Bark on command":           {Treats: 3, Difficulty: "Medium"},
		"Give a high five with paw": {Treats: 4, Difficulty: "Hard"},
		"Stand and twirl around":    {Treats: 5, Difficulty: "Expert"},
		"Balance treat on nose":     {Treats: 6, Difficulty: "Expert"},
	}

	fmt.Println("Tricks and their rewards:")
	for trick, reward := range OrderedByKey(trickRewards) {
		fmt.Printf("%-26s %d treats (%s)\n", trick+":", reward.Treats, reward.Difficulty)
	}

	// Output:
	// Tricks and their rewards:
	// Balance treat on nose:     6 treats (Expert)
	// Bark on command:           3 treats (Medium)
	// Give a high five with paw: 4 treats (Hard)
	// Retrieve thrown item:      2 treats (Easy)
	// Roll over on command:      3 treats (Medium)
	// Sit down on command:       1 treats (Easy)
	// Stand and twirl around:    5 treats (Expert)
}

func ExampleOrderedByValue() {
	olympicLongJump := map[string]float64{
		"Waffles": 72.4,
		"Biscuit": 83.1,
		"Toast":   64.0,
		"Muffin":  83.2,
		"Pancake": 76.5,
	}

	fmt.Println("Corgi jumping records (shortest to longest):")
	for name, distance := range OrderedByValue(olympicLongJump) {
		fmt.Printf("%s: %.1f cm\n", name, distance)
	}

	// Output:
	// Corgi jumping records (shortest to longest):
	// Toast: 64.0 cm
	// Waffles: 72.4 cm
	// Pancake: 76.5 cm
	// Biscuit: 83.1 cm
	// Muffin: 83.2 cm
}

func ExampleMapOrderedBy() {
	type form struct {
		Age        int
		Fluffiness string
	}
	formAge := func(f form) int { return f.Age }

	daycareRegistrations := map[string]form{
		"Biscuit": {Age: 3, Fluffiness: "Very"},
		"Waffles": {Age: 5, Fluffiness: "Extremely"},
		"Pancake": {Age: 2, Fluffiness: "Somewhat"},
		"Muffin":  {Age: 1, Fluffiness: "Moderately"},
	}

	fmt.Println("Corgis ordered by age:")
	for name, info := range MapOrderedBy(daycareRegistrations, formAge) {
		fmt.Printf("%-8s %d years old, %s fluffy\n", name+":", info.Age, info.Fluffiness)
	}

	// Output:
	// Corgis ordered by age:
	// Muffin:  1 years old, Moderately fluffy
	// Pancake: 2 years old, Somewhat fluffy
	// Biscuit: 3 years old, Very fluffy
	// Waffles: 5 years old, Extremely fluffy
}

func ExampleOrdered() {
	// Veterinarian's height measurements
	corgiHeights := []float64{26.7, 24.9, 28.4, 26.7, 22.1}

	fmt.Println("Corgi heights from shortest to tallest:")
	for i, height := range Ordered(corgiHeights) {
		fmt.Printf("#%d: %.1f cm\n", i+1, height)
	}

	// Output:
	// Corgi heights from shortest to tallest:
	// #1: 22.1 cm
	// #2: 24.9 cm
	// #3: 26.7 cm
	// #4: 26.7 cm
	// #5: 28.4 cm
}

func ExampleOrderedBy() {
	type corgi struct {
		Name   string
		Weight float64
	}

	dietPlan := []corgi{
		{"Biscuit", 10.2},
		{"Waffles", 9.0},
		{"Muffin", 11.5},
		{"Toast", 9.6},
	}

	fmt.Println("Corgis ordered by weight:")
	for idx, c := range OrderedBy(dietPlan, func(c corgi) float64 { return c.Weight }) {
		fmt.Printf("#%d: %s (%.1f kg)\n", idx+1, c.Name, c.Weight)
	}

	// Output:
	// Corgis ordered by weight:
	// #1: Waffles (9.0 kg)
	// #2: Toast (9.6 kg)
	// #3: Biscuit (10.2 kg)
	// #4: Muffin (11.5 kg)
}

func ExampleUnique() {
	// Pancake's talent show repertoire
	talentShow := []string{"sit", "roll", "fetch", "sit", "paw", "roll", "speak"}

	fmt.Println("Pancake's tricks:")
	for i, trick := range Unique(talentShow) {
		fmt.Printf("%dmin: %s\n", i+1, trick)
	}

	// Output:
	// Pancake's tricks:
	// 1min: sit
	// 2min: roll
	// 3min: fetch
	// 5min: paw
	// 7min: speak
}

func ExampleUniqueUsing() {
	// Muffin's toy box inventory
	type toy struct {
		Name     string
		Category string
	}
	toyCategory := func(t toy) string { return t.Category }

	toys := []toy{
		{"Squeaky Bone", "Squeak"},
		{"Tennis Ball", "Ball"},
		{"Rubber Ball", "Ball"},
		{"Plush Squirrel", "Plush"},
		{"Squeaky Duck", "Squeak"},
	}

	fmt.Println("Toy categories:")
	for _, toy := range UniqueUsing(toys, toyCategory) {
		fmt.Printf("%s, e.g. %s\n", toy.Category, toy.Name)
	}

	// Output:
	// Toy categories:
	// Squeak, e.g. Squeaky Bone
	// Ball, e.g. Tennis Ball
	// Plush, e.g. Plush Squirrel
}

func ExampleGroupBy() {
	type business struct {
		Name        string
		Owner       string
		Category    string
		Description string
		Location    string
	}

	yellowPages := []business{
		{"Woof & Wag Grooming", "Biscuit", "Services", "Premium fur styling", "By the Big Oak Tree"},
		{"Bark Bites Bakery", "Waffle", "Food", "Fresh-baked treats daily", "Near the Red Barn"},
		{"Paws & Relax Spa", "Pancake", "Services", "Pawdicures and massages", "By the Big Oak Tree"},
		{"Fetch Toy Emporium", "Toast", "Shopping", "All the best toys", "Next to the Duck Pond"},
		{"Corgi Cardio", "Muffin", "Fitness", "Exercise classes for short legs", "Near the Red Barn"},
		{"Sploot School", "Croissant", "Education", "Learn to sploot properly", "Next to the Duck Pond"},
		{"Fluffy Tails Daycare", "Bagel", "Services", "Playtime while you're away", "By the Big Oak Tree"},
		{"Treat Treasures", "Scone", "Food", "Artisanal dog biscuits", "On top of Treat Hill"},
	}

	fmt.Println("🔍 CORGI YELLOW PAGES - BUSINESS DIRECTORY 🔍")
	fmt.Println("============================================")

	for category, businesses := range GroupBy(yellowPages, func(b business) string { return b.Category }) {
		fmt.Printf("📋 %s:\n", category)
		for _, business := range businesses {
			fmt.Printf("  🏠 %s's %s\n", business.Owner, business.Name)
			fmt.Printf("     📝 %s\n", business.Description)
			fmt.Printf("     📍 %s\n\n", business.Location)
		}
	}

	// Output:
	// 🔍 CORGI YELLOW PAGES - BUSINESS DIRECTORY 🔍
	// ============================================
	// 📋 Services:
	//   🏠 Biscuit's Woof & Wag Grooming
	//      📝 Premium fur styling
	//      📍 By the Big Oak Tree
	//
	//   🏠 Pancake's Paws & Relax Spa
	//      📝 Pawdicures and massages
	//      📍 By the Big Oak Tree
	//
	//   🏠 Bagel's Fluffy Tails Daycare
	//      📝 Playtime while you're away
	//      📍 By the Big Oak Tree
	//
	// 📋 Food:
	//   🏠 Waffle's Bark Bites Bakery
	//      📝 Fresh-baked treats daily
	//      📍 Near the Red Barn
	//
	//   🏠 Scone's Treat Treasures
	//      📝 Artisanal dog biscuits
	//      📍 On top of Treat Hill
	//
	// 📋 Shopping:
	//   🏠 Toast's Fetch Toy Emporium
	//      📝 All the best toys
	//      📍 Next to the Duck Pond
	//
	// 📋 Fitness:
	//   🏠 Muffin's Corgi Cardio
	//      📝 Exercise classes for short legs
	//      📍 Near the Red Barn
	//
	// 📋 Education:
	//   🏠 Croissant's Sploot School
	//      📝 Learn to sploot properly
	//      📍 Next to the Duck Pond
}

func ExampleGroupItemsBy() {
	type corgi struct {
		Name  string
		Color string
	}
	corgiColor := func(c corgi) string { return c.Color }

	catalog := []corgi{
		{"Waffles", "Red"},
		{"Biscuit", "Red"},
		{"Toast", "Tri-color"},
		{"Muffin", "Red"},
		{"Pancake", "Tri-color"},
	}

	fmt.Println("Corgis grouped by color:")
	for color, colorGroup := range GroupItemsBy(catalog, corgiColor) {
		fmt.Printf("%s corgis:\n", color)

		// Further transform the group without affecting catalog
		slices.SortFunc(colorGroup, func(a, b corgi) int {
			return cmp.Compare(a.Name, b.Name)
		})
		for _, c := range colorGroup {
			fmt.Printf("  - %s\n", c.Name)
		}
	}

	// Output:
	// Corgis grouped by color:
	// Red corgis:
	//   - Biscuit
	//   - Muffin
	//   - Waffles
	// Tri-color corgis:
	//   - Pancake
	//   - Toast
}

func ExampleGroupAdjacentBy() {
	type activity struct {
		Name     string
		Type     string
		Duration int // minutes
	}

	activityType := func(a activity) string { return a.Type }

	// Fetched pre-sorted from the 'Barktivity Tracker' REST API
	schedule := []activity{
		{"Morning Walk", "Exercise", 20},
		{"Fetch", "Exercise", 15},
		{"Training", "Learning", 10},
		{"Tricks", "Learning", 10},
		{"Commands", "Learning", 5},
		{"Afternoon Nap", "Rest", 30},
		{"Evening Sleep", "Rest", 60},
	}

	fmt.Println("Toast's daily schedule:")
	for typ, activities := range GroupAdjacentBy(schedule, activityType) {
		fmt.Printf("%s activities:\n", typ)
		var totalTime int
		for _, activity := range activities {
			fmt.Printf("  %s (%d min)\n", activity.Name, activity.Duration)
			totalTime += activity.Duration
		}
		fmt.Printf("Total %s Time: %d minutes\n\n", typ, totalTime)
	}

	// Output:
	// Toast's daily schedule:
	// Exercise activities:
	//   Morning Walk (20 min)
	//   Fetch (15 min)
	// Total Exercise Time: 35 minutes
	//
	// Learning activities:
	//   Training (10 min)
	//   Tricks (10 min)
	//   Commands (5 min)
	// Total Learning Time: 25 minutes
	//
	// Rest activities:
	//   Afternoon Nap (30 min)
	//   Evening Sleep (60 min)
	// Total Rest Time: 90 minutes
}

type mapEntry[K comparable, V any] struct {
	Key   K
	Value V
}

func TestOrderedByKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]int
		want []mapEntry[string, int]
	}{
		{
			name: "nil map",
			in:   nil,
			want: []mapEntry[string, int]{},
		}, {
			name: "single key-value",
			in:   map[string]int{"woof": 1},
			want: []mapEntry[string, int]{{"woof", 1}},
		}, {
			name: "multiple key-values",
			in: map[string]int{
				"woof": 1,
				"bark": 2,
				"arf":  3,
			},
			want: []mapEntry[string, int]{
				{"arf", 3},
				{"bark", 2},
				{"woof", 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := make([]mapEntry[string, int], 0, len(c.want))
			for k, v := range OrderedByKey(c.in) {
				got = append(got, mapEntry[string, int]{k, v})
			}

			should.Equal(t, got, c.want)
		})
	}
}

func TestOrderedByValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]int
		want []mapEntry[string, int]
	}{
		{
			name: "nil map",
			in:   nil,
			want: []mapEntry[string, int]{},
		}, {
			name: "single key-value",
			in:   map[string]int{"woof": 1},
			want: []mapEntry[string, int]{{"woof", 1}},
		}, {
			name: "multiple key-values with unique values",
			in: map[string]int{
				"woof": 3,
				"bark": 2,
				"arf":  1,
			},
			want: []mapEntry[string, int]{
				{"arf", 1},
				{"bark", 2},
				{"woof", 3},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := make([]mapEntry[string, int], 0, len(c.want))
			for k, v := range OrderedByValue(c.in) {
				got = append(got, mapEntry[string, int]{k, v})
			}

			should.Equal(t, got, c.want)
		})
	}

	t.Run("multiple key-values with duplicate values", func(t *testing.T) {
		t.Parallel()

		input := map[string]int{
			"woof": 1,
			"bark": 2,
			"arf":  1,
		}

		got := make([]mapEntry[string, int], 0, len(input))
		for k, v := range OrderedByValue(input) {
			got = append(got, mapEntry[string, int]{k, v})
		}

		should.Equal(t, len(got), 3)

		// Check that values are in ascending order
		lastValue := -1
		for _, pair := range got {
			value := pair.Value
			should.True(t, value >= lastValue) // values not in ascending order
			lastValue = value
		}
	})
}

func TestMapOrderedBy(t *testing.T) {
	t.Parallel()

	type corgi struct {
		Name  string
		Age   int
		Fluff int
	}

	tests := []struct {
		name   string
		in     map[string]corgi
		mapper func(corgi) int
		want   []string // Expected order of keys
	}{
		{
			name:   "nil map",
			in:     nil,
			mapper: func(c corgi) int { return c.Age },
			want:   []string{},
		}, {
			name: "order by age",
			in: map[string]corgi{
				"Biscuit": {Name: "Biscuit", Age: 3, Fluff: 8},
				"Waffles": {Name: "Waffles", Age: 5, Fluff: 10},
				"Pancake": {Name: "Pancake", Age: 2, Fluff: 7},
			},
			mapper: func(c corgi) int { return c.Age },
			want:   []string{"Pancake", "Biscuit", "Waffles"},
		}, {
			name: "order by fluffiness",
			in: map[string]corgi{
				"Biscuit": {Name: "Biscuit", Age: 3, Fluff: 8},
				"Waffles": {Name: "Waffles", Age: 5, Fluff: 10},
				"Pancake": {Name: "Pancake", Age: 2, Fluff: 7},
			},
			mapper: func(c corgi) int { return c.Fluff },
			want:   []string{"Pancake", "Biscuit", "Waffles"},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			keys := make([]string, 0, len(c.want))
			for k := range MapOrderedBy(c.in, c.mapper) {
				keys = append(keys, k)
			}

			should.Equal(t, keys, c.want)
		})
	}
}

func TestOrdered(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []float64
		want []float64
	}{
		{
			name: "empty",
			in:   []float64{},
			want: []float64{},
		}, {
			name: "already sorted",
			in:   []float64{1.0, 2.0, 3.0, 4.0},
			want: []float64{1.0, 2.0, 3.0, 4.0},
		}, {
			name: "reverse sorted",
			in:   []float64{4.0, 3.0, 2.0, 1.0},
			want: []float64{1.0, 2.0, 3.0, 4.0},
		}, {
			name: "mixed order",
			in:   []float64{3.0, 1.0, 4.0, 2.0},
			want: []float64{1.0, 2.0, 3.0, 4.0},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			orig := make([]float64, len(c.in))
			copy(orig, c.in)

			got := make([]float64, 0, len(c.want))
			var wantI int
			for gotI, gotV := range Ordered(c.in) {
				got = append(got, gotV)
				should.Equal(t, gotI, wantI) // index
				wantI++
			}

			should.Equal(t, got, c.want)
			should.Equal(t, c.in, orig) // ensure original slice wasn't modified
		})
	}
}

func TestOrderedBy(t *testing.T) {
	t.Parallel()

	type corgi struct {
		Name   string
		Weight float64
	}

	tests := []struct {
		name string
		in   []corgi
		want []string // want order of names
	}{
		{
			name: "empty",
			in:   []corgi{},
			want: []string{},
		}, {
			name: "order by weight",
			in: []corgi{
				{"Biscuit", 10.2},
				{"Waffles", 9.0},
				{"Muffin", 11.5},
				{"Toast", 9.6},
			},
			want: []string{"Waffles", "Toast", "Biscuit", "Muffin"},
		}, {
			name: "with duplicate weights (stability)",
			in: []corgi{
				{"Biscuit", 10.2},
				{"Waffles", 9.0},
				{"Muffin", 10.2},
				{"Toast", 9.0},
			},
			want: []string{"Waffles", "Toast", "Biscuit", "Muffin"}, // Order should be stable
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			orig := make([]corgi, len(c.in))
			copy(orig, c.in)

			got := make([]string, 0, len(c.want))
			var wantI int
			for i, v := range OrderedBy(c.in, func(c corgi) float64 { return c.Weight }) {
				got = append(got, v.Name)
				should.Equal(t, i, wantI) // index
				wantI++
			}

			should.Equal(t, got, c.want)
			if should.Equal(t, c.in, orig) { // ensure original slice wasn't modified
				checked := make(map[float64]bool)
				for _, oc := range orig {
					for _, ic := range c.in {
						if oc.Weight == ic.Weight && !checked[ic.Weight] {
							should.Equal(t, oc, ic) // should be stable
							checked[ic.Weight] = true
						}
					}
				}
			}
		})
	}
}

func TestUnique(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want []mapEntry[int, string]
	}{
		{
			name: "empty",
			in:   []string{},
			want: []mapEntry[int, string]{},
		}, {
			name: "no duplicates",
			in:   []string{"woof", "bark", "arf"},
			want: []mapEntry[int, string]{{0, "woof"}, {1, "bark"}, {2, "arf"}},
		}, {
			name: "with duplicates",
			in:   []string{"woof", "bark", "woof", "arf", "bark"},
			want: []mapEntry[int, string]{{0, "woof"}, {1, "bark"}, {3, "arf"}},
		}, {
			name: "all duplicates",
			in:   []string{"woof", "woof", "woof"},
			want: []mapEntry[int, string]{{0, "woof"}},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			orig := make([]string, len(c.in))
			copy(orig, c.in)

			got := make([]mapEntry[int, string], 0, len(c.want))
			for i, v := range Unique(c.in) {
				got = append(got, mapEntry[int, string]{i, v})
			}

			should.Equal(t, got, c.want)
			should.Equal(t, c.in, orig) // ensure original slice wasn't modified
		})
	}
}

func TestUniqueUsing(t *testing.T) {
	t.Parallel()

	type corgiToy struct {
		Name     string
		Category string
	}

	tests := []struct {
		name string
		in   []corgiToy
		want []mapEntry[int, string]
	}{
		{
			name: "empty",
			in:   []corgiToy{},
			want: []mapEntry[int, string]{},
		}, {
			name: "no duplicates",
			in: []corgiToy{
				{"Squeaky Bone", "Squeak"},
				{"Tennis Ball", "Ball"},
				{"Plush Squirrel", "Plush"},
			},
			want: []mapEntry[int, string]{{0, "Squeaky Bone"}, {1, "Tennis Ball"}, {2, "Plush Squirrel"}},
		}, {
			name: "with duplicates",
			in: []corgiToy{
				{"Squeaky Bone", "Squeak"},
				{"Tennis Ball", "Ball"},
				{"Rubber Ball", "Ball"},
				{"Plush Squirrel", "Plush"},
				{"Squeaky Duck", "Squeak"},
			},
			want: []mapEntry[int, string]{{0, "Squeaky Bone"}, {1, "Tennis Ball"}, {3, "Plush Squirrel"}},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			orig := make([]corgiToy, len(c.in))
			copy(orig, c.in)

			got := make([]mapEntry[int, string], 0, len(c.want))
			for i, v := range UniqueUsing(c.in, func(t corgiToy) string { return t.Category }) {
				got = append(got, mapEntry[int, string]{i, v.Name})
			}

			should.Equal(t, got, c.want)
			should.Equal(t, c.in, orig) // ensure original slice wasn't modified
		})
	}
}

func TestGroupBy(t *testing.T) {
	t.Parallel()

	type business struct {
		Name     string
		Category string
	}

	yellowPages := []business{
		{"Woof & Wag Grooming", "Services"},
		{"Bark Bites Bakery", "Food"},
		{"Paws & Relax Spa", "Services"},
		{"Fetch Toy Emporium", "Shopping"},
		{"Corgi Cardio", "Fitness"},
		{"Sploot School", "Education"},
		{"Fluffy Tails Daycare", "Services"},
		{"Treat Treasures", "Food"},
	}

	tests := []struct {
		name       string
		in         []business
		wantGroups map[string][]string // map[category][]name
		wantOrder  []string            // expected order of categories
	}{
		{
			name:       "empty",
			in:         []business{},
			wantGroups: map[string][]string{},
			wantOrder:  []string{},
		}, {
			name: "single category",
			in: []business{
				{"Woof & Wag Grooming", "Services"},
				{"Paws & Relax Spa", "Services"},
			},
			wantGroups: map[string][]string{
				"Services": {"Woof & Wag Grooming", "Paws & Relax Spa"},
			},
			wantOrder: []string{"Services"},
		}, {
			name: "groups correctly by category",
			in:   yellowPages,
			wantGroups: map[string][]string{
				"Services":  {"Woof & Wag Grooming", "Paws & Relax Spa", "Fluffy Tails Daycare"},
				"Food":      {"Bark Bites Bakery", "Treat Treasures"},
				"Shopping":  {"Fetch Toy Emporium"},
				"Fitness":   {"Corgi Cardio"},
				"Education": {"Sploot School"},
			},
			wantOrder: []string{"Services", "Food", "Shopping", "Fitness", "Education"},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			orig := make([]business, len(c.in))
			copy(orig, c.in)

			gotGroups := make(map[string][]string, len(c.wantGroups))
			gotOrder := make([]string, 0, len(c.wantOrder))
			for category, businesses := range GroupBy(c.in, func(b business) string { return b.Category }) {
				gotOrder = append(gotOrder, category)

				gotNames := make([]string, 0, len(c.wantGroups[category]))
				for _, b := range businesses {
					gotNames = append(gotNames, b.Name)
				}
				gotGroups[category] = gotNames
			}

			should.Equal(t, gotGroups, c.wantGroups)
			should.Equal(t, gotOrder, c.wantOrder)
			should.Equal(t, c.in, orig) // ensure original slice wasn't modified
		})
	}
}

func TestGroupItemsBy(t *testing.T) {
	t.Parallel()

	type corgi struct {
		Name  string
		Color string
	}

	corgis := []corgi{
		{"Waffles", "Red"},
		{"Biscuit", "Red"},
		{"Toast", "Tri-color"},
		{"Muffin", "Red"},
		{"Pancake", "Tri-color"},
	}

	tests := []struct {
		name       string
		in         []corgi
		wantGroups map[string][]string // map[color][]name
	}{
		{
			name:       "empty",
			in:         []corgi{},
			wantGroups: map[string][]string{},
		}, {
			name: "single color",
			in: []corgi{
				{"Waffles", "Red"},
				{"Biscuit", "Red"},
			},
			wantGroups: map[string][]string{
				"Red": {"Waffles", "Biscuit"},
			},
		}, {
			name: "groups correctly by color",
			in:   corgis,
			wantGroups: map[string][]string{
				"Red":       {"Waffles", "Biscuit", "Muffin"},
				"Tri-color": {"Toast", "Pancake"},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			orig := make([]corgi, len(c.in))
			copy(orig, c.in)

			gotGroups := make(map[string][]string, len(c.wantGroups))
			for color, colorGroup := range GroupItemsBy(c.in, func(c corgi) string { return c.Color }) {
				gotNames := make([]string, 0, len(c.wantGroups[color]))
				for _, c := range colorGroup {
					gotNames = append(gotNames, c.Name)
				}
				gotGroups[color] = gotNames

				colorGroup[0] = colorGroup[1] // modify slice to test if it's a copy
			}

			should.Equal(t, gotGroups, c.wantGroups)
			should.Equal(t, c.in, orig) // ensure original slice wasn't modified
		})
	}
}

func TestGroupAdjacentBy(t *testing.T) {
	t.Parallel()

	type activity struct {
		Name     string
		Type     string
		Duration int
	}

	schedule := []activity{
		{"Morning Walk", "Exercise", 20},
		{"Fetch", "Exercise", 15},
		{"Training", "Learning", 10},
		{"Tricks", "Learning", 10},
		{"Commands", "Learning", 5},
		{"Afternoon Nap", "Rest", 30},
		{"Evening Sleep", "Rest", 60},
	}

	tests := []struct {
		name          string
		in            []activity
		wantGroups    map[string][]string // map[type]name
		wantDurations map[string]int      // map[type]totalDuration
	}{
		{
			name:          "empty",
			in:            []activity{},
			wantGroups:    map[string][]string{},
			wantDurations: map[string]int{},
		}, {
			name: "single activity type",
			in: []activity{
				{"Morning Walk", "Exercise", 20},
				{"Fetch", "Exercise", 15},
			},
			wantGroups: map[string][]string{
				"Exercise": {"Morning Walk", "Fetch"},
			},
			wantDurations: map[string]int{
				"Exercise": 35,
			},
		}, {
			name: "groups adjacent activities",
			in:   schedule,
			wantGroups: map[string][]string{
				"Exercise": {"Morning Walk", "Fetch"},
				"Learning": {"Training", "Tricks", "Commands"},
				"Rest":     {"Afternoon Nap", "Evening Sleep"},
			},
			wantDurations: map[string]int{
				"Exercise": 35,
				"Learning": 25,
				"Rest":     90,
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			orig := make([]activity, len(c.in))
			copy(orig, c.in)

			gotGroups := make(map[string][]string, len(c.wantGroups))
			gotDurations := make(map[string]int, len(c.wantDurations))
			for gotType, it := range GroupAdjacentBy(c.in, func(a activity) string { return a.Type }) {
				names := make([]string, 0, len(c.wantGroups[gotType]))
				var totalTime int
				for _, activity := range it {
					names = append(names, activity.Name)
					totalTime += activity.Duration
				}
				gotGroups[gotType] = names
				gotDurations[gotType] = totalTime
			}

			should.Equal(t, gotGroups, c.wantGroups)
			should.Equal(t, gotDurations, c.wantDurations)
			should.Equal(t, c.in, orig) // ensure original slice wasn't modified
		})
	}

	t.Run("early termination", func(t *testing.T) {
		t.Parallel()

		var secondIter bool
		for typ, it := range GroupAdjacentBy(schedule, func(a activity) string { return a.Type }) {
			if !secondIter {
				should.Equal(t, typ, "Exercise")
			} else {
				should.Equal(t, typ, "Learning")
				return
			}
			for range it {
				secondIter = true
				break
			}
		}
	})
}
