package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type AppConfig struct {
	data binding.ExternalStringList
}

type traceeEvent struct {
	Context struct {
		Timestamp   int64  `json:"timestamp"`
		EventName   string `json:"eventName"`
		HostName    string `json:"hostName"`
		ProcessId   string `json:"processId"`
		ProcessName string `json:"processName"`
	} `json:"Context"`
	SigMetadata struct {
		ID          string `json:"ID"`
		Description string `json:"Description"`
	}
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

	events := make(chan traceeEvent, 100)
	go generateTraceeEvents(events)
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

func generateTraceeEvents(events chan traceeEvent) {
	handleEventFunc := func(w http.ResponseWriter, r *http.Request) {
		log.Println("received an event")
		b, _ := io.ReadAll(r.Body)
		log.Println(string(b))
		var te traceeEvent
		if err := json.Unmarshal(b, &te); err != nil {
			log.Print(err)
		}
		events <- te
	}

	http.HandleFunc("/events", handleEventFunc)
	http.ListenAndServe(":8888", nil)
}

func addTraceeEvents(data binding.ExternalStringList, events chan traceeEvent) {
	for {
		select {
		case e := <-events:
			data.Append(fmt.Sprintf(`Timestamp: %d | EventName: %s | HostName: %s | ID: %s | Description: %s`, e.Context.Timestamp, e.Context.EventName, e.Context.HostName, e.SigMetadata.ID, e.SigMetadata.Description))
		default:
			log.Println("got nothing to do....")
			time.Sleep(time.Second)
		}
	}
}
