package chain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

var emptyAddr common.Address

type Case struct {
	input  string
	output interface{}
}

func addr(str string) common.Address {
	return common.HexToAddress(str)
}

func hash(str string) common.Hash {
	return common.HexToHash(str)
}

func TestGenesisAlloc(t *testing.T) {
	cases := []Case{
		{
			input: `{
				"0x0000000000000000000000000000000000000000": {
					"balance": "0x11"
				}
			}`,
			output: GenesisAlloc{
				emptyAddr: GenesisAccount{
					Balance: big.NewInt(17),
				},
			},
		},
		{
			input: `{
				"0x0000000000000000000000000000000000000000": {}
			}`,
			output: nil,
		},
		{
			input: `{
				"0x0000000000000000000000000000000000000000": {
					"balance": "0x11",
					"nonce": "0x100",
					"storage": {
						"` + hash("1").String() + `": "` + hash("3").String() + `",
						"` + hash("2").String() + `": "` + hash("4").String() + `"
					}
				}
			}`,
			output: GenesisAlloc{
				emptyAddr: GenesisAccount{
					Balance: big.NewInt(17),
					Nonce:   256,
					Storage: map[common.Hash]common.Hash{
						hash("1"): hash("3"),
						hash("2"): hash("4"),
					},
				},
			},
		},
		{
			input: `{
				"0x0000000000000000000000000000000000000000": {
					"balance": "0x11"
				},
				"0x0000000000000000000000000000000000000001": {
					"balance": "0x12"
				}
			}`,
			output: GenesisAlloc{
				addr("0"): GenesisAccount{
					Balance: big.NewInt(17),
				},
				addr("1"): GenesisAccount{
					Balance: big.NewInt(18),
				},
			},
		},
	}

	for _, c := range cases {
		var dec GenesisAlloc
		if err := json.Unmarshal([]byte(c.input), &dec); err != nil {
			if c.output != nil {
				t.Fatal(err)
			}
		} else if !reflect.DeepEqual(dec, c.output) {
			t.Fatal("bad")
		}
	}
}

func TestGenesis(t *testing.T) {
	cases := []Case{
		{
			input: `{
				"difficulty": "0x12",
				"gasLimit": "0x11",
				"alloc": {
					"0x0000000000000000000000000000000000000000": {
						"balance": "0x11"
					},
					"0x0000000000000000000000000000000000000001": {
						"balance": "0x12"
					}
				}
			}`,
			output: Genesis{
				Difficulty: big.NewInt(18),
				GasLimit:   17,
				Alloc: GenesisAlloc{
					emptyAddr: GenesisAccount{
						Balance: big.NewInt(17),
					},
					addr("1"): GenesisAccount{
						Balance: big.NewInt(18),
					},
				},
			},
		},
	}

	for _, c := range cases {
		var dec Genesis
		if err := json.Unmarshal([]byte(c.input), &dec); err != nil {
			fmt.Println(err)
			if c.output != nil {
				t.Fatal(err)
			}
		} else if !reflect.DeepEqual(dec, c.output) {
			t.Fatal("bad")
		}
	}
}

func TestChainFolder(t *testing.T) {
	// it should be able to parse all the chains in the ./chains folder
	files, err := ioutil.ReadDir("./chains")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		name := strings.TrimSuffix(f.Name(), ".json")
		if _, err := ImportFromName(name); err != nil {
			t.Fatalf("Failed to parse %s: %v", name, err)
		}
	}
}
