package test

import (
	"bytes"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	contracts "github.com/Filecoin-Titan/titan-game-sdk/contracts/api"
	"github.com/Filecoin-Titan/titan-game-sdk/contracts/client"
	"github.com/Filecoin-Titan/titan-game-sdk/vrf/filrpc"
	"github.com/Filecoin-Titan/titan-game-sdk/vrf/gamevrf"
	storage "github.com/Filecoin-Titan/titan-storage-sdk"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/filecoin-project/go-address"
	"github.com/google/uuid"
)

var (
	chainHeight = int64(1009859)

	replayData      = []byte{166, 151, 5, 61, 189, 201, 203, 69, 188, 20, 9, 50, 223, 153, 238, 59, 149, 71, 92, 205, 245, 57, 9, 168, 156, 163, 49, 215, 203, 159, 209, 245, 110, 78, 130, 62, 224, 136, 188, 64, 79, 245, 145, 21, 119, 13, 43, 8, 3, 231, 35, 65, 212, 42, 11, 44, 247, 146, 120, 206, 82, 252, 203, 131, 1, 13, 150, 229, 244, 12, 165, 170, 77, 27, 239, 148, 184, 106, 124, 46, 182, 222, 112, 241, 205, 168, 133, 58, 106, 104, 70, 68, 250, 70, 84, 27}
	contractAddress = "0x4b599a339A7b649C0fe641C2143dde42985602eD" // "0x5D7990C0487C57E3a0b57f2d3472600c37a5eE98"

	// nodeURL = "http://172.25.9.91:1251/rpc/v1"
	nodeURL = "https://api.calibration.node.glif.io/"

	sentCount     = 0
	receivedCount = 0
	gameInfoCount = 34
	messageCount  = 500
)

func TestOnGameServer(t *testing.T) {
	filPrivateKey := os.Getenv("FIL_PRIVATE_KEY")
	if len(filPrivateKey) == 0 {
		t.Fatal("Please set env FIL_PRIVATE_KEY")
	}

	gVRF := gamevrf.New(filrpc.NodeURLOption(nodeURL))

	var entropy []byte
	var gameRoundInfo = GameRoundInfo{
		GameID:    "abc-efg-hi",
		PlayerIDs: "a,b,c,d",
		RoundID:   uuid.NewString(),
		ReplayID:  uuid.NewString(),
	}

	buf := new(bytes.Buffer)
	err := gameRoundInfo.MarshalCBOR(buf)
	if err != nil {
		t.Fatal(err)
	}
	entropy = buf.Bytes()

	privateKey, err := gamevrf.FilBlsKeyFromString(filPrivateKey)
	if err != nil {
		t.Fatal("FilBlsKeyFromString error ", err)
	}

	vrfout, err := gVRF.GenerateVRF(gamevrf.DomainSeparationTag_GameBasic, privateKey, entropy)
	if err != nil {
		t.Fatal(err)
	}

	cid, err := storage.CalculateCid(bytes.NewReader(replayData))
	if err != nil {
		t.Fatal(err)
	}

	filPublicKey, err := gamevrf.FilBlsKey2PublicKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	addr, err := address.NewBLSAddress(filPublicKey)
	if err != nil {
		t.Fatal(err)
	}

	// playing Game, and generate game result
	replay := &contracts.GameRoundReplay{
		DomainSeparationTag: int64(gamevrf.DomainSeparationTag_GameBasic),
		VRFHeight:           uint64(chainHeight),
		HashFunc:            "blake2b",
		VRFProof:            vrfout.Proof,
		Address:             addr.String(),
		ReplayCID:           cid.String(),
		GameInfo:            gameRoundInfoToContractGameRoundInfo(gameRoundInfo),
		GameResults:         generateGameResults(&gameRoundInfo),
	}

	if err = sendContract(replay); err != nil {
		t.Fatal(err)
	}
}

func gameRoundInfoToContractGameRoundInfo(gameRoundInfo GameRoundInfo) contracts.GameRoundInfo {
	return contracts.GameRoundInfo{
		GameID:    gameRoundInfo.GameID,
		RoundID:   gameRoundInfo.RoundID,
		ReplayID:  gameRoundInfo.ReplayID,
		PlayerIDs: gameRoundInfo.PlayerIDs,
	}
}

func generateGameResults(gameRoundInfo *GameRoundInfo) []contracts.GameRoundResult {
	playerIDs := gameRoundInfo.PlayerIDs
	players := strings.Split(playerIDs, ",")

	results := make([]contracts.GameRoundResult, 0, len(players))
	for _, player := range players {
		result := contracts.GameRoundResult{
			PlayerID:     player,
			CurrentScore: 100,
			WinScore:     1,
		}
		results = append(results, result)
	}
	return results
}

// You have to deploy the contract before you can do that.
func sendContract(replay *contracts.GameRoundReplay) error {
	ethPrivateKey := os.Getenv("ETH_PRIVATE_KEY")
	if len(ethPrivateKey) == 0 {
		return fmt.Errorf("Please set env ETH_PRIVATE_KEY")
	}

	c, err := client.New(
		client.PrivateKeyOption(ethPrivateKey),
		client.EndpointOption(nodeURL),
	)
	if err != nil {
		return err
	}

	// replayID := uuid.NewString()
	gameContractAddress := common.HexToAddress(contractAddress)
	instance, err := contracts.NewGameReplayContract(gameContractAddress, c.EthClient())
	if err != nil {
		return err
	}

	result, err := c.InvokeContract(0, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return instance.SaveGameReplay(opts, []contracts.GameRoundReplay{*replay})
	})
	if err != nil {
		return err
	}

	fmt.Println("replay id: ", replay.GameInfo.ReplayID)
	fmt.Println("save game replay OK: ", string(result))
	fmt.Println("querying game replay...")

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-time.After(2 * time.Minute):
			return fmt.Errorf("query order timeout!")
		case <-ticker.C:
			order, err := instance.GetGameReplay(&bind.CallOpts{
				Pending: true,
			}, replay.GameInfo.ReplayID)
			if err != nil {
				fmt.Println("get game replay err ", err.Error())
				continue
			}

			fmt.Println("Query game replay OK: ", order)
			return nil
		}
	}
}

func TestOnGameUser(t *testing.T) {
	ethPrivateKey := os.Getenv("ETH_PRIVATE_KEY")
	if len(ethPrivateKey) == 0 {
		t.Fatal("Please set env ETH_PRIVATE_KEY")
	}

	gVRF := gamevrf.New(filrpc.NodeURLOption(nodeURL))

	c, err := client.New(
		client.PrivateKeyOption(ethPrivateKey),
		client.EndpointOption(nodeURL),
	)
	if err != nil {
		t.Fatal("new client err ", err.Error())
	}

	instance, err := contracts.NewGameReplayContract(common.HexToAddress(contractAddress), c.EthClient())
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

	end := int64(0)
	testNum := 10
	if length.Int64() > int64(testNum) {
		end = length.Int64() - int64(testNum)
	}

	for i := length.Int64() - 1; i >= end; i-- {
		replay, err := instance.GetGameReplayByIndex(nil, big.NewInt(int64(i)))
		if err != nil {
			t.Fatal("GetGameReplayByIndex ", err)
		}

		addr, err := address.NewFromString(replay.Address)
		if err != nil {
			t.Error("replay.Address ", replay.Address)
		}

		gameRoundInfo := contractGameRoundInfoToGameRoundInfo(&replay.GameInfo)
		buf := new(bytes.Buffer)
		err = gameRoundInfo.MarshalCBOR(buf)
		if err != nil {
			t.Fatal(err)
		}
		entropy := buf.Bytes()

		vrfout := &gamevrf.VRFOut{Height: replay.VRFHeight, Proof: replay.VRFProof}
		err = gVRF.VerifyVRF(gamevrf.DomainSeparationTag(replay.DomainSeparationTag), addr, entropy, vrfout)
		if err != nil {
			t.Fatal(err)
		}

	}

}

func contractGameRoundInfoToGameRoundInfo(roundInfo *contracts.GameRoundInfo) GameRoundInfo {
	return GameRoundInfo{
		GameID:    roundInfo.GameID,
		RoundID:   roundInfo.RoundID,
		ReplayID:  roundInfo.ReplayID,
		PlayerIDs: roundInfo.PlayerIDs,
	}
}

func TestSaveGameReplays(t *testing.T) {
	fClient := filrpc.New(
		filrpc.NodeURLOption(nodeURL),
	)

	tps, err := fClient.ChainHead()
	if err != nil {
		fmt.Println(err)
		return
	}

	var entropy []byte
	gameRoundInfo := GameRoundInfo{
		GameID:    "abc-efg-hi",
		PlayerIDs: "a,b,c,d",
		RoundID:   "gogogogo1",
		ReplayID:  "bilibili",
	}

	buf := new(bytes.Buffer)
	err = gameRoundInfo.MarshalCBOR(buf)
	if err != nil {
		fmt.Println(err)
		return
	}
	entropy = buf.Bytes()

	privateKey, err := gamevrf.FilBlsKeyFromString(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		t.Fatal("FilBlsKeyFromString error ", err)
	}

	vrfout, err := gamevrf.FilGenerateVRFByTipSet(gamevrf.DomainSeparationTag_GameBasic, privateKey, tps, entropy)
	if err != nil {
		fmt.Println(err)
		return
	}

	cid, err := storage.CalculateCid(bytes.NewReader(replayData))
	if err != nil {
		fmt.Println(err)
		return
	}

	filPublicKey, err := gamevrf.FilBlsKey2PublicKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	addr, err := address.NewBLSAddress(filPublicKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	// playing Game, and generate game resulPrintln(a)
	replay := &contracts.GameRoundReplay{
		DomainSeparationTag: int64(gamevrf.DomainSeparationTag_GameBasic),
		VRFHeight:           uint64(chainHeight),
		HashFunc:            "blake2b",
		VRFProof:            vrfout.Proof,
		Address:             addr.String(),
		ReplayCID:           cid.String(),
		GameInfo:            gameRoundInfoToContractGameRoundInfo(gameRoundInfo),
		GameResults:         generateGameResults(&gameRoundInfo),
	}

	c, err := client.New(
		client.PrivateKeyOption(os.Getenv("PRIVATE_KEY")),
		client.EndpointOption(nodeURL),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	nonce, err := c.Nonce()
	if err != nil {
		return
	}

	var gameMap sync.Map

	go watchMessage(&gameMap, c)

	for i := 0; i < messageCount; i++ {
		n := nonce

		key := fmt.Sprintf("r_%d", n)
		gameMap.Store(key, nil)

		saveGameReplyWithContract2(n, *replay)
		nonce++
	}

	select {}
}

func watchMessage(gameMap *sync.Map, c client.Client) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	gameContractAddress := common.HexToAddress(contractAddress)
	instance, err := contracts.NewGameReplayContract(gameContractAddress, c.EthClient())
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		select {
		case <-time.After(2 * time.Minute):
			return
		case <-ticker.C:
			gameMap.Range(func(key, value interface{}) bool {
				replayID := key.(string)
				_, err := instance.GetGameReplay(&bind.CallOpts{
					Pending: true,
				}, replayID)
				if err == nil {
					// 	fmt.Println(replay.GameInfo.ReplayID, " get game replay err ", err.Error())
					// } else {
					receivedCount++
					fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " Query game replay OK: ", replayID, " receivedCount: ", receivedCount)

					gameMap.Delete(replayID)
				}

				return true
			})

		}
	}
}

// You have to deploy the contract before you can do that.
func saveGameReplyWithContract2(nonce uint64, replay contracts.GameRoundReplay) error {
	// fmt.Println("send nonce :", nonce)
	c, err := client.New(
		client.PrivateKeyOption(os.Getenv("PRIVATE_KEY")),
		client.EndpointOption(nodeURL),
	)
	if err != nil {
		fmt.Println("err :", err)
		return err
	}

	// replayID := uuid.NewString()
	gameContractAddress := common.HexToAddress(contractAddress)
	instance, err := contracts.NewGameReplayContract(gameContractAddress, c.EthClient())
	if err != nil {
		fmt.Println("NewGameReplayContract :", err)
		return err
	}

	replay.GameInfo.ReplayID = fmt.Sprintf("r_%d", nonce)

	list := make([]contracts.GameRoundReplay, 0)
	for i := 0; i < gameInfoCount; i++ {
		list = append(list, replay)
	}

	_, err = c.InvokeContract(0, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		result, err := instance.SaveGameReplay(opts, list)
		if err != nil {
			fmt.Println("SaveGameReplay :", err)
			return nil, err
		}
		fmt.Println("querying Hash : ", result.Hash())
		return result, nil
	})
	if err != nil {
		fmt.Println("InvokeContract :", err)
		return err
	}

	sentCount++
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " send nonce :", nonce, " sentCount: ", sentCount)

	return nil
}

func TestPrivate2Address(t *testing.T) {
	filPrivateKey, err := gamevrf.FilBlsKeyFromString(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		t.Fatal("FilBlsKeyFromString error ", err)
	}

	publicKey, err := gamevrf.FilBlsKey2PublicKey(filPrivateKey)
	if err != nil {
		t.Fatal("FilBlsKey2PublicKey error ", err)
	}

	addr, err := address.NewBLSAddress(publicKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("address ", addr.String())
}
