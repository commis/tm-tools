package consensus

import (
	"github.com/commis/tm-tools/libs/op/hold"
	"github.com/commis/tm-tools/libs/op/store"
	oct "github.com/commis/tm-tools/oldver/consensus/types"
	otp "github.com/commis/tm-tools/oldver/types"
	"github.com/tendermint/tendermint/consensus"
	cstypes "github.com/tendermint/tendermint/consensus/types"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/types"
)

func FilterWalMessage(height int64, msg WALMessage) bool {
	switch m := msg.(type) {
	case EndHeightMessage:
		return m.Height >= height
	case otp.EventDataRoundState:
		if m.Step == oct.RoundStepNewHeight.String() {
			return m.Height > height
		}
		return m.Height >= height
	case timeoutInfo:
		return m.Height >= height
	case msgInfo:
		switch mi := m.Msg.(type) {
		case *BlockPartMessage:
			return mi.Height >= height
		case *CommitStepMessage:
			return mi.Height >= height
		case *HasVoteMessage:
			return mi.Height >= height
		case *NewRoundStepMessage:
			return mi.Height >= height
		case *ProposalHeartbeatMessage:
			return mi.Heartbeat.Height >= height
		case *ProposalMessage:
			return mi.Proposal.Height >= height
		case *ProposalPOLMessage:
			return mi.Height >= height
		case *VoteMessage:
			return mi.Vote.Height >= height
		case *VoteSetBitsMessage:
			return mi.Height >= height
		case *VoteSetMaj23Message:
			return mi.Height >= height
		default:
			break
		}
		break
	}

	return false
}

type BlockCsWal struct {
	blockDb db.DB
	//csMessage map[ConsensusMessage]interface{}
}

func CreateBlockCsWal(ldb db.DB) *BlockCsWal {
	return &BlockCsWal{
		blockDb: ldb,
	}
}

func (bc *BlockCsWal) ConvertWalMessage(message *TimedWALMessage) *consensus.TimedWALMessage {
	var msg WALMessage = nil
	switch m := message.Msg.(type) {
	case EndHeightMessage:
		msg = &consensus.EndHeightMessage{Height: m.Height}
		break
	case otp.EventDataRoundState:
		msg = &types.EventDataRoundState{Height: m.Height, Round: m.Round, Step: m.Step}
		break
	case timeoutInfo:
		msg = consensus.CvtTimeoutInfo(m.Duration, m.Height, m.Round, m.Step)
		break
	case msgInfo:
		var csMsg consensus.ConsensusMessage = nil
		switch mi := m.Msg.(type) {
		case *BlockPartMessage:
			csMsg = bc.cvtBlockPartMsg(mi)
			break
		case *CommitStepMessage:
			msg = bc.cvtCommitStepMsg(mi)
			break
		case *HasVoteMessage:
			msg = bc.cvtHasVoteMsg(mi)
			break
		case *NewRoundStepMessage:
			msg = bc.cvtNewRoundStepMsg(mi)
			break
		case *ProposalHeartbeatMessage:
			msg = bc.cvtProposalHeartbeatMsg(mi)
			break
		case *ProposalMessage:
			msg = bc.cvtProposalMsg(mi, m.PeerID)
			break
		case *ProposalPOLMessage:
			msg = bc.cvtProposalPolMsg(mi)
			break
		case *VoteMessage:
			msg = bc.cvtVoteMsg(mi)
			break
		case *VoteSetBitsMessage:
			msg = bc.cvtVoteSetBitsMsg(mi)
			break
		case *VoteSetMaj23Message:
			msg = bc.cvtVoteSetMaj23Msg(mi)
			break
		}

		if csMsg != nil {
			msg = consensus.CvtMsgInfo(csMsg, m.PeerID)
		}
		break
	}

	if msg != nil {
		return &consensus.TimedWALMessage{Time: message.Time, Msg: msg}
	}
	return nil
}

//// convert function ////
func (bc *BlockCsWal) getPart(height int64, partIndex int) *types.Part {
	return hold.LoadNewBlockPart(bc.blockDb, height, partIndex)
}

func (bc *BlockCsWal) getCommit(height int64) *types.Commit {
	return hold.LoadNewBlockCommit(bc.blockDb, height, "SC")
}

func (bc *BlockCsWal) cvtBlockPartMsg(old *BlockPartMessage) consensus.ConsensusMessage {
	part := bc.getPart(old.Height, old.Part.Index)
	if part != nil {
		return &consensus.BlockPartMessage{
			Height: old.Height,
			Round:  old.Round,
			Part:   part}
	}

	return nil
}

func (bc *BlockCsWal) cvtCommitStepMsg(old *CommitStepMessage) consensus.ConsensusMessage {
	commit := bc.getCommit(old.Height)
	if commit != nil {
		return &consensus.CommitStepMessage{
			Height:           old.Height,
			BlockPartsHeader: commit.BlockID.PartsHeader,
			BlockParts:       commit.BitArray(),
		}
	}

	return nil
}

func (bc *BlockCsWal) cvtHasVoteMsg(old *HasVoteMessage) consensus.ConsensusMessage {
	return &consensus.HasVoteMessage{
		Height: old.Height,
		Round:  old.Round,
		Type:   old.Type,
		Index:  old.Index}
}

func (bc *BlockCsWal) cvtNewRoundStepMsg(old *NewRoundStepMessage) consensus.ConsensusMessage {
	return &consensus.NewRoundStepMessage{
		Height:                old.Height,
		Round:                 old.Round,
		Step:                  cstypes.RoundStepType(old.Step),
		SecondsSinceStartTime: old.SecondsSinceStartTime,
		LastCommitRound:       old.LastCommitRound}
}

func (bc *BlockCsWal) cvtProposalHeartbeatMsg(old *ProposalHeartbeatMessage) consensus.ConsensusMessage {
	commit := bc.getCommit(old.Heartbeat.Height)
	if commit != nil {
		vote := commit.GetByIndex(old.Heartbeat.ValidatorIndex)
		heartBeatMsg := &consensus.ProposalHeartbeatMessage{
			Heartbeat: &types.Heartbeat{
				ValidatorAddress: vote.ValidatorAddress,
				ValidatorIndex:   vote.ValidatorIndex,
				Height:           old.Heartbeat.Height,
				Round:            old.Heartbeat.Round,
				Sequence:         old.Heartbeat.Sequence,
				Signature:        vote.Signature,
			}}

		return heartBeatMsg
	}

	return nil
}

func (bc *BlockCsWal) cvtProposalMsg(old *ProposalMessage, peer p2p.ID) consensus.ConsensusMessage {
	commit := bc.getCommit(old.Proposal.Height)
	if commit != nil {
		proposalMsg := &consensus.ProposalMessage{
			Proposal: &types.Proposal{
				Height:           old.Proposal.Height,
				Round:            old.Proposal.Round,
				Timestamp:        old.Proposal.Timestamp,
				BlockPartsHeader: commit.BlockID.PartsHeader,
				POLRound:         old.Proposal.POLRound,
				POLBlockID:       commit.BlockID,
				Signature:        commit.FirstPrecommit().Signature,
			}}
		if peer != "" {
			nodePv := store.GetNodePrivByPeer(string(peer))
			nodePv.SignProposal(nodePv.ChainID, proposalMsg.Proposal)
		}

		return proposalMsg
	}

	return nil
}

func (bc *BlockCsWal) cvtProposalPolMsg(old *ProposalPOLMessage) consensus.ConsensusMessage {
	commit := bc.getCommit(old.Height)
	if commit != nil {
		return &consensus.ProposalPOLMessage{
			Height:           old.Height,
			ProposalPOLRound: old.ProposalPOLRound,
			ProposalPOL:      commit.BitArray()}
	}

	return nil
}

func (bc *BlockCsWal) cvtVoteMsg(old *VoteMessage) consensus.ConsensusMessage {
	commit := bc.getCommit(old.Vote.Height)
	if commit != nil {
		vote := commit.GetByIndex(old.Vote.ValidatorIndex)
		nodePv := store.GetNodePrivByAddress(old.Vote.ValidatorAddress.String())
		nodePv.SignVote(nodePv.ChainID, vote)

		return &consensus.VoteMessage{Vote: vote}
	}

	return nil
}

func (bc *BlockCsWal) cvtVoteSetBitsMsg(old *VoteSetBitsMessage) consensus.ConsensusMessage {
	commit := bc.getCommit(old.Height)
	if commit != nil {
		return &consensus.VoteSetBitsMessage{
			Height:  old.Height,
			Round:   old.Round,
			Type:    commit.Type(),
			BlockID: commit.BlockID,
			Votes:   commit.BitArray()}
	}

	return nil
}

func (bc *BlockCsWal) cvtVoteSetMaj23Msg(old *VoteSetMaj23Message) consensus.ConsensusMessage {
	commit := bc.getCommit(old.Height)
	if commit != nil {
		return &consensus.VoteSetMaj23Message{
			Height:  old.Height,
			Round:   old.Round,
			Type:    commit.Type(),
			BlockID: commit.BlockID}
	}

	return nil
}
