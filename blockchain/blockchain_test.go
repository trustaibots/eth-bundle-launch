package blockchain

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestGenesis(t *testing.T) {
	b, close := NewTestBlockchain(t, nil)
	defer close()

	// no genesis block yet
	if _, err := b.Header(); err == nil {
		t.Fatal("it shoudl be empty")
	}

	// add genesis block
	genesis := &types.Header{Difficulty: big.NewInt(1), Number: big.NewInt(0)}
	if err := b.WriteGenesis(genesis); err != nil {
		t.Fatal(err)
	}

	header, err := b.Header()
	if err != nil {
		t.Fatal(err)
	}

	if header.Hash() != genesis.Hash() {
		t.Fatal("bad")
	}
}

type chain struct {
	headers map[byte]*types.Header
}

func (c *chain) add(h *header) error {
	if _, ok := c.headers[h.hash]; ok {
		return fmt.Errorf("hash already imported")
	}

	var parent common.Hash
	if h.number != 0 {
		p, ok := c.headers[h.parent]
		if !ok {
			return fmt.Errorf("parent not found %v", h.parent)
		}
		parent = p.Hash()
	}

	c.headers[h.hash] = &types.Header{
		ParentHash: parent,
		Number:     big.NewInt(int64(h.number)),
		Difficulty: big.NewInt(int64(h.diff)),
		Extra:      []byte{h.hash},
	}
	return nil
}

type header struct {
	hash   byte
	parent byte
	number uint64
	diff   uint64
}

func (h *header) Parent(parent byte) *header {
	h.parent = parent
	return h
}

func (h *header) Diff(d uint64) *header {
	h.diff = d
	return h
}

func (h *header) Number(d uint64) *header {
	h.number = d
	return h
}

func mock(number byte) *header {
	return &header{
		hash:   number,
		parent: number - 1,
		number: uint64(number),
		diff:   uint64(number),
	}
}

func TestInsertHeaders(t *testing.T) {
	var cases = []struct {
		Name    string
		History []*header
		Head    *header
		Forks   []*header
	}{
		{
			Name: "Genesis",
			History: []*header{
				mock(0x0),
			},
			Head: mock(0x0),
		},
		{
			Name: "Linear",
			History: []*header{
				mock(0x0),
				mock(0x1),
				mock(0x2),
			},
			Head: mock(0x2),
		},
		{
			Name: "Keep block with higher difficulty",
			History: []*header{
				mock(0x0),
				mock(0x1),
				mock(0x3).Parent(0x1).Diff(5),
				mock(0x2).Parent(0x1).Diff(3),
			},
			Head:  mock(0x3),
			Forks: []*header{mock(0x2)},
		},
		{
			Name: "Reorg",
			History: []*header{
				mock(0x0),
				mock(0x1),
				mock(0x2),
				mock(0x3),
				mock(0x4).Parent(0x1).Diff(10).Number(2),
				mock(0x5).Parent(0x4).Diff(11).Number(3),
				mock(0x6).Parent(0x3).Number(4),
			},
			Head:  mock(0x5),
			Forks: []*header{mock(0x6)},
		},
	}

	for _, cc := range cases {
		t.Run(cc.Name, func(tt *testing.T) {
			b, close := NewTestBlockchain(t, nil)
			defer close()

			chain := chain{
				headers: map[byte]*types.Header{},
			}
			for _, i := range cc.History {
				if err := chain.add(i); err != nil {
					tt.Fatal(err)
				}
			}

			// genesis is 0x0
			if err := b.WriteGenesis(chain.headers[0x0]); err != nil {
				tt.Fatal(err)
			}

			// run the history
			for i := 1; i < len(cc.History); i++ {
				if err := b.WriteHeader(chain.headers[cc.History[i].hash]); err != nil {
					tt.Fatal(err)
				}
			}

			head, err := b.Header()
			if err != nil {
				tt.Fatal(err)
			}

			expected, ok := chain.headers[cc.Head.hash]
			if !ok {
				tt.Fatal("bad")
			}

			if head.Hash() != expected.Hash() {
				tt.Fatal("bad2")
			}

			forks := b.GetForks()
			expectedForks := []common.Hash{}

			for _, i := range cc.Forks {
				expectedForks = append(expectedForks, chain.headers[i.hash].Hash())
			}

			if len(forks) == len(expectedForks) && len(forks) != 0 {
				if !reflect.DeepEqual(forks, expectedForks) {
					t.Fatal("forks dont match")
				}
			}
		})
	}
}
