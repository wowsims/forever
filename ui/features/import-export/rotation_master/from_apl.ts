// A WowSims APL (Action Priority List, proto-JSON) as a Rotation Master rotation. Pure: no DOM, no network; everything
// it needs comes in as arguments. It is the sim-side port of the website's importer (rotationmaster-web
// wowsims/import.mjs) and interprets the same rule table (mapping.json), so the site and the simulators agree on what
// an APL becomes.
//
// The input is the APL as proto-JSON (what `APLRotation.toJson` gives, and what the preset files hold). The output is
// a list of Rotation Master steps and a report that says, item by item, what became of every part of the original.
// Nothing is dropped silently: a part with no equivalent is reported, and a step that is not an exact copy is
// imported DISABLED, so nothing runs that the original did not do.
import rawMapping from './mapping.json';
import { type ClassKey, type Edition, FIDELITIES, type Fidelity, type Reason, worse } from './report';
import rawSupport from './rm_support.json';

type Json = any;
const mapping = rawMapping as Json;
const support = rawSupport as Json;
const own = (o: unknown, k: string): boolean => o !== null && typeof o === 'object' && Object.prototype.hasOwnProperty.call(o, k);
const isObj = (v: unknown): v is Record<string, Json> => v !== null && typeof v === 'object' && !Array.isArray(v);
const asList = (v: unknown): Json[] => (Array.isArray(v) ? v : []);

/** A Rotation Master condition tree as plain data. */
export type RmCondition = { type: string; [key: string]: Json };
/** A Rotation Master step as plain data. */
export interface RmStep {
	id: string;
	type: 'spell' | 'item';
	action: number | number[];
	ranked?: boolean;
	conditions?: RmCondition;
	disabled?: boolean;
}

export interface FromAplOptions {
	edition: Edition;
	cls: ClassKey;
	/** Import approximate steps enabled instead of disabled (default false). */
	enableApproximate?: boolean;
	/** Keep an APL item that is hidden in the simulator as a disabled step instead of leaving it out (default true). */
	keepHidden?: boolean;
	/** Item ids of this edition's mana potions: the simulator's resource potion becomes an item step with them. */
	potions?: ReadonlyArray<number>;
	/** Step id generator (default ws-1, ws-2, ...). */
	newId?: () => string;
	/** Whether Rotation Master can run a condition on this edition of the game. Default: the generated table (rm_support.json). */
	supports?: (tree: RmCondition) => { ok: true } | { ok: false; reason?: string; code?: string };
}

export interface FromAplItem {
	path: string;
	original: string;
	fidelity: Fidelity;
	outcome: 'enabled' | 'disabled' | 'skipped';
	reasons: Reason[];
	became: Array<{ type: string; action: number | number[]; disabled: boolean; conditions: string }>;
	prepull?: boolean;
	hidden?: boolean;
}

export interface FromAplSummary {
	items: number;
	exact: number;
	approximate: number;
	lossy: number;
	unsupported: number;
	enabled: number;
	disabled: number;
	skipped: number;
	prepull: number;
	steps: number;
	truncated: number;
}

export interface FromAplResult {
	steps: RmStep[];
	items: FromAplItem[];
	summary: FromAplSummary;
}

// ---- conditions ----------------------------------------------------------------------------------------------------
class Budget extends Error {}
type CondNode =
	| { k: 'true' | 'false' | 'drop' }
	| { k: 'cond'; tree: Json; dropped?: boolean }
	| { k: 'num'; tree: Json; unit: string; kind: string }
	| { k: 'const'; parsed?: { num: number; suffix: string }; bool?: boolean; truthy: boolean };
const TRUE: CondNode = { k: 'true' };
const FALSE: CondNode = { k: 'false' };
const dropNode = (): CondNode => ({ k: 'drop' });

function parseConst(text: unknown): { num: number; suffix: string; bool?: undefined } | { bool: boolean; num?: undefined; suffix?: undefined } | null {
	if (typeof text === 'number') return Number.isFinite(text) ? { num: text, suffix: '' } : null;
	if (typeof text === 'boolean') return { bool: text };
	if (typeof text !== 'string') return null;
	const s = text.trim().toLowerCase();
	if (s === 'true') return { bool: true };
	if (s === 'false') return { bool: false };
	const m = /^(-?\d+(?:\.\d+)?|-?\.\d+)(ms|s|m|%)?$/.exec(s);
	if (!m) return null;
	const num = Number(m[1]);
	return Number.isFinite(num) ? { num, suffix: m[2] || '' } : null;
}
const UNIT_BOUNDS: Record<string, [number, number]> = { seconds: [0, 3600], fraction: [0, 1], count: [0, 9999], raw: [-1e9, 1e9] };
// A constant in the unit a numeric value is measured in; null when it cannot be that, undefined when it is outside what RM stores.
function constIn(parsed: { num?: number; suffix?: string; bool?: boolean } | undefined, unit: string): number | null | undefined {
	if (!parsed || parsed.bool !== undefined) return null;
	let v = parsed.num as number;
	if (unit === 'seconds') {
		if (parsed.suffix === '%') return null;
		if (parsed.suffix === 'ms') v /= 1000;
		else if (parsed.suffix === 'm') v *= 60;
	} else if (unit === 'fraction') {
		if (parsed.suffix === '%') v /= 100;
		else if (parsed.suffix !== '' || v > 1) return null;
	} else if (parsed.suffix !== '') return null;
	const [lo, hi] = UNIT_BOUNDS[unit] || UNIT_BOUNDS.raw;
	return v >= lo && v <= hi ? v : undefined;
}
const fold = (op: string, a: number, b: number): boolean | null =>
	op === 'OpEq' ? a === b : op === 'OpNe' ? a !== b : op === 'OpLt' ? a < b : op === 'OpLe' ? a <= b : op === 'OpGt' ? a > b : op === 'OpGe' ? a >= b : null;
const compact = (v: unknown, limit: number): string => {
	let s = '';
	try {
		s = JSON.stringify(v) ?? '';
	} catch {
		s = '';
	}
	return s.length > limit ? s.slice(0, limit - 1) + '...' : s;
};

interface Scope {
	refVars: Map<string, { value: Json; scope: Scope }> | null;
	groupVars: Map<string, Json> | null;
}
interface ItemState {
	reasons: Reason[];
	nodes: number;
	varStack: string[];
}
interface Ctx {
	edition: Edition;
	cls: ClassKey;
	ranks: boolean;
	supports?: FromAplOptions['supports'];
	/** The id of the item a condition is about while its template is built (the `@item` directive). */
	itemId: number | null;
	rootVars: Map<string, Json>;
	rootScope: Scope;
	item: ItemState;
	totalNodes: number;
	entries: number;
	stop?: boolean;
	truncated?: number;
}

/** The generated table of the conditions Rotation Master refuses on an edition of the game, as the `supports` check. */
export function editionSupports(edition: Edition, cls: ClassKey): NonNullable<FromAplOptions['supports']> {
	const table = support.editions[edition];
	return tree => {
		const key: string = tree.type;
		const why = table?.all?.[key] ?? table?.byClass?.[cls]?.[key];
		return why ? { ok: false, reason: why, code: why } : { ok: true };
	};
}

function makeConverter(ctx: Ctx) {
	const lim = mapping.limits;
	const note = (code: string, level: Reason['level'], params: Reason['params'] = {}): void => {
		const r = ctx.item.reasons;
		if (r.length < 50) r.push({ code, level, params });
	};
	const tick = (): void => {
		if (++ctx.item.nodes > lim.expansion) throw new Budget();
		if (++ctx.totalNodes > lim.total) {
			ctx.stop = true;
			ctx.item.reasons.push({ code: 'too-complex', level: 'unsupported', params: {} });
			throw new Budget();
		}
	};
	const kindOf = (v: unknown): string | null => {
		if (!isObj(v)) return null;
		for (const k of Object.keys(v)) if (k !== 'uuid') return k;
		return null;
	};
	const baseKind = (k: string): string => (k.endsWith('WithReactionTime') ? k.slice(0, -16) : k);

	// An ActionID's spell id: {id} or {fail}.
	const spellOf = (arg: Json): { id: number } | { fail: string; other?: string } => {
		if (!isObj(arg)) return { fail: 'no-id' };
		if (typeof arg.spellId === 'number' && Number.isInteger(arg.spellId) && arg.spellId > 0 && arg.spellId <= mapping.validation.idMax)
			return { id: arg.spellId };
		if (own(arg, 'itemId')) return { fail: 'item-id' };
		if (own(arg, 'otherId')) return { fail: 'other-id', other: String(arg.otherId) };
		return { fail: 'no-id' };
	};
	const unitName = (ref: Json, dflt: string): { unit: string } | { fail: string } => {
		if (ref === undefined || ref === null) return { unit: mapping.units[dflt] };
		if (!isObj(ref)) return { fail: 'bad' };
		let type: Json = ref.type;
		if (typeof type === 'number') type = mapping.unitTypes[type];
		if (type === undefined || type === 'Unknown') return { unit: mapping.units[dflt] };
		if (typeof type !== 'string' || !own(mapping.units, type)) return { fail: type === undefined ? 'unknown' : String(type) };
		const entry = mapping.units[type];
		if (typeof entry === 'string') return { unit: entry };
		const index = typeof ref.index === 'number' ? ref.index : 0;
		return own(entry, String(index)) ? { unit: entry[String(index)] } : { fail: `${type} ${index}` };
	};

	// Resolves a template's directives against a value's arguments: {tree} or {fail}.
	function build(template: Json, body: Json, unitInfo: { unit: string } | null): { tree: Json } | { fail: { code: string; params: Reason['params'] } } {
		const out: Json = {};
		for (const [key, val] of Object.entries(template)) {
			if (val !== null && typeof val === 'object' && !Array.isArray(val)) {
				const inner = build(val, body, unitInfo);
				if ('fail' in inner) return inner;
				out[key] = inner.tree;
				continue;
			}
			if (typeof val !== 'string' || val[0] !== '@') {
				out[key] = val;
				continue;
			}
			const parts = val.split(':');
			if (parts[0] === '@item') {
				out[key] = [ctx.itemId];
				continue;
			}
			if (parts[0] === '@ranks') {
				if (ctx.ranks) out[key] = true;
				continue;
			}
			if (parts[0] === '@spell') {
				const r = spellOf(body?.[parts[1]]);
				if ('fail' in r) return { fail: { code: 'cond-id', params: { why: r.fail } } };
				out[key] = r.id;
				continue;
			}
			if (parts[0] === '@unit') {
				if (parts.length === 1 && unitInfo) {
					out[key] = unitInfo.unit;
					continue;
				}
				const r = unitName(body?.[parts[1]], parts[2] || 'Self');
				if ('fail' in r) return { fail: { code: 'cond-unit', params: { unit: r.fail } } };
				out[key] = r.unit;
				continue;
			}
			if (parts[0] === '@table') {
				const table = mapping.tables[parts[1]];
				const raw = body?.[parts[2]];
				const order = mapping.tableOrder?.[parts[1]];
				const name = typeof raw === 'number' ? (Array.isArray(order) ? order[raw] : undefined) : raw;
				if (!table || typeof name !== 'string' || !own(table, name)) return { fail: { code: 'cond-value', params: { what: parts[2] } } };
				out[key] = table[name];
				continue;
			}
			return { fail: { code: 'cond-value', params: { what: val } } };
		}
		return { tree: out };
	}

	// ---- the value converter ----
	function lookupVariable(name: string, scope: Scope): { value: Json; scope: Scope } | null {
		const ref = scope.refVars?.get(name);
		if (ref) return ref;
		if (scope.groupVars && scope.groupVars.has(name)) return { value: scope.groupVars.get(name), scope };
		if (ctx.rootVars.has(name)) return { value: ctx.rootVars.get(name), scope: ctx.rootScope };
		return null;
	}
	function convert(v: Json, scope: Scope, depth: number): CondNode {
		tick();
		if (depth > lim.depth) throw new Budget();
		let kind = isObj(v) ? kindOf(v) : null;
		if (!isObj(v) || kind === null) {
			note('cond-empty', 'lossy');
			return dropNode();
		}
		const body = v[kind];
		kind = baseKind(kind);
		switch (kind) {
			case 'const': {
				const c = parseConst(body?.val);
				if (!c) return dropNode();
				if (c.bool !== undefined) return { k: 'const', bool: c.bool, truthy: c.bool };
				return { k: 'const', parsed: c, truthy: c.num !== 0 };
			}
			case 'and':
			case 'or': {
				const vals = asList(body?.vals).map(x => convert(x, scope, depth + 1));
				return combine(kind, vals);
			}
			case 'not': {
				const inner = asBool(convert(body?.val, scope, depth + 1));
				if (inner.k === 'true') return FALSE;
				if (inner.k === 'false') return TRUE;
				if (inner.k === 'drop') return inner;
				return { k: 'cond', tree: { type: 'NOT', value: (inner as { tree: Json }).tree } };
			}
			case 'cmp':
				return compare(body || {}, scope, depth);
			case 'variableRef':
			case 'variablePlaceholder': {
				const name = typeof body?.name === 'string' ? body.name : null;
				let found: { value: Json; scope: Scope } | null = null;
				if (name !== null) {
					const ref = kind === 'variablePlaceholder' ? scope.refVars?.get(name) : lookupVariable(name, scope);
					found = ref ?? null;
				}
				if (!found || name === null) {
					note('cond-variable', 'lossy', { name: String(name).slice(0, 60) });
					return dropNode();
				}
				if (ctx.item.varStack.length >= lim.variableDepth || ctx.item.varStack.includes(name + '\0' + kind)) {
					note('cond-variable', 'lossy', { name: String(name).slice(0, 60) });
					return dropNode();
				}
				ctx.item.varStack.push(name + '\0' + kind);
				try {
					return convert(found.value, found.scope || scope, depth + 1);
				} finally {
					ctx.item.varStack.pop();
				}
			}
			default:
				break;
		}
		if (own(mapping.values, kind)) return leaf(kind, mapping.values[kind], body || {});
		const why = own(mapping.unsupportedValues, kind) ? mapping.unsupportedValues[kind] : 'unknown';
		note('cond-unsupported', 'lossy', { kind, why });
		return dropNode();
	}
	function leaf(kind: string, entry: Json, body: Json): CondNode {
		let template = entry.rm;
		let unitInfo: { unit: string } | null = null;
		if (entry.constTrue) {
			// Something Rotation Master decides for itself (it advises when you can act): always true.
			if (entry.fidelity === 'approximate') note('cond-approx', 'approximate', { kind, note: entry.note || kind });
			return TRUE;
		}
		// A condition about an item (a trinket's cooldown) instead of a spell.
		if (entry.itemRm) {
			const arg = body[entry.itemArg || 'spellId'];
			if (isObj(arg) && Number.isInteger(arg.itemId) && arg.itemId > 0 && arg.itemId <= mapping.validation.idMax) {
				template = entry.itemRm;
				ctx.itemId = arg.itemId;
			}
		}
		if (entry.byUnit) {
			const u = unitName(body[entry.byUnit.arg], entry.byUnit.default);
			if ('fail' in u) {
				note('cond-unit', 'lossy', { unit: u.fail });
				return dropNode();
			}
			unitInfo = u;
			template = u.unit === 'player' || u.unit === 'pet' ? entry.byUnit.friendly : entry.byUnit.hostile;
		}
		if (entry.self) {
			const u = unitName(body[entry.self], 'Self');
			if ('fail' in u || u.unit !== 'player') {
				note('cond-unit', 'lossy', { unit: 'fail' in u ? u.fail : u.unit });
				return dropNode();
			}
		}
		const built = build(template, body, unitInfo);
		ctx.itemId = null;
		if ('fail' in built) {
			note(built.fail.code, 'lossy', built.fail.params);
			return dropNode();
		}
		// A number's condition is checked with a placeholder comparison (the real operator and value come with the cmp).
		const sup = ctx.supports ? ctx.supports(entry.t === 'num' ? { ...built.tree, operator: 'GREATERTHAN', value: 0 } : built.tree) : null;
		if (sup && sup.ok === false) {
			// An edition that hides a value from addons (mana, health, the combat log on Forever) is a different thing from a
			// condition that simply is not offered there; the report says which.
			note('cond-unavailable', 'lossy', { type: built.tree.type, reason: sup.reason || sup.code || 'unsupported' });
			return dropNode();
		}
		if (entry.fidelity === 'approximate') note('cond-approx', 'approximate', { kind, note: entry.note || kind });
		if (entry.t === 'num') return { k: 'num', tree: built.tree, unit: entry.unit, kind };
		return entry.not ? { k: 'cond', tree: { type: 'NOT', value: built.tree } } : { k: 'cond', tree: built.tree };
	}
	const asBool = (node: CondNode): CondNode => {
		if (node.k === 'const') return node.truthy ? TRUE : FALSE;
		if (node.k === 'num') {
			// A number used as a condition means "is not zero".
			return { k: 'cond', tree: { ...node.tree, operator: 'NOTEQUALS', value: 0 } };
		}
		return node;
	};
	function combine(kind: string, vals: CondNode[]): CondNode {
		const nodes = vals.map(asBool);
		const isAnd = kind === 'and';
		const children: Json[] = [];
		let dropped = false;
		for (const n of nodes) {
			if (n.k === 'drop') {
				dropped = true;
				continue;
			}
			if (n.k === 'true') {
				if (!isAnd) return TRUE;
				continue;
			}
			if (n.k === 'false') {
				if (isAnd) return FALSE;
				continue;
			}
			const t = (n as { tree: Json }).tree;
			if (t.type === (isAnd ? 'AND' : 'OR')) children.push(...t.value);
			else children.push(t);
		}
		if (!children.length) return dropped ? dropNode() : isAnd ? TRUE : FALSE;
		if (children.length === 1) return { k: 'cond', tree: children[0] };
		return { k: 'cond', tree: { type: isAnd ? 'AND' : 'OR', value: children } };
	}
	// "Power against its maximum": rage >= max rage - 20, max mana - mana > 500, mana + 300 < max mana. Rotation Master
	// compares a negative number with how much the unit is MISSING from its maximum, so these are one condition.
	const SWAP: Record<string, string> = { OpLt: 'OpGt', OpLe: 'OpGe', OpGt: 'OpLt', OpGe: 'OpLe', OpEq: 'OpEq', OpNe: 'OpNe' };
	function foldMissing(op: string, lhs: Json, rhs: Json, scope: Scope, depth: number): CondNode | null {
		const maxOf: Record<string, string> | undefined = mapping.maxKinds;
		if (!maxOf || !isObj(lhs) || !isObj(rhs)) return null;
		const plain = (n: Json): number | null => {
			const c = kindOf(n) === 'const' ? parseConst(n.const?.val) : null;
			return c && c.bool === undefined && c.suffix === '' ? c.num : null;
		};
		const math = (n: Json): Json | null => (kindOf(n) === 'math' && isObj(n.math) ? n.math : null);
		const emptyBody = (n: Json, k: string): boolean => !isObj(n[k]) || Object.keys(n[k]).length === 0;
		const resource = (n: Json): string | null => {
			const k = kindOf(n);
			return k && Object.values(maxOf).includes(k) && emptyBody(n, k) ? k : null;
		};
		const maxKind = (n: Json): string | null => {
			const k = kindOf(n);
			return k && own(maxOf, k) && emptyBody(n, k) ? maxOf[k] : null;
		};
		let target: Json = null;
		let operator: string | null = null;
		let missing: number | null = null;
		const ml = math(lhs);
		const mr = math(rhs);
		if (resource(lhs) && mr && mr.op === 'OpSub' && maxKind(mr.lhs) === resource(lhs) && plain(mr.rhs) !== null) {
			// R op (max - C)
			target = lhs;
			operator = SWAP[op];
			missing = plain(mr.rhs);
		} else if (ml && ml.op === 'OpSub' && maxKind(ml.lhs) && resource(ml.rhs) === maxKind(ml.lhs) && plain(rhs) !== null) {
			// (max - R) op C
			target = ml.rhs;
			operator = op;
			missing = plain(rhs);
		} else if (ml && ml.op === 'OpAdd' && resource(ml.lhs) && maxKind(rhs) === resource(ml.lhs) && plain(ml.rhs) !== null) {
			// (R + C) op max
			target = ml.lhs;
			operator = SWAP[op];
			missing = plain(ml.rhs);
		}
		if (!target || !operator || !own(mapping.opNames, operator) || missing === null || !(missing >= 0)) return null;
		const r = convert(target, scope, depth + 1);
		if (r.k !== 'num' || r.unit !== 'raw') return null;
		note('cond-missing', 'exact');
		return { k: 'cond', tree: { ...r.tree, operator: mapping.opNames[operator], value: -missing } };
	}
	function compare(body: Json, scope: Scope, depth: number): CondNode {
		const op: Json = typeof body.op === 'number' ? mapping.ops[body.op] : body.op;
		if (typeof op !== 'string' || !own(mapping.opNames, op)) {
			note('cond-cmp', 'lossy', { why: 'operator' });
			return dropNode();
		}
		const folded = foldMissing(op, body.lhs, body.rhs, scope, depth);
		if (folded) return folded;
		const l = convert(body.lhs, scope, depth + 1);
		const r = convert(body.rhs, scope, depth + 1);
		if (l.k === 'drop' || r.k === 'drop') return dropNode();
		if (l.k === 'const' && r.k === 'const') {
			const a = l.parsed ? l.parsed.num : l.bool ? 1 : 0;
			const b = r.parsed ? r.parsed.num : r.bool ? 1 : 0;
			const res = fold(op, a, b);
			return res === null ? dropNode() : res ? TRUE : FALSE;
		}
		const numeric = l.k === 'num' ? l : r.k === 'num' ? r : null;
		if (numeric && l.k === 'num' && r.k === 'num') {
			note('cond-cmp', 'lossy', { why: 'nonconst', left: l.kind, right: r.kind });
			return dropNode();
		}
		if (numeric) {
			const other = numeric === l ? r : l;
			if (other.k !== 'const' || !other.parsed) {
				note('cond-cmp', 'lossy', { why: 'operand' });
				return dropNode();
			}
			const value = constIn(other.parsed, numeric.unit);
			if (value === null) {
				note('cond-cmp', 'lossy', { why: 'unit', kind: numeric.kind });
				return dropNode();
			}
			if (value === undefined) {
				note('cond-cmp', 'lossy', { why: 'range', kind: numeric.kind });
				return dropNode();
			}
			let operator: string = mapping.opNames[op];
			if (numeric === r) operator = mapping.opFlip[operator];
			return { k: 'cond', tree: { ...numeric.tree, operator, value } };
		}
		// A condition compared with true/false (or 1/0).
		const boolSide = l.k === 'const' ? r : l;
		const constSide = l.k === 'const' ? l : r;
		if (constSide.k === 'const' && boolSide.k !== 'const' && (op === 'OpEq' || op === 'OpNe')) {
			const want = constSide.parsed ? constSide.parsed.num !== 0 : !!constSide.bool;
			const same = op === 'OpEq' ? want : !want;
			const b = asBool(boolSide);
			if (b.k === 'drop') return b;
			if (b.k === 'true' || b.k === 'false') return same === (b.k === 'true') ? TRUE : FALSE;
			return same ? b : { k: 'cond', tree: { type: 'NOT', value: (b as { tree: Json }).tree } };
		}
		note('cond-cmp', 'lossy', { why: 'operand' });
		return dropNode();
	}

	// ---- actions ----
	function conditionOf(conds: Array<{ value?: Json; scope: Scope; tree?: Json }>): CondNode {
		// conds: AND-ed; returns true | false | cond | drop.
		const nodes: CondNode[] = [];
		for (const c of conds) {
			if (c.tree)
				nodes.push({ k: 'cond', tree: structuredClone(c.tree) }); // an RM condition the converter adds itself (a pre-pull action runs out of combat)
			else if (c.value !== undefined && c.value !== null) nodes.push(convert(c.value, c.scope, 0));
		}
		return nodes.length ? combine('and', nodes) : TRUE;
	}

	return { convert, conditionOf, spellOf, unitName, build, note, kindOf };
}

const dedupe = (list: Reason[]): Reason[] => {
	const seen = new Set<string>();
	const out: Reason[] = [];
	for (const r of list) {
		const key = r.code + JSON.stringify(r.params);
		if (!seen.has(key)) {
			seen.add(key);
			out.push(r);
		}
	}
	return out;
};

// ---- compact text of an RM condition (for the report) --------------------------------------------------------------
const SYMBOL: Record<string, string> = { EQUALS: '=', NOTEQUALS: '!=', LESSTHAN: '<', LESSTHANOREQUALS: '<=', GREATERTHAN: '>', GREATERTHANOREQUALS: '>=' };
export function conditionText(tree: Json, depth = 0): string {
	if (!tree || depth > 40) return '';
	if (tree.type === 'AND' || tree.type === 'OR') return `${tree.type}(${tree.value.map((c: Json) => conditionText(c, depth + 1)).join(', ')})`;
	if (tree.type === 'NOT') return `NOT(${conditionText(tree.value, depth + 1)})`;
	const parts: string[] = [tree.type];
	if (tree.unit) parts.push(tree.unit);
	if (typeof tree.spell === 'number') parts.push('#' + tree.spell);
	if (tree.operator) parts.push((SYMBOL[tree.operator] || tree.operator) + ' ' + tree.value);
	else if (tree.value !== undefined) parts.push(String(tree.value));
	return parts.join(' ');
}

// A pre-pull action applies only before the fight: RM's "in combat" condition, negated.
const OUT_OF_COMBAT = Object.freeze({ type: 'NOT', value: Object.freeze({ type: 'COMBAT', unit: 'player' }) });

interface Entry {
	item: Json;
	path: string;
	scope: Scope;
	inherited: Array<{ value?: Json; scope: Scope; tree?: Json }>;
	depth: number;
	stack: string[];
	prepull?: boolean;
	reasons?: Reason[];
}

/** Converts an APL (proto-JSON) for an edition and class. */
export function aplToRotation(apl: Json, options: FromAplOptions): FromAplResult {
	const { edition, cls } = options;
	const lim = mapping.limits;
	const keepHidden = options.keepHidden !== false;
	const ranks = mapping.editions[edition].ranks === true;
	const rootVars = new Map<string, Json>();
	for (const v of asList(apl?.valueVariables)) if (isObj(v) && typeof v.name === 'string' && !rootVars.has(v.name)) rootVars.set(v.name, v.value);
	const groups = new Map<string, Json>();
	for (const g of asList(apl?.groups)) if (isObj(g) && typeof g.name === 'string' && !groups.has(g.name)) groups.set(g.name, g);
	const rootScope: Scope = { refVars: null, groupVars: null };
	const ctx: Ctx = {
		edition,
		cls,
		ranks,
		supports: options.supports ?? editionSupports(edition, cls),
		itemId: null,
		rootVars,
		rootScope,
		item: { reasons: [], nodes: 0, varStack: [] },
		totalNodes: 0,
		entries: 0,
	};
	const conv = makeConverter(ctx);
	let counter = 0;
	const newId = options.newId || (() => `ws-${++counter}`);
	const steps: RmStep[] = [];
	const items: FromAplItem[] = [];
	const addItem = (item: FromAplItem): void => {
		if (items.length < lim.reports) items.push(item);
		else ctx.truncated = (ctx.truncated || 0) + 1;
	};
	// The original of an item, cut short; computed once per item however often a group puts it in the list.
	const snippets = new WeakMap<object, string>();
	const snippet = (v: Json): string => {
		if (v === null || typeof v !== 'object') return compact(v, lim.snippet);
		let s = snippets.get(v);
		if (s === undefined) {
			s = compact(v, lim.snippet);
			snippets.set(v, s);
		}
		return s;
	};

	interface Made {
		type: 'spell' | 'item';
		action: number | number[];
		extra?: Json;
		condNode: CondNode;
		tree?: Json | null;
	}

	// Turns one APL list entry into steps and a report item.
	function doEntry(entry: Entry): void {
		const { item, path, scope, inherited, depth } = entry;
		// Work is bounded however groups and references multiply the entries.
		if (++ctx.entries > lim.entries) {
			if (!ctx.stop) {
				ctx.stop = true;
				items.push({
					path,
					original: '',
					fidelity: 'unsupported',
					outcome: 'skipped',
					reasons: [{ code: 'item-limit', level: 'unsupported', params: { limit: lim.entries } }],
					became: [],
				});
			}
			return;
		}
		if (ctx.stop) return;
		const record: FromAplItem = { path, original: snippet(item), fidelity: 'exact', outcome: 'enabled', reasons: [], became: [] };
		if (entry.prepull) record.prepull = true;
		if (!isObj(item)) {
			record.fidelity = 'unsupported';
			record.outcome = 'skipped';
			record.reasons.push({ code: 'item-bad', level: 'unsupported', params: {} });
			addItem(record);
			return;
		}
		const hidden = item.hide === true;
		if (hidden) record.hidden = true;
		if (hidden && !keepHidden) {
			record.outcome = 'skipped';
			record.reasons.push({ code: 'hidden', level: 'exact', params: {} });
			addItem(record);
			return;
		}
		const action: Json = isObj(item.action) ? item.action : {};
		ctx.item = { reasons: entry.reasons ? entry.reasons.map(r => ({ ...r, params: { ...r.params } })) : [], nodes: 0, varStack: [] };
		if (hidden) ctx.item.reasons.push({ code: 'hidden', level: 'exact', params: {} });
		let created: Made[] = [];
		let handler: Json = null;
		let kind: string | null = null;
		try {
			for (const k of Object.keys(action))
				if (k !== 'condition' && k !== 'uuid') {
					kind = k;
					break;
				}
			if (kind === null) conv.note('action-empty', 'unsupported');
			else if (own(mapping.actions, kind)) handler = mapping.actions[kind];
			else if (own(mapping.unsupportedActions, kind)) conv.note('action-unsupported', 'unsupported', { kind, why: mapping.unsupportedActions[kind] });
			else conv.note('action-unsupported', 'unsupported', { kind, why: 'unknown' });
			if (kind !== null && handler && handler.handler === 'group') {
				const ref = action[kind];
				const name: string | null = typeof ref?.groupName === 'string' ? ref.groupName : null;
				const group = name !== null && groups.has(name) ? groups.get(name) : null;
				if (!group || name === null) conv.note('group-unknown', 'unsupported', { name: String(name).slice(0, 80) });
				else if (depth >= lim.groupDepth || (entry.stack || []).includes(name)) conv.note('group-cycle', 'unsupported', { name: name.slice(0, 80) });
				else {
					const refVars = new Map<string, { value: Json; scope: Scope }>();
					for (const v of asList(ref.variables)) if (isObj(v) && typeof v.name === 'string') refVars.set(v.name, { value: v.value, scope });
					const groupVars = new Map<string, Json>();
					for (const v of asList(group.variables))
						if (isObj(v) && typeof v.name === 'string' && !groupVars.has(v.name)) groupVars.set(v.name, v.value);
					const inner: Scope = { refVars, groupVars };
					const inheritedNext = [...inherited, { value: action.condition, scope }];
					asList(group.actions).forEach((child, i) => {
						if (ctx.stop) return;
						doEntry({
							item: hidden && isObj(child) ? { ...child, hide: true } : child,
							path: `${path} > ${name}[${i + 1}]`,
							scope: inner,
							inherited: inheritedNext,
							depth: depth + 1,
							stack: [...(entry.stack || []), name],
							prepull: entry.prepull,
							reasons: entry.reasons,
						});
					});
					return;
				}
			} else if (handler && handler.handler === 'noop') {
				// Nothing to create, but not nothing to say: what Rotation Master does instead.
				conv.note('action-noop', handler.fidelity || 'lossy', { kind: kind as string, note: handler.note });
			} else if (handler && kind !== null) {
				created = doAction(handler, kind, action, entry);
			}
		} catch (e) {
			if (e instanceof Budget) {
				ctx.item.reasons.push({ code: 'too-complex', level: 'unsupported', params: {} });
				created = [];
			} else throw e;
		}
		finish(record, created, handler, hidden);
		addItem(record);
	}

	// The steps an action makes: [{type, action, tree}] (conditions already merged).
	function doAction(handler: Json, kind: string, action: Json, entry: Entry): Made[] {
		const { scope, inherited } = entry;
		const body = action[kind] || {};
		const condNode = conv.conditionOf([...inherited, { value: action.condition, scope }]);
		const cast = (sub: Json, subKind: string): Omit<Made, 'condNode'> | null => {
			const node = isObj(sub) ? sub[subKind] : null;
			const r = conv.spellOf(node?.spellId);
			if ('fail' in r && r.fail === 'item-id') {
				const id = node.spellId.itemId;
				if (Number.isInteger(id) && id > 0 && id <= mapping.validation.idMax) return { type: 'item', action: [id] };
				conv.note('action-no-id', 'unsupported');
				return null;
			}
			if ('fail' in r && r.fail === 'other-id') {
				// The simulator's "resource potion" is whichever mana potion its consumables setting names; for a class that
				// uses mana, Rotation Master's nearest is an item step with this edition's mana potions.
				const potions = options.potions;
				if (r.other === 'OtherActionResourcePotion' && potions && potions.length && !(mapping.noManaClasses || []).includes(cls)) {
					conv.note('action-potion', 'approximate', { edition });
					return { type: 'item', action: potions.slice() };
				}
				conv.note('action-other-id', 'unsupported', { id: (r.other as string).slice(0, 60) });
				return null;
			}
			if ('fail' in r) {
				conv.note('action-no-id', 'unsupported');
				return null;
			}
			if (node.target !== undefined) {
				const u = conv.unitName(node.target, 'Self');
				if ('fail' in u || !['player', 'target'].includes(u.unit))
					conv.note('action-target', 'approximate', { unit: String('fail' in u ? u.fail : u.unit) });
			}
			return { type: 'spell', action: r.id };
		};
		if (handler.handler === 'cast') {
			let level: string | undefined = handler.fidelity;
			if (handler.lossyIf) {
				const used = handler.lossyIf.filter((k: string) => body[k] !== undefined && body[k] !== false && body[k] !== null);
				if (used.length) {
					conv.note('action-' + handler.note, 'lossy', { parts: used.join(', ') });
					level = undefined;
				}
			} else if (level) conv.note('action-' + handler.note, level as Reason['level'], {});
			const made = cast(action, kind);
			if (!made) return [];
			let extra: Json = null;
			if (handler.guard) {
				const g = conv.build(handler.guard, body, null);
				if ('fail' in g) conv.note(g.fail.code, 'lossy', g.fail.params);
				else extra = g.tree;
			}
			return [{ ...made, condNode, extra }];
		}
		if (handler.handler === 'sequence') {
			const out: Made[] = [];
			for (const sub of asList(body.actions)) {
				if (!isObj(sub)) continue;
				let sk: string | null = null;
				for (const k of Object.keys(sub))
					if (k !== 'condition' && k !== 'uuid') {
						sk = k;
						break;
					}
				if (sk === 'castSpell' || sk === 'castFriendlySpell' || sk === 'channelSpell') {
					const made = cast(sub, sk);
					if (!made) continue;
					const subCond = sub.condition !== undefined ? conv.conditionOf([{ value: sub.condition, scope }]) : TRUE;
					out.push({ ...made, condNode: combineNodes(condNode, subCond), extra: null });
				} else {
					conv.note('action-unsupported', 'lossy', {
						kind: String(sk),
						why: sk !== null && own(mapping.unsupportedActions, sk) ? mapping.unsupportedActions[sk] : 'unknown',
					});
				}
			}
			conv.note('action-sequence', 'lossy', { count: out.length });
			return out;
		}
		return [];
	}
	function combineNodes(a: CondNode, b: CondNode): CondNode {
		const known = (n: CondNode): CondNode => (n.k === 'cond' || n.k === 'true' || n.k === 'false' || n.k === 'drop' ? n : TRUE);
		const na = known(a);
		const nb = known(b);
		if (na.k === 'false' || nb.k === 'false') return FALSE;
		if (na.k === 'drop' || nb.k === 'drop') {
			const keep = na.k === 'cond' ? na : nb.k === 'cond' ? nb : null;
			return keep ? { k: 'cond', tree: keep.tree, dropped: true } : dropNode();
		}
		if (na.k === 'true') return nb;
		if (nb.k === 'true') return na;
		const ta = (na as { tree: Json }).tree;
		const tb = (nb as { tree: Json }).tree;
		return { k: 'cond', tree: { type: 'AND', value: [...(ta.type === 'AND' ? ta.value : [ta]), ...(tb.type === 'AND' ? tb.value : [tb])] } };
	}

	// Decides fidelity and outcome, builds the steps and fills the report record.
	function finish(record: FromAplItem, created: Made[], handler: Json, hidden: boolean): void {
		const reasons = dedupe(ctx.item.reasons);
		ctx.item.reasons = reasons;
		let level: Fidelity = 'exact';
		for (const r of reasons) if ((FIDELITIES as ReadonlyArray<string>).includes(r.level)) level = worse(level, r.level as Fidelity);
		const made = created.filter(Boolean);
		if (!made.length && handler && handler.handler === 'noop') {
			record.fidelity = handler.fidelity || 'lossy';
			record.outcome = 'skipped';
			record.reasons = reasons;
			return;
		}
		if (!made.length) {
			record.fidelity = 'unsupported';
			record.outcome = 'skipped';
			record.reasons = reasons;
			if (!reasons.some(r => r.level === 'unsupported')) record.reasons.push({ code: 'nothing-to-import', level: 'unsupported', params: {} });
			return;
		}
		if (level === 'unsupported') level = 'lossy';
		// A reason that is only a note (the pre-pull seconds are always dropped) does not make the step less trustworthy.
		let effective: Fidelity = 'exact';
		for (const r of reasons)
			if ((FIDELITIES as ReadonlyArray<string>).includes(r.level) && !(r.code === 'prepull-timing')) effective = worse(effective, r.level as Fidelity);
		if (effective === 'unsupported') effective = 'lossy';
		// Conditions: dropped parts show in the reasons already; always-false disables.
		let alwaysFalse = false;
		for (const m of made) {
			let node = m.condNode;
			if (m.extra && node.k !== 'false') node = combineNodes(node, { k: 'cond', tree: m.extra });
			m.tree = node.k === 'cond' ? node.tree : null;
			if (node.k === 'false') alwaysFalse = true;
		}
		if (alwaysFalse) reasons.push({ code: 'always-false', level: 'exact', params: {} });
		const policy = mapping.policy;
		const policyDisables =
			effective === 'exact' ? false : effective === 'approximate' ? policy.approximate === 'disable' && !options.enableApproximate : true;
		const disabled = alwaysFalse || policyDisables || hidden;
		record.fidelity = level;
		record.outcome = disabled ? 'disabled' : 'enabled';
		record.reasons = reasons;
		for (const m of made) {
			if (steps.length >= lim.steps) {
				record.reasons.push({ code: 'step-limit', level: 'unsupported', params: { limit: lim.steps } });
				record.outcome = 'skipped';
				record.fidelity = 'unsupported';
				ctx.stop = true;
				break;
			}
			const step: RmStep = { id: newId(), type: m.type, action: m.action };
			if (ranks && m.type === 'spell') step.ranked = false;
			if (m.tree) step.conditions = m.tree as RmCondition;
			if (disabled) step.disabled = true;
			steps.push(step);
			record.became.push({ type: step.type, action: step.action, disabled, conditions: conditionText(m.tree) });
		}
	}

	// Pre-pull actions run before the fight in the simulator, at a time before the pull. Rotation Master has no timeline,
	// but it has "not in combat": each becomes a step that only applies out of combat, ahead of the priority list, in the
	// order the simulator runs them (earliest first). The seconds are the one thing dropped, and the report says so.
	const prepull = asList(apl?.prepullActions)
		.slice(0, 200)
		.map((p, i) => {
			const raw = isObj(p?.doAtValue?.const) ? p.doAtValue.const.val : undefined;
			const parsed = parseConst(raw);
			const seconds =
				parsed && parsed.bool === undefined && Number.isFinite(parsed.num)
					? (parsed.num as number) * (parsed.suffix === 'ms' ? 0.001 : parsed.suffix === 'm' ? 60 : 1)
					: 0;
			return { p, i, when: typeof raw === 'string' ? raw.slice(0, 20) : '', seconds: Number.isFinite(seconds) ? seconds : 0 };
		})
		.sort((a, b) => a.seconds - b.seconds || a.i - b.i);
	for (const { p, i, when } of prepull) {
		if (ctx.stop) break;
		doEntry({
			item: p,
			path: `prepullActions[${i + 1}]`,
			scope: rootScope,
			inherited: [{ tree: OUT_OF_COMBAT, scope: rootScope }],
			depth: 0,
			stack: [],
			prepull: true,
			reasons: [{ code: 'prepull-timing', level: 'lossy', params: { time: when } }],
		});
	}
	// The priority list, in order.
	asList(apl?.priorityList)
		.slice(0, lim.items)
		.forEach((item, i) => {
			if (ctx.stop) return;
			doEntry({ item, path: `priorityList[${i + 1}]`, scope: rootScope, inherited: [], depth: 0, stack: [] });
		});
	if (asList(apl?.priorityList).length > lim.items) {
		addItem({
			path: 'priorityList',
			original: '',
			fidelity: 'unsupported',
			outcome: 'skipped',
			reasons: [{ code: 'item-limit', level: 'unsupported', params: { limit: lim.items } }],
			became: [],
		});
	}
	const summary: FromAplSummary = {
		items: 0,
		exact: 0,
		approximate: 0,
		lossy: 0,
		unsupported: 0,
		enabled: 0,
		disabled: 0,
		skipped: 0,
		prepull: 0,
		steps: steps.length,
		truncated: ctx.truncated || 0,
	};
	for (const r of items) {
		if (r.hidden && r.outcome === 'skipped') continue;
		if (r.prepull) summary.prepull++;
		summary.items++;
		summary[r.fidelity]++;
		summary[r.outcome]++;
	}
	return { steps, items, summary };
}
