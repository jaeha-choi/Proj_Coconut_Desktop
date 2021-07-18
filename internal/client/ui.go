package client

import (
	"errors"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"os"
	"path/filepath"
	"strconv"
)

const (
	appId = "dev.jaeha.coconut"
)

const (
	fileNameIdx = iota
	fileSizeWithUnitIdx
	fileStatusIdx
	fileFullPath
	fileSizeInBytes
)

var AssertFailed = errors.New("type assertion failed")

type UIStatus struct {
	builder       *gtk.Builder
	isFileTab     bool
	fileListOrder []int
	fileMap       map[string]struct{}
}

type preprocessing func(string) ([]interface{}, error)

func initUIStatus() (stat *UIStatus) {
	return &UIStatus{
		builder:       nil,
		isFileTab:     true,
		fileListOrder: []int{fileNameIdx, fileSizeWithUnitIdx, fileStatusIdx, fileFullPath, fileSizeInBytes},
		fileMap:       map[string]struct{}{},
	}
}

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
		var err error
		stat := initUIStatus()

		// Get the GtkBuilder ui definition in the glade file.
		stat.builder, err = gtk.BuilderNewFromFile(uiGladePath)
		//stat.builder, err = gtk.BuilderNewFromString(uiString)
		if err != nil {
			log.Debug(err)
			log.Error("Error in BuilderNewFromFile")
			return
		}

		// Map the handlers to callback functions, and connect the signals
		// to the Builder.
		signals := map[string]interface{}{
			"switchPage":     stat.switchPage,
			"addButtonClick": stat.addButtonClick,
			"keyPress":       stat.keyPress,
		}
		stat.builder.ConnectSignals(signals)

		win, err := stat.getWindowWithId("main_window")
		if err != nil {
			log.Fatal("Could not find main_window")
			return
		}

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

func (ui *UIStatus) addButtonClick() {
	log.Debug("Add button is clicked")
	win, err := ui.getWindowWithId("main_window")
	if err != nil {
		return
	}
	if ui.isFileTab {
		chooser, err := gtk.FileChooserNativeDialogNew("Select Files to Send", win, gtk.FILE_CHOOSER_ACTION_OPEN, "Select", "Cancel")
		if err != nil {
			log.Debug(err)
			log.Error("Error while showing file chooser dialog")
			return
		}
		chooser.SetSelectMultiple(true)
		// Run and wait until user selects file
		ret := chooser.Run()
		log.Debug("Chooser returned: ", ret)
		// If user clicks "Select"
		if gtk.ResponseType(ret) == gtk.RESPONSE_ACCEPT {
			filenames, err := chooser.GetFilenames()
			if err != nil {
				log.Debug(err)
				log.Error("Error while getting filenames")
				return
			}
			fileList, err := ui.getListStoreWithId("fileList")
			if err != nil {
				return
			}
			treeView, err := ui.getTreeViewWithId("fileListView")
			if err != nil {
				return
			}

			addToListStore(fileList, treeView, filenames, ui.fileListOrder, ui.getFileInfo)
		}

	} else {
		// Add contacts
	}
}

func (ui *UIStatus) keyPress(_ *gtk.Window, event *gdk.Event) {
	log.Debug("KeyPress function called")
	eventKey := gdk.EventKeyNewFromEvent(event)
	// If pressed key is "Delete" key, remove selected files from the list
	if eventKey.KeyVal() == gdk.KEY_Delete {
		if err := ui.removeSelected(); err != nil {
			log.Debug(err)
			log.Error("Error while removing selected")
			return
		}
	}
}

func (ui *UIStatus) removeSelected() (err error) {
	listStore, err := ui.getListStoreWithId("fileList")
	if err != nil {
		return err
	}
	treeView, err := ui.getTreeViewWithId("fileListView")
	if err != nil {
		return err
	}
	selected, err := treeView.GetSelection()
	if err != nil {
		log.Debug(err)
		log.Error("Error while getting selected files")
		return err
	}
	rows := selected.GetSelectedRows(listStore)
	// Reverse is called to preserve the head of the linked list for iter.
	// Without it, not all nodes will be deleted properly.
	reversed := rows.Reverse()
	reversed.Foreach(func(item interface{}) {
		path, err := isTreePath(item)
		if err != nil {
			return
		}
		iter, err := listStore.GetIter(path)
		if err != nil {
			log.Debug(err)
			log.Error("Error while getting iterator")
			return
		}
		value, err := listStore.GetValue(iter, fileFullPath)
		if err != nil {
			log.Debug(err)
			log.Error("Error while getting value")
			return
		}
		fullPath, err := value.GetString()
		if err != nil {
			log.Debug(err)
			log.Error("Error while getting string from *glib.Value")
			return
		}
		log.Debug("Full path: ", fullPath)

		// Delete from the map as well
		delete(ui.fileMap, fullPath)
		_ = listStore.Remove(iter)
	})
	return nil
}

func addToListStore(listStore *gtk.ListStore, treeView *gtk.TreeView, dataToAdd []string, order []int, f preprocessing) {
	for i, data := range dataToAdd {
		row, err := f(data)
		if err != nil {
			log.Debug(err)
			log.Error("Error while adding element. continuing")
			continue
		}
		// Row is empty if the elem is already added
		if row != nil {
			iter := listStore.Append()
			// Show file full path as a tooltip
			treeView.SetTooltipColumn(fileFullPath)
			if err = listStore.Set(iter, order, row); err != nil {
				log.Debug("Error while adding ", dataToAdd[i])
				continue
			}
		} else {
			log.Debug("Element already added; Skipping...")
		}
	}
}

func (ui *UIStatus) getFileInfo(fileName string) ([]interface{}, error) {
	if _, exist := ui.fileMap[fileName]; exist {
		return nil, nil
	}
	_, fName := filepath.Split(fileName)
	s, err := os.Stat(fileName)
	ui.fileMap[fileName] = struct{}{}
	if err != nil {
		log.Debug(err)
		log.Error("Error while getting stats")
		return nil, err
	}
	size := s.Size()
	return []interface{}{fName, sizeAddUnit(size), "Pending", fileName, size}, nil
}

func sizeAddUnit(size int64) (sizeStr string) {
	// Decimal points arent too important, so omit it for UI space
	if size < 1e+3 {
		sizeStr = strconv.Itoa(int(size)) + " B"
	} else if size < 1e+6 {
		sizeStr = strconv.Itoa(int(size/1e+3)) + " KB"
	} else if size < 1e+9 {
		sizeStr = strconv.Itoa(int(size/1e+6)) + " MB"
	} else {
		sizeStr = strconv.Itoa(int(size/1e+9)) + " MB"
	}
	return sizeStr
}

func (ui *UIStatus) switchPage() {
	log.Debug("Switch page clicked")
	ui.isFileTab = !ui.isFileTab
	log.Debug(ui.isFileTab)
}

func (ui *UIStatus) getButtonWithId(buttonId string) (button *gtk.Button, err error) {
	object, err := ui.builder.GetObject(buttonId)
	if err != nil {
		log.Debug(err)
		log.Errorf("Error while getting button with button id: %s", buttonId)
		return nil, err
	}
	button, ok := object.(*gtk.Button)
	if ok {
		return button, nil
	}
	log.Debug(AssertFailed)
	log.Error("object is not a button")
	return nil, AssertFailed
}

func (ui *UIStatus) getTreeViewWithId(treeViewId string) (treeView *gtk.TreeView, err error) {
	object, err := ui.builder.GetObject(treeViewId)
	if err != nil {
		log.Debug(err)
		log.Errorf("Error while getting treeView with treeView id: %s", treeViewId)
		return nil, err
	}
	treeView, ok := object.(*gtk.TreeView)
	if ok {
		return treeView, nil
	}
	log.Debug(AssertFailed)
	log.Error("object is not a treeView")
	return nil, AssertFailed
}

func (ui *UIStatus) getListStoreWithId(listStoreId string) (listStore *gtk.ListStore, err error) {
	object, err := ui.builder.GetObject(listStoreId)
	if err != nil {
		log.Debug(err)
		log.Errorf("Error while getting listStore with listStore id: %s", listStoreId)
		return nil, err
	}
	listStore, ok := object.(*gtk.ListStore)
	if ok {
		return listStore, nil
	}
	log.Debug(AssertFailed)
	log.Error("object is not a listStore")
	return nil, AssertFailed
}

func (ui *UIStatus) getWindowWithId(windowId string) (window *gtk.Window, err error) {
	object, err := ui.builder.GetObject(windowId)
	if err != nil {
		log.Debug(err)
		log.Errorf("Error while getting window with window id: %s", windowId)
		return nil, err
	}
	window, ok := object.(*gtk.Window)
	if ok {
		return window, nil
	}
	log.Debug(AssertFailed)
	log.Error("object is not a window")
	return nil, AssertFailed
}

func isTreePath(item interface{}) (*gtk.TreePath, error) {
	path, ok := item.(*gtk.TreePath)
	if ok {
		return path, nil
	}
	log.Debug(AssertFailed)
	log.Error("Item is not a TreePath")
	return nil, AssertFailed
}
