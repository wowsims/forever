package core

import (
	"fmt"
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
)

// Gives every action of an APL rotation the path of its config entry, counted from 1 with hidden entries
// included. Auto and Simple runs and the custom rotation of pets and targets get no entry paths; a rotation
// with no type runs as an APL and gets them.
func (rot *APLRotation) assignEntryPaths(config *proto.APLRotation) {
	if config.Type == proto.APLRotation_TypeAuto || config.Type == proto.APLRotation_TypeSimple {
		return
	}
	if len(rot.priorityList) == 1 {
		if _, ok := rot.priorityList[0].impl.(*APLActionCustomRotation); ok {
			return
		}
	}
	for i, action := range rot.prepullActions {
		rot.assignEntryPath(action, fmt.Sprintf("prepull %d", rot.prepullIdxMap[i]+1))
	}
	for i, action := range rot.priorityList {
		rot.assignEntryPath(action, fmt.Sprintf("priority %d", rot.priorityListIdxMap[i]+1))
	}
}

// A group's actions and a sequence's steps add their own segment; any other inner action runs for its
// parent's entry path. GetInnerActions lists every descendant, so the positional children go first.
func (rot *APLRotation) assignEntryPath(action *APLAction, entryPath string) {
	action.entryPath = entryPath
	action.rot = rot
	assignSteps := func(steps []*APLAction, stepIdx []int) {
		for i, step := range steps {
			rot.assignEntryPath(step, fmt.Sprintf("%s > step %d", entryPath, stepIdx[i]+1))
		}
	}
	switch impl := action.impl.(type) {
	case *APLActionGroupReference:
		if impl.group == nil {
			return
		}
		slot := slices.Index(rot.groups, impl.group)
		for i, groupAction := range impl.group.actions {
			rot.assignEntryPath(groupAction, fmt.Sprintf("%s > group %d.%d", entryPath, impl.group.configIdx+1, rot.groupListIdxMap[slot][i]+1))
		}
	case *APLActionSequence:
		assignSteps(impl.subactions, impl.subactionIdx)
	case *APLActionStrictSequence:
		assignSteps(impl.subactions, impl.subactionIdx)
	case *APLActionStrictMultidot:
		for _, target := range impl.actions {
			target.entryPath = entryPath
			target.rot = rot
		}
	}
	for _, inner := range action.impl.GetInnerActions() {
		if inner.entryPath == "" {
			rot.assignEntryPath(inner, entryPath)
		}
	}
}
