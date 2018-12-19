package chain

import "math/big"

// Params are all the set of params for the chain
type Params struct {
	Forks *Forks `json:"number"`
}

// Forks specifies when each fork is activated
type Forks struct {
	Homestead      *Fork `json:"homestead"`
	Byzantium      *Fork `json:"byzantium"`
	Constantinople *Fork `json:"constantinople"`
	EIP150         *Fork `json:"EIP150"`
	EIP158         *Fork `json:"EIP158"`
}

func (f *Forks) GasTable(num *big.Int) GasTable {
	if num == nil {
		return GasTableHomestead
	}

	n := num.Uint64()
	switch {
	case f.IsConstantinople(n):
		return GasTableConstantinople
	case f.IsEIP158(n):
		return GasTableEIP158
	case f.IsEIP150(n):
		return GasTableEIP150
	default:
		return GasTableHomestead
	}
}

func (f *Forks) active(ff *Fork, block uint64) bool {
	if ff == nil {
		return false
	}
	return ff.Active(block)
}

func (f *Forks) IsHomestead(block uint64) bool {
	return f.active(f.Homestead, block)
}

func (f *Forks) IsByzantium(block uint64) bool {
	return f.active(f.Byzantium, block)
}

func (f *Forks) IsConstantinople(block uint64) bool {
	return f.active(f.Constantinople, block)
}

func (f *Forks) IsEIP150(block uint64) bool {
	return f.active(f.EIP150, block)
}

func (f *Forks) IsEIP158(block uint64) bool {
	return f.active(f.EIP158, block)
}

func (f *Forks) At(block uint64) ForksInTime {
	return ForksInTime{
		Homestead:      f.active(f.Homestead, block),
		Byzantium:      f.active(f.Byzantium, block),
		Constantinople: f.active(f.Constantinople, block),
		EIP150:         f.active(f.EIP150, block),
		EIP158:         f.active(f.EIP158, block),
	}
}

type Fork uint64

func NewFork(n uint64) *Fork {
	f := Fork(n)
	return &f
}

func (f Fork) Active(block uint64) bool {
	return block >= uint64(f)
}

type ForksInTime struct {
	Homestead, Byzantium, Constantinople, EIP150, EIP158 bool
}
