package bot

import (
	"fmt"
	"reflect"
	"testing"
)

func TestIsValidSoFar(t *testing.T) {
	table := []struct {
		input []Recurser
		want  bool
	}{
		{
			input: []Recurser{
				{Id: "A", Config: UserConfig{
					Experience:        "hard",
					PairingDifficulty: []string{"hard"},
				}},
				{Id: "B", Config: UserConfig{
					Experience:        "easy",
					PairingDifficulty: []string{"easy"},
				}},
			},
			want: false,
		},
		{
			input: []Recurser{
				{Id: "A", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
				{Id: "B", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
			},
			want: true,
		},
		{
			input: []Recurser{
				{Id: "A", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
				{Id: "B", Config: UserConfig{
					Experience:        "easy",
					PairingDifficulty: []string{"easy"},
				}},
			},
			want: true,
		},
		{
			input: []Recurser{
				{Id: "A", Config: UserConfig{
					Experience:        "hard",
					PairingDifficulty: []string{"hard"},
				}},
				{Id: "B", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
				{Id: "C", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
			},
			want: true,
		},
		{
			input: []Recurser{
				{Id: "A", Config: UserConfig{
					Experience:        "hard",
					PairingDifficulty: []string{"hard"},
				}},
				{Id: "B", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
				{Id: "C", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
				{Id: "D", Config: UserConfig{
					Experience:        "hard",
					PairingDifficulty: []string{"hard"},
				}},
			},
			want: false,
		},
	}

	for i, test := range table {
		name := fmt.Sprintf("Test %v", i)
		t.Run(name, func(t *testing.T) {
			got := isValidSoFar(test.input)

			if got != test.want {
				t.Errorf("%s: Expected %v, got %v", name, test.want, got)
			}
		})
	}
}

func TestDetermineBestPath(t *testing.T) {
	table := []struct {
		input       []Recurser
		validOrders [][]string
		validPairs  int
	}{
		{
			input:       []Recurser{},
			validOrders: [][]string{},
			validPairs:  0,
		},
		{
			input: []Recurser{
				{Id: "A", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
			},
			validOrders: [][]string{},
			validPairs:  0,
		},
		{
			input: []Recurser{
				{Id: "A", Config: UserConfig{
					Experience:        "hard",
					PairingDifficulty: []string{"hard"},
				}},
				{Id: "B", Config: UserConfig{
					Experience:        "easy",
					PairingDifficulty: []string{"easy"},
				}},
			},
			validOrders: [][]string{},
			validPairs:  0,
		},
		{
			input: []Recurser{
				{Id: "A", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
				{Id: "B", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
			},
			validOrders: [][]string{{"A", "B"}, {"B", "A"}},
			validPairs:  1,
		},
		{
			input: []Recurser{
				{Id: "A", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
				{Id: "B", Config: UserConfig{
					Experience:        "easy",
					PairingDifficulty: []string{"easy"},
				}},
			},
			validOrders: [][]string{{"A", "B"}, {"B", "A"}},
			validPairs:  1,
		},
		{
			input: []Recurser{
				{Id: "A", Config: UserConfig{
					Experience:        "hard",
					PairingDifficulty: []string{"hard"},
				}},
				{Id: "B", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
				{Id: "C", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
			},
			validOrders: [][]string{{"B", "C", "A"}, {"C", "B", "A"}},
			validPairs:  1,
		},
		{
			input: []Recurser{
				{Id: "A", Config: UserConfig{
					Experience:        "hard",
					PairingDifficulty: []string{"hard"},
				}},
				{Id: "B", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
				{Id: "C", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
				{Id: "D", Config: UserConfig{
					Experience:        "hard",
					PairingDifficulty: []string{"hard"},
				}},
			},
			validOrders: [][]string{
				{"A", "D", "B", "C"},
				{"A", "D", "C", "B"},
				{"D", "A", "B", "C"},
				{"D", "A", "C", "B"},
				{"B", "C", "A", "D"},
				{"B", "C", "D", "A"},
				{"C", "B", "A", "D"},
				{"C", "B", "D", "A"},
			},
			validPairs: 2,
		},
		{
			input: []Recurser{
				{Id: "A", Config: UserConfig{
					Experience:        "easy",
					PairingDifficulty: []string{"easy"},
				}},
				{Id: "B", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"medium"},
				}},
				{Id: "C", Config: UserConfig{
					Experience:        "hard",
					PairingDifficulty: []string{"hard"},
				}},
				{Id: "D", Config: UserConfig{
					Experience:        "hard",
					PairingDifficulty: []string{"hard"},
				}},
			},
			validOrders: [][]string{
				{"C", "D", "A", "B"},
				{"C", "D", "B", "A"},
				{"D", "C", "A", "B"},
				{"D", "C", "B", "A"},
			},
			validPairs: 1,
		},
		{
			input: []Recurser{
				{Id: "A", Config: UserConfig{
					Experience:        "easy",
					PairingDifficulty: []string{"easy"},
				}},
				{Id: "B", Config: UserConfig{
					Experience:        "medium",
					PairingDifficulty: []string{"easy", "medium"},
				}},
				{Id: "C", Config: UserConfig{
					Experience:        "hard",
					PairingDifficulty: []string{"medium", "hard"},
				}},
				{Id: "D", Config: UserConfig{
					Experience:        "hard",
					PairingDifficulty: []string{"hard"},
				}},
			},
			validOrders: [][]string{
				{"C", "D", "A", "B"},
				{"C", "D", "B", "A"},
				{"D", "C", "A", "B"},
				{"D", "C", "B", "A"},
			},
			validPairs: 2,
		},
	}

	for i, test := range table {
		name := fmt.Sprintf("Test %v", i)
		t.Run(name, func(t *testing.T) {
			got := determineBestPath(test.input)

			order, valid := func() ([]string, bool) {
				order := []string{}
				for _, r := range got.order {
					order = append(order, r.Id)
				}

				if len(got.order) == 0 && len(test.validOrders) == 0 {
					return order, true
				}

				for _, t := range test.validOrders {
					if reflect.DeepEqual(t, order) {
						return order, true
					}
				}

				return order, false
			}()

			if !valid {
				t.Errorf("%s: Expected one of %v, got %v", name, test.validOrders, order)
			}

			if test.validPairs != got.validPairs {
				t.Errorf("%s: Expected %v, got %v", name, test.validPairs, got.validPairs)
			}
		})
	}
}
