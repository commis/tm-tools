package types

import gco "github.com/tendermint/go-crypto"

type ValidatorSet struct {
	// NOTE: persisted via reflect, must be exported.
	Validators       []*Validator `json:"validators"`
	Proposer         *Validator   `json:"proposer"`
	totalVotingPower int64
}

func (valSet *ValidatorSet) Copy() *ValidatorSet {
	validators := make([]*Validator, len(valSet.Validators))
	for i, val := range valSet.Validators {
		// NOTE: must copy, since IncrementAccum updates in place.
		validators[i] = val.Copy()
	}
	return &ValidatorSet{
		Validators:       validators,
		Proposer:         valSet.Proposer,
		totalVotingPower: valSet.totalVotingPower,
	}
}

type Validator struct {
	Address     Address    `json:"address"`
	PubKey      gco.PubKey `json:"pub_key"`
	VotingPower int64      `json:"voting_power"`
	Accum       int64      `json:"accum"`
}

func (v *Validator) Copy() *Validator {
	vCopy := *v
	return &vCopy
}
