import { useSimHost, useSpecPresets } from '@sim/context/SimHostContext';
import { useSimReady } from '@sim/hooks/useSimReady';
import type { IndividualSimHost } from '@sim/sim_host';
import type { IndividualSimUIConfig } from '@sim/spec_config';
import type { UseSavedDataResult } from '@ui-kit/hooks/useSavedData';
import type { SavedDataPanelEntry, SavedDataPanelProps } from '@ui-kit/SavedDataPanel';
import { useCallback, useMemo, useRef } from 'react';

import { trackEvent } from '../../tracking/utils';

export type SpecPresets = Readonly<IndividualSimUIConfig<any>['presets']>;

export interface SavedPanelOptions<T> {
	label: string;
	storage: Pick<UseSavedDataResult<T>, 'entries' | 'save' | 'remove'>;
	current: T | null;
	serialize: (data: T) => string;
	load: (entry: SavedDataPanelEntry<T>, host: IndividualSimHost<any>) => void;
	// The spec's presets for this slot, built once the sim is ready and again only when the host or
	// the spec's presets change. A module-level function, so its identity is stable.
	presets?: (specPresets: SpecPresets, host: IndividualSimHost<any>) => Array<SavedDataPanelEntry<T>>;
}

export type SavedPanelProps<T> = Pick<SavedDataPanelProps<T>, 'label' | 'presets' | 'userData' | 'currentJson' | 'onLoad' | 'onSave' | 'onDelete'>;

const NO_PRESETS: Array<never> = [];

/** Everything a `SavedDataPanel` needs but its labels: the presets, the user's entries and the three actions. */
export const useSavedPanel = <T>({ label, storage, current, serialize, load, presets: buildPresets }: SavedPanelOptions<T>): SavedPanelProps<T> => {
	const host = useSimHost();
	const specPresets = useSpecPresets();
	const ready = useSimReady();
	const { entries: userData, save, remove } = storage;

	const currentRef = useRef(current);
	currentRef.current = current;
	const loadRef = useRef(load);
	loadRef.current = load;

	const presets = useMemo(() => (ready && buildPresets ? buildPresets(specPresets, host) : NO_PRESETS), [ready, buildPresets, specPresets, host]);

	const currentJson = useMemo(() => (current ? serialize(current) : ''), [current, serialize]);

	const onLoad = useCallback(
		(entry: SavedDataPanelEntry<T>) => {
			loadRef.current(entry, host);
			trackEvent({ action: 'settings', category: 'load', label });
		},
		[label, host],
	);

	const onSave = useCallback(
		(name: string) => {
			const data = currentRef.current;
			if (data === null) return;
			save(name, data);
			trackEvent({ action: 'settings', category: 'save', label });
		},
		[label, save],
	);

	const onDelete = useCallback(
		(entry: SavedDataPanelEntry<T>) => {
			remove(entry.name);
			trackEvent({ action: 'settings', category: 'delete', label });
		},
		[label, remove],
	);

	return { label, presets, userData, currentJson, onLoad, onSave, onDelete };
};
