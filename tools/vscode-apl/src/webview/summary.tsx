import type { AplFieldSpec } from '@features/apl/model/field_specs';
import i18n from '@i18n/config';
import { Fragment, type ReactNode } from 'react';

import { IdChip } from './db';
import {
	enumNumber,
	type Family,
	fieldInfo,
	humanize,
	isObject,
	type Json,
	type JsonObject,
	type JsonPath,
	kindMessage,
	kindModel,
	kindOf,
	resolve,
} from './model';
import { callAt, callText, useStructure } from './structure';

const COMPARE: Record<string, string> = { OpEq: '=', OpNe: '≠', OpGe: '≥', OpGt: '>', OpLe: '≤', OpLt: '<' };
const MATH: Record<string, string> = { OpAdd: '+', OpSub: '−', OpMul: '×', OpDiv: '÷' };

// The site's unit labels (ui/features/apl/model/unit_values.ts).
const unitLabel = (key: string) => i18n.t(`rotation_tab.apl.helpers.unit_labels.${key}`);

// The short name a unit reference reads as; the kind's default unit has no reference and reads as nothing.
export function unitText(json: Json | undefined): string | undefined {
	if (!isObject(json) || typeof json.type !== 'string') {
		return undefined;
	}
	const index = typeof json.index === 'number' ? json.index : 0;
	switch (json.type) {
		case 'Self':
			return unitLabel('self');
		case 'CurrentTarget':
			return unitLabel('current_target');
		case 'PreviousTarget':
			return unitLabel('previous_target');
		case 'NextTarget':
			return unitLabel('next_target');
		case 'Target':
			return `${unitLabel('target')} ${index + 1}`;
		case 'Pet':
			return `${unitLabel('pet')} ${index + 1}`;
		case 'Player':
			return `${unitLabel('player')} ${index + 1}`;
		default:
			return json.type;
	}
}

// How tightly each operator binds, so only a looser operand gets brackets; `and` inside `or` and the other
// way round keep theirs, since readers do not all know which binds first.
const PRECEDENCE: Record<string, number> = { or: 1, and: 2, not: 3, cmp: 4, OpAdd: 5, OpSub: 5, OpMul: 6, OpDiv: 6 };
const ATOM = 7;

function precedence(json: Json | undefined): number {
	const kind = kindOf('value', json);
	if (kind === 'math' && isObject(json) && isObject(json.math)) {
		return PRECEDENCE[String(json.math.op ?? 'OpAdd')] ?? ATOM;
	}
	return (kind ? PRECEDENCE[kind] : undefined) ?? ATOM;
}

function needsBrackets(child: number, parent: number): boolean {
	return child < parent || (parent > 0 && child !== parent && child <= 2 && parent <= 2);
}

const op = (text: string) => <span className="op">{text}</span>;

export function joined(items: ReactNode[], separator: ReactNode): ReactNode {
	return items.map((item, i) => (
		<Fragment key={i}>
			{i > 0 ? separator : null}
			{item}
		</Fragment>
	));
}

// A path below `path`, or none where the value has no path.
const below = (path: JsonPath | undefined, ...keys: JsonPath) => (path ? [...path, ...keys] : undefined);

// One argument of a kind, as it reads inline; undefined when the field is unset and has nothing to say. `path` is the
// kind's body's.
function argument(family: Family, kind: string, body: JsonObject, spec: AplFieldSpec, path?: JsonPath): ReactNode | undefined {
	const json = body[spec.field];
	const at = below(path, spec.field);
	switch (spec.kind) {
		case 'actionId':
			return <IdChip json={json} />;
		case 'unit':
			return unitText(json) ? <span className="unit">{unitText(json)}</span> : undefined;
		case 'value':
			return json === undefined ? undefined : <ValueText json={json} path={at} />;
		case 'valueList':
			return Array.isArray(json)
				? joined(
						json.map((v, i) => <ValueText json={v} path={below(at, i)} />),
						', ',
					)
				: undefined;
		case 'action':
			return json === undefined ? undefined : <ActionText json={json} path={at} />;
		case 'actionList':
			return Array.isArray(json)
				? joined(
						json.map((a, i) => <ActionText json={a} path={below(at, i)} />),
						op(' → '),
					)
				: undefined;
		case 'boolean':
			return json === true ? <span className="flag">{spec.label ?? humanize(spec.field)}</span> : undefined;
		case 'groupVariables':
			return undefined;
		case 'enum': {
			const number = enumNumber(fieldInfo(kindMessage(family, kind), spec.field), json) ?? 0;
			const label = spec.table.options.find(option => option.value === number)?.label;
			return json === undefined && !label ? undefined : <span className="lit">{label ?? String(json)}</span>;
		}
		default:
			return json === undefined || json === '' ? undefined : <span className="lit">{String(json)}</span>;
	}
}

function args(family: Family, kind: string, body: JsonObject, path?: JsonPath): ReactNode[] {
	const model = kindModel(family, kind);
	if (!model) {
		return [];
	}
	return model.fields
		.map(descriptor => argument(family, kind, body, resolve(descriptor), below(path, kind)))
		.filter((node): node is ReactNode => node !== undefined);
}

// A value as it reads inline. With its path, a value the text writes as a def call reads as the call.
export function ValueText({ json, parent = 0, path }: { json: Json | undefined; parent?: number; path?: JsonPath }) {
	const call = callAt(useStructure(), path);
	if (call) {
		return (
			<span className="call def-call" title={`def ${call.name}`}>
				{callText(call.name, call.args)}
			</span>
		);
	}
	const kind = kindOf('value', json);
	if (!kind || !isObject(json)) {
		return <span className="lit invalid">{json === undefined || (isObject(json) && Object.keys(json).length === 0) ? 'none' : JSON.stringify(json)}</span>;
	}
	const body = isObject(json[kind]) ? (json[kind] as JsonObject) : {};
	const own = precedence(json);
	const wrap = (node: ReactNode) => (needsBrackets(own, parent) ? <span className="paren">({node})</span> : node);
	// The right operand of a non-associative operator binds one step tighter: a − (b − c) keeps its brackets.
	const right = own + 0.5;
	switch (kind) {
		case 'const':
			return <span className="lit">{String(body.val ?? '')}</span>;
		case 'variableRef':
			return <span className="var">${String(body.name ?? '')}</span>;
		case 'variablePlaceholder':
			return <span className="var">?{String(body.name ?? '')}</span>;
		case 'cmp':
			return wrap(
				<>
					<ValueText json={body.lhs} parent={right} path={below(path, kind, 'lhs')} /> {op(COMPARE[String(body.op ?? 'OpEq')] ?? String(body.op))}{' '}
					<ValueText json={body.rhs} parent={right} path={below(path, kind, 'rhs')} />
				</>,
			);
		case 'math':
			return wrap(
				<>
					<ValueText json={body.lhs} parent={own} path={below(path, kind, 'lhs')} /> {op(MATH[String(body.op ?? 'OpAdd')] ?? String(body.op))}{' '}
					<ValueText json={body.rhs} parent={right} path={below(path, kind, 'rhs')} />
				</>,
			);
		case 'and':
		case 'or': {
			const vals = Array.isArray(body.vals) ? body.vals : [];
			return wrap(
				<>
					{joined(
						vals.map((v, i) => <ValueText json={v} parent={own} path={below(path, kind, 'vals', i)} />),
						<> {op(kind)} </>,
					)}
				</>,
			);
		}
		case 'not':
			return wrap(
				<>
					{op('not ')}
					<ValueText json={body.val} parent={ATOM} path={below(path, kind, 'val')} />
				</>,
			);
	}
	const label = kindModel('value', kind)?.label ?? kind;
	const shown = args('value', kind, body, path);
	return (
		<span className="call">
			<span className="fn">{label}</span>
			{shown.length > 0 ? <>({joined(shown, ', ')})</> : null}
		</span>
	);
}

export function ActionText({ json, path }: { json: Json | undefined; path?: JsonPath }) {
	const kind = kindOf('action', json);
	if (!kind || !isObject(json)) {
		return (
			<span className="lit invalid">
				{json === undefined || (isObject(json) && Object.keys(json).length === 0) ? 'no action set' : JSON.stringify(json)}
			</span>
		);
	}
	const body = isObject(json[kind]) ? (json[kind] as JsonObject) : {};
	const label = kindModel('action', kind)?.label ?? kind;
	const shown = args('action', kind, body, path);
	return (
		<span className="action">
			<span className="fn">{label}</span>
			{shown.length > 0 ? <> {joined(shown, ' ')}</> : null}
		</span>
	);
}
