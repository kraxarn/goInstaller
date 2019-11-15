package main

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/cavaliercoder/grab"
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Application name
const appName = "APP_NAME"

// Base URL for downloaded files
// (%s gets replaced by current platform, windows/linux/darwin)
const baseUrl = "https://example.com/%s.zip"

// All files to download
// Only files ending with .zip are extracted
var files = []string{
	"%s.zip",
}

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
		return "/tmp/"
	}
	// Get full Windows temp path
	return fmt.Sprintf("C:/Users/%s/AppData/Local/Temp/", GetUsername())
}

func GetFileFromPath(path string) string {
	// Try to get last index of /
	lastIndex := strings.LastIndex(path, "/") + 1
	// -1 + 1 = 0, so lastIndex is 0 if failed
	if lastIndex == 0 {
		return path
	}
	// Return final string
	return path[lastIndex:]
}

// Starts download and updates progress bar 0-50
func Download(progress *widget.ProgressBar, status *widget.Label) error {
	// Create HTTP client
	client := grab.NewClient()
	// Create a new request for each file to download
	for i := 0; i < len(files); {
		// Get file we're downloading
		file := baseUrl + fmt.Sprintf(files[i], runtime.GOOS)
		fileName := GetFileFromPath(file)
		fmt.Println("Attempting to download:", file)
		status.SetText(fmt.Sprintf("[%d/%d] Downloading %s...", i + 1, len(files), fileName))
		// Create request
		request, err := grab.NewRequest(GetTempPath() + fileName, file)
		if err != nil {
			return err
		}
		// Get response
		response := client.Do(request)

		// Create ticker
		ticker := time.NewTicker(time.Second)

		// Create variable for when to run loop
		run := true
		for run {
			select {
			// Check for progress
			case <-ticker.C:
				progress.SetValue(response.Progress())
			// Check if we're done
			case <-response.Done:
				if err := response.Err(); err != nil {
					// Something went wrong, stop ticker and return error
					ticker.Stop()
					return err
				}
				// File downloaded, stop ticker and go to next file
				ticker.Stop()
				run = false
			}
		}

		i++
	}

	return nil
}
	}
}

func MakeContent(parent fyne.Window) fyne.CanvasObject {
	// Install progress
	progress := widget.NewProgressBar()
	// Status message
	status := widget.NewLabel("Waiting...")

	// Install button
	var btnInstall *widget.Button
	btnInstall = widget.NewButton("Install", func() {
		go func() {
			btnInstall.Disable()
			progress.SetValue(0)
			// Attempt download
			if err := Download(progress, status); err != nil  {
				dialog.ShowError(err, parent)
				status.SetText("Download failed")
			} else {
				progress.SetValue(1)
				status.SetText("Installation successful!")
			}
			btnInstall.Enable()
		}()
	})

	return widget.NewVBox(
		// Label with what to install
		widget.NewGroup(fmt.Sprintf("Welcome to the %s installer!", appName), status),
		// Install progress
		progress,
		// Install button
		layout.NewSpacer(),
		btnInstall,
	)
}

func LoadIcon() fyne.Resource {
	return fyne.NewStaticResource("icon.png", icon)
}

func main() {
	// License window to refer to later
	var licenseWindow fyne.Window

	// Create new main app
	mainApp := app.New()
	// Create window
	window := mainApp.NewWindow("Installer")
	window.Resize(fyne.Size{Width: 400, Height: 200})
	window.CenterOnScreen()
	window.SetIcon(LoadIcon())

	// Set window menu
	window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("About", func() {
				dialog.ShowInformation(
					"About",
					"goInstaller v0.1\nhttps://github.com/kraxarn/goInstaller\nLicensed under BSD-3", window)
			}),
			fyne.NewMenuItem("Licenses", func() {
				// Check if we already have a license window open
				if licenseWindow != nil {
					return
				}
				// Create window with content and reset on close
				licenseWindow = fyne.CurrentApp().NewWindow("Licenses")
				licenseWindow.Resize(fyne.Size{Width: 0, Height: 800})
				licenseWindow.CenterOnScreen()
				licenseWindow.SetPadded(true)
				licenseWindow.SetContent(widget.NewScrollContainer(widget.NewLabel(licenses)))
				licenseWindow.Show()
				licenseWindow.SetOnClosed(func() {
					licenseWindow = nil
				})
			}),
		),
	))

	// Set what to show in the window
	window.SetContent(MakeContent(window))
	// Show window
	window.ShowAndRun()
}