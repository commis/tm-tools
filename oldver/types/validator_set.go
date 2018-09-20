package types

import gco "github.com/tendermint/go-crypto"

type ValidatorSet struct {
	// NOTE: persisted via reflect, must be exported.
	Validators       []*Validator `json:"validators"`
	Proposer         *Validator   `json:"proposer"`
	totalVotingPower int64
}

type Validator struct {
	Address     Address    `json:"address"`
	PubKey      gco.PubKey `json:"pub_key"`
	VotingPower int64      `json:"voting_power"`
	Accum       int64      `json:"accum"`
}
