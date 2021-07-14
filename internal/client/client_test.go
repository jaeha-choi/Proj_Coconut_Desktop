package client

import (
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"os"
)

func init() {
	log.Init(os.Stdout, log.DEBUG)
}
