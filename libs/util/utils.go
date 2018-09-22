package util

import cmn "github.com/tendermint/tendermint/libs/common"

func CalcValidatorsKey(height int64) []byte {
	return []byte(cmn.Fmt("validatorsKey:%v", height))
}

func CalcConsensusParamsKey(height int64) []byte {
	return []byte(cmn.Fmt("consensusParamsKey:%v", height))
}

func CalcABCIResponsesKey(height int64) []byte {
	return []byte(cmn.Fmt("abciResponsesKey:%v", height))
}
