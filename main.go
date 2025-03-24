package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var isServiceRunning bool
var config = GetConfigArray()

func main() {
	a := app.New()
	w := a.NewWindow(APP_NAME)
	a.Lifecycle().SetOnStarted(func() {
		setActivationPolicy()
	})

	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu(APP_NAME,
			fyne.NewMenuItem("Show Application", func() {
				w.Show()
			}))

		var ServiceStatus *fyne.MenuItem
		ServiceStatus = fyne.NewMenuItem("Service Status", func() {
			if isServiceRunning {
				dialog.ShowInformation("Service Status", "Service is running", w)
			} else {
				dialog.ShowInformation("Service Status", "Service is not running", w)
			}
		})
		var StartStopService *fyne.MenuItem
		StartStopService = fyne.NewMenuItem("Start Service", func() {
			if isServiceRunning {
				isServiceRunning = false
				StopService()
				StartStopService.Label = "Start Service"
				m.Refresh()
			} else {
				isServiceRunning = true
				StartService()
				StartStopService.Label = "Stop Service"
				m.Refresh()
			}
		})
		var Settings = fyne.NewMenuItem("Settings", func() {
			showSettings(a)
		})
		var OpenFolder = fyne.NewMenuItem("Open Folder", func() {
			OpenLocalFolder()
		})

		m.Items = append(m.Items, ServiceStatus)
		m.Items = append(m.Items, StartStopService)
		m.Items = append(m.Items, Settings)
		m.Items = append(m.Items, OpenFolder)

		desk.SetSystemTrayMenu(m)
	}

	w.SetContent(createMainScreen(a))
	w.Resize(fyne.NewSize(1200, 400))
	w.SetCloseIntercept(func() {
		w.Hide()
	})
	w.ShowAndRun()
}

func createMainScreen(a fyne.App) fyne.CanvasObject {
	var startStopButton *widget.Button
	startStopButton = widget.NewButton("Start Service", func() {
		if isServiceRunning {
			isServiceRunning = false
			startStopButton.SetText("Start Service")
			StopService()
		} else {
			isServiceRunning = true
			startStopButton.SetText("Stop Service")
			StartService()
		}
	})

	settingsButton := widget.NewButton("Settings", func() {
		showSettings(a)
	})

	startOnBootCheck := widget.NewCheck("Start On Boot", func(value bool) {
		// TODO: Implement start on boot logic
	})

	hideFromDockCheck := widget.NewCheck("Hide App From Dock", func(value bool) {
		// TODO: Implement hide from dock logic
	})

	return container.NewVBox(
		startStopButton,
		settingsButton,
		startOnBootCheck,
		hideFromDockCheck,
	)
}

func showSettings(a fyne.App) {
	w := a.NewWindow("Settings")

	intervalEntry := widget.NewEntry()
	intervalEntry.SetPlaceHolder("Interval (in seconds)")

	// Local target widgets
	localPathEntry := widget.NewEntry()
	localPathEntry.SetPlaceHolder("Local path")
	browseButton := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return
			}
			localPathEntry.SetText(uri.Path())
		}, w)
	})
	localFileRetentionEntry := widget.NewEntry()
	localFileRetentionEntry.SetPlaceHolder("Max File Retention")
	localContent := container.NewVBox(
		widget.NewLabel("Local Path:"),
		container.NewBorder(nil, nil, nil, browseButton, localPathEntry),
		widget.NewLabel("Max File Retention:"),
		localFileRetentionEntry,
	)

	// HTTP target widgets
	httpEndpointEntry := widget.NewEntry()
	httpEndpointEntry.SetPlaceHolder("HTTP Endpoint")
	httpContent := container.NewVBox(
		widget.NewLabel("HTTP Endpoint:"),
		httpEndpointEntry,
	)

	// S3 target widgets
	s3EndpointEntry := widget.NewEntry()
	s3EndpointEntry.SetPlaceHolder("S3 Endpoint")
	s3AccessKeyEntry := widget.NewEntry()
	s3AccessKeyEntry.SetPlaceHolder("Access Key")
	s3SecretKeyEntry := widget.NewEntry()
	s3SecretKeyEntry.SetPlaceHolder("Secret Key")
	s3BucketEntry := widget.NewEntry()
	s3BucketEntry.SetPlaceHolder("Bucket Name")
	s3Content := container.NewVBox(
		widget.NewLabel("S3 Endpoint:"),
		s3EndpointEntry,
		widget.NewLabel("Access Key:"),
		s3AccessKeyEntry,
		widget.NewLabel("Secret Key:"),
		s3SecretKeyEntry,
		widget.NewLabel("Bucket Name:"),
		s3BucketEntry,
	)

	// Target selection
	targetContent := container.NewVBox()
	targetSelect := widget.NewSelect([]string{"Local", "HTTP", "S3"}, func(value string) {
		targetContent.Objects = nil // Clear previous content
		switch value {
		case "Local":
			targetContent.Add(localContent)
		case "HTTP":
			targetContent.Add(httpContent)
		case "S3":
			targetContent.Add(s3Content)
		}
		targetContent.Refresh()
	})

	// Load config
	config := GetConfig()
	// {[{interval 4} {target Local} {local_path 4}]}
	entryMap := map[string]interface{}{
		"interval":             intervalEntry,
		"target":               targetSelect,
		"local_path":           localPathEntry,
		"http_endpoint":        httpEndpointEntry,
		"s3_endpoint":          s3EndpointEntry,
		"s3_access_key":        s3AccessKeyEntry,
		"s3_secret_key":        s3SecretKeyEntry,
		"s3_bucket_name":       s3BucketEntry,
		"local_file_retention": localFileRetentionEntry,
	}

	for _, conf := range config.Configurations {
		if entry, exists := entryMap[conf.Key]; exists {
			switch entry := entry.(type) {
			case *widget.Entry:
				entry.SetText(conf.Value)
			case *widget.Select:
				entry.SetSelected(conf.Value)
			}
		}
	}

	// Submit button
	submitButton := widget.NewButton("Save", func() {
		conf := Configurations{
			Configurations: []Configuration{
				{Key: "interval", Value: intervalEntry.Text},
				{Key: "target", Value: targetSelect.Selected},
			},
		}

		// Add target-specific configurations
		targetConfigs := map[string][]Configuration{
			"Local": {{Key: "local_path", Value: localPathEntry.Text}, {Key: "local_file_retention", Value: localFileRetentionEntry.Text}},
			"HTTP":  {{Key: "http_endpoint", Value: httpEndpointEntry.Text}},
			"S3": {
				{Key: "s3_endpoint", Value: s3EndpointEntry.Text},
				{Key: "s3_access_key", Value: s3AccessKeyEntry.Text},
				{Key: "s3_secret_key", Value: s3SecretKeyEntry.Text},
				{Key: "s3_bucket_name", Value: s3BucketEntry.Text},
			},
		}

		// Append target-specific configurations if they exist
		if configs, exists := targetConfigs[targetSelect.Selected]; exists {
			conf.Configurations = append(conf.Configurations, configs...)
		}
		err := WriteConfigurationFile(conf)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		// Update global config

		dialog.ShowInformation("Success", "Settings saved successfully", w)

	})

	content := container.NewVBox(
		widget.NewLabel("Interval:"),
		intervalEntry,
		widget.NewLabel("Target:"),
		targetSelect,
		targetContent,
		submitButton,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(1200, 500))
	w.Show()
}
