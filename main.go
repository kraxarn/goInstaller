package main

import (
	"fmt"
	"github.com/andlabs/ui"
)

// Application name
const appName = "APP_NAME"

// Base URL for downloaded files
// ({platform} gets replaced by current platform, windows/linux/darwin)
const baseUrl = "https://example.com/{platform}.zip"

// Main window
var mainWindow *ui.Window

/// Gets the username from whoami
func GetUsername() string {
	// Figure out what command to run
	name := "whoami"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	// Create the command and stdout pipe
	cmd := exec.Command(name)
	stdout, _ := cmd.StdoutPipe()

	// Start and check for errors
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Read first output
	// (it is the only output we expect)
	result, _ := ioutil.ReadAll(stdout)

	// Convert byte[] to string, trim and return
	return strings.Trim(fmt.Sprintf("%s", result), "\n ")
}

func GetTempPath() string {
	// If we're not on windows
	if runtime.GOOS != "windows" {
		return "/tmp"
	}
	// Get full Windows temp path
	return fmt.Sprintf("C:/Users/%s/AppData/Local/Temp", GetUsername())
}
func MakePage() ui.Control {
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
	// About option
	btnAbout := ui.NewButton("?")
	btnAbout.OnClicked(func(button *ui.Button) {
		btnAbout.Disable()

		aboutWindow := ui.NewWindow("About", 300, 300, false)

		licenseContent := ui.NewMultilineEntry()
		licenseContent.Append(licenses)

		tabs := ui.NewTab()
		tabs.Append("About", ui.NewLabel("goInstaller v0.1\nhttps://github.com/kraxarn/goInstaller\nLicensed under BSD-3"))
		tabs.SetMargined(0, true)
		tabs.Append("Licenses", licenseContent)
		tabs.SetMargined(1, true)
		aboutWindow.SetChild(tabs)

		aboutWindow.SetMargined(true)
		aboutWindow.Show()

		aboutWindow.OnClosing(func(window *ui.Window) bool {
			window.Hide()
			btnAbout.Enable()
			return true
		})
	})

	// Option buttons
	grid := ui.NewGrid()
	grid.SetPadded(true)
	grid.Append(btnCancel, 0, 0, 1, 1, true, ui.AlignFill, false, ui.AlignFill)
	grid.Append(ui.NewButton("Install"), 1, 0, 1, 1, true, ui.AlignFill, false, ui.AlignFill)
	vBox.Append(grid, false)

	return vBox
}

func SetupUi() {
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
	mainWindow.SetChild(MakePage())
	mainWindow.SetMargined(true)
	mainWindow.Show()
}

func main() {
	// Start application
	err := ui.Main(SetupUi)
	// Check if something went wrong
	if err != nil {
		fmt.Println("Error: ", err)
	}
}