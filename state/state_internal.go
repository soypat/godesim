package state

import "fmt"

func (s *State) xCreateIfNotExist(sym Symbol) {
	if _, ok := s.varmap[sym]; !ok {
		s.x = append(s.x, 0)
		s.varmap[sym] = len(s.x) - 1
	}
}

func (s *State) uCreateIfNotExist(sym Symbol) {
	if _, ok := s.inputmap[sym]; !ok {
		s.u = append(s.u, 0)
		s.inputmap[sym] = len(s.u) - 1
	}
}

func throwf(s string, i ...interface{}) {
	panic(fmt.Sprintf(s, i...))
}

func (s *State) has(v string, sym Symbol) bool {
	switch v {
	case "X":
		if _, ok := s.varmap[Symbol(sym)]; !ok {
			return false
		}
		return true
	case "U":
		if _, ok := s.inputmap[Symbol(sym)]; !ok {
			return false
		}
		return true
	}
	panic("unreachable code")
}
