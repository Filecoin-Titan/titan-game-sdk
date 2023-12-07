# titan game contracts
Uploading basic information about the game, game results, certifiable random numbers
## Build contracts

### install solcjs

    npm install -g solcjs

### build contracts to abi and bin
    solcjs contracts/GameReplay.sol --output-dir ./build --bin --abi

### generate go from abi and bin
    go install github.com/ethereum/go-ethereum/cmd/abigen@latest
    abigen --abi build/contracts_GameReplay_sol_GameReplayContract.abi --bin build/contracts_GameReplay_sol_GameReplayContract.bin --pkg contracts --type GameReplayContract --out ./api/game_replay.go

### deploy contract
Here it is recommended to use proxy to call the following contract, so as not to lead to the back can not be updated, the specific program please refer to the [official](https://ethereum.org/en/developers/docs/smart-contracts/upgrading/) documentation

    c, err := client.New(
		client.PrivateKeyOption(os.Getenv("PRIVATE_KEY")),
		client.EndpointOption(endpoint),
	)
	if err != nil {
		log.Fatal("new client err", err.Error())
	}

	result, err := c.InvokeContract(0, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		addr, tr, _, err := contracts.DeployGameReplayContract(opts, c.EthClient())
		if err != nil {
			return nil, err
		}

		log.Printf("deploy contract %s", addr.Hex())
		return tr, nil
	})
	if err != nil {
		log.Fatal("deploy contracts err ", err.Error())
	}

	log.Println("deploy OK: ", string(result))

### invoker contract
    contractAddress := "YOUR-CONTRACT-ADDRESS"
    c, err := client.New(
		client.PrivateKeyOption(os.Getenv("PRIVATE_KEY")),
		client.EndpointOption(endpoint),
	)
	if err != nil {
		log.Fatal("new client err ", err.Error())
	}

	replayID := uuid.NewString()
	gameContractAddress := common.HexToAddress(contractAddress)
	instance, err := contracts.NewGameReplayContract(gameContractAddress, c.EthClient())
	if err != nil {
		log.Fatal("new contract instance err ", err.Error())
	}

	result, err := c.InvokeContract(0, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		results := make([]contracts.GameRoundResult, 0, 4)
		
        gameReplay = contracts.GameRoundReplay{
            // TODO: please implement GameRoundReplay
        }
		return instance.SaveGameReplay(opts, []contracts.GameRoundReplay{gameReplay})
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("save game replay OK: ", string(result))
