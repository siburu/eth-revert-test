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
	"github.com/ethereum/go-ethereum/rpc"
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

func parseError(errorData []byte) (string, error) {
	if revertReason, err := abi.UnpackRevert(errorData); err == nil {
		return fmt.Sprintf("traditional revert: %s", revertReason), nil
	}

	abi, err := contracts.CMetaData.GetAbi()
	if err != nil {
		return "", fmt.Errorf("failed to parse ABI: %v", err)
	}

	for errName, errABI := range abi.Errors {
		if v, err := errABI.Unpack(errorData); err == nil {
			s, err := errorToJSON(v, errABI)
			if err != nil {
				return "", fmt.Errorf("failed to convert error into JSON string: %v", err)
			}
			return fmt.Sprintf("custom error: name=%s, inputs=%s", errName, s), nil
		}
	}

	return "", fmt.Errorf("unparseable error data")
}

func errorToJSON(errVal interface{}, errABI abi.Error) (string, error) {
	m := make(map[string]interface{})
	for i, v := range errVal.([]interface{}) {
		m[errABI.Inputs[i].Name] = v
	}
	bz, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(bz), nil
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
			de, ok := err.(rpc.DataError)
			if ok {
				errorData := de.ErrorData()
				if errorData != nil {
					e, err := parseError(common.FromHex(errorData.(string)))
					if err != nil {
						panic(err)
					}
					fmt.Printf("DataError: parsed=%s\n", e)
				} else {
					fmt.Printf("DataError: unparsed=%s\n", de.Error())
				}
			} else {
				fmt.Printf("not DataError: value=%v, type=%T\n", err, err)
			}
			continue Out
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

		e, err := parseError(cf.Output)
		if err != nil {
			fmt.Printf("failed to parse: %v\n", err)
		} else {
			fmt.Printf("parsed error: %s\n", e)
		}
	}
}
