package bot

import (
	"fmt"
	"reflect"
	"testing"
)

func TestIsValidSoFar(t *testing.T) {
	a := Recurser{Id: "A", Config: UserConfig{
		Experience:        "easy",
		PairingDifficulty: []string{"easy"},
	}}
	b := Recurser{Id: "B", Config: UserConfig{
		Experience:        "medium",
		PairingDifficulty: []string{"easy", "medium"},
	}}
	c := Recurser{Id: "C", Config: UserConfig{
		Experience:        "medium",
		PairingDifficulty: []string{"medium"},
	}}
	d := Recurser{Id: "D", Config: UserConfig{
		Experience:        "hard",
		PairingDifficulty: []string{"medium", "hard"},
	}}
	e := Recurser{Id: "E", Config: UserConfig{
		Experience:        "hard",
		PairingDifficulty: []string{"hard"},
	}}

	table := []struct {
		input []Recurser
		want  bool
	}{
		{
			input: []Recurser{a, d},
			want:  false,
		},
		{
			input: []Recurser{b, c},
			want:  true,
		},
		{
			input: []Recurser{e, d, c},
			want:  true,
		},
		{
			input: []Recurser{a, b, c, d},
			want:  true,
		},
		{
			input: []Recurser{b, c, a, e},
			want:  false,
		},
	}

	for i, test := range table {
		name := fmt.Sprintf("Test %v", i)
		t.Run(name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				got := isValidSoFar(test.input)

				if got != test.want {
					t.Errorf("%s: Expected %v, got %v", name, test.want, got)
					break
				}
			}
		})
	}
}

func TestDetermineBestPath(t *testing.T) {
	a := Recurser{Id: "A", Config: UserConfig{
		Experience:        "easy",
		PairingDifficulty: []string{"easy"},
	}}
	b := Recurser{Id: "B", Config: UserConfig{
		Experience:        "medium",
		PairingDifficulty: []string{"easy", "medium"},
	}}
	c := Recurser{Id: "C", Config: UserConfig{
		Experience:        "medium",
		PairingDifficulty: []string{"easy", "medium"},
	}}
	d := Recurser{Id: "D", Config: UserConfig{
		Experience:        "medium",
		PairingDifficulty: []string{"medium"},
	}}
	e := Recurser{Id: "E", Config: UserConfig{
		Experience:        "hard",
		PairingDifficulty: []string{"hard"},
	}}
	f := Recurser{Id: "F", Config: UserConfig{
		Experience:        "hard",
		PairingDifficulty: []string{"medium, hard"},
	}}

	table := []struct {
		input      []Recurser
		validPairs int
	}{
		{
			input:      []Recurser{},
			validPairs: 0,
		},
		{
			input:      []Recurser{a},
			validPairs: 0,
		},
		{
			input:      []Recurser{a, b},
			validPairs: 1,
		},
		{
			input:      []Recurser{a, e},
			validPairs: 0,
		},
		{
			input:      []Recurser{b, c},
			validPairs: 1,
		},
		{
			input:      []Recurser{a, d, e},
			validPairs: 0,
		},
		{
			input:      []Recurser{a, b, c, d},
			validPairs: 2,
		},
		{
			input:      []Recurser{a, b, c, d, e},
			validPairs: 2,
		},
		{
			input:      []Recurser{a, b, c, d, e, f},
			validPairs: 3,
		},
	}

	for i, test := range table {
		name := fmt.Sprintf("Test %v", i)
		t.Run(name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				got, _ := determineBestPath(test.input)

				if len(got.order) != len(test.input) {
					t.Errorf("%s: Expected length %v, got length %v", name, len(test.input), len(got.order))
					break
				}

				if got.validPairs != test.validPairs {
					t.Errorf("%s: Expected %v, got %v", name, test.validPairs, got.validPairs)
					break
				}
			}
		})
	}
}

func TestDeterminePairs(t *testing.T) {
	a := Recurser{Id: "A", Config: UserConfig{
		Experience:        "easy",
		PairingDifficulty: []string{"easy"},
	}}
	b := Recurser{Id: "B", Config: UserConfig{
		Experience:        "medium",
		PairingDifficulty: []string{"easy", "medium"},
	}}
	c := Recurser{Id: "C", Config: UserConfig{
		Experience:        "medium",
		PairingDifficulty: []string{"easy", "medium"},
	}}
	d := Recurser{Id: "D", Config: UserConfig{
		Experience:        "medium",
		PairingDifficulty: []string{"medium"},
	}}
	e := Recurser{Id: "E", Config: UserConfig{
		Experience:        "hard",
		PairingDifficulty: []string{"hard"},
	}}
	f := Recurser{Id: "F", Config: UserConfig{
		Experience:        "hard",
		PairingDifficulty: []string{"medium, hard"},
	}}

	table := []struct {
		input         Path
		pairedList    []Recurser
		notPairedList []Recurser
	}{
		{
			input:         Path{[]Recurser{a, b}, 1},
			pairedList:    []Recurser{a, b},
			notPairedList: []Recurser{},
		},
		{
			input:         Path{[]Recurser{a, d}, 0},
			pairedList:    []Recurser{},
			notPairedList: []Recurser{a, d},
		},
		{
			input:         Path{[]Recurser{a, c, e}, 1},
			pairedList:    []Recurser{a, c},
			notPairedList: []Recurser{e},
		},
		{
			input:         Path{[]Recurser{a, b, c, d}, 2},
			pairedList:    []Recurser{a, b, c, d},
			notPairedList: []Recurser{},
		},
		{
			input:         Path{[]Recurser{a, b, c, d, e}, 2},
			pairedList:    []Recurser{a, b, c, d},
			notPairedList: []Recurser{e},
		},
		{
			input:         Path{[]Recurser{a, d, b, f}, 0},
			pairedList:    []Recurser{b, f},
			notPairedList: []Recurser{a, d},
		},
	}

	for i, test := range table {
		name := fmt.Sprintf("Test %v", i)
		t.Run(name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				pairedList, notPairedList, _ := determinePairs(test.input)

				if (len(pairedList) != 0 && len(test.pairedList) != 0) && !reflect.DeepEqual(pairedList, test.pairedList) {
					t.Errorf("%s: Expected %v, got %v", name, test.pairedList, pairedList)
					break
				}

				if (len(notPairedList) != 0 && len(test.notPairedList) != 0) && !reflect.DeepEqual(notPairedList, test.notPairedList) {
					t.Errorf("%s: Expected %v, got %v", name, test.notPairedList, notPairedList)
					break
				}
			}
		})
	}
}
