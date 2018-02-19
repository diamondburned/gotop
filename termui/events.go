package termui

import (
	"strconv"

	tb "github.com/nsf/termbox-go"
)

var eventStream = EventStream{
	make(map[string]func(Event)),
	"",
	make(chan bool, 1),
	make(chan tb.Event),
}

type EventStream struct {
	eventHandlers map[string]func(Event)
	prevKey       string
	stopLoop      chan bool
	eventQueue    chan tb.Event
}

type Event struct {
	Key    string
	Width  int
	Height int
	MouseX int
	MouseY int
}

// handleEvent calls the approriate callback function if there is one.
func handleEvent(e tb.Event) {
	if e.Type == tb.EventError {
		panic(e.Err)
	}

	ne := convertTermboxEvent(e)

	if val, ok := eventStream.eventHandlers[ne.Key]; ok {
		val(ne)
		eventStream.prevKey = ""
	} else { // check if the last 2 keys form a key combo with a handler
		// if this is a keyboard event and the previous event was unhandled
		if e.Type == tb.EventKey && eventStream.prevKey != "" {
			combo := eventStream.prevKey + ne.Key
			if val, ok := eventStream.eventHandlers[combo]; ok {
				ne.Key = combo
				val(ne)
				eventStream.prevKey = ""
			} else {
				eventStream.prevKey = ne.Key
			}
		} else {
			eventStream.prevKey = ne.Key
		}
	}
}

// Loop gets events from termbox and passes them off to handleEvent.
// Stops when StopLoop is called.
func Loop() {
	go func() {
		for {
			eventStream.eventQueue <- tb.PollEvent()
		}
	}()

	for {
		select {
		case <-eventStream.stopLoop:
			return
		case e := <-eventStream.eventQueue:
			handleEvent(e)
		}
	}
}

// StopLoop stops the events Loop
func StopLoop() {
	eventStream.stopLoop <- true
}

// On assigns event names to their handlers. Takes a string, strings, or a slice of strings, and a function.
func On(things ...interface{}) {
	function := things[len(things)-1].(func(Event))
	for _, thing := range things {
		if value, ok := thing.(string); ok {
			eventStream.eventHandlers[value] = function
		}
		if value, ok := thing.([]string); ok {
			for _, name := range value {
				eventStream.eventHandlers[name] = function
			}
		}
	}
}

// convertTermboxKeyValue converts a termbox keyboard event to a more friendly string format.
// Combines modifiers into the string instead of having them as additional fields in an event.
func convertTermboxKeyValue(e tb.Event) string {
	k := string(e.Ch)
	pre := ""
	mod := ""

	if e.Mod == tb.ModAlt {
		mod = "M-"
	}
	if e.Ch == 0 {
		if e.Key > 0xFFFF-12 {
			k = "<f" + strconv.Itoa(0xFFFF-int(e.Key)+1) + ">"
		} else if e.Key > 0xFFFF-25 {
			ks := []string{"<insert>", "<delete>", "<home>", "<end>", "<previous>", "<next>", "<up>", "<down>", "<left>", "<right>"}
			k = ks[0xFFFF-int(e.Key)-12]
		}

		if e.Key <= 0x7F {
			pre = "C-"
			k = string('a' - 1 + int(e.Key))
			kmap := map[tb.Key][2]string{
				tb.KeyCtrlSpace:     {"C-", "<space>"},
				tb.KeyBackspace:     {"", "<backspace>"},
				tb.KeyTab:           {"", "<tab>"},
				tb.KeyEnter:         {"", "<enter>"},
				tb.KeyEsc:           {"", "<escape>"},
				tb.KeyCtrlBackslash: {"C-", "\\"},
				tb.KeyCtrlSlash:     {"C-", "/"},
				tb.KeySpace:         {"", "<space>"},
				tb.KeyCtrl8:         {"C-", "8"},
			}
			if sk, ok := kmap[e.Key]; ok {
				pre = sk[0]
				k = sk[1]
			}
		}
	}

	return pre + mod + k
}

// convertTermboxMouseValue turns termbox mouse events into strings
func convertTermboxMouseValue(e tb.Event) string {
	switch e.Key {
	case tb.MouseLeft:
		return "MouseLeft"
	case tb.MouseMiddle:
		return "MouseMiddle"
	case tb.MouseRight:
		return "MouseRight"
	case tb.MouseWheelUp:
		return "MouseWheelUp"
	case tb.MouseWheelDown:
		return "MouseWheelDown"
	case tb.MouseRelease:
		return "MouseRelease"
	}
	return ""
}

// convertTermboxEvent turns a termbox event into a termui event
func convertTermboxEvent(e tb.Event) Event {
	ne := Event{} // new event

	switch e.Type {
	case tb.EventKey:
		ne.Key = convertTermboxKeyValue(e)
	case tb.EventMouse:
		ne.Key = convertTermboxMouseValue(e)
		ne.MouseX = e.MouseX
		ne.MouseY = e.MouseY
	case tb.EventResize:
		ne.Key = "resize"
		ne.Width = e.Width
		ne.Height = e.Height
	}

	return ne
}
