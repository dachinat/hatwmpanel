package native

/*
#cgo CFLAGS: -D_GNU_SOURCE -I${SRCDIR}/../protocols
#include "panel.h"
*/
import "C"

import "hatwmpanel/internal/hatwm/workspaces"

func (p *Panel) SetWorkspaces(snapshot workspaces.WorkspaceSnapshot) {
	if p == nil || p.ptr == nil {
		return
	}

	entries := snapshot.Workspaces
	if len(entries) == 0 && snapshot.Count > 0 {
		entries = make([]workspaces.IPCWorkspace, snapshot.Count)
		for i := range entries {
			entries[i] = workspaces.IPCWorkspace{Number: i + 1, Active: i+1 == snapshot.Current}
		}
	}

	cWorkspaces := make([]C.HatWMPanelWorkspace, len(entries))
	for i, workspace := range entries {
		active := 0
		if workspace.Active || workspace.Number == snapshot.Current {
			active = 1
		}
		focused := 0
		if workspace.Focused {
			focused = 1
		}
		urgent := 0
		if workspace.Urgent {
			urgent = 1
		}
		cWorkspaces[i] = C.HatWMPanelWorkspace{
			number:  C.int(workspace.Number),
			active:  C.int(active),
			focused: C.int(focused),
			urgent:  C.int(urgent),
			windows: C.int(workspace.Windows),
		}
	}

	cWindows := make([]C.HatWMPanelWindow, len(snapshot.Windows))
	for i, window := range snapshot.Windows {
		mapped, focused, fullscreen := 0, 0, 0
		if window.Mapped {
			mapped = 1
		}
		if window.Focused {
			focused = 1
		}
		if window.Fullscreen {
			fullscreen = 1
		}
		cWindows[i] = C.HatWMPanelWindow{
			id:         C.uint64_t(window.ID),
			workspace:  C.int(window.Workspace),
			mapped:     C.int(mapped),
			focused:    C.int(focused),
			fullscreen: C.int(fullscreen),
			x:          C.int(window.X),
			y:          C.int(window.Y),
			width:      C.int(window.Width),
			height:     C.int(window.Height),
		}
	}

	output := C.HatWMPanelOutput{
		x:             C.int(snapshot.Output.X),
		y:             C.int(snapshot.Output.Y),
		width:         C.int(snapshot.Output.Width),
		height:        C.int(snapshot.Output.Height),
		usable_x:      C.int(snapshot.Output.UsableX),
		usable_y:      C.int(snapshot.Output.UsableY),
		usable_width:  C.int(snapshot.Output.UsableWidth),
		usable_height: C.int(snapshot.Output.UsableHeight),
	}

	var workspacePtr *C.HatWMPanelWorkspace
	if len(cWorkspaces) > 0 {
		workspacePtr = &cWorkspaces[0]
	}
	var windowPtr *C.HatWMPanelWindow
	if len(cWindows) > 0 {
		windowPtr = &cWindows[0]
	}
	C.hatwm_panel_set_workspaces(
		p.ptr,
		workspacePtr, C.int(len(cWorkspaces)),
		windowPtr, C.int(len(cWindows)),
		&output,
	)
}

func (p *Panel) TakeWorkspaceClick() int {
	if p == nil || p.ptr == nil {
		return 0
	}
	return int(C.hatwm_panel_take_workspace_click(p.ptr))
}
