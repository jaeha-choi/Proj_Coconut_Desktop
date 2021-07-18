package main

import (
	//_ "embed"
	"github.com/jaeha-choi/Proj_Coconut_Desktop/internal/client"
)

//var uiString []byte

func main() {
	client.Start("./data/ui/UI.glade")
	//client.Start(string(uiString))
}
