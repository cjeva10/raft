package store

import (
	"testing"
)

func TestSet(t *testing.T) {
	key := "hello"
	val := 2

	s := New()

    set := s.set(key, val)
    if set != val {
        t.Errorf("set returned the wrong value: expected %v, got %v\n", val, set)
    }
	if s.data[key] != val {
        t.Errorf("Failed to set %v=%v: expected %v, got %v\n", key, val, val, s.data[key])
	}
}

func TestGet(t *testing.T) {
	key := "hello"
	val := 2
	s := New()
	s.data[key] = val

	actual := s.get(key)
	if actual != val {
		t.Errorf("Failed to get %v: expected %v, got %v\n", key, val, actual)
	}
}

func TestParseCommands(t *testing.T) {
	tests := []struct {
		input     []string
		expected  []int
		shouldErr bool
	}{
		{
			[]string{
				"",
			},
			[]int{0},
			true,
		},
		{
			[]string{
				"sets a",
			},
			[]int{0},
			true,
		},
		{
			[]string{
				"set a a",
				"get a a a",
			},
			[]int{0, 0},
			true,
		},
		{
			[]string{
				"set a",
			},
			[]int{0},
			true,
		},
		{
			[]string{
				"set a 2",
				"get a",
				"get b",
				"set b 3",
				"get b",
			},
			[]int{2, 2, 0, 3, 3},
			false,
		},
		{
			[]string{
				"get a",
			},
			[]int{0},
			false,
		},
	}

	for _, tt := range tests {
		s := New()
		var val int
		var err error
		errCount := 0

		for i, in := range tt.input {
			val, err = s.Apply(in)
			if err != nil {
				errCount++
				if !tt.shouldErr {
					t.Errorf("Shouldn't produce an error\n")
				}
			}
			if val != tt.expected[i] {
				t.Errorf("Wrong value: expected %v, got %v\n", tt.expected[i], val)
			}
		}

		if tt.shouldErr && errCount == 0 {
			t.Errorf("Didn't get an error\n")
		}
	}
}
