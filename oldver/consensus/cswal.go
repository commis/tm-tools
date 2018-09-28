package consensus

import (
	"github.com/commis/tm-tools/libs/log"
	ts "github.com/commis/tm-tools/oldver/consensus/types"
	his "github.com/commis/tm-tools/oldver/types"
)

func FilterOldEventBlockHeight(height int64, msg WALMessage) bool {
	switch m := msg.(type) {
	case EndHeightMessage:
		return height <= m.Height
	case his.EventDataRoundState:
		if m.Step == ts.RoundStepNewHeight.String() {
			return height < m.Height
		}
		return height <= m.Height
	case timeoutInfo:
		return height <= msg.(timeoutInfo).Height
	case msgInfo:
		switch mi := m.Msg.(type) {
		case *BlockPartMessage:
			return height <= mi.Height
		case *CommitStepMessage:
			return height <= mi.Height
		case *HasVoteMessage:
			return height <= mi.Height
		case *NewRoundStepMessage:
			return height <= mi.Height
		case *ProposalHeartbeatMessage:
			return height <= mi.Heartbeat.Height
		case *ProposalMessage:
			return height <= mi.Proposal.Height
		case *ProposalPOLMessage:
			return height <= mi.Height
		case *VoteMessage:
			return height <= mi.Vote.Height
		case *VoteSetBitsMessage:
			return height <= mi.Height
		case *VoteSetMaj23Message:
			return height <= mi.Height
		default:
			log.Errorf("event: %v", msg)
			break
		}
		break
	}

	return false
}
