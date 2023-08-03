package store

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// implement the key value store so that when `apply()` is called,
// we can update our machine

type Store struct {
	data map[string]int
}

func (s *Store) set(key string, val int) int {
	s.data[key] = val
    return val
}

func (s *Store) get(key string) int {
	return s.data[key]
}

func New() *Store {
	return &Store{
        data: make(map[string]int),
    }
}

// apply the given command to the given store
// return an error if the command is unable to be applied
func (s *Store) Apply(command string) (int, error) {
	words := strings.Split(command, " ")

	if len(words) == 0 {
		return 0, errors.New("Empty command")
	}

	switch words[0] {
	case "get":
        val := words[1]
        if val == "" {
            err := fmt.Sprintf("Attempted to get empty value\n")
            return 0, errors.New(err)
        }
        return s.get(val), nil
    case "set":
        if len(words) < 3 {
            err := fmt.Sprintf("Missing key/val pair for set\n")
            return 0, errors.New(err)
        }
        key := words[1]
        val, err := strconv.Atoi(words[2])
        if err != nil {
            err := fmt.Sprintf("Failed to parse val to integer: %v\n", words[2])
            return 0, errors.New(err)
        }
        return s.set(key, val), nil
	default:
		err := fmt.Sprintf("Didn't recognize command: %v\n", words[0])
		return 0, errors.New(err)
	}
}
