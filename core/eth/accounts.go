package eth

import (
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/swarleynunez/superfog/core/utils"
)

func LoadKeystore(keydir string) (ks *keystore.KeyStore) {

	ks = keystore.NewKeyStore(keydir, keystore.StandardScryptN, keystore.StandardScryptP)

	return
}

func CreateAccount(ks *keystore.KeyStore, passphrase string) (a accounts.Account) {

	a, err := ks.NewAccount(passphrase)
	utils.CheckError(err, utils.FatalMode)

	return
}
