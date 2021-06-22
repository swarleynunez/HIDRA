package eth

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/swarleynunez/superfog/core/utils"
)

func Connect(url string) (ethc *ethclient.Client) {

	ethc, err := ethclient.Dial(url)
	utils.CheckError(err, utils.FatalMode)

	return
}
