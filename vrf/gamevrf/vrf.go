package gamevrf

import (
	"sync"
	"time"

	"github.com/Filecoin-Titan/titan-game-sdk/vrf/filrpc"

	"github.com/filecoin-project/go-address"
	"golang.org/x/xerrors"
)

const (
	// FILECOIN_EPOCH_DURATION represents the duration of a Filecoin epoch in seconds
	FILECOIN_EPOCH_DURATION = 30
	// GAME_CHAIN_EPOCH_LOOKBACK represents the number of epochs to look back when fetching tipsets for the game chain
	GAME_CHAIN_EPOCH_LOOKBACK = 10
)

// GameVRF represents a VRF implementation for the game
type GameVRF struct {
	rpcOptions []filrpc.Option

	lck             sync.Mutex
	isCacheValid    bool // use cache to reduce 'ChainHead' calls
	cachedEpoch     uint64
	cachedTimestamp time.Time
}

// New creates a new instance of GameVRF with the specified RPC options
func New(options ...filrpc.Option) *GameVRF {
	return &GameVRF{
		rpcOptions: options,
	}
}

// getTipsetByHeight retrieves a non-empty tipset at the specified height or within the lookback window
func (g *GameVRF) getTipsetByHeight(height uint64) (*filrpc.TipSet, error) {
	client := filrpc.New(g.rpcOptions...)

	iheight := int64(height)
	for i := 0; i < GAME_CHAIN_EPOCH_LOOKBACK && iheight > 0; i++ {
		tps, err := client.ChainGetTipSetByHeight(iheight)
		if err != nil {
			return nil, err
		}

		if len(tps.Blocks()) > 0 {
			return tps, nil
		}

		iheight--
	}

	return nil, xerrors.Errorf("getTipsetByHeight can't found a non-empty tipset from height: %d", height)
}

// getChainHead retrieves the current chain head height
func (g *GameVRF) getChainHead() (uint64, error) {
	client := filrpc.New(g.rpcOptions...)
	tps, err := client.ChainHead()
	if err != nil {
		return 0, xerrors.Errorf("getChainHead ChainHead call failed: %w", err)
	}

	return tps.Height(), nil
}

// ForceUpdateCachedEpoch forces an update of the cached epoch and returns the new epoch
func (g *GameVRF) ForceUpdateCachedEpoch() (uint64, error) {
	g.lck.Lock()
	defer g.lck.Unlock()

	g.isCacheValid = false
	g.cachedTimestamp = time.Now()
	h, err := g.getChainHead()
	if err != nil {
		return 0, err
	}

	g.cachedEpoch = h
	g.isCacheValid = true

	return h, nil
}

// getGameEpoch retrieves the current game epoch, updating the cache if necessary
func (g *GameVRF) getGameEpoch() (uint64, error) {
	g.lck.Lock()
	defer g.lck.Unlock()

	if !g.isCacheValid {
		g.cachedTimestamp = time.Now()
		h, err := g.getChainHead()
		if err != nil {
			return 0, err
		}

		g.cachedEpoch = h
		g.isCacheValid = true
	}

	duration := time.Since(g.cachedTimestamp)
	if duration < 0 {
		return 0, xerrors.Errorf("current time is not correct with negative duration: %s", duration)
	}

	elapseEpoch := int64(duration.Seconds()) / FILECOIN_EPOCH_DURATION

	return g.cachedEpoch + uint64(elapseEpoch), nil
}

// GenerateVRF generates a VRF output given the domain separation tag, Filecoin BLS private key, and entropy
func (g *GameVRF) GenerateVRF(pers DomainSeparationTag, filBlsPrivateKey []byte, entropy []byte) (*VRFOut, error) {
	height, err := g.getGameEpoch()
	if err != nil {
		return nil, xerrors.Errorf("GenerateVRF getGameEpoch failed: %w", err)
	}

	if height <= GAME_CHAIN_EPOCH_LOOKBACK {
		return nil, xerrors.Errorf("GenerateVRF getGameEpoch return invalid height: %d", height)
	}

	lookback := height - GAME_CHAIN_EPOCH_LOOKBACK
	tps, err := g.getTipsetByHeight(lookback)
	if err != nil {
		return nil, xerrors.Errorf("GenerateVRF getTipsetByHeight failed: %w", err)
	}

	return FilGenerateVRFByTipSet(pers, filBlsPrivateKey, tps, entropy)
}

// VerifyVRF verifies a VRF output given the domain separation tag, worker address, entropy, and the VRF output
func (g *GameVRF) VerifyVRF(pers DomainSeparationTag, worker address.Address, entropy []byte, vrf *VRFOut) error {
	tps, err := g.getTipsetByHeight(vrf.Height)
	if err != nil {
		return xerrors.Errorf("VerifyVRF getTipsetByHeight failed: %w", err)
	}

	return FilVerifyVRFByTipSet(pers, worker, tps, entropy, vrf)
}
