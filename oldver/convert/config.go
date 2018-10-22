package convert

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/commis/tm-tools/libs/op/hold"
	otp "github.com/commis/tm-tools/oldver/types"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
)

func UpgradeGenesisJSON(oFile, nFile string) string {
	jsonBytes, err := ioutil.ReadFile(oFile)
	if err != nil {
		cmn.Exit(err.Error())
	}

	oGen, err := otp.GenesisDocFromJSON(jsonBytes)
	if err == nil {
		nGen := toNewGenesisDoc(oGen)
		if err := nGen.SaveAs(nFile); err != nil {
			cmn.Exit(err.Error())
		}
		return nGen.ChainID
	}

	return ""
}

func UpgradePrivValidatorJson(oFile, nFile string) (string, *privval.FilePV) {
	old := otp.LoadPrivValidator(oFile)
	newPrivVal := privval.GenFilePV(nFile)

	/*后续根据cs.wal消息更新*/
	newPrivVal.LastHeight = old.LastHeight
	newPrivVal.LastRound = old.LastRound
	newPrivVal.LastStep = old.LastStep
	newPrivVal.LastSignature = CvtNewSignature(old.LastSignature)
	newPrivVal.LastSignBytes = old.LastSignBytes.Bytes()

	newPrivVal.PubKey = CvtNewPubKey(old.PubKey)
	newPrivVal.PrivKey = CvtNewPrivKey(old.PrivKey)
	newPrivVal.Address = newPrivVal.PubKey.Address()

	newPrivVal.Save()
	return old.Address.String(), newPrivVal
}

func LoadOldGenesisDoc(db dbm.DB) *otp.GenesisDoc {
	bytes := db.Get([]byte(hold.GenesisDoc))
	if len(bytes) == 0 {
		return nil
	}

	var genDoc *otp.GenesisDoc
	err := json.Unmarshal(bytes, &genDoc)
	if err != nil {
		log.Printf("Failed to load genesis doc due to unmarshaling error: %v (bytes: %X)", err, bytes)
	}
	return genDoc
}

func toNewGenesisDoc(old *otp.GenesisDoc) *types.GenesisDoc {
	newGenesisDoc := new(types.GenesisDoc)
	newGenesisDoc.AppHash = old.AppHash.Bytes()
	newGenesisDoc.ChainID = old.ChainID
	newGenesisDoc.GenesisTime = old.GenesisTime
	newGenesisDoc.Validators = []types.GenesisValidator{}

	// 修正老版本错误的时间
	/*if newGenesisDoc.GenesisTime.IsZero() {
		newGenesisDoc.GenesisTime = time.Now().UTC()
	}*/

	for _, val := range old.Validators {
		one := types.GenesisValidator{}
		one.Power = val.Power
		one.Name = val.Name
		one.PubKey = CvtNewPubKey(val.PubKey)

		newGenesisDoc.Validators = append(newGenesisDoc.Validators, one)
	}

	return newGenesisDoc
}
