package client

import (
	"errors"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"os"
)

const (
	appId = "dev.jaeha.coconut"
)

// Client contains information about each client devices
type Client struct {
	PublicIp    string `json:"publicIp"`
	PrivateIp   string `json:"privateIp"`
	PublicPort  string `json:"publicPort"`
	PrivatePort string `json:"privatePort"`
}

var AssertFailed = errors.New("type assertion failed")

func Start(uiGladePath string) {
	log.Init(os.Stdout, log.DEBUG)

	// Create a new application.
	application, err := gtk.ApplicationNew(appId, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Debug(err)
		log.Error("Error while creating application")
		return
	}

	// Connect function to application startup event, this is not required.
	//application.Connect("startup", func() {
	//	log.Println("application startup")
	//})

	// Connect function to application activate event
	application.Connect("activate", func() {
		// Get the GtkBuilder ui definition in the glade file.
		builder, err := gtk.BuilderNewFromFile(uiGladePath)
		if err != nil {
			log.Debug(err)
			log.Error("Error in BuilderNewFromFile")
			return
		}

		// Map the handlers to callback functions, and connect the signals
		// to the Builder.
		//signals := map[string]interface{}{
		//	"resize": windResize,
		//}
		//builder.ConnectSignals(signals)

		// Get the object with the id of "main_window".
		obj, err := builder.GetObject("main_window")
		if err != nil {
			log.Debug(err)
			log.Error("Error in GetObject")
			return
		}

		// Verify that the object is a pointer to a gtk.ApplicationWindow.
		win, err := isWindow(obj)
		if err != nil {
			log.Debug(err)
			log.Error("Error in isWindow")
			return
		}
		//win.SetDecorated(false)

		// Show the Window and all of its components.
		win.Show()

		application.AddWindow(win)
	})

	// Connect function to application shutdown event, this is not required.
	//application.Connect("shutdown", func() {
	//	log.Println("application shutdown")
	//})

	// Launch the application
	os.Exit(application.Run(os.Args))
}

func isWindow(obj glib.IObject) (*gtk.Window, error) {
	// Make type assertion (as per gtk.go).
	if win, ok := obj.(*gtk.Window); ok {
		return win, nil
	}
	return nil, AssertFailed
}
