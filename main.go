package main

import (
	"archive/zip"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/cavaliercoder/grab"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Application name
// (preferably without spaces as this will be used in install path)
const appName = "APP_NAME"

// Base URL for downloaded files
// (%s gets replaced by current platform, windows/linux/darwin)
const baseUrl = "https://example.com/%s.zip"

// All files to download
// Only files ending with .zip are extracted
var files = []string{
	"%s.zip",
}

// Gets the username from whoami
// TODO: Cache this as username probably doesn't change during execution
func GetUsername() string {
	// Create the command and stdout pipe
	cmd := exec.Command("whoami")
	stdout, _ := cmd.StdoutPipe()

	// Start and check for errors
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Read first output
	// (it is the only output we expect)
	result, _ := ioutil.ReadAll(stdout)

	// Get a nice string
	username := strings.Trim(string(result), "\n ")

	// windows prints it as "<pcName>/<user>"
	if strings.Contains(username, "/") {
		username = username[strings.LastIndex(username, "/"):]
	}

	// Convert byte[] to string, trim and return
	return username
}

func GetTempPath() string {
	// darwin doesn't need username (probably)
	if runtime.GOOS == "darwin" {
		return "/tmp/"
	}
	// Try to match platform
	var dir string
	if runtime.GOOS == "windows" {
		dir = "C:/Users/%s/AppData/Local/Temp/"
	} else {
		dir = "/home/%s/.cache/"
	}
	// Get full temp path
	return fmt.Sprintf(dir, GetUsername())
}

// Gets the install path
func GetInstallPath() string {
	// Default current directory
	dir := "%s/%s/"
	// Try to match platform
	switch runtime.GOOS {
	case "windows":
		dir = "C:/Users/%s/AppData/Local/%s/"
	case "linux":
		dir = "/home/%s/.local/share/%s/"
	case "darwin":
		dir = "/home/%s/Applications/%s/"
	}
	// Return formatted string
	return fmt.Sprintf(dir, GetUsername(), appName)
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
		file := baseUrl
		// Check if we need the os
		if strings.Contains(files[i], "%s") {
			file += fmt.Sprintf(files[i], runtime.GOOS)
		} else {
			file += files[i]
		}
		fileName := GetFileFromPath(file)
		fmt.Println("Download:\t", file)
		status.SetText(fmt.Sprintf("[%d/%d] Downloading %s...", i + 1, len(files), fileName))
		// Create request
		request, err := grab.NewRequest(GetTempPath() + fileName, file)
		if err != nil {
			return err
		}
		// Get response
		response := client.Do(request)

		// Create ticker
		ticker := time.NewTicker(time.Millisecond)

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

// Attempts to extract input zip file to output destination
func Extract(input, output string, progress *widget.ProgressBar) error {
	// Try to open file
	reader, err := zip.OpenReader(input)
	if err != nil {
		return err
	}
	// Close reader when we're done
	defer func() {
		if err := reader.Close(); err != nil {
			panic(err)
		}
	}()
	// Helper function to extract each file in a zip
	extractAndWrite := func(file *zip.File) error {
		// Open file for reading
		readCloser, err := file.Open()
		if err != nil {
			return err
		}
		// Close file when we're done
		defer func() {
			if err := readCloser.Close(); err != nil {
				panic(err)
			}
		}()
		// Get full output path
		path := filepath.Join(output, file.Name)
		// If it's just a directory, create it only
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				return err
			}
		// If it's a file, actually extract it
		} else {
			// Create directory for file if needed
			if err := os.MkdirAll(filepath.Dir(path), file.Mode()); err != nil {
				return err
			}
			// Create output file
			outFile, err := os.OpenFile(path, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, file.Mode())
			if err != nil {
				return err
			}
			// Close output file after we're done
			defer func() {
				if err := outFile.Close(); err != nil {
					panic(err)
				}
			}()
			// Copy to output file
			_, err = io.Copy(outFile, readCloser)
			if err != nil {
				return err
			}
		}
		// Nothing went wrong, no error
		return nil
	}
	// Loop through all files in zip
	for i := 0; i < len(reader.File); i++ {
		// Get current file
		file := reader.File[i]
		// Update progress
		progress.SetValue(float64(i + 1) / float64(len(reader.File)))
		// Attempt to extract file
		err := extractAndWrite(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func Install(progress *widget.ProgressBar, status *widget.Label) error {
	// Create install directory if needed
	if err := os.MkdirAll(GetInstallPath(), 0700); err != nil {
		return err
	}
	// Loop over all files hopefully downloaded
	for i := 0; i < len(files); i++ {
		// Get file we're installing
		file := GetTempPath()
		// Check if we need os
		if strings.Contains(files[i], "%s") {
			file += fmt.Sprintf(files[i], runtime.GOOS)
		} else {
			file += files[i]
		}
		fileName := GetFileFromPath(file)
		fmt.Println("Install:\t", file)
		status.SetText(fmt.Sprintf("[%d/%d] Installing %s...", i + 1, len(files), fileName))
		// Check if zip file
		if strings.HasSuffix(fileName, ".zip") {
			// It's a zip file, extract it first
			if err := Extract(file, GetInstallPath(), progress); err != nil {
				return err
			}
			// Delete file after extracting
			if err := os.Remove(file); err != nil {
				return err
			}
		} else {
			// Any other file, just move it
			if err := os.Rename(file, GetInstallPath() + fileName); err != nil {
				return err
			}
		}
	}
	// Everything is fine
	return nil
}

func CreateShortcut() error {
	// darwin doesn't use shortcuts
	if runtime.GOOS == "darwin" {
		return nil
	}
	// linux uses a simple desktop file
	if runtime.GOOS == "linux" {
		// Create initial shortcut text
		// (this assumes executable contains os and icon doesn't)
		content := fmt.Sprintf("[Desktop Entry]\nName=%s\nType=Application\nTerminal=false\nExec=%s\nIcon=%s",
			appName, fmt.Sprintf("%s%s", GetInstallPath(), fmt.Sprintf(files[0], runtime.GOOS)),
			fmt.Sprintf("%s%s", GetInstallPath(), files[1]))
		// Try to write to file
		if err := ioutil.WriteFile(fmt.Sprintf("/home/%s/.local/share/applications/%s.desktop",
			GetUsername(), strings.ToLower(appName)), []byte(content), 0700); err != nil {
			return err
		}
	// windows uses annoying binary lnk files
	} else if runtime.GOOS == "windows" {
		// C:/Users/<user>/Roaming/Microsoft/Windows/Start Menu/Programs
		// We need to create a temporary Visual Basic file and then execute it
		location := fmt.Sprintf("C:/Users/%s/Roaming/Microsoft/Windows/Start Menu/Programs/%s.lnk", GetUsername(), appName)
		target   := fmt.Sprintf("%s%s", GetInstallPath(), fmt.Sprintf(files[0], runtime.GOOS))
		icon     := fmt.Sprintf("%s%s", GetInstallPath(), files[1])
		vbs := fmt.Sprintf("Set link = WScript.CreateObject(\"WScript.Shell\").CreateShortcut(\"%s\")\nlink.TargetPath = \"%s\"\nlink.IconLocation = \"%s\"\nlink.Description = \"%s\"\n link.Save", location, target, icon, appName)
		// Write vbs to file
		scriptFile := GetTempPath() + "CreateShortcut.vbs"
		if err := ioutil.WriteFile(scriptFile, []byte(vbs), 0777); err != nil {
			return err
		}
		// Execute script
		cmd := exec.Command("wscript", scriptFile)
		if err := cmd.Run(); err != nil {
			return err
		}
		if err := cmd.Wait(); err != nil {
			return err
		}
		if err := os.Remove(scriptFile); err != nil {
			return err
		}
	}

	return nil
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
			// Attempt install
			} else if err := Install(progress, status); err != nil {
				dialog.ShowError(err, parent)
				status.SetText("Install failed")
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