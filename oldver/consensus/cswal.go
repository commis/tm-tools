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
		}
	}

	return false
}

type BlockCsWal struct {
	blockDb db.DB
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
	case otp.EventDataRoundState:
		msg = &types.EventDataRoundState{Height: m.Height, Round: m.Round, Step: m.Step}
	case timeoutInfo:
		msg = consensus.CvtTimeoutInfo(m.Duration, m.Height, m.Round, m.Step)
	case msgInfo:
		var csMsg consensus.ConsensusMessage = nil
		switch mi := m.Msg.(type) {
		case *BlockPartMessage:
			csMsg = bc.cvtBlockPartMsg(mi)
		case *CommitStepMessage:
			msg = bc.cvtCommitStepMsg(mi)
		case *HasVoteMessage:
			msg = bc.cvtHasVoteMsg(mi)
		case *NewRoundStepMessage:
			msg = bc.cvtNewRoundStepMsg(mi)
		case *ProposalHeartbeatMessage:
			msg = bc.cvtProposalHeartbeatMsg(mi)
		case *ProposalMessage:
			msg = bc.cvtProposalMsg(mi, m.PeerID)
		case *ProposalPOLMessage:
			msg = bc.cvtProposalPolMsg(mi)
		case *VoteMessage:
			msg = bc.cvtVoteMsg(mi)
		case *VoteSetBitsMessage:
			msg = bc.cvtVoteSetBitsMsg(mi)
		case *VoteSetMaj23Message:
			msg = bc.cvtVoteSetMaj23Msg(mi)
		}

		if csMsg != nil {
			msg = consensus.CvtMsgInfo(csMsg, m.PeerID)
		}
	}

	if msg != nil {
		return &consensus.TimedWALMessage{Time: message.Time, Msg: msg}
	}
	return nil
}

func (bc *BlockCsWal) UpdateOldVoteToPrivVal(message *TimedWALMessage) {
	switch m := message.Msg.(type) {
	case msgInfo:
		switch mi := m.Msg.(type) {
		case *ProposalMessage:
			bc.updateOldProposalMsg(mi, m.PeerID)
		case *VoteMessage:
			bc.updateOldVoteMsg(mi)
		}
	}
}

func (bc *BlockCsWal) UpdateNewVoteToPrivVal(message *consensus.TimedWALMessage) {
	switch m := message.Msg.(type) {
	case msgInfo:
		switch mi := m.Msg.(type) {
		case *consensus.ProposalMessage:
			bc.updateNewProposalMsg(mi, m.PeerID)
		case *consensus.VoteMessage:
			bc.updateNewVoteMsg(mi)
		}
	}
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

func (bc *BlockCsWal) updateOldProposalMsg(old *ProposalMessage, peer p2p.ID) {
	if peer != "" {
		nodePv := store.GetNodePrivByPeer(string(peer))
		height, round, step := old.Proposal.Height, old.Proposal.Round, store.StepPropose
		signBytes := old.Proposal.SignBytes(nodePv.ChainID)
		nodePv.SaveOldSigned(height, round, step, signBytes, old.Proposal.Signature)
	}
}

func (bc *BlockCsWal) updateNewProposalMsg(msg *consensus.ProposalMessage, peer p2p.ID) {
	if peer != "" {
		nodePv := store.GetNodePrivByPeer(string(peer))
		nodePv.SignProposal(nodePv.ChainID, msg.Proposal)
	}
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

func (bc *BlockCsWal) updateOldVoteMsg(old *VoteMessage) {
	nodePv := store.GetNodePrivByAddress(old.Vote.ValidatorAddress.String())
	height, round, step := old.Vote.Height, old.Vote.Round, nodePv.VoteToStep(old.Vote.Type)
	signBytes := old.Vote.SignBytes(nodePv.ChainID)
	nodePv.SaveOldSigned(height, round, step, signBytes, old.Vote.Signature)
}

func (bc *BlockCsWal) updateNewVoteMsg(old *consensus.VoteMessage) {
	nodePv := store.GetNodePrivByAddress(old.Vote.ValidatorAddress.String())
	height, round, step := old.Vote.Height, old.Vote.Round, nodePv.VoteToStep(old.Vote.Type)
	signBytes := old.Vote.SignBytes(nodePv.ChainID)
	nodePv.SaveSigned(height, round, step, signBytes, old.Vote.Signature)
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
