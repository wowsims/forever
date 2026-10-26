// A Rotation Master rotation as a WowSims APL. Pure: no DOM, no network; everything it needs comes in as arguments.
// It is the sim-side port of the website's exporter (rotationmaster-web wowsims/export.mjs) and reads the same rule
// table (mapping.json, its `export` section), so the site and the simulators agree on what a rotation becomes.
//
// The input is a rotation body as plain data: {rotation:[step...], cooldowns:[step...]}, a step being
// {id, type, action, conditions, disabled, mode, onTrue, onFalse, ...} with conditions as plain trees. The output is an
// APLRotation as proto-JSON (what the simulators' preset files hold) and a report saying, step by step, what became of
// it. What the structure becomes:
//  - a branch is flattened: each arm's steps carry the branch condition (or its negation) as an extra condition;
//  - a Stop with conditions C adds NOT C to every step after it; an unconditional Stop ends the list;
//  - the Cooldowns list (independent in Rotation Master) becomes the first items of the priority list;
//  - a step that applies only out of combat becomes a pre-pull action (the seconds are assumed);
//  - a condition with no APL counterpart is left out of its AND (the step then fires more often) and listed; one that
//    cannot be left out of an OR makes the whole OR unknown, which is left out the same way.
import rawMapping from './mapping.json';
import { type ClassKey, type Edition, FIDELITIES, type Fidelity, type Ladders, type Reason, type SimKinds, worse } from './report';

type Json = any;
const mapping = rawMapping as Json;
const own = (o: unknown, k: string): boolean => o !== null && typeof o === 'object' && Object.prototype.hasOwnProperty.call(o, k);
const isObj = (v: unknown): v is Record<string, Json> => v !== null && typeof v === 'object' && !Array.isArray(v);

export interface ToAplOptions {
	edition: Edition;
	cls: ClassKey;
	/** The APL kinds the target simulator reads. Absent: every kind of the rule table is assumed to be readable. */
	kinds?: SimKinds;
	/** Rank ladders (editions with spell ranks): an id becomes the highest rank of its ladder. */
	ladders?: Ladders;
	/** Whether the simulator's ActionID carries a spell rank. Default: whether the edition has ranks. */
	actionIdHasRank?: boolean;
	/** Keep a disabled Rotation Master step as a hidden APL item instead of leaving it out (default true). */
	keepDisabled?: boolean;
	/** Item ids of this edition's mana potions: a step that names only these becomes the simulator's resource potion. */
	potions?: ReadonlyArray<number>;
	/** Hide the items that lost a condition or more (fidelity lossy), so nothing runs more often than the original (default false, as on the website). */
	hideLossy?: boolean;
}

export interface ToAplItem {
	path: string;
	section: 'rotation' | 'cooldowns';
	label: string;
	type: string;
	name: string;
	fidelity: Fidelity;
	outcome: 'exported' | 'hidden' | 'dropped' | 'skipped';
	reasons: Reason[];
	became: number;
	/** Where its actions ended up in the priority list (indexes). */
	at: number[];
}

export interface ToAplSummary {
	items: number;
	exact: number;
	approximate: number;
	lossy: number;
	unsupported: number;
	exported: number;
	hidden: number;
	dropped: number;
	skipped: number;
	actions: number;
	prepull: number;
	truncated: number;
}

export type ToAplResult =
	| { ok: true; apl: Json; items: ToAplItem[]; summary: ToAplSummary; kinds: string[] }
	| { ok: false; error: { code: string; [key: string]: string } };

const TRUE = { k: 'true' } as const;
const FALSE = { k: 'false' } as const;
const DROP = { k: 'drop' } as const;
const trim = (n: number): string => String(Number(Number(n).toFixed(3)));
const FORMAT: Record<string, (v: number) => string> = {
	seconds: v => trim(v) + 's',
	fraction: v => trim(v * 100) + '%',
	count: v => String(Math.round(v)),
	raw: v => trim(v),
};
const MAX_STEPS = 2000;
const MAX_FLAT = 4000;

/** A condition tree in plain form: a missing or empty one is null. */
const cleanTree = (tree: Json): Json | null => (isObj(tree) && typeof tree.type === 'string' ? tree : null);
const and = (...trees: Array<Json | null>): Json | null => {
	const list = trees.filter(Boolean);
	if (!list.length) return null;
	if (list.length === 1) return list[0];
	return { type: 'AND', value: list };
};
// NOT of a NOT is the inner tree; the negation of a missing (always true) tree is never true.
const not = (tree: Json | null): Json =>
	!tree ? { type: 'NOT', value: { type: 'AND', value: [] } } : tree.type === 'NOT' && isObj(tree.value) ? tree.value : { type: 'NOT', value: tree };
const listOf = (v: Json): Json[] => (Array.isArray(v) ? v : isObj(v) ? Object.values(v) : []);

// A tree is "out of combat" when it is exactly NOT COMBAT(player), or an AND that has it among its parts.
const isNotCombat = (t: Json | null): boolean =>
	!!t && t.type === 'NOT' && isObj(t.value) && t.value.type === 'COMBAT' && (t.value.unit === undefined || t.value.unit === 'player');
const conjuncts = (tree: Json, out: Json[] = [], depth = 0): Json[] => {
	const t = cleanTree(tree);
	if (!t || depth > 32) return out;
	if (t.type === 'AND' && t.disabled !== true) for (const c of listOf(t.value)) conjuncts(c, out, depth + 1);
	else out.push(t);
	return out;
};
const splitOutOfCombat = (tree: Json | null): { out: boolean; rest: Json | null } => {
	const parts = conjuncts(tree);
	// "Not in combat" together with "in combat" can never be true; that is not a pre-pull action.
	if (!parts.some(isNotCombat) || parts.some(p => p.type === 'COMBAT' && (p.unit === undefined || p.unit === 'player'))) return { out: false, rest: tree };
	return { out: true, rest: and(...parts.filter(t => !isNotCombat(t))) };
};

/** Exports a rotation body for an edition and class. */
export function rotationToApl(body: Json, options: ToAplOptions): ToAplResult {
	const { edition, cls } = options;
	if (!isObj(body)) return { ok: false, error: { code: 'not-rotation' } };
	const EXP = mapping.export;
	if (!own(mapping.editions, edition)) return { ok: false, error: { code: 'edition-unsupported', edition: String(edition) } };
	const hasRanks = mapping.editions[edition].ranks === true;
	const idHasRank = options.actionIdHasRank ?? hasRanks;
	const ladders = options.ladders;
	const keepDisabled = options.keepDisabled !== false;
	const usedKinds = new Set<string>();
	const canRead = (kind: string): boolean => !options.kinds || options.kinds.values.has(kind);

	const items: ToAplItem[] = [];
	const prepull: Array<{ id: Json; item: ToAplItem }> = [];
	let priority: Array<{ action: Json; item: ToAplItem; hidden: boolean }> = [];
	let flat = 0;
	let item: ToAplItem | null = null;
	const note = (code: string, level: Reason['level'], params: Reason['params'] = {}): void => {
		if (item && item.reasons.length < 40) item.reasons.push({ code, level, params });
	};

	// ---- ids ----
	// An RM spell id as the simulator's ActionID. Rotation Master's Built Ins name a spell by its first rank and let it match
	// every rank (ranked: false); the simulator needs the rank its character casts, which for a character at the level cap
	// is the highest. Any other rank is taken as chosen on purpose (a downranked spell) and kept, with its rank.
	const actionId = (raw: Json, kind = 'spell'): { id: Json } | { fail: string } => {
		if (typeof raw !== 'number' || !Number.isInteger(raw) || raw < 1) return { fail: 'id' };
		if (kind === 'item') return { id: { itemId: raw } };
		const ladder = hasRanks ? ladders?.get(raw) : undefined;
		if (ladder && ladder.length) {
			const mine = ladder.find(x => x[0] === raw);
			const top = ladder[ladder.length - 1];
			if (mine && mine[0] === ladder[0][0] && top[0] !== raw) {
				note('x-rank', 'approximate', { from: raw, to: top[0], rank: top[1] });
				return { id: idHasRank ? { spellId: top[0], rank: top[1] } : { spellId: top[0] } };
			}
			return { id: idHasRank && mine ? { spellId: raw, rank: mine[1] } : { spellId: raw } };
		}
		return { id: { spellId: raw } };
	};
	const unitRef = (unit: Json): { ref: Json } | { fail: string } => {
		const u = unit === undefined || unit === null ? 'player' : unit;
		if (!own(EXP.units, u)) return { fail: String(u).slice(0, 30) };
		return { ref: EXP.units[u] };
	};

	// ---- conditions ----
	type Node = { k: 'true' } | { k: 'false' } | { k: 'drop' } | { k: 'val'; v: Json };
	const val = (v: Json): Node => ({ k: 'val', v });
	const dropLeaf = (tree: Json, why: string, extra: Reason['params'] = {}): Node => {
		note('x-cond-dropped', 'lossy', { type: tree.type, why, ...extra });
		return DROP;
	};
	function build(args: Json, tree: Json): { args: Json } | { fail: string; unit?: string } {
		const out: Json = {};
		for (const [key, spec] of Object.entries(args || {})) {
			if (typeof spec !== 'string' || spec[0] !== '@') {
				out[key] = spec;
				continue;
			}
			const parts = spec.split(':');
			if (parts[0] === '@id') {
				const r = actionId(tree[parts[1]]);
				if ('fail' in r) return { fail: 'id' };
				out[key] = r.id;
				continue;
			}
			if (parts[0] === '@target') {
				// A dot's target defaults to your current target, so only another unit is written.
				const r = unitRef(tree[parts[1]]);
				if ('fail' in r) return { fail: 'unit', unit: r.fail };
				if (r.ref && r.ref.type !== 'CurrentTarget') out[key] = structuredClone(r.ref);
				continue;
			}
			if (parts[0] === '@unit') {
				const r = unitRef(tree[parts[1]]);
				if ('fail' in r) return { fail: 'unit', unit: r.fail };
				if (r.ref) out[key] = structuredClone(r.ref);
				continue;
			}
			if (parts[0] === '@table') {
				const table = EXP.tables[parts[1]];
				const raw = tree[parts[2]];
				const k = typeof raw === 'boolean' ? String(raw) : raw === undefined && parts[1].startsWith('swing') ? 'false' : String(raw);
				if (!table || !own(table, k)) return { fail: 'value' };
				out[key] = table[k];
				continue;
			}
			return { fail: 'value' };
		}
		return { args: out };
	}
	function valueNode(tree: Json, entry: Json): { kind: string; args: Json; approx: string | null } | { fail: string; unit?: string } {
		let kind: string = entry.kind;
		let args: Json = entry.args;
		let approx: string | null = entry.fidelity === 'approximate' ? entry.note : null;
		if (entry.variants) {
			const v = entry.variants.find((x: Json) => !x.when || Object.entries(x.when).every(([k, want]) => tree[k] === want));
			if (!v) return { fail: 'no-equivalent' };
			kind = v.kind;
			args = v.args;
			if (v.fidelity === 'approximate') approx = v.note;
		}
		if (entry.handler === 'power') {
			if (!own(EXP.power, cls)) return { fail: 'class' };
			kind = EXP.power[cls];
			args = {};
			if (own(EXP.powerApprox, cls)) approx = EXP.powerApprox[cls];
			if (tree.unit && tree.unit !== 'player') return { fail: 'unit', unit: String(tree.unit) };
		} else if (entry.handler === 'points') {
			if (!own(EXP.points, cls)) return { fail: 'class' };
			kind = EXP.points[cls];
			args = {};
			if (tree.unit && tree.unit !== 'player') return { fail: 'unit', unit: String(tree.unit) };
		} else if (entry.handler === 'pet') {
			if (!own(EXP.petKinds, cls)) return { fail: 'class' };
			kind = EXP.petKinds[cls];
			args = {};
		}
		if (entry.self && tree.unit && tree.unit !== 'player') return { fail: 'unit', unit: String(tree.unit) };
		if (entry.not) for (const [k, want] of Object.entries(entry.not)) if (tree[k] === want) return { fail: 'no-equivalent' };
		if (entry.only)
			for (const [k, want] of Object.entries(entry.only))
				if ((tree[k] ?? (k === 'unit' ? 'player' : undefined)) !== want) return { fail: 'no-equivalent' };
		const built = build(args, tree);
		if ('fail' in built) return built;
		return { kind, args: built.args, approx };
	}
	function leaf(tree: Json): Node {
		const type: string = tree.type;
		if (own(EXP.trueIf, type)) {
			note('x-cond-approx', 'approximate', { type, note: EXP.trueIf[type] });
			return TRUE;
		}
		// What the simulator decides itself: it only chooses an action when you are free to act (never casting or
		// channeling), and its target is a boss.
		if (own(EXP.falseIf, type) && (tree.unit === undefined || tree.unit === 'player')) {
			note('x-cond-approx', 'approximate', { type, note: EXP.falseIf[type] });
			return FALSE;
		}
		if (type === 'CLASSIFICATION' && (tree.unit === undefined || tree.unit === 'target') && typeof tree.value === 'string' && EXP.classification) {
			const word = tree.value.toLowerCase();
			if (EXP.classification.boss.includes(word) || EXP.classification.other.includes(word)) {
				note('x-cond-approx', 'approximate', { type, note: 'sim-boss' });
				return EXP.classification.boss.includes(word) ? TRUE : FALSE;
			}
		}
		if (!own(EXP.conditions, type)) return dropLeaf(tree, 'no-equivalent');
		const entry = EXP.conditions[type];
		const made = valueNode(tree, entry);
		if ('fail' in made) return dropLeaf(tree, made.fail, made.unit ? { unit: made.unit } : {});
		if (!canRead(made.kind)) return dropLeaf(tree, 'sim-lacks', { kind: made.kind });
		usedKinds.add(made.kind);
		const node = { [made.kind]: made.args };
		if (made.approx) note('x-cond-approx', 'approximate', { type, note: made.approx });
		if (entry.t === 'bool') return val(entry.cmp ? { cmp: { op: entry.cmp.op, lhs: node, rhs: { const: { val: entry.cmp.val } } } } : node);
		const op = own(EXP.ops, tree.operator) ? EXP.ops[tree.operator] : null;
		if (!op) return dropLeaf(tree, tree.operator === undefined ? 'incomplete' : 'operator');
		if (typeof tree.value !== 'number' || !Number.isFinite(tree.value)) return dropLeaf(tree, 'incomplete');
		if (tree.value < 0 && (type === 'POINT' || type === 'HEALTH' || type === 'MANA' || type === 'POWER')) return dropLeaf(tree, 'value');
		return val({ cmp: { op, lhs: node, rhs: { const: { val: FORMAT[entry.unit](tree.value) } } } });
	}
	function convert(treeIn: Json, depth: number): Node {
		if (depth > 64) return dropLeaf({ type: 'AND' }, 'depth');
		const tree = cleanTree(treeIn);
		if (!tree || tree.disabled === true) return TRUE;
		if (tree.type === 'AND' || tree.type === 'OR') {
			const kids = listOf(tree.value).map(c => convert(c, depth + 1));
			const isAnd = tree.type === 'AND';
			const vals: Json[] = [];
			let unknown = false;
			for (const k of kids) {
				if (k.k === 'true') {
					if (!isAnd) return TRUE;
					continue;
				}
				if (k.k === 'false') {
					if (isAnd) return FALSE;
					continue;
				}
				if (k.k === 'drop') {
					unknown = true;
					continue;
				}
				vals.push(k.v);
			}
			if (!isAnd && unknown) return DROP; // an unknown alternative could be the true one
			if (!vals.length) return unknown ? DROP : isAnd ? TRUE : FALSE;
			if (vals.length === 1) return val(vals[0]);
			return val({ [isAnd ? 'and' : 'or']: { vals } });
		}
		if (tree.type === 'NOT') {
			const inner = convert(tree.value, depth + 1);
			if (inner.k === 'true') return FALSE;
			if (inner.k === 'false') return TRUE;
			if (inner.k === 'drop') return DROP;
			return val({ not: { val: inner.v } });
		}
		return leaf(tree);
	}

	// ---- actions ----
	// A step's casts: [{spellId...}|{itemId}] or a failure code.
	function castsOf(step: Json): { casts: Json[] } | { fail: string } {
		const type = step.type;
		if (type === 'spell' || type === 'any') {
			const r = actionId(step.action);
			if ('fail' in r) return { fail: 'x-action-id' };
			return { casts: [r.id] };
		}
		if (type === 'item') {
			const list = Array.isArray(step.action) ? step.action : typeof step.action === 'number' ? [step.action] : null;
			if (!list) return { fail: 'x-action-itemset' };
			const ids = list.filter((x: Json) => typeof x === 'number' && Number.isInteger(x) && x > 0);
			if (!ids.length) return { fail: 'x-action-itemset' };
			// Every id a mana potion of this edition: the simulator's resource potion placeholder, as the import made it.
			const potions = options.potions ?? [];
			if (potions.length && ids.every((id: number) => potions.includes(id)) && !(mapping.noManaClasses || []).includes(cls)) {
				note('x-potion', 'approximate', {});
				return { casts: [{ otherId: EXP.resourcePotion }] };
			}
			if (ids.length > 1) note('x-item-list', 'approximate', { count: ids.length });
			return { casts: ids.map((id: number) => ({ itemId: id })) };
		}
		if (type === 'pet') return { fail: 'x-action-pet' };
		if (type === 'assisted') return { fail: 'x-action-assisted' };
		return { fail: 'x-action-unknown' };
	}
	const labelOf = (step: Json): string => {
		if (step.type === 'item') {
			const ids: number[] = Array.isArray(step.action) ? step.action.filter((x: Json) => typeof x === 'number') : [];
			return ids.map(id => '#' + id).join(',');
		}
		return typeof step.action === 'number' ? '#' + step.action : '';
	};
	const NOTES: Record<string, string> = {
		'x-rank': 'rank raised',
		'x-prepull-cond': 'its other conditions were lost as a pre-pull action',
		'x-cond-dropped': 'a condition has no WowSims equivalent and was left out',
		'x-cond-approx': 'a condition was converted approximately',
		'x-cooldown-first': 'a Rotation Master cooldown, placed first',
		'x-prepull-time': 'pre-pull time assumed',
		'x-item-list': 'one of several items',
		'x-cond-never': 'its condition can never be true in the simulator',
	};
	const noteText = (reasons: Reason[]): string | null => {
		const words = [...new Set(reasons.filter(r => r.level !== 'exact' && NOTES[r.code]).map(r => NOTES[r.code]))];
		return words.length ? 'Rotation Master: ' + words.join('; ') + '.' : null;
	};

	const newItem = (path: string, step: Json, section: 'rotation' | 'cooldowns'): ToAplItem => ({
		path,
		section,
		label: labelOf(step),
		type: String(step.type),
		name: typeof step.name === 'string' ? step.name.slice(0, 120) : '',
		fidelity: 'exact',
		outcome: 'exported',
		reasons: [],
		became: 0,
		at: [],
	});
	const finishItem = (rec: ToAplItem): void => {
		let level: Fidelity = 'exact';
		for (const r of rec.reasons) if ((FIDELITIES as ReadonlyArray<string>).includes(r.level)) level = worse(level, r.level as Fidelity);
		rec.fidelity = level;
		items.push(rec);
	};

	interface WalkState {
		ended: boolean;
		stops: Json[];
	}
	// Walks a list of steps. guards: RM trees every step in the list also needs; state.stops: the NOT(stop condition)
	// trees that apply to everything after a Stop; state.ended: an unconditional Stop was passed.
	function walk(list: Json, path: string, section: 'rotation' | 'cooldowns', guards: Array<Json | null>, state: WalkState, depth: number): void {
		const steps: Json[] = Array.isArray(list) ? list : [];
		steps.forEach((step, i) => {
			if (!isObj(step)) return;
			if (++flat > MAX_FLAT) return;
			const here = `${path}[${i + 1}]`;
			item = newItem(here, step, section);
			// A disabled Stop does nothing; a disabled action is kept hidden unless asked not to.
			if (step.disabled === true && (!keepDisabled || step.type === 'stop')) {
				item.outcome = 'skipped';
				note('x-disabled', 'exact');
				finishItem(item);
				item = null;
				return;
			}
			if (state.ended) {
				item.outcome = 'skipped';
				note('x-unreachable', 'exact', { why: 'stop' });
				finishItem(item);
				item = null;
				return;
			}
			if (step.type === 'branch') {
				const guard = cleanTree(step.conditions);
				// A branch with no condition is always true; a forced mode leaves the other arm unreachable.
				let mode: string = step.mode || 'evaluate';
				if (mode === 'evaluate' && !guard) mode = 'true';
				item.type = 'branch';
				item.label = '';
				note('x-branch', 'exact', { mode });
				finishItem(item);
				item = null;
				const arms = [
					{ list: step.onTrue, name: '.onTrue', reach: mode !== 'false', guards: mode === 'true' ? [] : [guard] },
					{ list: step.onFalse, name: '.onFalse', reach: mode !== 'true', guards: mode === 'false' ? [] : [not(guard)] },
				];
				for (const arm of arms) {
					if (!arm.reach) {
						markUnreachable(arm.list, here + arm.name, section);
						continue;
					}
					const armState: WalkState = { ended: false, stops: [...state.stops] };
					walk(arm.list, here + arm.name, section, [...guards, ...arm.guards], armState, depth + 1);
					// A Stop inside an arm ends the whole list when the arm was taken: it carries on as a condition on everything after.
					for (const stop of armState.stops.slice(state.stops.length)) state.stops.push(stop);
					if (armState.ended) {
						const full = and(...guards, ...arm.guards);
						if (full) state.stops.push(not(full));
						else state.ended = true;
					}
				}
				return;
			}
			if (step.type === 'stop') {
				const cond = cleanTree(step.conditions);
				note('x-stop', 'exact');
				finishItem(item);
				item = null;
				// Unconditional: everything after it in this list is unreachable. Conditional: it needs "not (this condition)".
				if (!cond) state.ended = true;
				else state.stops.push(not(and(...guards, cond)));
				return;
			}
			exportStep(step, here, section, [...guards, ...state.stops]);
		});
	}
	function markUnreachable(list: Json, path: string, section: 'rotation' | 'cooldowns'): void {
		(Array.isArray(list) ? list : []).forEach((step: Json, i: number) => {
			if (!isObj(step)) return;
			const rec = newItem(`${path}[${i + 1}]`, step, section);
			rec.outcome = 'skipped';
			item = rec;
			note('x-unreachable', 'exact', { why: 'mode' });
			finishItem(rec);
			item = null;
			if (step.type === 'branch') {
				markUnreachable(step.onTrue, `${path}[${i + 1}].onTrue`, section);
				markUnreachable(step.onFalse, `${path}[${i + 1}].onFalse`, section);
			}
		});
	}

	function exportStep(step: Json, path: string, section: 'rotation' | 'cooldowns', guards: Array<Json | null>): void {
		const rec = item as ToAplItem;
		const casts = castsOf(step);
		if ('fail' in casts) {
			rec.outcome = 'dropped';
			rec.fidelity = 'unsupported';
			note(casts.fail, 'unsupported');
			finishItem(rec);
			item = null;
			return;
		}
		const ownTree = cleanTree(step.conditions);
		const hidden = step.disabled === true;
		if (hidden) {
			rec.outcome = 'hidden';
			note('x-hidden', 'exact');
		}
		// A step that applies only out of combat is a pre-pull action to the simulator. Pre-pull actions carry no
		// conditions, so any other condition on the step is lost (and said so), and the seconds are assumed.
		const split = splitOutOfCombat(and(...guards, ownTree));
		if (split.out) {
			if (split.rest) {
				const converted = convert(split.rest, 0);
				if (converted.k !== 'true') note('x-prepull-cond', 'lossy');
			}
			note('x-prepull-time', 'lossy');
			for (const id of casts.casts) prepull.push({ id, item: rec });
			rec.became = casts.casts.length;
			finishItem(rec);
			item = null;
			return;
		}
		const cond = convert(and(...guards, ownTree), 0);
		if (cond.k === 'false') {
			rec.outcome = 'dropped';
			note('x-cond-never', 'lossy');
			finishItem(rec);
			item = null;
			return;
		}
		if (section === 'cooldowns') note('x-cooldown-first', 'approximate');
		for (const id of casts.casts) {
			if (priority.length >= MAX_STEPS) break;
			const action: Json = {};
			if (cond.k === 'val') action.condition = cond.v;
			action.castSpell = { spellId: id };
			priority.push({ action, item: rec, hidden });
		}
		rec.became = casts.casts.length;
		finishItem(rec);
		item = null;
	}

	const rotationState: WalkState = { ended: false, stops: [] };
	// Cooldowns are independent in Rotation Master: no Stop or branch touches them.
	walk(body.cooldowns, 'cooldowns', 'cooldowns', [], { ended: false, stops: [] }, 0);
	const cooldownItems = priority;
	priority = [];
	walk(body.rotation, 'rotation', 'rotation', [], rotationState, 0);
	const all = [...cooldownItems, ...priority];

	// An item that lost a condition runs more often than the original did; on request it is hidden instead.
	if (options.hideLossy) {
		for (const entry of all) {
			if (entry.hidden || entry.item.fidelity === 'exact' || entry.item.fidelity === 'approximate') continue;
			entry.hidden = true;
			if (entry.item.outcome === 'exported') {
				entry.item.outcome = 'hidden';
				entry.item.reasons.push({ code: 'x-hidden-lossy', level: 'exact', params: {} });
			}
		}
	}
	// Where each report item's actions ended up in the priority list.
	all.forEach((entry, i) => entry.item.at.push(i));
	// An item's notes carry its own reasons, once the item is final.
	const priorityList = all.map(entry => {
		const out: Json = {};
		if (entry.hidden) out.hide = true;
		const text = noteText(entry.item.reasons);
		if (text) out.notes = text;
		out.action = entry.action;
		return out;
	});
	// Pre-pull: earliest first, one second apart, the last one a second before the pull.
	const prepullActions = prepull.map((p, i) => ({ action: { castSpell: { spellId: p.id } }, doAtValue: { const: { val: `-${prepull.length - i}s` } } }));

	const apl = { type: 'TypeAPL', prepullActions, priorityList };
	const summary: ToAplSummary = {
		items: 0,
		exact: 0,
		approximate: 0,
		lossy: 0,
		unsupported: 0,
		exported: 0,
		hidden: 0,
		dropped: 0,
		skipped: 0,
		actions: priorityList.length,
		prepull: prepullActions.length,
		truncated: flat > MAX_FLAT ? flat - MAX_FLAT : 0,
	};
	for (const r of items) {
		if (r.type === 'branch' || r.type === 'stop') continue;
		if (r.outcome === 'skipped') {
			summary.skipped++;
			continue;
		}
		summary.items++;
		summary[r.fidelity]++;
		summary[r.outcome]++;
	}
	return { ok: true, apl, items, summary, kinds: [...usedKinds].sort() };
}
