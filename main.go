package main

import (
	"fmt"
	"github.com/andlabs/ui"
)

// Application preferences
const appName = "APP_NAME"

// Main window
var mainWindow *ui.Window

func makePage() ui.Control {
	// Main vertical layout
	vBox := ui.NewVerticalBox()
	vBox.SetPadded(true)

	// Title
	group := ui.NewGroup("This will install:")
	vBox.Append(group, false)

	// Text
	label := ui.NewLabel("\t∙ " + appName)
	vBox.Append(label, false)

	// Progress
	progress := ui.NewProgressBar()
	vBox.Append(progress, false)

	// Cancel option
	btnCancel := ui.NewButton("Cancel")
	btnCancel.OnClicked(func(button *ui.Button) {
		ui.Quit()
	})

	// Install option
	// TODO

	// Option buttons
	grid := ui.NewGrid()
	grid.SetPadded(true)
	grid.Append(btnCancel, 0, 0, 1, 1, true, ui.AlignFill, false, ui.AlignFill)
	grid.Append(ui.NewButton("Install"), 1, 0, 1, 1, true, ui.AlignFill, false, ui.AlignFill)
	vBox.Append(grid, false)

	return vBox
}

func setupUi() {
	// Create the main window
	mainWindow = ui.NewWindow("Installer", 300, 120, false)

	// Setup closing stuff
	mainWindow.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		mainWindow.Destroy()
		return true
	})

	// Add child to main window and show it
	mainWindow.SetChild(makePage())
	mainWindow.SetMargined(true)
	mainWindow.Show()
}

func main() {
	// Start application
	err := ui.Main(setupUi)
	// Check if something went wrong
	if err != nil {
		fmt.Println("Error: ", err)
	}
}