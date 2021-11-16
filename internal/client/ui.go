package client

import (
	"errors"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/jaeha-choi/Proj_Coconut_Utility/cryptography"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	appId = "dev.jaeha.coconut"
)

// File tree view index
const (
	fileNameIdx = iota
	fileSizeWithUnitIdx
	fileStatusIdx
	fileFullPath
	fileSizeInBytes
)

// Contact tree view index
const (
	keyName = iota
	keyDate
	keyFingerprint
)

// AssertFailed is returned when the type assertion fails.
var AssertFailed = errors.New("type assertion failed")

type UIStatus struct {
	builder        *gtk.Builder
	isFileTab      bool
	onlineStatus   bool
	fileListOrder  []int
	keyListOrder   []int
	totalFileSize  int64
	totalFileCount int
	fileMap        map[string]struct{}
	client         *Client
}

func init() {
	log.Init(os.Stdout, log.DEBUG)
}

// initUIStatus returns default UIStatus settings
func initUIStatus() (stat *UIStatus) {
	return &UIStatus{
		builder:        nil,
		isFileTab:      true,
		onlineStatus:   false,
		fileListOrder:  []int{fileNameIdx, fileSizeWithUnitIdx, fileStatusIdx, fileFullPath, fileSizeInBytes},
		keyListOrder:   []int{keyName, keyDate, keyFingerprint},
		totalFileSize:  0,
		totalFileCount: 0,
		fileMap:        map[string]struct{}{},
		client:         nil,
	}
}

// Start initializes all configurations and starts main UI
func Start(uiGladePath string, client *Client) {
	var stat *UIStatus

	// Create a new application.
	application, err := gtk.ApplicationNew(appId, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Debug(err)
		log.Error("Error while creating application")
		return
	}

	// Connect function to application startup event
	application.Connect("startup", func() {
		log.Debug("Application starting up...")

		stat = initUIStatus()
		stat.client = client
	})

	// Connect function to application activate event
	application.Connect("activate", func() {
		var err error
		// Open RSA Keys
		pubBlock, err := cryptography.OpenKeysAsBlock(client.KeyPath, "key.pub")
		//privBlock, err := cryptography.OpenPrivKey(client.KeyPath, "key.priv")
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		stat.client.pubKeyBlock = pubBlock
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

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
			"switchPage":         stat.handleSwitchPage,
			"addButtonClick":     stat.handleAddButtonClick,
			"keyPressFileList":   stat.handleKeyPressFileList,
			"statusClick":        stat.handleStatusClick,
			"addCodeDone":        stat.handleAddCodeDone,
			"clickEmptySpotFile": stat.handleClickEmptySpotFile,
			"activateExpander":   stat.handleActivateExpander,
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

	// Connect function to application shutdown event
	application.Connect("shutdown", func() {
		log.Debug("Application shutdown...")
		// Close connection if not already
		if stat.client.conn != nil {
			if err = stat.client.Disconnect(); err != nil {
				log.Debug(err)
				log.Error("Error while closing connection")
				return
			}
		}
	})

	// Launch the application
	os.Exit(application.Run(nil))
}

func (ui *UIStatus) handleSwitchPage() {
	//log.Debug("handleSwitchPage called")
	ui.isFileTab = !ui.isFileTab
}

// handleAddButtonClick handles event when "+" button is clicked.
// Behavior depends on current viewing tab ("Files"/"Contacts")
// If "Files" tab is activated, this will show a file chooser
// If "Contacts" tab is activated, this will prompt for receiver's Add Code
func (ui *UIStatus) handleAddButtonClick() {
	//log.Debug("handleAddButtonClick called")

	// Behavior depends on current tab
	if ui.isFileTab {
		// Add files

		// Get main window
		win, err := ui.getWindowWithId("main_window")
		if err != nil {
			return
		}

		// FileChooserNative does not get recognized by gotk3 and we need to create
		// it via FileChooserNativeDialogNew.
		chooser, err := gtk.FileChooserNativeDialogNew("Select File(s) to Send", win, gtk.FILE_CHOOSER_ACTION_OPEN, "Select", "Cancel")
		if err != nil {
			log.Debug(err)
			log.Error("Error while showing file chooser dialog")
			return
		}

		// Allow users to select multiple files
		chooser.SetSelectMultiple(true)
		// Run and wait until user selects file
		ret := chooser.Run()

		log.Debug("Chooser returned: ", ret)

		// If user chose "Select" (accept) add files to the list
		if gtk.ResponseType(ret) == gtk.RESPONSE_ACCEPT {
			filenames, err := chooser.GetFilenames()
			if err != nil {
				log.Debug(err)
				log.Error("Error while getting filenames")
				return
			}
			ui.addFilesToListStore(filenames)
		}
	} else {
		// Add contacts
		popover, err := ui.getPopoverWithId("addCodeEntry")
		if err != nil {
			return
		}
		popover.Popup()
	}
}

// handleActivateExpander handles event when the expander button on "Contacts" tab is clicked.
// Behavior depends on the current online/offline status, and current expander status.
// If current status is offline, do nothing. Enable grayed out expander, otherwise.
// If expander is opened, register device and get Add Code then display it.
// If expander is closed, remove device from the Add Code list from the server.
func (ui *UIStatus) handleActivateExpander(expander *gtk.Expander) {
	//log.Debug("handleActivateExpander called")

	// If the current user is offline, do nothing
	if !ui.onlineStatus {
		expander.SetExpanded(true)
		return
	}

	// Get grid containing labels for Add Code digits
	addCodeGrid, err := ui.getGridWithId("addCodeGrid")
	if err != nil {
		return
	}

	// Get expander label
	addCodeExpanderLabel, err := ui.getLabelWithId("expanderLabel")
	if err != nil {
		return
	}

	// When the process is ongoing, "gray out" expander
	expander.SetSensitive(false)

	if expander.GetExpanded() {
		// If expander was expanded, remove current device from the Add Code list
		//log.Debug("Expander no longer revealed")
		go func() {
			log.Debugf("Removing Add Code: %s", ui.client.addCode)
			if err = ui.client.DoRemoveAddCode(); err != nil {
				log.Debug(err)
				log.Error("Error while removing Add Code from the server")
				_ = glib.IdleAdd(func() {
					expander.SetExpanded(false)
					addCodeExpanderLabel.SetLabel("Error. Try again in a bit")
					time.Sleep(5 * time.Second)
					addCodeExpanderLabel.SetLabel("Click to activate Add Code")
					expander.SetSensitive(true)
				})
				return
			}
			time.Sleep(1 * time.Second)
			// Replace each Add Code character to "-"
			_ = glib.IdleAdd(func() {
				children := addCodeGrid.GetChildren()
				children.Foreach(func(item interface{}) {
					label, _ := gtk.WidgetToLabel(item.(*gtk.Widget))
					label.SetLabel("<span size=\"xx-large\" weight=\"bold\">-</span>")
				})

				addCodeExpanderLabel.SetLabel("Click to activate Add Code")
				expander.SetSensitive(true)
			})
		}()
	} else {
		// If expander was not expanded, add current device to the Add Code list
		//log.Debug("Expander revealed")
		go func() {
			if err = ui.client.DoGetAddCode(); err != nil {
				log.Debug(err)
				log.Error("Error while getting Add Code from the server")
				_ = glib.IdleAdd(func() {
					expander.SetExpanded(true)
					addCodeExpanderLabel.SetLabel("Error. Try again in a bit")
					time.Sleep(5 * time.Second)
					addCodeExpanderLabel.SetLabel("Click to deactivate Add Code")
					expander.SetSensitive(true)
				})
				return
			}
			log.Debugf("Received Add Code: %s", ui.client.addCode)
			time.Sleep(1 * time.Second)

			// Replace each Add Code character to corresponding digit
			_ = glib.IdleAdd(func() {
				children := addCodeGrid.GetChildren()
				// Add Code digit index (GetChildren returns labels from the right side,
				// and Add Codes are always 6 digit long; so start from the last index)
				idx := 5
				//children = children.Reverse()
				children.Foreach(func(item interface{}) {
					label, _ := gtk.WidgetToLabel(item.(*gtk.Widget))
					label.SetLabel("<span size=\"xx-large\" weight=\"bold\">" + string(ui.client.addCode[idx]) + "</span>")
					idx--
				})

				addCodeExpanderLabel.SetLabel("Click to deactivate Add Code")
				expander.SetSensitive(true)
			})
		}()
	}
}

// handleStatusClick handles event when the status button ("Online"/"Offline") is clicked.
// This event also affects the Add Code expander and expander label.
// If switched to Online, expander is no longer grayed out.
// If switched to Offline, expander is grayed out.
func (ui *UIStatus) handleStatusClick(_ *gtk.EventBox, event *gdk.Event) {
	//log.Debug("Status label clicked")
	eventButton := gdk.EventButtonNewFromEvent(event)
	// If user right-clicks the status label
	if eventButton.Button() == gdk.BUTTON_PRIMARY {
		label, err := ui.getLabelWithId("connStatusLabel")
		if err != nil {
			return
		}
		listBoxRow, err := ui.getListBoxRowWithId("statusBoxRow")
		if err != nil {
			return
		}
		addCodeExpander, err := ui.getExpanderWithId("addCodeExpander")
		if err != nil {
			return
		}
		addCodeExpanderLabel, err := ui.getLabelWithId("expanderLabel")
		if err != nil {
			return
		}

		// Make status label un-clickable
		listBoxRow.SetSensitive(false)

		var markAsError = func() {
			label.SetMarkup("<span foreground=\"red\">ERROR</span>")
			listBoxRow.SetSensitive(true)
		}

		// If currently online
		if ui.onlineStatus {
			label.SetMarkup("<span foreground=\"orange\">Disconnecting...</span>")
			go func() {
				if err = ui.client.Disconnect(); err != nil {
					log.Debug(err)
					log.Error("Error while connecting to the server")
					_ = glib.IdleAdd(markAsError)
					return
				}
				_ = glib.IdleAdd(func() {
					label.SetMarkup("<span foreground=\"red\">Offline</span>")
					ui.onlineStatus = false

					if addCodeExpander.GetExpanded() {
						// Get grid containing labels for Add Code digits
						addCodeGrid, err := ui.getGridWithId("addCodeGrid")
						if err != nil {
							return
						}
						children := addCodeGrid.GetChildren()
						children.Foreach(func(item interface{}) {
							label, _ := gtk.WidgetToLabel(item.(*gtk.Widget))
							label.SetLabel("<span size=\"xx-large\" weight=\"bold\">-</span>")
						})
						addCodeExpander.SetExpanded(false)
					}

					addCodeExpanderLabel.SetLabel("Switch online to get Add Code")
					addCodeExpander.SetSensitive(false)
					listBoxRow.SetSensitive(true)
				})
			}()

		} else {
			label.SetMarkup("<span foreground=\"orange\">Connecting...</span>")
			go func() {
				if err = ui.client.Connect(); err != nil {
					log.Debug(err)
					log.Error("Error while connecting to the server")
					_ = glib.IdleAdd(markAsError)
					return
				}
				_ = glib.IdleAdd(func() {
					label.SetMarkup("<span foreground=\"green\">Online</span>")
					ui.onlineStatus = true
					addCodeExpanderLabel.SetLabel("Activate Add Code")
					addCodeExpander.SetSensitive(true)
					listBoxRow.SetSensitive(true)
				})
			}()
		}
	}
}

// handleAddCodeDone handles event when Add Code was entered
// TODO: WIP
func (ui *UIStatus) handleAddCodeDone(entry *gtk.Entry, event *gdk.Event) {
	log.Debug("addCodeDone called")
	eventKey := gdk.EventKeyNewFromEvent(event)
	key := eventKey.KeyVal()
	if key == gdk.KEY_Return || key == gdk.KEY_KP_Enter {
		text, err := entry.GetText()
		if err != nil {
			return
		}
		intVal, err := strconv.ParseInt(text, 10, 32)
		if err != nil {
			// text is not an integer
			log.Debug("Not an integer: ", text)
			return
		}
		popover, err := ui.getPopoverWithId("addCodeEntry")
		if err != nil {
			return
		}
		popover.Popdown()
		entry.SetText("")
		// TODO: Get pubkey based on intVal
		log.Debug(intVal)
	}
}

// handleClickEmptySpotFile handles event when empty space is clicked.
// If empty spot is clicked, selected files are deselected.
func (ui *UIStatus) handleClickEmptySpotFile(_ *gtk.EventBox, event *gdk.Event) {
	log.Debug("clickEmptySpotFile called")
	eventButton := gdk.EventButtonNewFromEvent(event)
	// // If user right-clicks empty space
	if eventButton.Button() == gdk.BUTTON_PRIMARY {
		fileTreeView, err := ui.getTreeViewWithId("fileListView")
		if err != nil {
			return
		}
		ui.unselectAll(fileTreeView)
	}
}

// handleKeyPressFileList handles event when key is pressed while FileList is focused.
// If delete key is pressed, selected files are removed from the list.
func (ui *UIStatus) handleKeyPressFileList(fileTreeView *gtk.TreeView, event *gdk.Event) {
	log.Debug("KeyPress function called")
	eventKey := gdk.EventKeyNewFromEvent(event)
	// If pressed key is "Delete" key, remove selected files from the list
	if eventKey.KeyVal() == gdk.KEY_Delete {
		if err := ui.removeSelectedFiles(fileTreeView); err != nil {
			log.Debug(err)
			log.Error("Error while removing selected")
			return
		}
	}
}

// unselectAll deselects selections in treeView
func (ui *UIStatus) unselectAll(treeView *gtk.TreeView) {
	selection, err := treeView.GetSelection()
	if err != nil {
		log.Debug(err)
		log.Error("Error while getting selected files")
	}
	selection.UnselectAll()
}

// removeSelectedFiles removes selected files from the list
// This function also affects the InfoBox (total file count, total file size)
func (ui *UIStatus) removeSelectedFiles(fileTreeView *gtk.TreeView) (err error) {
	listStore, err := ui.getListStoreWithId("fileList")
	if err != nil {
		return err
	}
	selected, err := fileTreeView.GetSelection()
	if err != nil {
		log.Debug(err)
		log.Error("Error while getting selected files")
		return err
	}
	rows := selected.GetSelectedRows(listStore)
	// Reverse is called to preserve the head of the linked list for iter.
	// Without Reserve, not all nodes will be deleted properly.
	reversed := rows.Reverse()
	reversed.Foreach(func(item interface{}) {
		// Convert interface to *gtk.TreePath
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
			log.Error("Error while getting full path from iterator")
			return
		}
		// Get full path as a string (full path works as a unique key for map)
		fullPath, err := value.GetString()
		if err != nil {
			log.Debug(err)
			log.Error("Error while getting string from *glib.Value")
			return
		}
		log.Debug("Full path: ", fullPath)
		value, err = listStore.GetValue(iter, fileSizeInBytes)
		if err != nil {
			log.Debug(err)
			log.Error("Error while getting value from GetValue")
			return
		}
		goValue, err := value.GoValue()
		if err != nil {
			log.Debug(err)
			log.Error("Error while getting value from GoValue")
			return
		}
		// Get file size in bytes
		size, ok := goValue.(int64)
		if !ok {
			log.Debug(AssertFailed)
			log.Error("Returned value is not int64")
			return
		}
		// Delete the file from the map
		delete(ui.fileMap, fullPath)
		// Delete the file from the UI list
		_ = listStore.Remove(iter)

		// Update total file info
		ui.totalFileSize -= size
		ui.totalFileCount -= 1
	})

	// Update InfoBox with new file count/size values
	ui.updateInfoBox()
	return nil
}

// updateInfoBox updates total file count and size
func (ui *UIStatus) updateInfoBox() {
	fileCountLabel, err := ui.getLabelWithId("infoFileCount")
	if err != nil {
		return
	}
	fileSizeLabel, err := ui.getLabelWithId("infoFileSize")
	if err != nil {
		return
	}
	fileCountLabel.SetLabel(strconv.Itoa(ui.totalFileCount))
	fileSizeLabel.SetLabel(sizeAddUnit(ui.totalFileSize))
}

// addFilesToListStore adds files in fileNames to the fileList
// This function also affects the InfoBox (total file count, total file size)
func (ui *UIStatus) addFilesToListStore(fileNames []string) {
	fileList, err := ui.getListStoreWithId("fileList")
	if err != nil {
		return
	}
	// For each files
	for _, fileName := range fileNames {
		// Check if file already exist
		if _, exist := ui.fileMap[fileName]; exist {
			log.Debug("File is already added; Skipping...")
			continue
		}

		// Add to set
		ui.fileMap[fileName] = struct{}{}

		// Get file name (without path)
		_, fName := filepath.Split(fileName)

		// Get file stat to get file size
		s, err := os.Stat(fileName)
		if err != nil {
			log.Debug(err)
			log.Error("Error while getting stats; Skipping...")
			continue
		}

		// Get file size in bytes
		size := s.Size()

		// Define new row
		// Only the first three values are visible;
		// 4th value is full file path that is used to keep track of files and provide a tooltip
		// 5th value is a full size used for InfoBox size to calculate correct values
		row := []interface{}{fName, sizeAddUnit(size), "Pending", fileName, size}

		// Add new row to the list
		iter := fileList.Append()
		if err = fileList.Set(iter, ui.fileListOrder, row); err != nil {
			log.Debug("Error while adding ", fileName)
			continue
		}

		// Update total file info
		ui.totalFileSize += size
		ui.totalFileCount += 1
	}
	// Update total file labels
	ui.updateInfoBox()
}

// sizeAddUnit converts size in bytes to size with appropriate file unit
func sizeAddUnit(size int64) (sizeStr string) {
	// Decimal points arent too important, so omit it for space in UI
	if size < 1e+3 {
		sizeStr = strconv.Itoa(int(size)) + " B"
	} else if size < 1e+6 {
		sizeStr = strconv.Itoa(int(size/1e+3)) + " KB"
	} else if size < 1e+9 {
		sizeStr = strconv.Itoa(int(size/1e+6)) + " MB"
	} else {
		sizeStr = strconv.Itoa(int(size/1e+9)) + " GB"
	}
	return sizeStr
}

// getPopoverWithId returns Popover with a provided id. If found, err != nil.
func (ui *UIStatus) getPopoverWithId(popoverId string) (popover *gtk.Popover, err error) {
	object, err := ui.builder.GetObject(popoverId)
	if err != nil {
		log.Debug(err)
		log.Errorf("Error while getting popover with popover id: %s", popoverId)
		return nil, err
	}
	popover, ok := object.(*gtk.Popover)
	if ok {
		return popover, nil
	}
	log.Debug(AssertFailed)
	log.Error("object is not a popover")
	return nil, AssertFailed
}

// getListBoxRowWithId returns ListBox with a provided id. If found, err != nil.
func (ui *UIStatus) getListBoxRowWithId(listBoxRowId string) (listBoxRow *gtk.ListBoxRow, err error) {
	object, err := ui.builder.GetObject(listBoxRowId)
	if err != nil {
		log.Debug(err)
		log.Errorf("Error while getting listBoxRow with listBoxRow id: %s", listBoxRowId)
		return nil, err
	}
	listBoxRow, ok := object.(*gtk.ListBoxRow)
	if ok {
		return listBoxRow, nil
	}
	log.Debug(AssertFailed)
	log.Error("object is not a listBoxRow")
	return nil, AssertFailed
}

// getLabelWithId returns Label with a provided id. If found, err != nil.
func (ui *UIStatus) getLabelWithId(labelId string) (label *gtk.Label, err error) {
	object, err := ui.builder.GetObject(labelId)
	if err != nil {
		log.Debug(err)
		log.Errorf("Error while getting label with label id: %s", labelId)
		return nil, err
	}
	label, ok := object.(*gtk.Label)
	if ok {
		return label, nil
	}
	log.Debug(AssertFailed)
	log.Error("object is not a label")
	return nil, AssertFailed
}

// getButtonWithId returns Button with a provided id. If found, err != nil.
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

// getTreeViewWithId returns TreeView with a provided id. If found, err != nil.
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

// getListStoreWithId returns ListStore with a provided id. If found, err != nil.
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

// getWindowWithId returns Window with a provided id. If found, err != nil.
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

// getExpanderWithId returns Expander with a provided id. If found, err != nil.
func (ui *UIStatus) getExpanderWithId(expanderId string) (expander *gtk.Expander, err error) {
	object, err := ui.builder.GetObject(expanderId)
	if err != nil {
		log.Debug(err)
		log.Errorf("Error while getting expander with expander id: %s", expanderId)
		return nil, err
	}
	expander, ok := object.(*gtk.Expander)
	if ok {
		return expander, nil
	}
	log.Debug(AssertFailed)
	log.Error("object is not an expander")
	return nil, AssertFailed
}

// getGridWithId returns Grid with a provided id. If found, err != nil.
func (ui *UIStatus) getGridWithId(gridId string) (grid *gtk.Grid, err error) {
	object, err := ui.builder.GetObject(gridId)
	if err != nil {
		log.Debug(err)
		log.Errorf("Error while getting grid with grid id: %s", gridId)
		return nil, err
	}
	grid, ok := object.(*gtk.Grid)
	if ok {
		return grid, nil
	}
	log.Debug(AssertFailed)
	log.Error("object is not an grid")
	return nil, AssertFailed
}

// isTreePath converts interface to TreePath.
func isTreePath(item interface{}) (*gtk.TreePath, error) {
	path, ok := item.(*gtk.TreePath)
	if ok {
		return path, nil
	}
	log.Debug(AssertFailed)
	log.Error("Item is not a TreePath")
	return nil, AssertFailed
}
