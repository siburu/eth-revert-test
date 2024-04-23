package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/siburu/eth-revert-test/go/hoge"
)

const (
	gethURL = "http://localhost:8545"
	chainID = 12345
)

const (
	revertTypeRequire      uint8 = 0
	revertTypeRevert             = 1
	revertTypeCustomRevert       = 2
)

func parseAddress(path string) (common.Address, error) {
	f, err := os.Open(path)
	if err != nil {
		return common.Address{}, err
	}
	defer f.Close()

	bz, err := io.ReadAll(f)
	if err != nil {
		return common.Address{}, err
	}

	return common.HexToAddress(string(bz)), nil
}

func main() {
	ctx := context.Background()

	addr, err := parseAddress("../share/address.txt")
	if err != nil {
		panic(err)
	}

	cl, err := ethclient.DialContext(ctx, gethURL)
	if err != nil {
		panic(err)
	}

	h, err := hoge.NewHoge(addr, cl)
	if err != nil {
		panic(err)
	}

	bzPrivateKey := make([]byte, 32)
	bzPrivateKey[31] = 1
	privateKey, err := crypto.ToECDSA(bzPrivateKey)
	if err != nil {
		panic(err)
	}

	txOpts, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainID))
	if err != nil {
		panic(err)
	}

	// set GasLimit to avoid executing eth_estimateGas
	txOpts.GasLimit = 5000000

	for _, revertType := range []uint8{
		revertTypeRequire,
		revertTypeRevert,
		revertTypeCustomRevert,
	} {
		tx, err := h.Test(txOpts, revertType, big.NewInt(10))
		if err != nil {
			panic(err)
		}

		for {
			receipt, err := cl.TransactionReceipt(ctx, tx.Hash())
			if err == nil {
				if receipt.Status != types.ReceiptStatusFailed {
					bz, err := json.MarshalIndent(receipt, "", "\t")
					if err != nil {
						panic(err)
					}
					panic(fmt.Sprintf("unexpected successful receipt: %s", string(bz)))
				}
				break
			} else if err != ethereum.NotFound {
				panic(err)
			}

			time.Sleep(time.Second)
		}

		cf, err := traceTransaction(ctx, cl, tx.Hash())
		if err != nil {
			panic(err)
		}

		if revertReason, err := abi.UnpackRevert(cf.Output); err == nil {
			fmt.Printf("parsed revert reason: %s\n", revertReason)
			fmt.Printf("RevertReason: %s\n", cf.RevertReason)
		} else if abi, err := hoge.HogeMetaData.GetAbi(); err != nil {
			panic(err)
		} else {
			for _, errABI := range abi.Errors {
				if v, err := errABI.Unpack(cf.Output); err == nil {
					vmap := make(map[int]interface{})
					for i, v := range v.([]interface{}) {
						vmap[i] = v
					}
					fmt.Printf("parsed error: %v\n", vmap)
					break
				}
			}
		}
	}
}
