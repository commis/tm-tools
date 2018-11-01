package store

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	otp "github.com/commis/tm-tools/oldver/types"
	gco "github.com/tendermint/go-crypto"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"
)

type NodePrivList map[string]*NodePriv

const (
	StepNone      int8 = 0 // Used to distinguish the initial state
	StepPropose   int8 = 1
	StepPrevote   int8 = 2
	StepPrecommit int8 = 3
)

var (
	topHeightPv *privval.FilePV = nil
	//key: old validator address
	nodePvs = make(NodePrivList, 0)
)

type NodePriv struct {
	Index      int
	ChainID    string
	Address    cmn.HexBytes /*pv address*/
	PrivVal    *privval.FilePV
	OldPrivVal *otp.PrivValidatorFS
}

func (np *NodePriv) VoteToStep(voteType byte) int8 {
	switch voteType {
	case types.VoteTypePrevote:
		return StepPrevote
	case types.VoteTypePrecommit:
		return StepPrecommit
	default:
		cmn.PanicSanity("Unknown vote type")
		return 0
	}
}

func (np *NodePriv) SaveSigned(height int64, round int, step int8, signBytes []byte, sig []byte) {
	np.PrivVal.LastHeight = height
	np.PrivVal.LastRound = round
	np.PrivVal.LastStep = step
	np.PrivVal.LastSignature = sig
	np.PrivVal.LastSignBytes = signBytes
	np.PrivVal.Save()
}

func (np *NodePriv) SaveOldSigned(height int64, round int, step int8, signBytes []byte, sig gco.Signature) {
	np.OldPrivVal.LastHeight = height
	np.OldPrivVal.LastRound = round
	np.OldPrivVal.LastStep = step
	np.OldPrivVal.LastSignature = sig
	np.OldPrivVal.LastSignBytes = signBytes
	np.OldPrivVal.Save()
}

func (np *NodePriv) SignVote(chainID string, vote *types.Vote) {
	height, round, step := vote.Height, vote.Round, np.VoteToStep(vote.Type)
	signBytes := vote.SignBytes(chainID)

	sig, err := np.PrivVal.PrivKey.Sign(signBytes)
	if err != nil {
		cmn.Exit(fmt.Sprintf("failed to sign vote %v", err))
	}

	np.SaveSigned(height, round, step, signBytes, sig)
	vote.Signature = sig
}

func (np *NodePriv) SignProposal(chainID string, proposal *types.Proposal) {
	height, round, step := proposal.Height, proposal.Round, StepPropose
	signBytes := proposal.SignBytes(chainID)

	sig, err := np.PrivVal.PrivKey.Sign(signBytes)
	if err != nil {
		cmn.Exit(fmt.Sprintf("failed to sign proposal %v", err))
	}
	np.SaveSigned(height, round, step, signBytes, sig)
	proposal.Signature = sig
}

// Sort validators by address
type NodePrivByAddress []*NodePriv

func (ns NodePrivByAddress) Len() int {
	return len(ns)
}

func (ns NodePrivByAddress) Less(i, j int) bool {
	return bytes.Compare(ns[i].Address, ns[j].Address) == -1
}

func (ns NodePrivByAddress) Swap(i, j int) {
	it := ns[i]
	ns[i] = ns[j]
	ns[j] = it
}

func AddNodePriv(address, chainID string, pv *privval.FilePV, old *otp.PrivValidatorFS) {
	npv := &NodePriv{
		Index:      0,
		ChainID:    chainID,
		PrivVal:    pv,
		OldPrivVal: old,
	}

	if old != nil {
		npv.Address = old.Address.Bytes()
	} else if pv != nil {
		npv.Address = pv.Address.Bytes()
	}
	nodePvs[address] = npv
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

	//fmt.Println(nodePvs)
}

func GetNodePrivByAddress(address string) *NodePriv {
	pv, found := nodePvs[address]
	if !found {
		cmn.Exit(fmt.Sprintf("can't find PrivVal file %s", address))
	}
	return pv
}

func GetNodePrivByPeer(p2pID string) *NodePriv {
	address := strings.ToUpper(p2pID)
	return GetNodePrivByAddress(address)
}
