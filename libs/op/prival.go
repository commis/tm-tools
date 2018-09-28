package op

import (
	his "github.com/commis/tm-tools/oldver/types"
	"github.com/tendermint/tendermint/privval"
)

type TmCfgPrival struct {
	pvFile string
}

func CreateTmCfgPrival(cfgRoot string) *TmCfgPrival {
	return &TmCfgPrival{
		pvFile: cfgRoot + "/priv_validator.json"}
}

func (t *TmCfgPrival) ResetOldPVHeight(height int64) {
	pfs := his.LoadPrivValidator(t.pvFile)
	pfs.LastHeight = height
	pfs.Save()
}

func (t *TmCfgPrival) ResetNewPVHeight(height int64) {
	fpv := privval.LoadFilePV(t.pvFile)
	fpv.LastHeight = height
	fpv.Save()
}
