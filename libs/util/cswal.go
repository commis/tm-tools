package util

import (
	"github.com/commis/tm-tools/libs/log"
	"github.com/tendermint/tendermint/consensus"
	cs "github.com/tendermint/tendermint/consensus/types"
	"github.com/tendermint/tendermint/types"
)

func FilterNewEventBlockHeight(height int64, msg consensus.WALMessage) bool {
	switch m := msg.(type) {
	case consensus.EndHeightMessage:
		return height <= m.Height
	case types.EventDataRoundState:
		if m.Step == cs.RoundStepNewHeight.String() {
			return height < m.Height
		}
		return height <= m.Height
	case consensus.TimeoutInfo:
		return height <= msg.(consensus.TimeoutInfo).Height
	case consensus.MsgInfo:
		switch mi := m.Msg.(type) {
		case *consensus.BlockPartMessage:
			return height <= mi.Height
		case *consensus.CommitStepMessage:
			return height <= mi.Height
		case *consensus.HasVoteMessage:
			return height <= mi.Height
		case *consensus.NewRoundStepMessage:
			return height <= mi.Height
		case *consensus.ProposalHeartbeatMessage:
			return height <= mi.Heartbeat.Height
		case *consensus.ProposalMessage:
			return height <= mi.Proposal.Height
		case *consensus.ProposalPOLMessage:
			return height <= mi.Height
		case *consensus.VoteMessage:
			return height <= mi.Vote.Height
		case *consensus.VoteSetBitsMessage:
			return height <= mi.Height
		case *consensus.VoteSetMaj23Message:
			return height <= mi.Height
		default:
			log.Errorf("event: %v", msg)
			break
		}
		break
	}

	return false
}
