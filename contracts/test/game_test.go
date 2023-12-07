package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	contracts "github.com/Filecoin-Titan/titan-game-sdk/contracts/api"
	"github.com/Filecoin-Titan/titan-game-sdk/contracts/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

const (
	// endpoint  = "http://172.25.9.91:1251/rpc/v1"
	// networkID = 2371761377286762
	// networkID = 314159

	endpoint = "https://api.calibration.node.glif.io/"

	contractAddress = "0x4b599a339A7b649C0fe641C2143dde42985602eD" //"0x759c4Bb3BB2182A478e1A95e0A0D906E23716F54"
)

func TestDeployGame(t *testing.T) {
	c, err := client.New(
		client.PrivateKeyOption(os.Getenv("PRIVATE_KEY")),
		client.EndpointOption(endpoint),
	)
	if err != nil {
		t.Fatal("new client err", err.Error())
	}

	result, err := c.InvokeContract(0, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		addr, tr, _, err := contracts.DeployGameReplayContract(opts, c.EthClient())
		if err != nil {
			return nil, err
		}

		t.Logf("deploy contract %s", addr.Hex())
		return tr, nil
	})
	if err != nil {
		t.Fatal("deploy contracts err ", err.Error())
	}

	t.Log("deploy OK: ", string(result))
}

func TestSaveGameReplay(t *testing.T) {
	c, err := client.New(
		client.PrivateKeyOption(os.Getenv("PRIVATE_KEY")),
		client.EndpointOption(endpoint),
	)
	if err != nil {
		t.Fatal("new client err ", err.Error())
	}

	replayID := uuid.NewString()
	gameContractAddress := common.HexToAddress(contractAddress)
	instance, err := contracts.NewGameReplayContract(gameContractAddress, c.EthClient())
	if err != nil {
		t.Fatal("new contract instance err ", err.Error())
	}

	result, err := c.InvokeContract(0, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		results := make([]contracts.GameRoundResult, 0, 4)
		for i := 1; i <= 4; i++ {
			result := contracts.GameRoundResult{
				PlayerID:     fmt.Sprintf("%d", i),
				CurrentScore: 100,
				WinScore:     -1,
			}
			results = append(results, result)

		}

		info := contracts.GameRoundInfo{
			GameID:    "123",
			RoundID:   "1",
			ReplayID:  replayID,
			PlayerIDs: "1,2,3,4,5",
		}
		gameReplay := contracts.GameRoundReplay{
			DomainSeparationTag: 1,
			VRFHeight:           1,
			HashFunc:            "hash256",
			VRFProof:            []byte("abcd"),
			Address:             "1211211212121212",
			ReplayCID:           "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			GameInfo:            info,
			GameResults:         results,
		}
		return instance.SaveGameReplay(opts, []contracts.GameRoundReplay{gameReplay})
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("replay id: ", replayID)
	t.Log("save game replay OK: ", string(result))
	t.Log("querying game replay...")

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-time.After(2 * time.Minute):
			t.Log("query order timeout!")
			return
		case <-ticker.C:
			order, err := instance.GetGameReplay(&bind.CallOpts{
				Pending: true,
			}, replayID)
			if err != nil {
				t.Logf("get game replay err %s", err.Error())
				continue
			}

			t.Log("Query game replay OK: ", order)
			return
		}
	}
}

// a86582cf975e1c5ffecc5f4e8816a71a2055fd3a3e682113dd00f42f27c111e8
func TestTransferOwner(t *testing.T) {
	c, err := client.New(
		client.PrivateKeyOption(os.Getenv("PRIVATE_KEY")),
		client.EndpointOption(endpoint),
	)
	if err != nil {
		t.Fatal("new client err ", err.Error())
	}

	myAddress, err := c.Address()
	if err != nil {
		t.Fatal("get address err ", err.Error())
	}

	gameContractAddress := common.HexToAddress(contractAddress)
	instance, err := contracts.NewGameReplayContract(gameContractAddress, c.EthClient())
	if err != nil {
		t.Fatal("new contract instance err ", err.Error())
	}

	if instance == nil {
		t.Fatal("instance == nil")
	}

	result, err := c.InvokeContract(0, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return instance.TransferOwnership(opts, myAddress)
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("TransferOwnership OK: ", string(result))
}

func TestGetGameReplay(t *testing.T) {
	roundID := "59ba7216-db8d-42c2-b92e-fff22b07a6e3"
	c, err := client.New(
		client.PrivateKeyOption(os.Getenv("PRIVATE_KEY")),
		client.EndpointOption(endpoint),
	)
	if err != nil {
		t.Fatal("new client err ", err.Error())
	}

	gameContractAddress := common.HexToAddress(contractAddress)
	instance, err := contracts.NewGameReplayContract(gameContractAddress, c.EthClient())
	if err != nil {
		t.Fatal("new contract instance err ", err.Error())
	}

	if instance == nil {
		t.Fatal("instance == nil")
	}

	replay, err := instance.GetGameReplay(nil, roundID)
	if err != nil {
		t.Fatal("new client err ", err.Error())
	}

	t.Log("game replay: ", replay)
}

func TestForeachGameReplay(t *testing.T) {
	c, err := client.New(
		client.PrivateKeyOption(os.Getenv("PRIVATE_KEY")),
		client.EndpointOption(endpoint),
	)
	if err != nil {
		t.Fatal("new client err ", err.Error())
	}

	gameContractAddress := common.HexToAddress(contractAddress)
	instance, err := contracts.NewGameReplayContract(gameContractAddress, c.EthClient())
	if err != nil {
		t.Fatal("new contract instance err ", err.Error())
	}

	if instance == nil {
		t.Fatal("instance == nil")
	}

	length, err := instance.GetGameReplayLength(nil)
	if err != nil {
		t.Fatal("new client err ", err.Error())
	}

	t.Log("game replay length: ", length)

	for i := 0; i < int(length.Int64()); i++ {
		replay, err := instance.GetGameReplayByIndex(nil, big.NewInt(int64(i)))
		if err != nil {
			t.Fatal("GetGameReplayByIndex ", err)
		}

		buf, err := json.Marshal(replay)
		if err != nil {
			t.Fatal("Marshal ", err)
		}
		t.Log("game replay: ", string(buf))
		// t.Logf("gameID %s,replayCID %s, replayID %s, roundID %s", replay.GameInfo.GameID, replay.ReplayCID, replay.GameInfo.ReplayID, replay.GameInfo.RoundID)
	}

	// t.Log("game replay: ", replay)
}
