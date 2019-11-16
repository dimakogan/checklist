package boosted

import (
  "math/rand"
)

type Set map[int]bool

type SetKey struct {
  univSize int
  delta int
  set map[int]bool
}

type PuncSetKey struct {
  univSize int
  delta int
  set map[int]bool
}

// Go's % operator follows C semantics and can produce
// negative values if it's given a negative argument. 
// We need an arithmetic mod operator.
func MathMod(x int, mod int) int {
  out := x % mod

  // TODO: This is not a constant-time operation.
  if out < 0 {
    out = out + mod
  }

  return out
}

func (s Set) RandomValue(src *rand.Rand) int {
  keys := make([]int, len(s))
  i := 0
  for k := range s {
    keys[i] = k
    i += 1
  }

  choose := src.Intn(len(keys))
  return keys[choose]
}

func SetGen(src *rand.Rand, univSize int, setSize int) *SetKey {
  if (univSize < setSize) {
     panic("Set size too large.")
  }

  if (univSize < 1) {
     panic("Universe size too small.")
  }

  // TODO: Implement this more efficiently
  out := make(map[int]bool)
  for len(out) < setSize {
    out[src.Intn(univSize)] = true
  }

  delta := src.Intn(univSize)
  return &SetKey{univSize, delta, out}
}

func SetGenWith(src *rand.Rand, univSize int, setSize int, val int) *SetKey {
  key := SetGen(src, univSize, setSize)
  key.delta = 0

  // TODO: Implement this more efficiently.
  set := key.Eval()
  choose := set.RandomValue(src)
  key.delta = MathMod(val - choose, univSize)

  return key
}

func (key *SetKey) Shift(amount int) {
  key.delta = MathMod(key.delta + amount, key.univSize)
}

func (key *SetKey) Punc(idx int) *PuncSetKey {
  puncAt := MathMod(idx - key.delta, key.univSize)

  if !key.set[puncAt] {
    panic("Can't puncture at this point!")
  }

  out := make(map[int]bool)
  for i := range key.set {
    if i != puncAt {
      out[i] = true
    }
  }

  return &PuncSetKey{key.univSize, key.delta, out}
}

func (key *SetKey) Eval() Set {
  return evalMap(key.univSize, key.delta, key.set)
}

func (key *PuncSetKey) Eval() Set {
  return evalMap(key.univSize, key.delta, key.set)
}

func evalMap(univSize int, delta int, m map[int]bool) Set {
  out := make(Set)

  for k := range m {
    out[MathMod(k + delta, univSize)] = true
  }

  return out
}

