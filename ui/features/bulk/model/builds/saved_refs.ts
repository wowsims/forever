import type { IMessageType, JsonReadOptions } from '@protobuf-ts/runtime';
import { fromJsonLenient } from '@sim/proto/lenient_json';
import type { IndividualSimHost } from '@sim/sim_host';
import { savedDataEntries } from '@ui-kit/hooks/useSavedData';

export interface SavedRef {
	name: string;
	preset: boolean;
	group?: string; // Only gear set refs carry one, empty or missing when ungrouped.
}

export const sameRef = (a: SavedRef, b: SavedRef) => a.name === b.name && a.preset === b.preset && (a.group ?? '') === (b.group ?? '');

const isRefChecked = (checked: readonly SavedRef[], ref: SavedRef) => checked.some(item => sameRef(item, ref));

// The checked refs, rebuilt from every pickable ref in list order, so the stored order follows the list, a ref the list
// shows no row for keeps its state, and refs to deleted entries go.
export const checkedRefsAfter = <R extends SavedRef>(stored: readonly R[], ref: R, checked: boolean, pickable: readonly R[]): R[] =>
	pickable.filter(item => (sameRef(item, ref) ? checked : isRefChecked(stored, item)));

// Each proto object is keyed once, so a list that shows again keys only the entries it has not seen.
export const keyedOnce = <T extends object>(key: (value: T) => string): ((value: T) => string) => {
	const keys = new WeakMap<T, string>();
	return value => {
		if (!keys.has(value)) keys.set(value, key(value));
		return keys.get(value)!;
	};
};

// The record a Saved panel writes under `key`, read leniently so an entry with fields the protos no longer list still
// loads, as the Saved Encounters and Saved Rotations panels read theirs.
export const storedSavedData = <T extends object>(
	host: IndividualSimHost<any>,
	key: string,
	type: IMessageType<T>,
	readOptions?: Partial<JsonReadOptions>,
): Array<{ name: string; data: T }> => {
	let record: unknown;
	try {
		record = JSON.parse(host.sim.env.storage.getItem(key) ?? '{}');
	} catch {
		return [];
	}
	return savedDataEntries(record, { toJson: data => type.toJson(data), fromJson: json => fromJsonLenient(type, json, readOptions) });
};
