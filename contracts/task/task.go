package task

import (
	"fmt"
	"os"
	"sync"
	"time"

	contracts "github.com/Filecoin-Titan/titan-game-sdk/contracts/api"
	"github.com/Filecoin-Titan/titan-game-sdk/contracts/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const uploadBatch = 1

type ContractConfig struct {
	PrivateKey      string
	FilNodeURL      string
	ContractAddress string
	UploadBatch     int
}

type Contract contracts.GameRoundReplay

type Task struct {
	config      *ContractConfig
	contractCli client.Client
	contracts   []*Contract
	lock        *sync.Mutex
}

func NewTask(config *ContractConfig) (*Task, error) {
	if len(config.PrivateKey) == 0 {
		config.PrivateKey = os.Getenv("PRIVATE_KEY")
	}
	if len(config.FilNodeURL) == 0 {
		config.FilNodeURL = os.Getenv("NODE_URL")
	}
	if len(config.ContractAddress) == 0 {
		config.ContractAddress = os.Getenv("CONTRACT_ADDRESS")
	}

	if config.UploadBatch == 0 {
		config.UploadBatch = uploadBatch
	}

	privateKeyOption := client.PrivateKeyOption(config.PrivateKey)
	endpointOption := client.EndpointOption(config.FilNodeURL)

	c, err := client.New(privateKeyOption, endpointOption)
	if err != nil {
		return nil, err
	}

	t := &Task{config: config, contractCli: c, contracts: make([]*Contract, 0), lock: &sync.Mutex{}}
	go t.run()

	return t, nil
}

func (t *Task) AddContract(c *Contract) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.contracts = append(t.contracts, c)
}

// pop the maximum number of contracts is UploadBatch
func (t *Task) popContracts() []*Contract {
	t.lock.Lock()
	defer t.lock.Unlock()

	if len(t.contracts) == 0 {
		return nil
	}

	if len(t.contracts) <= t.config.UploadBatch {
		contracts := t.contracts
		t.contracts = make([]*Contract, 0)
		return contracts

	}

	contracts := t.contracts[0:t.config.UploadBatch]
	t.contracts = t.contracts[t.config.UploadBatch:]
	return contracts
}

func (t *Task) run() {
	for {
		for len(t.contracts) > 0 {
			contracts := t.popContracts()
			if len(contracts) == 0 {
				fmt.Println("pop contract == nil")
				continue
			}

			if err := t.sendContracts(contracts...); err != nil {
				fmt.Printf("sendContracts error %s", err)
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func (t *Task) sendContracts(cs ...*Contract) error {
	gameReplays := make([]contracts.GameRoundReplay, 0, len(cs))
	for _, c := range cs {
		gameReplays = append(gameReplays, contracts.GameRoundReplay(*c))
	}

	gameContractAddress := common.HexToAddress(t.config.ContractAddress)
	instance, err := contracts.NewGameReplayContract(gameContractAddress, t.contractCli.EthClient())
	if err != nil {
		return err
	}

	_, err = t.contractCli.InvokeContract(0, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return instance.SaveGameReplay(opts, gameReplays)
	})

	fmt.Printf("invoke contract error= %#v", err)

	return err
}
