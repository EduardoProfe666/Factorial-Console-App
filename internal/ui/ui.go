﻿package ui

import (
	"factorial/internal/database"
	"factorial/internal/logic"
	"factorial/internal/ui/components"
	utils2 "factorial/internal/ui/utils"
	"factorial/internal/utils"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	url2 "net/url"
	"strconv"
	"sync"
	"time"
)

func SetupUI(w fyne.Window) {
	database.InitDB()
	utils.LogInfo("Factorial App Started")

	w.Resize(fyne.NewSize(500, 400))

	titleText := canvas.NewText("Factorial Calculator", theme.ForegroundColor())
	titleText.TextSize = 24
	titleText.Alignment = fyne.TextAlignCenter

	description := widget.NewLabelWithStyle("Enter a number or a range to calculate the factorial.", fyne.TextAlignCenter, fyne.TextStyle{})

	numberInput := widget.NewLabel("Number Input")
	rangeInput := widget.NewLabel("Range Input")
	rangeInput.Hide()

	entry := components.NewCustomEntry()
	entry.SetPlaceHolder("Enter a number")

	lowerRangeEntry := components.NewCustomEntry()
	lowerRangeEntry.SetPlaceHolder("Enter lower limit")
	lowerRangeEntry.Hide()

	upperRangeEntry := components.NewCustomEntry()
	upperRangeEntry.SetPlaceHolder("Enter upper limit")
	upperRangeEntry.Hide()

	calculateRangeCheckbox := widget.NewCheck("Calculate Range", func(checked bool) {
		if checked {
			entry.Hide()
			lowerRangeEntry.Show()
			upperRangeEntry.Show()
			numberInput.Hide()
			rangeInput.Show()
		} else {
			entry.Show()
			lowerRangeEntry.Hide()
			upperRangeEntry.Hide()
			numberInput.Show()
			rangeInput.Hide()
		}
	})

	resultLabel := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	resultLabel.Wrapping = fyne.TextWrapOff

	copyButton := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), nil)
	copyButton.OnTapped = func() {
		result := resultLabel.Text
		if len(result) > 8 {
			result = result[8:]
		}
		w.Clipboard().SetContent(result)
		copyButton.SetIcon(theme.ContentPasteIcon())
		time.AfterFunc(1*time.Second, func() {
			copyButton.SetIcon(theme.ContentCopyIcon())
		})
	}
	copyButton.Hide()

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Stop()
	progressBar.Hide()

	calculateButton := widget.NewButton("Calculate Factorial", func() {
		resultLabel.SetText("")
		copyButton.Hide()
		progressBar.Show()
		progressBar.Start()

		go func() {
			defer func() {
				progressBar.Stop()
				progressBar.Hide()
			}()

			if calculateRangeCheckbox.Checked {
				lowerLimit, err1 := strconv.Atoi(lowerRangeEntry.Text)
				upperLimit, err2 := strconv.Atoi(upperRangeEntry.Text)
				if err1 != nil || err2 != nil || lowerLimit > upperLimit {
					resultLabel.SetText("Invalid range values")
					utils.LogError("Invalid range values")
					return
				}

				var wg sync.WaitGroup
				for i := lowerLimit; i <= upperLimit; i++ {
					wg.Add(1)
					go func(num int) {
						defer wg.Done()
						_, err := database.GetFactorial(num)
						if err != nil {
							result := logic.Factorial(num)
							err = database.SaveResult(num, result.String())
							if err != nil {
								utils.LogError("Error saving result to database: " + err.Error())
							}
						}
					}(i)
				}
				wg.Wait()
				resultLabel.SetText("Range factorial calculation completed.")
			} else {
				number, err := strconv.Atoi(entry.Text)
				if err != nil {
					resultLabel.SetText("Invalid input")
					utils.LogError("Invalid input")
					return
				}

				result, err := database.GetFactorial(number)
				if err == nil {
					utils.LogResult(number, true)
					resultLabel.SetText("Result: " + utils2.TruncateString(result, 100))
					copyButton.Show()
				} else {
					result := logic.Factorial(number)
					utils.LogResult(number, false)
					resultLabel.SetText("Result: " + utils2.TruncateString(result.String(), 100))
					copyButton.Show()

					err = database.SaveResult(number, result.String())
					if err != nil {
						resultLabel.SetText("Error saving result to database")
						utils.LogError("Error saving result to database: " + err.Error())
					}
				}
			}
		}()
	})

	inputContainer := container.NewVBox(
		numberInput,
		entry,
		rangeInput,
		lowerRangeEntry,
		upperRangeEntry,
	)

	resultContainer := container.NewHBox(
		container.NewCenter(resultLabel),
		container.NewCenter(copyButton),
	)

	image := canvas.NewImageFromFile("resources/icon.png")
	image.SetMinSize(fyne.NewSize(100, 100))
	image.FillMode = canvas.ImageFillContain
	emojiText := canvas.NewText("🎩", theme.ForegroundColor())
	emojiText.TextSize = 24
	emojiText.Alignment = fyne.TextAlignCenter

	emojiLink := widget.NewButton("", func() {
		url, _ := url2.Parse("https://eduardoprofe666.github.io/")
		fyne.CurrentApp().OpenURL(url)
	})
	emojiLink.SetText("🎩")

	titleContainer := container.NewHBox(titleText, emojiLink)
	titleContainer = container.NewCenter(titleContainer)

	menu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Export to CSV", func() {
				progressBar.Show()
				progressBar.Start()

				go func() {
					defer func() {
						progressBar.Stop()
						progressBar.Hide()
					}()

					err := database.ExportToCSV("results.csv")
					if err != nil {
						utils.LogError("Error exporting to CSV: " + err.Error())
						resultLabel.SetText("Error exporting to CSV")
					} else {
						resultLabel.SetText("Results exported to results.csv")
					}
					copyButton.Hide()
				}()
			}),
			fyne.NewMenuItem("View Data", func() {
				progressBar.Show()
				progressBar.Start()

				go func() {
					defer func() {
						progressBar.Stop()
						progressBar.Hide()
					}()

					results, err := database.GetResults()
					if err != nil {
						utils.LogError("Error retrieving results: " + err.Error())
						resultLabel.SetText("Error retrieving results")
						return
					}

					var items []fyne.CanvasObject
					for _, result := range results {
						label := widget.NewLabelWithStyle(strconv.Itoa(result.Number)+": "+utils2.TruncateString(result.Result, 20), fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
						label.Wrapping = fyne.TextWrapOff
						copyButton := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
							w.Clipboard().SetContent(result.Result)
						})
						items = append(items, container.NewHBox(label, copyButton))
					}

					totalLabel := widget.NewLabel("Total results: " + strconv.Itoa(len(results)))
					scrollContainer := container.NewVScroll(container.NewVBox(append([]fyne.CanvasObject{totalLabel}, items...)...))
					scrollContainer.SetMinSize(fyne.NewSize(400, 300))

					dialog.NewCustom("Saved Results", "Close", scrollContainer, w).Show()
				}()
			}),
			fyne.NewMenuItem("Delete Data", func() {
				dialog.NewConfirm("Warning", "Are you sure you want to delete all data?", func(confirmed bool) {
					if confirmed {
						progressBar.Show()
						progressBar.Start()

						go func() {
							defer func() {
								progressBar.Stop()
								progressBar.Hide()
							}()

							err := database.ClearResults()
							if err != nil {
								utils.LogError("Error deleting data: " + err.Error())
								resultLabel.SetText("Error deleting data")
							} else {
								resultLabel.SetText("All data deleted")
							}
						}()
					}
				}, w).Show()
			}),
		),
		fyne.NewMenu("Settings",
			fyne.NewMenuItem("Preferences", func() {
				showPreferencesDialog(w)
			}),
		),
	)
	w.SetMainMenu(menu)

	progressBarContainer := container.NewHBox(
		layout.NewSpacer(),
		progressBar,
		layout.NewSpacer(),
	)

	content := container.NewVBox(
		titleContainer,
		container.NewCenter(image),
		description,
		inputContainer,
		calculateRangeCheckbox,
		calculateButton,
		resultContainer,
		container.NewCenter(progressBarContainer),
	)

	w.SetContent(content)

	w.SetOnClosed(func() {
		utils.LogInfo("Factorial App Exited")
	})
}

func showPreferencesDialog(w fyne.Window) {
	darkModeCheckbox := widget.NewCheck("Dark Mode", func(checked bool) {
		if checked {
			fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
		} else {
			fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
		}
	})
	fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
	darkModeCheckbox.Checked = true

	languageSelect := widget.NewSelect([]string{"English", "Spanish", "French"}, func(selected string) {
		// Handle language selection
	})
	languageSelect.SetSelected("English")

	preferencesContent := container.NewVBox(
		widget.NewLabel("Preferences"),
		darkModeCheckbox,
		widget.NewLabel("Language:"),
		languageSelect,
	)

	footer := container.NewBorder(nil, nil, nil, nil,
		container.NewVBox(
			widget.NewSeparator(),
			widget.NewLabelWithStyle("©️ "+strconv.Itoa(time.Now().Year())+" EduardoProfe666 🎩", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		),
	)

	emojiLink := widget.NewButton("", func() {
		url, _ := url2.Parse("https://eduardoprofe666.github.io/")
		fyne.CurrentApp().OpenURL(url)
	})
	emojiLink.SetText("🎩")

	preferencesDialog := dialog.NewCustom("Preferences", "Close", container.NewVBox(preferencesContent, footer), w)
	preferencesDialog.Show()
}
