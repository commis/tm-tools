package store

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/tendermint/tendermint/types"

	"github.com/tendermint/tendermint/privval"
	cmn "github.com/tendermint/tmlibs/common"
)

type NodePrivList map[string]*NodePriv

const (
	stepNone      int8 = 0 // Used to distinguish the initial state
	stepPropose   int8 = 1
	stepPrevote   int8 = 2
	stepPrecommit int8 = 3
)

var (
	topHeightPv *privval.FilePV = nil
	//key: old validator address
	nodePvs = make(NodePrivList, 0)
)

type NodePriv struct {
	Index   int
	ChainID string
	PrivVal *privval.FilePV
}

func (np *NodePriv) voteToStep(vote *types.Vote) int8 {
	switch vote.Type {
	case types.VoteTypePrevote:
		return stepPrevote
	case types.VoteTypePrecommit:
		return stepPrecommit
	default:
		cmn.PanicSanity("Unknown vote type")
		return 0
	}
}

func (np *NodePriv) saveSigned(height int64, round int, step int8, signBytes []byte, sig []byte) {
	np.PrivVal.LastHeight = height
	np.PrivVal.LastRound = round
	np.PrivVal.LastStep = step
	np.PrivVal.LastSignature = sig
	np.PrivVal.LastSignBytes = signBytes
	np.PrivVal.Save()
}

func (np *NodePriv) SignVote(chainID string, vote *types.Vote) {
	height, round, step := vote.Height, vote.Round, np.voteToStep(vote)
	signBytes := vote.SignBytes(chainID)

	sig, err := np.PrivVal.PrivKey.Sign(signBytes)
	if err != nil {
		cmn.Exit(fmt.Sprintf("failed to sign vote %v", err))
	}

	np.saveSigned(height, round, step, signBytes, sig)
	vote.Signature = sig
}

func (np *NodePriv) SignProposal(chainID string, proposal *types.Proposal) {
	height, round, step := proposal.Height, proposal.Round, stepPropose
	signBytes := proposal.SignBytes(chainID)

	sig, err := np.PrivVal.PrivKey.Sign(signBytes)
	if err != nil {
		cmn.Exit(fmt.Sprintf("failed to sign proposal %v", err))
	}
	np.saveSigned(height, round, step, signBytes, sig)
	proposal.Signature = sig
}

// Sort validators by address
type NodePrivByAddress []*NodePriv

func (ns NodePrivByAddress) Len() int {
	return len(ns)
}

func (ns NodePrivByAddress) Less(i, j int) bool {
	return bytes.Compare(ns[i].PrivVal.Address, ns[j].PrivVal.Address) == -1
}

func (ns NodePrivByAddress) Swap(i, j int) {
	it := ns[i]
	ns[i] = ns[j]
	ns[j] = it
}

func AddNodePriv(address, chainID string, pv *privval.FilePV) {
	if pv != nil {
		npv := &NodePriv{
			Index:   0,
			ChainID: chainID,
			PrivVal: pv,
		}
		nodePvs[address] = npv
	}
}

func SortNodePriv() {
	pvList := make([]*NodePriv, 0, len(nodePvs))
	for _, v := range nodePvs {
		pvList = append(pvList, v)
	}
	sort.Sort(NodePrivByAddress(pvList))

	for idx, v := range pvList {
		v.Index = idx
	}
}

func GetNodePrivByAddress(address string) *NodePriv {
	pv, found := nodePvs[address]
	if !found {
		cmn.Exit("can't find PrivVal file")
	}
	return pv
}

func GetNodePrivByPeer(p2pID string) *NodePriv {
	address := strings.ToUpper(p2pID)
	return GetNodePrivByAddress(address)
}
