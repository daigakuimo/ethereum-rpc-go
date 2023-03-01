package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	L2StandardERC20 "github.com/daigakuimo/ethereum-rpc-go/contracts"
)

type TransferEvent struct {
	From  common.Address // なんでか知らんけど使えない
	To    common.Address // 0x0000...が入ってる
	Value *big.Int
}

type balance struct {
	Address common.Address
	Amount  *big.Int
}

func main() {
	ethClient, err := ethclient.Dial("https://rpc.sandverse.oasys.games/")
	if err != nil {
		log.Fatal(err)
	}

	// Create a filter query
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(1000), // 負荷軽減のためにトークンデプロイした時のブロックから検索
		Addresses: []common.Address{
			common.HexToAddress("0x60E21183813719C7A78B403c3B0C5BdcA6ceDEb8"), // ここにBPCのアドレス入れる
		},
		Topics: [][]common.Hash{
			{common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
		},
	}

	// Call the eth_getLogs RPC method
	logs, err := ethClient.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	contractAbi, err := abi.JSON(strings.NewReader(string(L2StandardERC20.L2StandardERC20ABI)))
	if err != nil {
		log.Fatal(err)
	}

	balances := map[common.Address]*big.Int{}

	for _, vLog := range logs {
		fromAddress := common.HexToAddress(vLog.Topics[1].Hex()) // vLog.Topics[1].Hex() トークン送信者
		toAddress := common.HexToAddress(vLog.Topics[2].Hex())   // vLog.Topics[2].Hex() トークン受信者

		// Transferイベントを取得
		var event TransferEvent
		err := contractAbi.UnpackIntoInterface(&event, "Transfer", vLog.Data)
		if err != nil {
			log.Fatal(err)
		}

		if balances[fromAddress] == nil {
			balances[fromAddress] = big.NewInt(0)
		}

		if balances[toAddress] == nil {
			balances[toAddress] = big.NewInt(0)
		}

		// トークンの取引額を計算
		balances[fromAddress] = big.NewInt(0).Sub(balances[fromAddress], event.Value)
		balances[toAddress] = big.NewInt(0).Add(balances[toAddress], event.Value)
	}

	// 所持数順にソート
	// キーと値を格納するスライスを作成
	var sortBalances []balance
	for address, amount := range balances {
		sortBalances = append(sortBalances, balance{address, amount})
	}

	// スライスをソート
	sort.Slice(sortBalances, func(i, j int) bool {
		return sortBalances[i].Amount.Cmp(sortBalances[j].Amount) > 0
	})

	for index, balance := range sortBalances {
		fmt.Println(index, ":", balance.Address, ":", balance.Amount)
	}
}
