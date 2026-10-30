import './style.css';

import { createContext, type CSSProperties, Fragment, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { createRoot } from 'react-dom/client';

import { differenceText, perFight, standardError } from '../sim_text';
import type { AplDatabase, AplEdit, AplStructure, Finding, FromPanel, Named, PanelState, Registered, ToPanel } from '../protocol';
import { applyEdit, shiftKeys } from './apply';
import { CommitInput } from './commit_input';
import { DbContext } from './db';
import { type EditorApi, EditorContext, useEditor } from './editor_context';
import { ActionEditor, type ListContext, ValueEditor } from './forms';
import { Icon } from './icons';
import { isObject, type Json, type JsonObject, type JsonPath, newKindJson, setClassName } from './model';
import { KindPicker } from './pickers';
import { idKey, RegisteredContext } from './registered';
import { blocksAt, blocksIn, conditionOf, emptyStructure, joinedBy, key, ownConjuncts, rowAt, StructureContext, useStructure } from './structure';
import { ActionText, joined, ValueText } from './summary';

type SimState = Extract<ToPanel, { type: 'sim' }>;

// The file's last sim, which rows show as counts.
const SimContext = createContext<SimState>({ type: 'sim' });

function simRun(state: SimState) {
	return state.sim && (state.sim.result.compare ?? state.sim.result);
}

// How often the last sim ran the row per fight, or that it never did.
function RowCount({ rowKey }: { rowKey: string }) {
	const state = useContext(SimContext);
	const stats = simRun(state)?.rows[rowKey];
	if (!stats || stats.hidden) return null;
	const n = stats.executions;
	const stale = state.sim?.stale ? ' stale' : '';
	if (n === 0)
		return (
			<span className={`count-badge dead${stale}`} title="The last sim never ran this row">
				never ran
			</span>
		);
	return (
		<span
			className={`count-badge${stale}`}
			title={`The last sim ran this row ${n.toFixed(2)} times per fight${state.sim?.stale ? ' (before the latest edits)' : ''}`}>
			{perFight(n)}×
		</span>
	);
}

function SimBar() {
	const state = useContext(SimContext);
	const run = simRun(state);
	const result = state.sim?.result;
	const send = (compare: boolean) => vscode.postMessage({ type: 'sim', compare });
	return (
		<div className="sim-bar">
			<button type="button" className="sim-button" disabled={state.running} onClick={() => send(false)} title="Sim this rotation, unsaved edits included">
				▶ Sim
			</button>
			<button
				type="button"
				className="sim-button"
				disabled={state.running}
				onClick={() => send(true)}
				title="Sim the saved file (or HEAD) and these edits on the same seed">
				⇄ Compare
			</button>
			{state.running ? <span className="sim-text">simming…</span> : null}
			{!state.running && run && result ? (
				<span
					className={`sim-text${state.sim?.stale ? ' stale' : ''}`}
					title={`${result.character.source}: ${result.character.race}, talents ${result.character.talents}${result.character.gear ? `, gear ${result.character.gear}` : ''}`}>
					<strong>{run.dps.mean.toFixed(1)} DPS</strong> ± {standardError(run).toFixed(1)}
					{result.difference ? (
						<span className={result.difference.delta >= 0 ? 'better' : 'worse'}>
							{' '}
							{differenceText(result.difference)} vs {result.baseline}
						</span>
					) : null}
					{state.sim?.stale ? ' · stale' : ''}
				</span>
			) : null}
		</div>
	);
}

// The order the last sim ran the rows in, for one fight: click an event to find its row.
function SimDetails() {
	const state = useContext(SimContext);
	const { reveal, groupNames } = useEditor();
	const db = useContext(DbContext);
	const run = simRun(state);
	if (!run) return null;
	const pathOf = (key: string) => key.split('/').map(part => (/^\d+$/.test(part) ? Number(part) : part));
	// A row as the panel numbers it: "priority 6", "prepull 1", or the group's name and the row in it.
	const rowName = (key: string | undefined) => {
		const parts = key?.split('/') ?? [];
		if (parts[0] === 'priorityList') return `priority ${Number(parts[1]) + 1}`;
		if (parts[0] === 'prepullActions') return `prepull ${Number(parts[1]) + 1}`;
		if (parts[0] === 'groups') return `${groupNames[Number(parts[1])] ?? `group ${Number(parts[1]) + 1}`} ${Number(parts[3]) + 1}`;
		return key ?? '';
	};
	const spellName = (id: { spellId?: number; itemId?: number; otherId?: string }, name?: string) =>
		name ?? (id.itemId ? db.items[id.itemId]?.[0] : id.spellId ? db.spells[id.spellId]?.[0] : id.otherId) ?? JSON.stringify(id);
	return (
		<section>
			<details>
				<summary>
					<h2>
						Timeline <span className="count">{run.timeline?.length ?? 0} row runs in one fight</span>
					</h2>
				</summary>
				<ol className="timeline">
					{(run.timeline ?? []).map((event, i) => (
						<li key={i} onClick={() => reveal(pathOf(event.path))} title={event.path}>
							<span className="t">{(event.timeMs / 1000).toFixed(2)}s</span>
							<span className="where">{rowName(event.path)}</span>
							<span className="what">
								{(event.label ?? '').replace(/\{ItemID: (\d+)\}/g, (_, id) => db.items[Number(id)]?.[0] ?? `item ${id}`)}
							</span>
							{event.via ? <span className="via">via {rowName(event.via)}</span> : null}
						</li>
					))}
				</ol>
			</details>
			<details>
				<summary>
					<h2>
						Damage <span className="count">by spell</span>
					</h2>
				</summary>
				<table className="spells">
					<tbody>
						{[...run.spells]
							.sort((a, b) => b.dps - a.dps)
							.map((spell, i) => (
								<tr key={i}>
									<td>
										{spellName(spell.id, spell.name)}
										{spell.pet ? ` (${spell.pet})` : ''}
									</td>
									<td className="num">{spell.dps.toFixed(1)} DPS</td>
									<td className="num">{spell.casts.toFixed(1)} casts</td>
								</tr>
							))}
					</tbody>
				</table>
			</details>
		</section>
	);
}

// The lint findings for the file, which rows show as markers: all of them, and those under each path, its own and
// those of anything nested in it.
interface Findings {
	all: Finding[];
	under: Map<string, Finding[]>;
}
const FindingsContext = createContext<Findings>({ all: [], under: new Map() });

function indexFindings(all: Finding[]): Findings {
	const under = new Map<string, Finding[]>();
	for (const finding of all) {
		for (let n = 0; n <= finding.path.length; n++) {
			const at = key(finding.path.slice(0, n));
			under.set(at, [...(under.get(at) ?? []), finding]);
		}
	}
	return { all, under };
}

const RANK = { error: 0, warning: 1, info: 2 } as const;

// The findings under a row: its own and those of anything nested in it.
function RowFindings({ rowKey }: { rowKey: string }) {
	const findings = useContext(FindingsContext).under.get(rowKey) ?? [];
	if (findings.length === 0) return null;
	const worst = findings.reduce((a, b) => (RANK[b.severity] < RANK[a.severity] ? b : a)).severity;
	return (
		<span className={`finding finding-${worst}`} title={findings.map(f => `${f.severity}: ${f.message} (${f.code})`).join('\n')}>
			{worst === 'info' ? 'i' : '!'}
			{findings.length > 1 ? <sub>{findings.length}</sub> : null}
		</span>
	);
}

function FindingsBadge() {
	const findings = useContext(FindingsContext).all;
	if (findings.length === 0) return null;
	const count = (severity: Finding['severity']) => findings.filter(f => f.severity === severity).length;
	return (
		<span className="badge" title={findings.map(f => `${f.path.join('/')}: ${f.message}`).join('\n')}>
			{[
				['error', count('error')],
				['warning', count('warning')],
				['info', count('info')],
			]
				.filter(([, n]) => n)
				.map(([severity, n]) => `${n} ${severity}${n === 1 ? '' : 's'}`)
				.join(' · ')}
		</span>
	);
}

declare function acquireVsCodeApi(): { postMessage(message: FromPanel): void; setState(state: PanelState): void };
const vscode = acquireVsCodeApi();

function collectIds(json: Json, into: Set<string>): Set<string> {
	if (Array.isArray(json)) {
		json.forEach(item => collectIds(item, into));
	} else if (isObject(json)) {
		const key = idKey(json);
		if (key) into.add(key);
		Object.values(json).forEach(value => collectIds(value, into));
	}
	return into;
}

// The "when" of a prepull action: a constant reads as its time, anything else as the value it is.
function DoAt({ json }: { json: Json | undefined }) {
	const constant = isObject(json) && isObject(json.const) ? json.const.val : undefined;
	return <span className="do-at">{constant !== undefined ? String(constant) : json === undefined ? '—' : <ValueText json={json} />}</span>;
}

interface ListProps {
	title: string;
	path: JsonPath;
	items: Json[];
	ctx: ListContext;
	selected?: string;
	expanded: Set<string>;
	toggle: (rowKey: string) => void;
	open: (rowKey: string) => void;
}

type Block = ReturnType<typeof blocksIn>[number];

// A when block's line: its condition, which a click edits, and what it does to its rows.
function WhenHead({
	block,
	blocks,
	path,
	ctx,
	collapsed,
	onCollapse,
}: {
	block: Block;
	blocks: Block[];
	path: JsonPath;
	ctx: ListContext;
	collapsed: boolean;
	onCollapse: () => void;
}) {
	const { edit, revealRange } = useEditor();
	const structure = useStructure();
	const [editing, setEditing] = useState(false);
	const at: JsonPath = ['whens', block.id];
	// The block and those around it, the outermost first.
	const chain: Block[] = [];
	for (let b: Block | undefined = block; b; b = blocks.find(x => x.id === b!.parent)) chain.unshift(b);
	const hidden = chain.some(b => b.hide);
	const add = (kind: string) => {
		const action = { [kind]: newKindJson('action', kind), condition: conditionOf(joinedBy(chain)) };
		edit({ op: 'insert', path, index: block.last + 1, value: hidden ? { action, hide: true } : { action } });
	};
	const rows = block.last - block.first + 1;
	return (
		<li className={`when-head${block.hide ? ' hidden' : ''}`} data-when={block.id} style={{ '--depth': block.depth } as CSSProperties}>
			<div className="row-head">
				<button type="button" className="icon-btn chevron" onClick={onCollapse} title={collapsed ? 'Show its rows' : 'Fold its rows'}>
					<Icon name={collapsed ? 'right' : 'down'} />
				</button>
				<div
					className="summary"
					tabIndex={0}
					title="Click to edit the condition its rows share"
					onClick={() => setEditing(!editing)}
					onDoubleClick={() => revealRange(structure.whenRanges[block.id])}
					onKeyDown={e => {
						if (e.key !== 'Enter' && e.key !== ' ') return;
						setEditing(!editing);
						e.preventDefault();
					}}>
					<span className="op">{block.hide ? '# when' : 'when'}</span> <ValueText json={block.condition} path={[...at, 'condition']} />
					{collapsed ? <span className="count"> {rows} rows</span> : null}
				</div>
				<span className="row-buttons">
					<KindPicker family="action" isPrepull={false} isGroup={ctx.isGroup} onPick={add} trigger={<span title="Add a row to the block">+</span>} />
					<button
						type="button"
						className="icon-btn"
						title={block.hide ? 'Show the block (the sim runs its rows)' : 'Hide the block (the sim skips its rows)'}
						onClick={() => edit({ op: 'set', path: [...at, 'hide'], value: block.hide ? undefined : true })}>
						<Icon name={block.hide ? 'eyeOff' : 'eye'} />
					</button>
					<button type="button" className="icon-btn" title="Jump to the text" onClick={() => revealRange(structure.whenRanges[block.id])}>
						<Icon name="jump" />
					</button>
					<button
						type="button"
						className="icon-btn"
						title="Take the block apart: its rows keep the condition, written on each"
						onClick={() => edit({ op: 'set', path: at })}>
						<Icon name="unblock" />
					</button>
					<button
						type="button"
						className="icon-btn"
						title="Delete the block and its rows"
						onClick={() => edit(...Array.from({ length: rows }, (_, k) => ({ op: 'set' as const, path: [...path, block.last - k] })))}>
						<Icon name="trash" />
					</button>
				</span>
			</div>
			{editing ? (
				<div className="row-body">
					<ValueEditor json={block.condition} path={[...at, 'condition']} ctx={ctx} />
				</div>
			) : null}
		</li>
	);
}

// The rows a selection covers, which a when block can be put around.
interface Selection {
	anchor: number;
	from: number;
	to: number;
}

function ItemList({ title, path, items, ctx, selected, expanded, toggle, open }: ListProps) {
	const { edit, reveal } = useEditor();
	const structure = useStructure();
	const [dragFrom, setDragFrom] = useState<number>();
	const [dropAt, setDropAt] = useState<number>();
	const [collapsed, setCollapsed] = useState<Set<number>>(new Set());
	const [picked, setPicked] = useState<Selection>();
	const list = useRef<HTMLOListElement>(null);
	// The row a keyboard move put a row at, which takes the focus once it shows there.
	const focusAt = useRef<number>(undefined);

	useEffect(() => {
		if (focusAt.current === undefined) return;
		list.current?.querySelectorAll<HTMLElement>(':scope > li > .row-head > .summary')[focusAt.current]?.focus();
		focusAt.current = undefined;
	});

	const blocks = useMemo(() => blocksIn(structure, path), [structure, path]);
	const rowInfo = (i: number) => rowAt(structure, [...path, i]);
	const conditionAt = (i: number) => {
		const row = items[i];
		return isObject(row) && isObject(row.action) ? row.action.condition : undefined;
	};
	// A row's own conjuncts, after those its blocks join.
	const ownAt = (i: number) => {
		const info = rowInfo(i);
		return ownConjuncts(conditionAt(i), info?.inherited ?? 0, info?.own ?? true);
	};

	const move = (from: number, to: number) => to >= 0 && to < items.length && from !== to && edit({ op: 'move', path, from, to });

	// A row dropped on another goes into that row's blocks, with its own conjuncts joined to theirs.
	const drop = (from: number, to: number) => {
		if (to < 0 || to >= items.length || from === to) return false;
		const into = blocksAt(blocks, to);
		if (key(blocksAt(blocks, from).map(b => b.id)) === key(into.map(b => b.id))) return move(from, to);
		const edits: AplEdit[] = [
			{ op: 'move', path, from, to },
			{ op: 'set', path: [...path, to, 'action', 'condition'], value: conditionOf([...joinedBy(into), ...ownAt(from)]) },
		];
		if (into.some(b => b.hide)) edits.push({ op: 'set', path: [...path, to, 'hide'], value: true });
		return edit(...edits);
	};

	const add = (kind: string) => {
		const action = { [kind]: newKindJson('action', kind) };
		edit({ op: 'insert', path, index: items.length, value: ctx.isPrepull ? { action, doAtValue: { const: { val: '-1s' } } } : { action } });
		open(key([...path, items.length]));
	};

	// Puts the picked rows in a new block. Its condition is the first conjunct the rows share, else true to edit.
	const wrap = (sel: Selection) => {
		const owns = Array.from({ length: sel.to - sel.from + 1 }, (_, k) => ownAt(sel.from + k));
		const first = owns[0][0];
		const shared = first !== undefined && owns.every(own => own.length > 0 && JSON.stringify(own[0]) === JSON.stringify(first));
		edit({ op: 'wrap', path, from: sel.from, to: sel.to, value: shared ? first : { const: { val: 'true' } } });
		setPicked(undefined);
	};

	const pick = (i: number, extend: boolean) =>
		setPicked(current => {
			if (extend && current) return { anchor: current.anchor, from: Math.min(current.anchor, i), to: Math.max(current.anchor, i) };
			return current?.from === i && current.to === i ? undefined : { anchor: i, from: i, to: i };
		});

	const folded = (b: Block) => blocks.some(outer => collapsed.has(outer.id) && outer.depth < b.depth && outer.first <= b.first && b.first <= outer.last);
	const heads = (i: number) =>
		blocks
			.filter(b => b.first === i && !folded(b))
			.sort((a, b) => a.depth - b.depth)
			.map(b => (
				<WhenHead
					key={`when-${b.id}`}
					block={b}
					blocks={blocks}
					path={path}
					ctx={ctx}
					collapsed={collapsed.has(b.id)}
					onCollapse={() =>
						setCollapsed(current => {
							const next = new Set(current);
							if (!next.delete(b.id)) next.add(b.id);
							return next;
						})
					}
				/>
			));

	return (
		<section>
			<h2>
				{title} <span className="count">{items.length}</span>
			</h2>
			{picked && !ctx.isPrepull ? (
				<div className="pick-bar">
					<span>
						{picked.from === picked.to ? `Row ${picked.from + 1}` : `Rows ${picked.from + 1}–${picked.to + 1}`} picked; shift-click a number to pick
						more
					</span>
					<button type="button" className="add" onClick={() => wrap(picked)}>
						Put in a when block
					</button>
					<button type="button" className="icon-btn" title="Clear the pick" onClick={() => setPicked(undefined)}>
						<Icon name="close" />
					</button>
				</div>
			) : null}
			<ol className="items" ref={list} onDragLeave={() => setDropAt(undefined)}>
				{items.flatMap((item, i) => {
					const rowPath = [...path, i];
					const rowKey = key(rowPath);
					const row = isObject(item) ? item : {};
					const hidden = row.hide === true;
					const isOpen = expanded.has(rowKey);
					const around = blocksAt(blocks, i);
					if (around.some(b => collapsed.has(b.id))) return heads(i);
					const info = rowInfo(i);
					const condition = isObject(row.action) ? row.action.condition : undefined;
					const own = info ? ownConjuncts(condition, info.inherited, info.own) : undefined;
					return [
						...heads(i),
						<li
							key={i}
							data-row={rowKey}
							style={{ '--depth': around.length } as CSSProperties}
							className={[
								'row',
								hidden && 'hidden',
								selected === rowKey && 'selected',
								picked && picked.from <= i && i <= picked.to && 'picked',
								dropAt === i && dragFrom !== undefined && (dragFrom < i ? 'drop-below' : 'drop-above'),
							]
								.filter(Boolean)
								.join(' ')}
							onDragOver={e => {
								if (dragFrom === undefined) return;
								e.preventDefault();
								setDropAt(i);
							}}
							onDrop={e => {
								e.preventDefault();
								if (dragFrom !== undefined) drop(dragFrom, i);
								setDragFrom(undefined);
								setDropAt(undefined);
							}}>
							<div className="row-head">
								<span
									className="grip"
									draggable
									title="Drag to reorder, into or out of a when block"
									onDragStart={e => {
										e.dataTransfer.effectAllowed = 'move';
										e.dataTransfer.setData('text/plain', rowKey);
										setDragFrom(i);
									}}
									onDragEnd={() => {
										setDragFrom(undefined);
										setDropAt(undefined);
									}}>
									<Icon name="grip" />
								</span>
								<button type="button" className="icon-btn chevron" onClick={() => toggle(rowKey)} title={isOpen ? 'Collapse' : 'Edit'}>
									<Icon name={isOpen ? 'down' : 'right'} />
								</button>
								<span
									className="index"
									title={ctx.isPrepull ? undefined : 'Click to pick the row, shift-click to pick a run of rows'}
									onClick={e => !ctx.isPrepull && pick(i, e.shiftKey)}>
									{i + 1}
								</span>
								<div
									className="summary"
									tabIndex={0}
									onClick={() => reveal(rowPath)}
									onDoubleClick={() => toggle(rowKey)}
									onKeyDown={e => {
										if (e.altKey && (e.key === 'ArrowUp' || e.key === 'ArrowDown')) {
											const to = e.key === 'ArrowUp' ? i - 1 : i + 1;
											if (move(i, to)) focusAt.current = to;
										} else if (e.key === 'Enter' || e.key === ' ') toggle(rowKey);
										else return;
										e.preventDefault();
									}}>
									<div className="summary-main">
										{ctx.isPrepull ? <DoAt json={row.doAtValue} /> : null}
										<ActionText json={row.action} path={[...rowPath, 'action']} />
									</div>
									{own ? (
										own.length > 0 ? (
											<div className="summary-cond">
												<span className="op">if</span>{' '}
												{joined(
													own.map((c, k) => (
														<ValueText
															json={c}
															parent={2}
															path={[...rowPath, 'action', 'condition', 'and', 'vals', info!.inherited + k]}
														/>
													)),
													<>
														{' '}
														<span className="op">and</span>{' '}
													</>,
												)}
											</div>
										) : null
									) : condition !== undefined ? (
										<div className="summary-cond">
											<span className="op">if</span> <ValueText json={condition} path={[...rowPath, 'action', 'condition']} />
										</div>
									) : null}
									{typeof row.notes === 'string' && row.notes ? <div className="summary-notes">{row.notes}</div> : null}
								</div>
								<RowFindings rowKey={rowKey} />
								<RowCount rowKey={rowKey} />
								<span className="row-buttons">
									<button
										type="button"
										className="icon-btn"
										title={hidden ? 'Show (the sim runs it)' : 'Hide (the sim skips it)'}
										onClick={() => edit({ op: 'set', path: [...rowPath, 'hide'], value: hidden ? undefined : true })}>
										<Icon name={hidden ? 'eyeOff' : 'eye'} />
									</button>
									<button
										type="button"
										className="icon-btn"
										title="Duplicate"
										onClick={() => edit({ op: 'insert', path, index: i + 1, value: item })}>
										<Icon name="copy" />
									</button>
									<button
										type="button"
										className="icon-btn"
										title="Delete"
										onClick={() => edit({ op: 'set', path: rowPath, value: undefined })}>
										<Icon name="trash" />
									</button>
								</span>
							</div>
							{isOpen ? (
								<div className="row-body">
									{ctx.isPrepull ? (
										<div className="field wide">
											<label title="When to do it; a negative time before the pull">Do at</label>
											<div className="field-input">
												<ValueEditor json={row.doAtValue} path={[...rowPath, 'doAtValue']} ctx={{ ...ctx, isPrepull: false }} />
											</div>
										</div>
									) : null}
									<ActionEditor json={row.action} path={[...rowPath, 'action']} ctx={ctx} inherited={info ? joinedBy(around) : undefined} />
									{ctx.isPrepull ? null : (
										<div className="field wide">
											<label>Notes</label>
											<div className="field-input">
												<CommitInput
													value={typeof row.notes === 'string' ? row.notes : ''}
													placeholder="A note for the reader"
													onCommit={v => edit({ op: 'set', path: [...rowPath, 'notes'], value: v || undefined })}
												/>
											</div>
										</div>
									)}
								</div>
							) : null}
						</li>,
					];
				})}
				{heads(items.length)}
			</ol>
			<div className="list-adds">
				<KindPicker family="action" isPrepull={ctx.isPrepull} isGroup={ctx.isGroup} onPick={add} trigger={<span className="add">+ Add action</span>} />
				<button
					type="button"
					className="add"
					title="Insert a common pattern: refresh a DoT, keep a debuff up, an execute block…"
					onClick={() => vscode.postMessage({ type: 'pattern', path: [...path, items.length - 1] })}>
					+ pattern
				</button>
			</div>
		</section>
	);
}

// The file's defs: each one's name, which renames its calls too, its parameters and its value, as SIML.
function Defs() {
	const { edit, revealRange } = useEditor();
	const { defs, defRanges } = useStructure();
	const params = (text: string) =>
		text
			.split(',')
			.map(p => p.trim())
			.filter(Boolean);
	return (
		<section className="defs">
			<h2>
				Defs <span className="count">{defs.length}</span>
			</h2>
			{defs.map((def, i) => (
				<div className="def-line" key={def.name} data-def={def.name}>
					<span className="op">def</span>
					<CommitInput
						className="def-name-input"
						value={def.name}
						onCommit={name => name.trim() && edit({ op: 'defRename', name: def.name, text: name.trim() })}
					/>
					<span className="op">(</span>
					<CommitInput
						className="def-params"
						value={def.params.join(', ')}
						placeholder="no parameters"
						onCommit={v => edit({ op: 'defSet', name: def.name, params: params(v), text: def.value })}
					/>
					<span className="op">) =</span>
					<CommitInput className="def-value" value={def.value} onCommit={v => edit({ op: 'defSet', name: def.name, params: def.params, text: v })} />
					{def.used ? null : (
						<span className="badge muted" title="No row, var, when line or def calls it">
							unused
						</span>
					)}
					<button type="button" className="icon-btn" title="Jump to the text" onClick={() => revealRange(defRanges[i])}>
						<Icon name="jump" />
					</button>
					<button
						type="button"
						className="icon-btn"
						title="Delete the def; each call is written as the value it reads as"
						onClick={() => edit({ op: 'defDelete', name: def.name })}>
						<Icon name="trash" />
					</button>
				</div>
			))}
			<button
				type="button"
				className="add"
				title="A def names a value, which its calls read as"
				onClick={() =>
					edit({
						op: 'defInsert',
						name: freeName(
							'def',
							defs.map(d => d.name),
						),
						text: 'true',
					})
				}>
				+ def
			</button>
		</section>
	);
}

function Variables({ path, variables, ctx, title }: { path: JsonPath; variables: Json[]; ctx: ListContext; title: string }) {
	const { edit, reveal } = useEditor();
	return (
		<div className="variables">
			<h3>{title}</h3>
			{variables.map((variable, i) => {
				const v = isObject(variable) ? variable : {};
				return (
					<div className="variable" key={i} data-row={key([...path, i])}>
						<span className="var-name" onClick={() => reveal([...path, i])}>
							$<CommitInput value={String(v.name ?? '')} onCommit={name => edit({ op: 'set', path: [...path, i, 'name'], value: name })} />
						</span>
						<span className="op">=</span>
						<ValueEditor json={v.value} path={[...path, i, 'value']} ctx={ctx} />
						<RowFindings rowKey={key([...path, i])} />
						<button type="button" className="icon-btn" title="Delete" onClick={() => edit({ op: 'set', path: [...path, i], value: undefined })}>
							<Icon name="trash" />
						</button>
					</div>
				);
			})}
			<button
				type="button"
				className="add"
				onClick={() =>
					edit({
						op: 'insert',
						path,
						index: variables.length,
						value: {
							name: freeName(
								'var',
								variables.map(v => (isObject(v) ? String(v.name ?? '') : '')),
							),
							value: { const: { val: '0' } },
						},
					})
				}>
				+ variable
			</button>
		</div>
	);
}

function RegisteredBadge() {
	const registered = useContext(RegisteredContext);
	if (registered.status === 'loading') return null;
	if (registered.status === 'unavailable')
		return (
			<span className="badge" title={registered.reason}>
				no names to offer: {registered.reason}
			</span>
		);
	return <span className="badge muted">names {registered.spec} registers</span>;
}

// The first `<prefix>N` no name in `taken` has.
function freeName(prefix: string, taken: string[]): string {
	for (let n = 1; ; n++) if (!taken.includes(`${prefix}${n}`)) return `${prefix}${n}`;
}

interface AppProps {
	db: AplDatabase;
	fileName: string;
	rotation: JsonObject;
	readOnly?: string;
	cursor: JsonPath; // The rotation path at the text cursor.
	send: (edits: AplEdit[]) => void;
}

function App({ db, fileName, rotation, readOnly, cursor, send }: AppProps) {
	const [expanded, setExpanded] = useState<Set<string>>(new Set());

	const groups = useMemo(() => (Array.isArray(rotation.groups) ? rotation.groups.filter(isObject) : []), [rotation]);
	const variables = useMemo(() => (Array.isArray(rotation.valueVariables) ? rotation.valueVariables : []), [rotation]);

	// Every edit of the panel: open rows follow it, and it is made on the panel's rotation and sent to the extension.
	// Nothing is sent while the panel is read-only.
	const edit = useCallback(
		(...edits: AplEdit[]) => {
			if (readOnly || edits.length === 0) return false;
			setExpanded(current => edits.reduce<Set<string>>(shiftKeys, current));
			send(edits);
			return true;
		},
		[readOnly, send],
	);

	const api: EditorApi = useMemo(
		() => ({
			edit,
			reveal: path => vscode.postMessage({ type: 'reveal', path }),
			revealRange: range => vscode.postMessage({ type: 'revealRange', range }),
			groupNames: groups.map(g => String(g.name ?? '')),
			variableNames: variables.filter(isObject).map(v => String(v.name ?? '')),
			usedIds: [...collectIds(rotation, new Set())],
			groupVariables: Object.fromEntries(
				groups.map(g => [String(g.name ?? ''), (Array.isArray(g.variables) ? g.variables : []).filter(isObject).map(v => String(v.name ?? ''))]),
			),
		}),
		[edit, rotation, groups, variables],
	);

	// The row the text cursor is in: a list item is its list, its index, and for a group's list two more steps.
	const selected = useMemo(() => {
		const rowLength = cursor[0] === 'groups' ? 4 : 2;
		return cursor.length >= rowLength ? key(cursor.slice(0, rowLength)) : undefined;
	}, [cursor]);

	useEffect(() => {
		if (selected) document.querySelector(`[data-row="${CSS.escape(selected)}"]`)?.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
	}, [selected]);

	const toggle = useCallback((rowKey: string) => {
		setExpanded(current => {
			const next = new Set(current);
			if (!next.delete(rowKey)) next.add(rowKey);
			return next;
		});
	}, []);

	const open = useCallback((rowKey: string) => setExpanded(current => new Set(current).add(rowKey)), []);

	const type = typeof rotation.type === 'string' ? rotation.type : 'TypeUnknown';
	const listProps = { selected, expanded, toggle, open };

	return (
		<DbContext.Provider value={db}>
			<EditorContext.Provider value={api}>
				<datalist id="apl-group-names">
					{api.groupNames.map(name => (
						<option key={name} value={name} />
					))}
				</datalist>
				<datalist id="apl-variable-names">
					{api.variableNames.map(name => (
						<option key={name} value={name} />
					))}
				</datalist>
				<header>
					<h1>{fileName}</h1>
					<RegisteredBadge />
					<FindingsBadge />
					<SimBar />
					{type !== 'TypeAPL' ? <span className="badge">{type}: the sim only runs the lists below for TypeAPL</span> : null}
				</header>
				{readOnly ? <div className="error">{readOnly}</div> : null}
				<fieldset disabled={!!readOnly}>
					<ItemList
						title="Prepull"
						path={['prepullActions']}
						items={Array.isArray(rotation.prepullActions) ? rotation.prepullActions : []}
						ctx={{ isPrepull: true, isGroup: false }}
						{...listProps}
					/>
					<ItemList
						title="Priority list"
						path={['priorityList']}
						items={Array.isArray(rotation.priorityList) ? rotation.priorityList : []}
						ctx={{ isPrepull: false, isGroup: false }}
						{...listProps}
					/>
					<section>
						<h2>
							Groups <span className="count">{groups.length}</span>
						</h2>
						{groups.map((group, g) => (
							<div className="group" key={g}>
								<div className="group-head">
									<CommitInput
										className="group-name"
										value={String(group.name ?? '')}
										onCommit={name => api.edit({ op: 'set', path: ['groups', g, 'name'], value: name })}
									/>
									<RowFindings rowKey={key(['groups', g, 'name'])} />
									<button type="button" className="icon-btn" title="Jump to the text" onClick={() => api.reveal(['groups', g])}>
										<Icon name="jump" />
									</button>
									<button
										type="button"
										className="icon-btn"
										title="Delete the group"
										onClick={() => api.edit({ op: 'set', path: ['groups', g], value: undefined })}>
										<Icon name="trash" />
									</button>
								</div>
								{Array.isArray(group.variables) && group.variables.length > 0 ? (
									<Variables
										title="Placeholders"
										path={['groups', g, 'variables']}
										variables={group.variables}
										ctx={{ isPrepull: false, isGroup: true }}
									/>
								) : null}
								<ItemList
									title="Actions"
									path={['groups', g, 'actions']}
									items={Array.isArray(group.actions) ? group.actions : []}
									ctx={{ isPrepull: false, isGroup: true }}
									{...listProps}
								/>
							</div>
						))}
						<button
							type="button"
							className="add"
							onClick={() =>
								api.edit({
									op: 'insert',
									path: ['groups'],
									index: groups.length,
									value: { name: freeName('Group ', api.groupNames), actions: [] },
								})
							}>
							+ group
						</button>
					</section>
					<section>
						<Variables title="Variables" path={['valueVariables']} variables={variables} ctx={{ isPrepull: false, isGroup: false }} />
					</section>
					<Defs />
					<SimDetails />
				</fieldset>
			</EditorContext.Provider>
		</DbContext.Provider>
	);
}

// The rotation the extension last read, and what it says about it.
interface Doc {
	version: number;
	rotation: JsonObject;
	readOnly?: string;
	findings: Finding[];
	structure: AplStructure;
}

function Root() {
	const [init, setInit] = useState<Extract<ToPanel, { type: 'init' }>>();
	const [doc, setDoc] = useState<Doc>();
	// The panel's edits the extension has not answered yet: the panel shows them made on the rotation it last read.
	const [pending, setPending] = useState<Array<{ seq: number; edits: AplEdit[] }>>([]);
	const seq = useRef(0);
	const [registered, setRegistered] = useState<Registered>({ status: 'loading' });
	const [sim, setSim] = useState<SimState>({ type: 'sim' });
	const [cursor, setCursor] = useState<JsonPath>([]);

	useEffect(() => {
		const onMessage = (event: MessageEvent<ToPanel>) => {
			const message = event.data;
			if (message.type === 'init') {
				vscode.setState({ uri: message.uri });
				setInit(message);
			} else if (message.type === 'rotation') {
				if (message.className) setClassName(message.className);
				setDoc(current => ({
					version: message.version,
					rotation: message.rotation ?? current?.rotation ?? {},
					readOnly: message.readOnly,
					findings: message.findings,
					structure: message.structure ?? current?.structure ?? emptyStructure,
				}));
				setPending(current => current.filter(p => p.seq > message.handled));
			} else if (message.type === 'registered') {
				setRegistered(message.registered);
			} else if (message.type === 'sim') {
				setSim(message);
			} else if (message.type === 'cursor') {
				setCursor(message.path);
			}
		};
		window.addEventListener('message', onMessage);
		vscode.postMessage({ type: 'ready' });
		return () => window.removeEventListener('message', onMessage);
	}, []);

	const version = doc?.version;
	const send = useCallback(
		(edits: AplEdit[]) => {
			if (version === undefined) return;
			const next = ++seq.current;
			setPending(current => [...current, { seq: next, edits }]);
			vscode.postMessage({ type: 'edit', version, seq: next, edits });
		},
		[version],
	);
	const rotation = useMemo(() => doc && pending.reduce((r, p) => p.edits.reduce(applyEdit, r), doc.rotation), [doc, pending]);
	const findings = useMemo(() => indexFindings(doc?.findings ?? []), [doc?.findings]);

	// The table's names fill in the spells the database has no entry for, with no icon.
	const db = useMemo(() => {
		if (!init || registered.status !== 'ready') return init?.db;
		const { spells, auras, targetAuras, petAuras, targetSpells } = registered.registered;
		const named = Object.fromEntries(
			[...spells, ...auras, ...targetAuras, ...petAuras, ...targetSpells].flatMap((entry): Array<[number, Named]> =>
				entry.id.spellId === undefined ? [] : [[entry.id.spellId, [entry.name, '']]],
			),
		);
		return { ...init.db, spells: { ...named, ...init.db.spells } };
	}, [init, registered]);

	if (!init || !doc || !rotation || !db) {
		return <div className="loading">Loading…</div>;
	}
	return (
		<RegisteredContext.Provider value={registered}>
			<FindingsContext.Provider value={findings}>
				<StructureContext.Provider value={doc.structure}>
					<SimContext.Provider value={sim}>
						<App db={db} fileName={init.fileName} rotation={rotation} readOnly={doc.readOnly} cursor={cursor} send={send} />
					</SimContext.Provider>
				</StructureContext.Provider>
			</FindingsContext.Provider>
		</RegisteredContext.Provider>
	);
}

createRoot(document.getElementById('root')!).render(<Root />);
