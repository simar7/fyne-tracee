package main

import (
	"fmt"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
	"github.com/tjarratt/babble"
)

type AppConfig struct {
	data binding.ExternalStringList
}

type traceeEvent struct {
	Timestamp int64
	EventName string
	HostName  string
}

var traceeEvents []string

func main() {
	myApp := app.New()
	ac := AppConfig{}
	myWindow := myApp.NewWindow("Tracee Events")
	myWindow.Resize(fyne.Size{
		Width:  640,
		Height: 480,
	})
	myWindow.CenterOnScreen()

	data := binding.BindStringList(
		&traceeEvents,
	)
	ac.data = data

	list := widget.NewListWithData(data,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})

	events := generateTraceeEvents()
	go addTraceeEvents(data, events)

	entry := xwidget.NewCompletionEntry([]string{})
	entry.OnChanged = func(input string) {
		events, err := ac.data.Get()
		if err != nil {
			entry.HideCompletion()
			return
		}

		matches := fuzzy.Find(input, events)
		if len(matches) <= 0 {
			entry.HideCompletion()
			return
		}

		entry.SetOptions(matches[:5])
		entry.ShowCompletion()
	}

	myWindow.SetContent(container.NewBorder(entry, nil, nil, nil, list))
	myWindow.ShowAndRun()
}

func generateTraceeEvents() (events []traceeEvent) {
	babbler := babble.NewBabbler()
	for i := 0; i < 100; i++ {
		time.Sleep(time.Microsecond)
		events = append(events, traceeEvent{
			Timestamp: time.Now().Unix(),
			EventName: babbler.Babble(),
			HostName:  babbler.Babble(),
		})
	}
	return
}

func addTraceeEvents(data binding.ExternalStringList, events []traceeEvent) {
	for {
		if len(events) > 0 {
			for i := 0; i < len(events); i++ {
				data.Append(fmt.Sprintf(`Timestamp: %d | EventName: %s | HostName: %s`, events[i].Timestamp, events[i].EventName, events[i].HostName))
			}
		}
		time.Sleep(time.Second * 1)
	}
}
