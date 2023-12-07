# titan-game-sdk
A Trustworthy Proof Library for Gaming Applications on TitanNetwork

**Overview**
titan-game-sdk is a specialized Software Development Kit (SDK) tailored for TitanNetwork, aimed at providing verifiable proofs for gaming applications. The kit utilizes Verifiable Random Functions (VRF), on-chain traceability, and oracle technology to ensure both fairness and transparency in the gaming experience.

**Key Features**
Leverages VRF for secure and unbiased random number generation.
Enables on-chain auditability for game sessions to retrospectively validate fairness.
Incorporates oracle services for the acquisition and verification of off-chain data.

By combining these advanced blockchain technologies, titan-game-sdk provides a robust foundation for developers seeking to integrate trust, transparency, and fairness into their gaming platforms on TitanNetwork.

### Generating VRF　and verify VRF 
	nodeURL := "https://api.calibration.node.glif.io/"

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

    privateKey, err := gamevrf.FilBlsKeyFromString(os.Getenv("FIL_PRIVATE_KEY"))
	if err != nil {
		t.Fatal("FilBlsKeyFromString error ", err)
	}

	vrfout, err := gVRF.GenerateVRF(gamevrf.DomainSeparationTag_GameBasic, privateKey, entropy)
	if err != nil {
		t.Fatal(err)
	}

	publicKey, err := gamevrf.FilBlsKey2PublicKey(privateKey)
	if err != nil {
		return err
	}

	addr, err := address.NewBLSAddress(publicKey)
	if err != nil {
		return err
	}

    err = gVRF.VerifyVRF(gamevrf.DomainSeparationTag_GameBasic, addr, entropy, vrfout)
	if err != nil {
		t.Fatal(err)
	}

### Upload game data to blockchain
    nodeURL = "https://api.calibration.node.glif.io/"
	contractAddress = "YOUR_CONTRACT_ADDRESS"

    c, err := client.New(
		client.PrivateKeyOption(os.Getenv("ETH_PRIVATE_KEY")),
		client.EndpointOption(nodeURL),
	)
	if err != nil {
		return err
	}

	gameContractAddress := common.HexToAddress(contractAddress)
	instance, err := contracts.NewGameReplayContract(gameContractAddress, c.EthClient())
	if err != nil {
		return err
	}

	result, err := c.InvokeContract(0, func(opts *bind.TransactOpts) (*types.Transaction, error) {
        replay = contracts.GameRoundReplay{
            // TODO: please implement GameRoundReplay
        }
		return instance.SaveGameReplay(opts, []contracts.GameRoundReplay{replay})
	})
	if err != nil {
		return err
	}