package consensus

import (
	oldcstype "github.com/commis/tm-tools/oldver/consensus/types"
	oldtype "github.com/commis/tm-tools/oldver/types"
	tdcns "github.com/tendermint/tendermint/consensus"
	"github.com/tendermint/tendermint/types"
)

//var cdc = amino.NewCodec()
//
//func init() {
//	consensus.RegisterConsensusMessages(cdc)
//
//}

func FilterBlockWalMessage(height int64, msg WALMessage) bool {
	switch m := msg.(type) {
	case EndHeightMessage:
		return m.Height >= height
	case oldtype.EventDataRoundState:
		if m.Step == oldcstype.RoundStepNewHeight.String() {
			return m.Height > height
		}
		return m.Height >= height
	case timeoutInfo:
		return msg.(timeoutInfo).Height >= height
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

func ConvertWalMessage(message *TimedWALMessage) *tdcns.TimedWALMessage {
	var msg tdcns.WALMessage = nil
	switch m := message.Msg.(type) {
	case EndHeightMessage:
		msg = tdcns.EndHeightMessage{Height: m.Height}
		break
	case oldtype.EventDataRoundState:
		msg = types.EventDataRoundState{Height: m.Height, Round: m.Round, Step: m.Step}
		break
		/*case timeoutInfo:
			break
		case msgInfo:
			switch mi := m.Msg.(type) {
			case *BlockPartMessage:
				break
			case *CommitStepMessage:
				break
			case *HasVoteMessage:
				break
			case *NewRoundStepMessage:
				break
			case *ProposalHeartbeatMessage:
				break
			case *ProposalMessage:
				break
			case *ProposalPOLMessage:
				break
			case *VoteMessage:
				break
			case *VoteSetBitsMessage:
				break
			case *VoteSetMaj23Message:
				break
			}
			break*/
	}

	if &msg == nil {
		return nil
	}
	return &tdcns.TimedWALMessage{Time: message.Time, Msg: msg}
}
