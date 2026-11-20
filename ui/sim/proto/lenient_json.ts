import type { FieldInfo, IMessageType, JsonObject, JsonReadOptions, JsonValue } from '@protobuf-ts/runtime';
import { isJsonObject } from '@protobuf-ts/runtime';

// Reads saved or imported JSON, skipping the fields and enum names the protos do not list. With
// ignoreUnknownFields, protobuf-ts picks a oneof's case before it reads the value, so an enum name it
// doesn't know (an APL placeholder the protos do not list, in an ActionID, say) leaves the case set
// with an undefined value, which throws once the message is written to binary. Those oneofs are
// cleared here. A bool saved in a field the protos now declare as an enum reads as that enum's 0 or 1, so
// a Judgement of the Crusader saved as true loads as TristateEffectRegular.
export function fromJsonLenient<T extends object>(type: IMessageType<T>, json: JsonValue, options?: Partial<JsonReadOptions>): T {
	return clearUnreadOneofs(type, type.fromJson(enumsFromBools(type, json), { ...options, ignoreUnknownFields: true }));
}

export function fromJsonStringLenient<T extends object>(type: IMessageType<T>, json: string, options?: Partial<JsonReadOptions>): T {
	return fromJsonLenient(type, JSON.parse(json), options);
}

function clearUnreadOneofs<T extends object>(type: IMessageType<T>, message: T): T {
	const fields = message as Record<string, any>;
	for (const field of type.fields) {
		let value;
		if (field.oneof) {
			const group = fields[field.oneof];
			if (group?.oneofKind !== field.localName) continue;
			value = group[field.localName];
			if (value === undefined) {
				fields[field.oneof] = { oneofKind: undefined };
				continue;
			}
		} else {
			value = fields[field.localName];
		}

		if (field.kind === 'message') {
			const subType = field.T();
			if (Array.isArray(value)) value.forEach(item => clearUnreadOneofs(subType, item));
			else if (value) clearUnreadOneofs(subType, value);
		} else if (field.kind === 'map' && field.V.kind === 'message' && value) {
			const subType = field.V.T();
			Object.values(value).forEach(item => clearUnreadOneofs(subType, item as object));
		}
	}
	return message;
}

// A copy of the JSON with every bool in an enum field turned into its number; the parts holding none
// are the caller's own objects.
function enumsFromBools(type: IMessageType<object>, json: JsonValue): JsonValue {
	if (!isJsonObject(json)) return json;
	let copy: JsonObject | undefined;
	for (const field of type.fields) {
		for (const key of new Set([field.jsonName, field.name])) {
			const value = json[key];
			if (value === undefined) continue;
			const read = enumFieldsFromBools(field, value);
			if (read !== value) (copy ??= { ...json })[key] = read;
		}
	}
	return copy ?? json;
}

function enumFieldsFromBools(field: FieldInfo, value: JsonValue): JsonValue {
	switch (field.kind) {
		case 'enum':
			return typeof value === 'boolean' ? Number(value) : value;
		case 'message':
			return Array.isArray(value) ? mapIfChanged(value, item => enumsFromBools(field.T(), item)) : enumsFromBools(field.T(), value);
		case 'map': {
			const subType = field.V.kind === 'message' ? field.V.T() : undefined;
			if (!subType || !isJsonObject(value)) return value;
			let copy: JsonObject | undefined;
			for (const [key, item] of Object.entries(value)) {
				const read = enumsFromBools(subType, item);
				if (read !== item) (copy ??= { ...value })[key] = read;
			}
			return copy ?? value;
		}
		default:
			return value;
	}
}

function mapIfChanged(values: JsonValue[], read: (value: JsonValue) => JsonValue): JsonValue[] {
	const mapped = values.map(read);
	return mapped.every((value, i) => value === values[i]) ? values : mapped;
}
