package filrpc

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/ipfs/go-cid"
	"github.com/minio/blake2b-simd"
	"golang.org/x/xerrors"
)

// TipSet represents a set of blocks at the same height in the blockchain
type TipSet struct {
	cids   []cid.Cid
	blks   []*BlockHeader
	height uint64
}

// ExpTipSet is an exported version of TipSet for JSON marshaling
type ExpTipSet struct {
	Cids   []cid.Cid
	Blocks []*BlockHeader
	Height uint64
}

// Ticket represents the VRF proof associated with a block
type Ticket struct {
	VRFProof []byte
}

// BlockHeader represents the header of a block in the blockchain
type BlockHeader struct {
	Ticket *Ticket // 1 unique per block/miner: should be a valid VRF
	//ParentWeight          BigInt            // 6 identical for all blocks in same tipset
	Height uint64 // 7 identical for all blocks in same tipset
}

// LastTicket returns the VRF proof associated with the block
func (blk *BlockHeader) LastTicket() *Ticket {
	return blk.Ticket
}

// MarshalJSON serializes a TipSet to JSON format
func (ts *TipSet) MarshalJSON() ([]byte, error) {
	// why didnt i just export the fields? Because the struct has methods with the
	// same names already
	return json.Marshal(ExpTipSet{
		Cids:   ts.cids,
		Blocks: ts.blks,
		Height: ts.height,
	})
}

// UnmarshalJSON deserializes a TipSet from JSON format
func (ts *TipSet) UnmarshalJSON(b []byte) error {
	var ets ExpTipSet
	if err := json.Unmarshal(b, &ets); err != nil {
		return err
	}

	ots, err := NewTipSet(ets.Blocks)
	if err != nil {
		return err
	}

	*ts = *ots

	return nil
}

// tipsetSortFunc returns a sorting function based on the VRF proof of blocks
func tipsetSortFunc(blks []*BlockHeader) func(i, j int) bool {
	return func(i, j int) bool {
		ti := blks[i].LastTicket()
		tj := blks[j].LastTicket()

		// if ti.Equals(tj) {
		// 	return bytes.Compare(blks[i].Cid().Bytes(), blks[j].Cid().Bytes()) < 0
		// }

		return ti.Less(tj)
	}
}

// NewTipSet creates a new TipSet from the given blocks
func NewTipSet(blks []*BlockHeader) (*TipSet, error) {
	if len(blks) == 0 {
		return nil, xerrors.Errorf("NewTipSet called with zero length array of blocks")
	}

	sort.Slice(blks, tipsetSortFunc(blks))

	var ts TipSet
	ts.blks = blks

	ts.height = blks[0].Height

	return &ts, nil
}

// MinTicket returns the minimum VRF proof in the TipSet
func (ts *TipSet) MinTicket() *Ticket {
	return ts.MinTicketBlock().Ticket
}

// MinTicketBlock returns the block with the minimum VRF proof in the TipSet
func (ts *TipSet) MinTicketBlock() *BlockHeader {
	blks := ts.Blocks()

	min := blks[0]

	for _, b := range blks[1:] {
		if b.LastTicket().Less(min.LastTicket()) {
			min = b
		}
	}

	return min
}

// Blocks returns the blocks in the TipSet
func (ts *TipSet) Blocks() []*BlockHeader {
	return ts.blks
}

// Height returns the height of the TipSet
func (ts *TipSet) Height() uint64 {
	return ts.height
}

// Less compares two VRF proofs and returns true if the first is less than the second
func (t *Ticket) Less(o *Ticket) bool {
	tDigest := blake2b.Sum256(t.VRFProof)
	oDigest := blake2b.Sum256(o.VRFProof)
	return bytes.Compare(tDigest[:], oDigest[:]) < 0
}

// Equals compares two VRF proofs and returns true if they are equal
func (t *Ticket) Equals(ot *Ticket) bool {
	return bytes.Equal(t.VRFProof, ot.VRFProof)
}
