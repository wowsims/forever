import type { IMessageType } from '@protobuf-ts/runtime';
import type { IndividualSimHost } from '@sim/sim_host';
import { savedDataEntries } from '@ui-kit/hooks/useSavedData';

export interface SavedRef {
	name: string;
	preset: boolean;
}

export const sameRef = (a: SavedRef, b: SavedRef) => a.name === b.name && a.preset === b.preset;

export const isRefChecked = (checked: readonly SavedRef[], ref: SavedRef) => checked.some(item => sameRef(item, ref));

// The rows a group shows, in list order. One equal to the current value is left out, since the current one runs anyway;
// a checked one equal to an earlier checked row is flagged. A checked ref whose entry no longer exists has no row.
export const checkedRefList = <E extends { ref: SavedRef }>(
	entries: readonly E[],
	checked: readonly SavedRef[],
	equalsCurrent: (entry: E) => boolean,
	same: (a: E, b: E) => boolean,
): Array<E & { checked: boolean; duplicate: boolean }> => {
	const picked: E[] = [];
	return entries
		.filter(entry => !equalsCurrent(entry))
		.map(entry => {
			const isChecked = isRefChecked(checked, entry.ref);
			const duplicate = isChecked && picked.some(earlier => same(earlier, entry));
			if (isChecked) picked.push(entry);
			return { ...entry, checked: isChecked, duplicate };
		});
};

// The checked refs, rebuilt from every pickable ref (rows hidden as equal to the current value included), so the stored
// order follows the list, a hidden row keeps its state, and refs to deleted entries go.
export const checkedRefsAfter = <R extends SavedRef>(stored: readonly R[], ref: R, checked: boolean, pickable: readonly R[]): R[] =>
	pickable.filter(item => (sameRef(item, ref) ? checked : isRefChecked(stored, item)));

// The record a Saved panel writes under `key`, read as that panel reads it.
export const storedSavedData = <T extends object>(host: IndividualSimHost<any>, key: string, type: IMessageType<T>): Array<{ name: string; data: T }> => {
	let record: unknown;
	try {
		record = JSON.parse(host.sim.env.storage.getItem(key) ?? '{}');
	} catch {
		return [];
	}
	return savedDataEntries(record, type);
};
