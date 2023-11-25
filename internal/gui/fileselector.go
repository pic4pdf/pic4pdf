package gui

import (
	"image/color"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/adrg/xdg"
	usbdrivedetector "github.com/deepakjois/gousbdrivedetector"
	"github.com/fsnotify/fsnotify"
)

type FileItem struct {
	widget.BaseWidget

	obj *fyne.Container

	LabelIcon   *widget.Icon
	Label       *widget.Label
	LabelButton *widget.Button
	IconButton  *widget.Button
	Overlay     *canvas.Rectangle
}

func (itm *FileItem) overlayColor() color.Color {
	s := fyne.CurrentApp().Settings()
	return s.Theme().Color(
		theme.ColorNameSelection,
		s.ThemeVariant(),
	)
}

func NewFileItem(label string, onTapped func(), icon fyne.Resource, onIconTapped func()) *FileItem {
	itm := &FileItem{
		LabelIcon:   widget.NewIcon(nil),
		Label:       widget.NewLabel(label),
		LabelButton: widget.NewButton("", onTapped),
		IconButton:  widget.NewButtonWithIcon("", icon, onIconTapped),
	}
	itm.ExtendBaseWidget(itm)
	return itm
}

func (itm *FileItem) ExtendBaseWidget(w fyne.Widget) {
	itm.BaseWidget.ExtendBaseWidget(w)
	itm.Label.Truncation = fyne.TextTruncateEllipsis
	itm.Overlay = canvas.NewRectangle(itm.overlayColor())
	itm.Overlay.Hide()
	spaceL := canvas.NewRectangle(color.Transparent)
	spaceL.SetMinSize(fyne.NewSize(2, 0))
	itm.obj = container.NewStack(
		itm.LabelButton,
		container.NewBorder(
			nil, nil, spaceL, nil,
			container.NewBorder(
				nil, nil,
				itm.LabelIcon, itm.IconButton,
				itm.Label,
			),
		),
		itm.Overlay,
	)
}

func (itm *FileItem) Refresh() {
	itm.Overlay.FillColor = itm.overlayColor()
	itm.obj.Refresh()
}

func (itm *FileItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(itm.obj)
}

type fileList struct {
	widget.BaseWidget

	entries []fs.DirEntry
	list    *widget.List

	OnDirTapped func(fs.DirEntry)
	IsSelected  func(id int, ent fs.DirEntry) bool
	OnAdded     func(fs.DirEntry)
	OnRemoved   func(fs.DirEntry)
}

func newFileList(onDirTapped func(fs.DirEntry), isSelected func(id int, ent fs.DirEntry) bool, onAdded func(fs.DirEntry), onRemoved func(fs.DirEntry)) *fileList {
	fl := &fileList{
		OnDirTapped: onDirTapped,
		IsSelected:  isSelected,
		OnAdded:     onAdded,
		OnRemoved:   onRemoved,
	}
	fl.ExtendBaseWidget(fl)
	return fl
}

func (fl *fileList) ExtendBaseWidget(w fyne.Widget) {
	fl.BaseWidget.ExtendBaseWidget(w)
	fl.list = widget.NewList(
		func() int {
			return len(fl.entries)
		}, func() fyne.CanvasObject {
			item := NewFileItem(
				"PLACEHOLDER",
				nil,
				theme.ContentAddIcon(),
				nil,
			)
			item.LabelButton.Disable()
			return item
		}, func(id widget.ListItemID, obj fyne.CanvasObject) {
			item := obj.(*FileItem)
			entry := fl.entries[id]
			item.Label.SetText(entry.Name())
			if fl.entries[id].IsDir() {
				item.LabelIcon.SetResource(theme.FolderIcon())
				item.LabelButton.Enable()
				item.IconButton.Hide()
			} else {
				item.LabelIcon.SetResource(theme.FileImageIcon())
				item.LabelButton.Disable()
				item.IconButton.Show()
			}
			item.LabelButton.OnTapped = func() {
				fl.OnDirTapped(entry)
			}
			var add, remove func()
			var updateSelected func()
			updateSelected = func() {
				add = func() {
					fl.OnAdded(entry)
					updateSelected()
				}
				remove = func() {
					fl.OnRemoved(entry)
					updateSelected()
				}
				sel := fl.IsSelected(id, entry)
				if sel {
					item.Overlay.Show()
					item.IconButton.SetIcon(theme.ContentRemoveIcon())
					item.IconButton.OnTapped = remove
				} else {
					item.Overlay.Hide()
					item.IconButton.SetIcon(theme.ContentAddIcon())
					item.IconButton.OnTapped = add
				}
			}
			updateSelected()
		},
	)
}

func (fl *fileList) SetEntries(ents []fs.DirEntry) {
	fl.entries = ents
	fl.Refresh()
}

func (fl *fileList) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(fl.list)
}

type FileSelector struct {
	widget.BaseWidget

	pathEntry     *widget.Entry
	backButton    *widget.Button
	forwardButton *widget.Button
	showHidden    *widget.Check
	filter        *widget.Entry
	quickAccess   *fyne.Container
	list          *fileList
	listMessage   *widget.Label
	obj           *fyne.Container

	OnSelected   func(path string)
	OnUnselected func(path string)

	path                 string
	next                 []string
	selected             map[string]struct{}
	watcher              *fsnotify.Watcher
	preferencesID        string
	prevQuickAccessPaths []string
}

func (f *FileSelector) refreshList() {
	if f.watcher != nil {
		for _, v := range f.watcher.WatchList() {
			f.watcher.Remove(v)
		}
		f.watcher.Add(f.path)
	}
	ents, err := os.ReadDir(f.path)
	if err != nil {
		f.listMessage.Show()
		f.listMessage.SetText("Error: " + err.Error())
		f.list.SetEntries(nil)
		return
	}
	f.listMessage.Hide()
	var res []fs.DirEntry
	for _, ent := range ents {
		name := ent.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if !f.showHidden.Checked && strings.HasPrefix(name, ".") {
			continue
		}
		if f.filter.Text != "" {
			if !strings.Contains(strings.ToLower(name), strings.ToLower(f.filter.Text)) {
				continue
			}
		}
		if ent.IsDir() || ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp" {
			res = append(res, ent)
		}
	}
	f.list.SetEntries(res)
	if len(res) == 0 {
		f.listMessage.Show()
		if len(ents) == 0 {
			f.listMessage.SetText("This folder is empty.")
		} else {
			f.listMessage.SetText("This folder contains no matching files.")
		}
	}
	f.pathEntry.Text = f.path
	f.pathEntry.CursorColumn = len(f.path)
	f.pathEntry.Refresh()
	if len(f.next) == 0 {
		f.forwardButton.Disable()
	} else {
		f.forwardButton.Enable()
	}
	if f.path == filepath.Dir(f.path) {
		// Already at root.
		f.backButton.Disable()
	} else {
		f.backButton.Enable()
	}
}

func (f *FileSelector) refreshQuickAccess() {
	var quickAccessButtons []fyne.CanvasObject
	userDirPaths := []string{xdg.Home, xdg.UserDirs.Pictures, xdg.UserDirs.Documents, xdg.UserDirs.Download, xdg.UserDirs.Desktop}
	drivePaths, err := usbdrivedetector.Detect()
	if err != nil {
		l := widget.NewLabel("Error detecting removable drives: " + err.Error())
		l.Wrapping = fyne.TextWrapWord
		quickAccessButtons = append(quickAccessButtons, l)
		drivePaths = nil
	}
	sort.Strings(drivePaths)

	allPaths := append(userDirPaths, drivePaths...)

	if len(f.prevQuickAccessPaths) == len(allPaths) {
		pathsEq := true
		for i := range allPaths {
			if f.prevQuickAccessPaths[i] != allPaths[i] {
				pathsEq = false
				break
			}
		}
		if pathsEq {
			return
		}
	}
	f.prevQuickAccessPaths = allPaths

	f.quickAccess.RemoveAll()
	for i, p := range allPaths {
		p := p
		var icon fyne.Resource
		if i == 0 {
			icon = theme.HomeIcon()
		} else if i < len(userDirPaths) {
			icon = theme.FolderIcon()
		} else {
			icon = theme.StorageIcon()
		}
		l := widget.NewLabel(filepath.Base(p))
		l.Truncation = fyne.TextTruncateEllipsis
		f.quickAccess.Add(
			container.NewStack(
				widget.NewButton("", func() {
					f.cd(p)
				}),
				container.NewPadded(container.NewBorder(nil, nil, widget.NewIcon(icon), nil, l)),
			),
		)
	}
	f.quickAccess.Refresh()
}

func (f *FileSelector) cd(name string) {
	if filepath.IsAbs(name) {
		f.path = name
	} else {
		f.path = path.Join(f.path, name)
	}
	f.next = nil
	if f.preferencesID != "" {
		fyne.CurrentApp().Preferences().SetString("FileSelectorPath"+f.preferencesID, f.path)
	}
	f.filter.SetText("")
	f.refreshList()
}

func (f *FileSelector) back() {
	if f.path == filepath.Dir(f.path) {
		// Do nothing if already at root.
		return
	}
	f.next = append(f.next, f.path)
	f.path = filepath.Dir(f.path)
	f.refreshList()
}

func (f *FileSelector) forward() {
	if len(f.next) == 0 {
		return
	}
	f.path = f.next[len(f.next)-1]
	f.next = f.next[:len(f.next)-1]
	f.refreshList()
}

func NewFileSelector() *FileSelector {
	f := &FileSelector{
		path: xdg.Home,
	}
	f.ExtendBaseWidget(f)
	return f
}

func NewFileSelectorPersistent(preferencesID string) *FileSelector {
	f := &FileSelector{
		path:          fyne.CurrentApp().Preferences().StringWithFallback("FileSelectorPath"+preferencesID, xdg.Home),
		preferencesID: preferencesID,
	}
	f.ExtendBaseWidget(f)
	return f
}

func (f *FileSelector) ExtendBaseWidget(w fyne.Widget) {
	f.BaseWidget.ExtendBaseWidget(w)
	f.selected = make(map[string]struct{})
	f.pathEntry = widget.NewEntry()
	f.pathEntry.Validator = func(s string) error { return nil }
	f.pathEntry.OnChanged = func(s string) {
		if s != f.path {
			_, err := os.Stat(s)
			if err == nil {
				f.cd(s)
			}
			f.pathEntry.SetValidationError(err)
		}
	}
	f.backButton = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		f.back()
		f.filter.SetText("")
	})
	f.forwardButton = widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		f.forward()
		f.filter.SetText("")
	})
	f.showHidden = widget.NewCheck("Show Hidden", func(b bool) {
		f.refreshList()
	})
	f.filter = widget.NewEntry()
	searchIcon := widget.NewIcon(theme.SearchIcon())
	clearButton := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		f.filter.SetText("")
	})
	clearButton.Hide()
	f.filter.ActionItem = container.NewStack(searchIcon, clearButton)
	f.filter.PlaceHolder = "Filter"
	f.filter.OnChanged = func(s string) {
		if s == "" {
			searchIcon.Show()
			clearButton.Hide()
		} else {
			searchIcon.Hide()
			clearButton.Show()
		}
		f.refreshList()
	}
	f.quickAccess = container.NewGridWithColumns(3)
	quickAccessAccordion := widget.NewAccordion(widget.NewAccordionItem(
		"Quick Access",
		f.quickAccess,
	))
	quickAccessAccordion.Open(0)
	f.list = newFileList(func(de fs.DirEntry) {
		f.cd(de.Name())
	}, func(id int, ent fs.DirEntry) bool {
		_, ok := f.selected[path.Join(f.path, ent.Name())]
		return ok
	}, func(ent fs.DirEntry) {
		path := path.Join(f.path, ent.Name())
		f.selected[path] = struct{}{}
		if f.OnSelected != nil {
			f.OnSelected(path)
		}
	}, func(ent fs.DirEntry) {
		path := path.Join(f.path, ent.Name())
		delete(f.selected, path)
		if f.OnUnselected != nil {
			f.OnUnselected(path)
		}
	})
	f.listMessage = widget.NewLabel("")
	f.listMessage.Wrapping = fyne.TextWrapWord
	f.listMessage.Alignment = fyne.TextAlignCenter
	f.listMessage.Hide()
	f.listMessage.Resize(fyne.NewSize(280, 40))
	f.listMessage.Move(fyne.NewPos(-140, 0))
	f.obj = container.NewBorder(
		container.NewVBox(
			container.NewBorder(
				nil,
				nil,
				nil,
				container.NewBorder(nil, nil, nil, container.NewHBox(
					f.showHidden,
					f.backButton,
					f.forwardButton,
				)),
				f.pathEntry,
			),
			f.filter,
			quickAccessAccordion,
		),
		nil, nil, nil,
		container.NewStack(f.list, container.NewCenter(container.NewWithoutLayout(f.listMessage))),
	)
	f.refreshQuickAccess()
	f.refreshList()
}

func (f *FileSelector) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(f.obj)
}

// Watches the current directory and removable drives for changes.
//
// User must call close() when the watcher is no longer required, as well
// as provide an error handler.
//
// See CreateSimpleWatcher for a simpler version.
func (f *FileSelector) CreateWatcher(onError func(error)) (close func() error, err error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	func() {
		go func() {
			for {
				select {
				case _, ok := <-f.watcher.Events:
					if !ok {
						return
					}
					f.refreshList()
				case err, ok := <-f.watcher.Errors:
					if !ok {
						return
					}
					onError(err)
				}
			}
		}()
	}()

	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-time.Tick(500 * time.Millisecond):
				f.refreshQuickAccess()
			case <-quit:
				return
			}
		}
	}()

	f.watcher = w
	f.watcher.Add(f.path)

	return func() error {
		quit <- struct{}{}
		return w.Close()
	}, nil
}

// Watches the current directory and removable drives for changes.
//
// User MUST call close() on the returned watcher when it is no longer in use.
func (f *FileSelector) CreateSimpleWatcher() (close func()) {
	complexClose, err := f.CreateWatcher(func(err error) {
		log.Println("FileSelector: Watcher:", err)
	})
	if err != nil {
		log.Println("FileSelector.CreateSimpleWatcher:", err)
		return func() {}
	}

	return func() {
		if err := complexClose(); err != nil {
			log.Println("FileSelector: Close Watcher:", err)
		}
	}
}

func (f *FileSelector) IsSelected(path string) bool {
	_, ok := f.selected[path]
	return ok
}

func (f *FileSelector) Select(path string) {
	f.selected[path] = struct{}{}
	if f.OnSelected != nil {
		f.OnSelected(path)
	}
	f.refreshList()
}

func (f *FileSelector) Unselect(path string) {
	delete(f.selected, path)
	if f.OnUnselected != nil {
		f.OnUnselected(path)
	}
	f.refreshList()
}

func (f *FileSelector) Selected() (paths []string) {
	paths = make([]string, 0, len(f.selected))
	for p := range f.selected {
		paths = append(paths, p)
	}
	return
}

func (f *FileSelector) SetPath(path string) {
	f.path = path
	f.refreshList()
}
