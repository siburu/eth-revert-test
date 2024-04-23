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
	"github.com/siburu/eth-revert-test/go/contracts"
)

const (
	gethURL = "http://localhost:8545"
	chainID = 12345
)

const (
	revertTypeRequire      uint8 = 0
	revertTypeRevert             = 1
	revertTypeCustomRevert       = 2
	revertTypeOverflow           = 3
	revertTypeNone               = 4
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

	c, err := contracts.NewC(addr, cl)
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

Out:
	for _, revertType := range []uint8{
		revertTypeRequire,
		revertTypeRevert,
		revertTypeCustomRevert,
		revertTypeOverflow,
		revertTypeNone,
		100,
	} {
		tx, err := c.Test(txOpts, revertType, big.NewInt(3))
		if err != nil {
			panic(err)
		}

	In:
		for {
			receipt, err := cl.TransactionReceipt(ctx, tx.Hash())
			if err == nil {
				switch receipt.Status {
				case types.ReceiptStatusSuccessful:
					fmt.Printf("OK: RevertType=%d\n", revertType)
					continue Out
				case types.ReceiptStatusFailed:
					fmt.Printf("NG: RevertType=%d\n", revertType)
					break In
				default:
					panic(fmt.Sprintf("unexpected receipt status: %d", receipt.Status))
				}
			} else if err == ethereum.NotFound {
				time.Sleep(time.Second)
			} else {
				panic(err)
			}
		}

		cf, err := traceTransaction(ctx, cl, tx.Hash())
		if err != nil {
			panic(err)
		}
		fmt.Printf("cf: output=%x, error=%s, revertReason=%s\n", cf.Output, cf.Error, cf.RevertReason)

		if revertReason, err := abi.UnpackRevert(cf.Output); err == nil {
			fmt.Printf("parsed revert reason: %s\n", revertReason)
		} else if abi, err := contracts.CMetaData.GetAbi(); err != nil {
			panic(err)
		} else {
			for errName, errABI := range abi.Errors {
				if v, err := errABI.Unpack(cf.Output); err == nil {
					vmap := make(map[int]interface{})
					for i, v := range v.([]interface{}) {
						vmap[i] = v
					}
					fmt.Printf("parsed custom error: name=%s, values=%v\n", errName, vmap)
					continue Out
				}
			}

			bzCf, err := json.MarshalIndent(cf, "", "\t")
			if err != nil {
				panic(err)
			}
			fmt.Printf("unknown error: %v\n", string(bzCf))
		}
	}
}
