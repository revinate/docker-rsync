package main

import (
	"github.com/fsnotify/fsevents"
	"sort"
	"time"
	"strings"
)

var noteDescription = map[fsevents.EventFlags]string{
	fsevents.MustScanSubDirs: "MustScanSubdirs",
	fsevents.UserDropped:     "UserDropped",
	fsevents.KernelDropped:   "KernelDropped",
	fsevents.EventIDsWrapped: "EventIDsWrapped",
	fsevents.HistoryDone:     "HistoryDone",
	fsevents.RootChanged:     "RootChanged",
	fsevents.Mount:           "Mount",
	fsevents.Unmount:         "Unmount",

	fsevents.ItemCreated:       "Created",
	fsevents.ItemRemoved:       "Removed",
	fsevents.ItemInodeMetaMod:  "InodeMetaMod",
	fsevents.ItemRenamed:       "Renamed",
	fsevents.ItemModified:      "Modified",
	fsevents.ItemFinderInfoMod: "FinderInfoMod",
	fsevents.ItemChangeOwner:   "ChangeOwner",
	fsevents.ItemXattrMod:      "XAttrMod",
	fsevents.ItemIsFile:        "IsFile",
	fsevents.ItemIsDir:         "IsDir",
	fsevents.ItemIsSymlink:     "IsSymLink",
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.Contains(a, b) {
			return true
		}
	}
	return false
}

func throttle(interval time.Duration, input chan []fsevents.Event, eventHandler func(id uint64, path string, flags []string)) {
	var (
		item []fsevents.Event
	)
OuterLoop:
	for {
		select {
		case item = <-input:
			//do nothing
		case <-time.After(interval):
			for _, event := range item {
				if (len(event.Path) > 0 && stringInSlice(event.Path, []string{".git", ".idea"})) {
					continue;
				}
				flags := make([]string, 0)
				for bit, description := range noteDescription {
					if event.Flags&bit == bit {
						flags = append(flags, description)
					}
				}
				sort.Sort(sort.StringSlice(flags))
				go eventHandler(event.ID, event.Path, flags)
				item = nil
				continue OuterLoop
			}
		}
	}
}


func Watch(path string, eventHandler func(id uint64, path string, flags []string)) {
	dev, _ := fsevents.DeviceForPath(path)
	fsevents.EventIDForDeviceBeforeTime(dev, time.Now())

	es := &fsevents.EventStream{
		Paths:   []string{path},
		Latency: 50 * time.Millisecond,
		Device:  dev,
		Flags:   fsevents.FileEvents | fsevents.WatchRoot}
	es.Start()
	ec := es.Events
	throttle(150*time.Millisecond, ec, eventHandler)
}
