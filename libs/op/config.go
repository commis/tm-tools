package op

import (
	"fmt"
	"os"

	"github.com/commis/tm-tools/libs/log"
	"github.com/commis/tm-tools/libs/op/store"
	"github.com/commis/tm-tools/libs/util"
	cvt "github.com/commis/tm-tools/oldver/convert"
	otp "github.com/commis/tm-tools/oldver/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/privval"
)

func ResetPrivValHeight(ver TMVersionType, tmPath string, height int64) {
	var tm *TmConfig = nil
	switch ver {
	case TMVer0180:
		tm = CreateTmConfig(tmPath, "")
		break
	case TMVer0231:
		tm = CreateTmConfig("", tmPath)
		break
	}

	if tm != nil {
		tm.ResetPvHeight(height)
	}
}

type TmConfig struct {
	oldPath string
	newPath string
	config  string
	genesis string
	privVal string

	maxHeightNode string
	maxPrivVal    *privval.FilePV
}

//oTmPath, nTmPath is /home/share/chaindata/peerx/tendermint
func CreateTmConfig(oTmPath, nTmPath string) *TmConfig {
	if oTmPath == "" && nTmPath == "" {
		cmn.Exit("create parameter is empty")
	}

	data := "/data"
	tm := &TmConfig{
		config:  "tendermint/config",
		genesis: "genesis.json",
		privVal: "priv_validator.json",
	}

	if oTmPath != "" {
		tm.oldPath = util.GetParentDir(oTmPath, 2)
	}

	if nTmPath != "" {
		dataPath := nTmPath + data
		util.CreateDirAll(dataPath)
		tm.newPath = util.GetParentDir(nTmPath, 2)
	}

	return tm
}

func (tm *TmConfig) UpgradeAllGenesisAndPvFile() {
	nodes := util.GetChildDir(tm.oldPath)
	for _, node := range nodes {
		flagFile := tm.getFilePath(tm.newPath, node, ".config")

		oGenFile := tm.getFilePath(tm.oldPath, node, tm.genesis)
		nGenFile := tm.getFilePath(tm.newPath, node, tm.genesis)

		oPvFile := tm.getFilePath(tm.oldPath, node, tm.privVal)
		nPvFile := tm.getFilePath(tm.newPath, node, tm.privVal)

		var nPrivVal *privval.FilePV = nil
		if util.Exist(flagFile) { //存在则直接加载到内存中
			oGen, _ := otp.MakeGenesisDocFromFile(oGenFile)
			opv := otp.LoadPrivValidator(oPvFile)

			nPrivVal = privval.LoadFilePV(nPvFile)
			store.AddNodePriv(opv.Address.String(), oGen.ChainID, nPrivVal)
		} else {
			util.CreateDirAll(tm.getTmDir(tm.newPath, node, tm.config))

			var oldAddress string
			chainID := cvt.UpgradeGenesisJSON(oGenFile, nGenFile)
			oldAddress, nPrivVal = cvt.UpgradePrivValidatorJson(oPvFile, nPvFile)
			store.AddNodePriv(oldAddress, chainID, nPrivVal)

			util.CreateDirAll(flagFile)
		}

		//update maxHeight PrivVal
		if tm.maxPrivVal == nil || nPrivVal.LastHeight > tm.maxPrivVal.LastHeight {
			tm.maxHeightNode = node
			tm.maxPrivVal = nPrivVal
		}
	}
	store.SortNodePriv()
}

func (tm *TmConfig) UpdateAllPrivValFile() {
	nodes := util.GetChildDir(tm.oldPath)
	for _, node := range nodes {
		flagFile := tm.getFilePath(tm.newPath, node, ".config")
		if util.Exist(flagFile) {
			nPvFile := tm.getFilePath(tm.newPath, node, tm.privVal)
			nPrivVal := privval.LoadFilePV(nPvFile)

			//update last vote info
			nPrivVal.LastHeight = tm.maxPrivVal.LastHeight
			nPrivVal.LastRound = tm.maxPrivVal.LastRound
			nPrivVal.LastStep = tm.maxPrivVal.LastStep
			nPrivVal.LastSignature = tm.maxPrivVal.LastSignature
			nPrivVal.LastSignBytes = tm.maxPrivVal.LastSignBytes

			nPrivVal.Save()
		}
	}
	tm.saveMaxHeightNode()
}

func (tm *TmConfig) saveMaxHeightNode() {
	maxNodeFile := tm.oldPath + "/topNode.txt"
	rd, err := os.Create(maxNodeFile)
	if err != nil {
		log.Errorf("failed to open heightNode file: %v", err)
		return
	}

	rd.WriteString(fmt.Sprintf("%s height:%d", tm.maxHeightNode, tm.maxPrivVal.LastHeight))
	rd.Close()
}

func (tm *TmConfig) GetTopPrivVal() *privval.FilePV {
	return tm.maxPrivVal
}

func (tm *TmConfig) ResetPvHeight(height int64) {
	if tm.oldPath != "" {
		tm.setOldPvHeight(height)
	} else if tm.newPath != "" {
		tm.setNewPvHeight(height)
	}
}

func (tm *TmConfig) getTmDir(rootPath, node, tmDir string) string {
	return rootPath + "/" + node + "/" + tmDir
}

func (tm *TmConfig) getFilePath(rootPath, node, fileName string) string {
	return rootPath + "/" + node + "/" + tm.config + "/" + fileName
}

func (tm *TmConfig) setOldPvHeight(height int64) {
	nodes := util.GetChildDir(tm.oldPath)
	for _, node := range nodes {
		pvFile := tm.getFilePath(tm.oldPath, node, tm.privVal)
		flagFile := tm.getFilePath(tm.oldPath, node, "."+tm.privVal)
		if util.Exist(flagFile) {
			continue
		}

		pfs := otp.LoadPrivValidator(pvFile)
		pfs.LastHeight = height
		pfs.LastStep = 3
		pfs.Save()
	}
}

func (tm *TmConfig) setNewPvHeight(height int64) {
	nodes := util.GetChildDir(tm.newPath)
	for _, node := range nodes {
		pvFile := tm.getFilePath(tm.newPath, node, tm.privVal)
		flagFile := tm.getFilePath(tm.newPath, node, "."+tm.privVal)
		if util.Exist(flagFile) {
			continue
		}

		fpv := privval.LoadFilePV(pvFile)
		fpv.LastHeight = height
		fpv.LastStep = 3
		fpv.Save()
	}
}
