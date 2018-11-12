package syncer

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
)

// -- list --

type flat struct {
	block uint64
	t     itemType
}

func checkList(t *testing.T, s *list, f []flat) {
	ff := []flat{}
	elem := s.front
	for elem != nil {
		ff = append(ff, flat{elem.block, elem.t})
		elem = elem.next
	}

	if !reflect.DeepEqual(ff, f) {
		t.Fatal("dont match")
	}
}

func TestListNew(t *testing.T) {
	s := newList(0, 100)

	checkList(t, s, []flat{
		{0, empty},
		{100, empty},
	})
}

func TestListAddItem(t *testing.T) {
	s := newList(0, 1000)

	s.GetQuerySlot()
	s.GetQuerySlot()

	checkList(t, s, []flat{
		{0, pending},
		{100, pending},
		{200, empty},
		{1000, empty},
	})
}

func TestListAddItemWithUpdate(t *testing.T) {
	s := newList(0, 1000)

	i := s.GetQuerySlot()
	s.GetQuerySlot()

	if err := s.UpdateSlot(i.id, failed, nil); err != nil {
		t.Fatal(err)
	}

	checkList(t, s, []flat{
		{0, failed},
		{100, pending},
		{200, empty},
		{1000, empty},
	})

	// ask for another slot, it should return the failed one
	i1 := s.GetQuerySlot()
	if i1.id != i.id {
		t.Fatal("id should match the failed one")
	}

	checkList(t, s, []flat{
		{0, pending},
		{100, pending},
		{200, empty},
		{1000, empty},
	})
}

func TestListUpdateWrongItem(t *testing.T) {
	s := newList(0, 1000)
	s.GetQuerySlot()

	if err := s.UpdateSlot(100, failed, nil); err == nil {
		t.Fatal("it should fail")
	}
}

func TestListCommit(t *testing.T) {
	s := newList(0, 1000)

	dummy := &types.Header{}

	i0 := s.GetQuerySlot()
	i1 := s.GetQuerySlot()

	if err := s.UpdateSlot(i0.id, completed, []*types.Header{dummy}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateSlot(i1.id, completed, []*types.Header{dummy}); err != nil {
		t.Fatal(err)
	}

	checkList(t, s, []flat{
		{0, completed},
		{100, completed},
		{200, empty},
		{1000, empty},
	})

	headers := s.commitData()
	if len(headers) != 2 {
		t.Fatal("it should retrieve 2 headers")
	}

	checkList(t, s, []flat{
		{200, empty},
		{1000, empty},
	})
}
