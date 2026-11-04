import { actionKinds } from '@features/apl/model/action_kinds';
import type { APLFieldDescriptor } from '@features/apl/model/field_descriptors';
import { type AplFieldSpec, newKindImpl, resolveField } from '@features/apl/model/field_specs';
import { type UNIT_SET, unitSets } from '@features/apl/model/unit_sets';
import { valueKinds } from '@features/apl/model/value_kinds';
import { APLAction, APLValue } from '@generated/proto/apl';
import { Class, UnitReference } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { FieldInfo, IMessageType } from '@protobuf-ts/runtime';
import type { Player } from '@sim/player/player';

import type { JsonObject, JsonPath, JsonValue as Json } from '../protocol';

export type { Json, JsonObject, JsonPath };

export type Family = 'action' | 'value';

interface KindModel {
	kind: string;
	label: string;
	submenu?: string[];
	shortDescription: string;
	fullDescription?: string;
	includeIf?: (player: Player<any>, isPrepull: boolean, isGroup: boolean) => boolean;
	newValue: () => unknown;
	fields: APLFieldDescriptor[];
}

const KINDS: Record<Family, Record<string, KindModel>> = {
	action: Object.fromEntries(Object.entries(actionKinds).map(([kind, model]) => [kind, { kind, ...model } as KindModel])),
	value: Object.fromEntries(Object.entries(valueKinds).map(([kind, model]) => [kind, { kind, ...model } as KindModel])),
};

const ONEOF: Record<Family, { type: IMessageType<any>; oneof: string }> = {
	action: { type: APLAction, oneof: 'action' },
	value: { type: APLValue, oneof: 'value' },
};

export function kindModel(family: Family, kind: string): KindModel | undefined {
	return KINDS[family][kind];
}

// The kind a JSON action or value holds: its one key that names a kind (an action also carries `condition`).
export function kindOf(family: Family, json: Json | undefined): string | undefined {
	if (!json || typeof json !== 'object' || Array.isArray(json)) {
		return undefined;
	}
	return Object.keys(json).find(key => key in KINDS[family]);
}

// The message type a kind's body is, read off the oneof in the proto.
export function kindMessage(family: Family, kind: string): IMessageType<any> | undefined {
	const field = ONEOF[family].type.fields.find(f => f.oneof === ONEOF[family].oneof && f.localName === kind);
	return field?.kind === 'message' ? field.T() : undefined;
}

export function fieldInfo(message: IMessageType<any> | undefined, field: string): FieldInfo | undefined {
	return message?.fields.find(f => f.localName === field || f.jsonName === field);
}

// The panel has no sim, so the site's pickers read a stand-in player. It answers the class, which picks the wand or
// melee auto attack lists, from the file's spec line; and it has one pet, one target and one other player in its raid,
// so a unit set lists each kind of unit it offers once.
let className: string | undefined;

export function setClassName(name: string | undefined): void {
	className = name;
}

const one = { asList: () => [{}] };
const player = {
	getClass: () =>
		className
			? ((Class as unknown as Record<string, number>)[`Class${className[0].toUpperCase()}${className.slice(1)}`] ?? Class.ClassUnknown)
			: Class.ClassUnknown,
	getPlayerSpec: () => ({ isHealingSpec: false }),
	getPetMetadatas: () => one,
	sim: { raid: { getActivePlayers: () => [{ getRaidIndex: () => 0 }] }, encounter: { targetsMetadata: one } },
} as unknown as Player<any>;

// The units a field of `set` takes, each as the JSON of its first one: a pet carries its owner, as the sim needs.
export function unitChoices(set: UNIT_SET): JsonObject[] {
	try {
		return unitSets[set]
			.getUnits(player)
			.filter((unit): unit is UnitReference => unit !== undefined)
			.map(unit => UnitReference.toJson(unit) as JsonObject);
	} catch {
		return [];
	}
}

export function resolve(descriptor: APLFieldDescriptor): AplFieldSpec {
	return resolveField(descriptor, player);
}

// Whether the site offers the kind in this list. A test that asks the sim something the panel cannot answer
// keeps the kind on offer.
function offered(model: KindModel, isPrepull: boolean, isGroup: boolean): boolean {
	try {
		return model.includeIf?.(player, isPrepull, isGroup) ?? true;
	} catch {
		return true;
	}
}

export function kindsFor(family: Family, isPrepull: boolean, isGroup: boolean): KindModel[] {
	return Object.values(KINDS[family]).filter(model => offered(model, isPrepull, isGroup));
}

export function submenuLabel(segment: string): string {
	const key = `rotation_tab.apl.submenus.${segment}`;
	const translated = i18n.t(key);
	return translated === key ? segment : translated;
}

// Proto JSON for a newly picked kind, with the defaults the site's pickers write and without the editor uuids
// the site keys its rows by, which the APL files do not carry.
export function newKindJson(family: Family, kind: string): JsonObject {
	const model = kindModel(family, kind)!;
	const message = kindMessage(family, kind);
	const impl = newKindImpl(model, player)();
	return withoutUuids(message ? (message.toJson(message.create(impl as object)) as Json) : {}) as JsonObject;
}

function withoutUuids(json: Json): Json {
	if (Array.isArray(json)) {
		return json.map(withoutUuids);
	}
	if (json && typeof json === 'object') {
		return Object.fromEntries(
			Object.entries(json)
				.filter(([key]) => key !== 'uuid')
				.map(([key, value]) => [key, withoutUuids(value)]),
		);
	}
	return json;
}

// The JSON name of a proto enum value: protobuf-ts strips a shared prefix from the TypeScript names and
// keeps it as the third element of the field's type tuple.
export function enumJsonName(info: FieldInfo | undefined, value: number): string | number {
	if (info?.kind !== 'enum') {
		return value;
	}
	const [, values, prefix] = info.T();
	const name = (values as Record<number, string>)[value];
	return name === undefined ? value : `${prefix ?? ''}${name}`;
}

export function enumNumber(info: FieldInfo | undefined, json: Json | undefined): number | undefined {
	if (typeof json === 'number') {
		return json;
	}
	if (info?.kind !== 'enum' || typeof json !== 'string') {
		return undefined;
	}
	const [, values, prefix] = info.T();
	const name = prefix && json.startsWith(prefix) ? json.slice(prefix.length) : json;
	const number = (values as Record<string, number>)[name];
	return typeof number === 'number' ? number : undefined;
}

// The enum values a proto field takes, by JSON name, for a field with no table of its own.
export function enumChoices(info: FieldInfo | undefined): Array<{ value: number; label: string }> {
	if (info?.kind !== 'enum') {
		return [];
	}
	const [, values, prefix] = info.T();
	return Object.entries(values as Record<string, number | string>)
		.filter(([, value]) => typeof value === 'number')
		.map(([name, value]) => ({ value: value as number, label: `${prefix ?? ''}${name}` }));
}

// "minIcdSeconds" reads as "Min icd seconds" where the site gives a field no label.
export function humanize(field: string): string {
	const words = field.replace(/([a-z0-9])([A-Z])/g, '$1 $2').toLowerCase();
	return words.charAt(0).toUpperCase() + words.slice(1);
}

export function isObject(json: Json | undefined): json is JsonObject {
	return !!json && typeof json === 'object' && !Array.isArray(json);
}
