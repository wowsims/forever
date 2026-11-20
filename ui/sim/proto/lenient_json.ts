import type { FieldInfo, IMessageType, JsonObject, JsonReadOptions, JsonValue } from '@protobuf-ts/runtime';
import { isJsonObject } from '@protobuf-ts/runtime';

// Reads saved or imported JSON, skipping the fields and enum names the protos do not list. With
// ignoreUnknownFields, protobuf-ts picks a oneof's case before it reads the value, so an enum name it
// doesn't know (an APL placeholder the protos do not list, in an ActionID, say) leaves the case set
// with an undefined value, which throws once the message is written to binary. Those oneofs are
// cleared here. A save can hold a bool in a TristateEffect field, which protobuf-ts refuses; it reads as
// Missing or Regular, so a Judgement of the Crusader saved as true loads as TristateEffectRegular.
export function fromJsonLenient<T extends object>(type: IMessageType<T>, json: JsonValue, options?: Partial<JsonReadOptions>): T {
	return clearUnreadOneofs(type, type.fromJson(tristatesFromBools(type, json), { ...options, ignoreUnknownFields: true }));
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

// A copy of the JSON with every bool in a TristateEffect field turned into its number: false reads as
// Missing (0) and true as Regular (1).
function tristatesFromBools(type: IMessageType<object>, json: JsonValue): JsonValue {
	if (!isJsonObject(json)) return json;
	const copy: JsonObject = { ...json };
	for (const field of type.fields) {
		for (const key of field.jsonName === field.name ? [field.name] : [field.jsonName, field.name]) {
			if (copy[key] !== undefined) copy[key] = tristateFieldFromBools(field, copy[key]);
		}
	}
	return copy;
}

function tristateFieldFromBools(field: FieldInfo, value: JsonValue): JsonValue {
	switch (field.kind) {
		case 'enum':
			return typeof value === 'boolean' && field.T()[0] === 'proto.TristateEffect' ? Number(value) : value;
		case 'message':
			return Array.isArray(value) ? value.map(item => tristatesFromBools(field.T(), item)) : tristatesFromBools(field.T(), value);
		case 'map': {
			const valueType = field.V.kind === 'message' ? field.V.T() : undefined;
			if (!valueType || !isJsonObject(value)) return value;
			return Object.fromEntries(Object.entries(value).map(([key, item]) => [key, tristatesFromBools(valueType, item)]));
		}
		default:
			return value;
	}
}
