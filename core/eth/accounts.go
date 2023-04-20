package eth

import (
	"errors"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/swarleynunez/hidra/core/utils"
)

var (
	//errNotFoundAddr  = errors.New("ethereum address not found in keystore")
	ErrMalformedAddr = errors.New("malformed ethereum address")
)

func LoadKeystore(keydir string) (ks *keystore.KeyStore) {

	ks = keystore.NewKeyStore(keydir, keystore.StandardScryptN, keystore.StandardScryptP)

	return
}

func CreateAccount(ks *keystore.KeyStore, passphrase string) (from accounts.Account) {

	from, err := ks.NewAccount(passphrase)
	utils.CheckError(err, utils.WarningMode)

	if err == nil {
		// Save the created address
		// TODO: test --> utils.SetEnv("NODE_ADDR", from.Address.String())
	}

	return
}

func LoadAccount(ks *keystore.KeyStore, passphrase string) (from accounts.Account) {

	// Unlock the loaded account
	from = ks.Accounts()[0]
	err := ks.Unlock(from, passphrase)
	utils.CheckError(err, utils.FatalMode)

	return
}

/*func LoadAccount(ks *keystore.KeyStore, addr, passphrase string) (from accounts.Account) {

	if len(ks.Accounts()) == 0 {
		from = CreateAccount(ks, passphrase)
	} else {
		if utils.ValidEthAddress(addr) {
			ksa := ks.Accounts()
			for i := range ksa {
				if ksa[i].Address == common.HexToAddress(addr) {
					from = ksa[i]
					break
				}
			}

			if from == (accounts.Account{}) {
				utils.CheckError(errNotFoundAddr, utils.FatalMode)
			}
		} else {
			utils.CheckError(ErrMalformedAddr, utils.FatalMode)
		}
	}

	// Unlock the loaded account
	err := ks.Unlock(from, passphrase)
	utils.CheckError(err, utils.FatalMode)

	return
}*/
