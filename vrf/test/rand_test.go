package test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/Filecoin-Titan/titan-game-sdk/vrf/filrpc"
	"github.com/Filecoin-Titan/titan-game-sdk/vrf/gamevrf"
	"github.com/Filecoin-Titan/titan-game-sdk/vrf/trand"

	"github.com/filecoin-project/go-address"
)

var (
	chainHeight = int64(3303964)
	// filecoin bls private key
	filPrivateKey = []byte{152, 112, 35, 145, 21, 31, 99, 206, 204, 113, 33, 99, 241, 180, 157, 194, 91, 224, 34, 186, 137, 10, 136, 38, 133, 32, 109, 255, 59, 81, 45, 26}
	// filecoin bls public key
	filPublicKey = []byte{146, 209, 52, 147, 166, 127, 130, 148, 172, 13, 162, 254, 17, 85, 254, 151, 93, 182, 28, 218, 103, 106, 200, 115, 178, 101, 156, 74, 25, 214, 220, 136, 167, 32, 147, 231, 40, 250, 149, 109, 229, 58, 7, 135, 214, 93, 55, 169}

	filProof = []byte{166, 151, 5, 61, 189, 201, 203, 69, 188, 20, 9, 50, 223, 153, 238, 59, 149, 71, 92, 205, 245, 57, 9, 168, 156, 163, 49, 215, 203, 159, 209, 245, 110, 78, 130, 62, 224, 136, 188, 64, 79, 245, 145, 21, 119, 13, 43, 8, 3, 231, 35, 65, 212, 42, 11, 44, 247, 146, 120, 206, 82, 252, 203, 131, 1, 13, 150, 229, 244, 12, 165, 170, 77, 27, 239, 148, 184, 106, 124, 46, 182, 222, 112, 241, 205, 168, 133, 58, 106, 104, 70, 68, 250, 70, 84, 27}
)

func TestVRFGenVerify(t *testing.T) {
	nodeURL := "http://api.node.glif.io/rpc/v1"

	client := filrpc.New(
		filrpc.NodeURLOption(nodeURL),
	)

	tps, err := client.ChainGetTipSetByHeight(chainHeight)
	if err != nil {
		t.Fatal(err)
	}

	privateKey := filPrivateKey
	publicKey := filPublicKey

	var entropy []byte
	var gameRoundInfo = GameRoundInfo{
		GameID:    "abc-efg-hi",
		PlayerIDs: "a,b,c,d",
		RoundID:   "gogogogo1",
		ReplayID:  "bilibili",
	}

	buf := new(bytes.Buffer)
	err = gameRoundInfo.MarshalCBOR(buf)
	if err != nil {
		t.Fatal(err)
	}
	entropy = buf.Bytes()

	vrfout, err := gamevrf.FilGenerateVRFByTipSet(gamevrf.DomainSeparationTag_GameBasic, privateKey, tps, entropy)
	if err != nil {
		t.Fatal(err)
	}

	for i, p := range vrfout.Proof {
		if p != filProof[i] {
			t.Fatalf("proof not equal %d != %d, pos:%d", p, filProof[i], i)
		}
	}

	addr, err := address.NewBLSAddress(publicKey)
	if err != nil {
		t.Fatal(err)
	}

	err = gamevrf.FilVerifyVRFByTipSet(gamevrf.DomainSeparationTag_GameBasic, addr, tps, entropy, vrfout)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVRFGenVerify2(t *testing.T) {
	nodeURL := "http://api.node.glif.io/rpc/v1"

	client := filrpc.New(
		filrpc.NodeURLOption(nodeURL),
	)

	tps, err := client.ChainGetTipSetByHeight(chainHeight)
	if err != nil {
		t.Fatal(err)
	}

	publicKey := filPublicKey

	var entropy []byte
	var gameRoundInfo = GameRoundInfo{
		GameID:    "abc-efg-hi",
		PlayerIDs: "a,b,c,d",
		RoundID:   "gogogogo1",
		ReplayID:  "bilibili",
	}

	buf := new(bytes.Buffer)
	err = gameRoundInfo.MarshalCBOR(buf)
	if err != nil {
		t.Fatal(err)
	}
	entropy = buf.Bytes()

	vrfout := &gamevrf.VRFOut{
		Height: uint64(chainHeight),
		Proof:  filProof,
	}

	addr, err := address.NewBLSAddress(publicKey)
	if err != nil {
		t.Fatal(err)
	}

	err = gamevrf.FilVerifyVRFByTipSet(gamevrf.DomainSeparationTag_GameBasic, addr, tps, entropy, vrfout)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVRFGenVerify3(t *testing.T) {
	nodeURL := "http://api.node.glif.io/rpc/v1"

	gg := gamevrf.New(filrpc.NodeURLOption(nodeURL))

	privateKey := filPrivateKey
	publicKey := filPublicKey

	var entropy []byte
	var gameRoundInfo = GameRoundInfo{
		GameID:    "abc-efg-hi",
		PlayerIDs: "a,b,c,d",
		RoundID:   "gogogogo1",
		ReplayID:  "bilibili",
	}

	buf := new(bytes.Buffer)
	err := gameRoundInfo.MarshalCBOR(buf)
	if err != nil {
		t.Fatal(err)
	}
	entropy = buf.Bytes()

	vrfout, err := gg.GenerateVRF(gamevrf.DomainSeparationTag_GameBasic, privateKey, entropy)
	if err != nil {
		t.Fatal(err)
	}

	addr, err := address.NewBLSAddress(publicKey)
	if err != nil {
		t.Fatal(err)
	}

	err = gg.VerifyVRF(gamevrf.DomainSeparationTag_GameBasic, addr, entropy, vrfout)
	if err != nil {
		t.Fatal(err)
	}

	var sb strings.Builder
	rng := trand.NewRng(vrfout.Sum256(), trand.RNGType_Normal)
	for i := 0; i < 10; i++ {
		sb.WriteString(fmt.Sprintf("%d,", rng.Intn(100)))
	}
	t.Logf("RNGType_Normal: %s", sb.String())

	var sb2 strings.Builder
	rng2 := trand.NewRng(vrfout.Sum256(), trand.RNGType_Cipher)
	for i := 0; i < 10; i++ {
		sb2.WriteString(fmt.Sprintf("%d,", rng2.Intn(100)))
	}
	t.Logf("RNGType_Cipher: %s", sb2.String())
}
