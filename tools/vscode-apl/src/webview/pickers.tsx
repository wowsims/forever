import type { ACTION_ID_SET } from '@features/apl/model/action_id_sets';
import type { UNIT_SET } from '@features/apl/model/unit_sets';
import { type KeyboardEvent, type ReactNode, useEffect, useMemo, useRef, useState } from 'react';

import type { AplDatabase } from '../protocol';
import { CommitInput } from './commit_input';
import { iconUrl, IdChip, useDb } from './db';
import { useEditor } from './editor_context';
import {
	enumChoices,
	enumJsonName,
	enumNumber,
	type Family,
	type fieldInfo,
	isObject,
	type Json,
	type JsonObject,
	kindsFor,
	submenuLabel,
	unitChoices,
} from './model';
import { idKey, offeredFor, useRegistered } from './registered';
import { unitText } from './summary';

export interface PickerOption {
	key: string;
	label: string;
	group?: string;
	hint?: string;
	icon?: string;
	value: Json;
}

const MAX_SHOWN = 80;

// A text box that filters a long option list: arrows move, Enter picks, Escape closes.
export function SearchPicker({
	trigger,
	options,
	onPick,
	placeholder,
	note,
}: {
	trigger: ReactNode;
	options: () => PickerOption[];
	onPick: (value: Json) => void;
	placeholder: string;
	note?: string;
}) {
	const [open, setOpen] = useState(false);
	const [query, setQuery] = useState('');
	const [active, setActive] = useState(0);
	const input = useRef<HTMLInputElement>(null);
	const box = useRef<HTMLDivElement>(null);

	const all = useMemo(() => (open ? options() : []), [open, options]);
	const shown = useMemo(() => {
		const words = query.toLowerCase().split(/\s+/).filter(Boolean);
		const hits = words.length === 0 ? all : all.filter(o => words.every(w => `${o.label} ${o.hint ?? ''} ${o.group ?? ''}`.toLowerCase().includes(w)));
		return hits.slice(0, MAX_SHOWN);
	}, [all, query]);

	useEffect(() => {
		if (!open) return;
		input.current?.focus();
		const close = (event: MouseEvent) => {
			if (!box.current?.contains(event.target as Node)) setOpen(false);
		};
		document.addEventListener('mousedown', close);
		return () => document.removeEventListener('mousedown', close);
	}, [open]);

	useEffect(() => setActive(0), [query]);

	const pick = (option: PickerOption | undefined) => {
		if (!option) return;
		setOpen(false);
		setQuery('');
		onPick(option.value);
	};

	const onKey = (event: KeyboardEvent) => {
		if (event.key === 'ArrowDown') setActive(i => Math.min(i + 1, shown.length - 1));
		else if (event.key === 'ArrowUp') setActive(i => Math.max(i - 1, 0));
		else if (event.key === 'Enter') pick(shown[active]);
		else if (event.key === 'Escape') setOpen(false);
		else return;
		event.preventDefault();
		event.stopPropagation();
	};

	return (
		<div className="picker" ref={box}>
			<button type="button" className="picker-trigger" onClick={() => setOpen(!open)}>
				{trigger}
				<span className="caret">▾</span>
			</button>
			{open ? (
				<div className="picker-pop">
					<input ref={input} value={query} placeholder={placeholder} onChange={e => setQuery(e.target.value)} onKeyDown={onKey} />
					{note ? <div className="picker-note">{note}</div> : null}
					<ul>
						{shown.map((option, i) => (
							<li
								key={option.key}
								className={i === active ? 'active' : undefined}
								onMouseEnter={() => setActive(i)}
								onMouseDown={e => {
									e.preventDefault();
									pick(option);
								}}
								title={option.hint}>
								{option.icon ? <img src={iconUrl(option.icon)} alt="" /> : null}
								<span className="opt-label">{option.label}</span>
								{option.group ? <span className="opt-group">{option.group}</span> : null}
							</li>
						))}
						{shown.length === 0 ? <li className="empty">No match</li> : null}
					</ul>
				</div>
			) : null}
		</div>
	);
}

// An id as a picker option, with the game name and icon the database gives it, and its name in the table.
function dbOption(db: AplDatabase, id: JsonObject, group: string, named: string, hint?: string): PickerOption {
	const kind = typeof id.itemId === 'number' ? 'item' : typeof id.otherId === 'string' ? 'other' : 'spell';
	const number = kind === 'item' ? id.itemId : id.spellId;
	const [name, icon] = kind === 'other' ? [named, undefined] : ((kind === 'item' ? db.items : db.spells)[number as number] ?? [named, undefined]);
	const tag = typeof id.tag === 'number' && id.tag ? ` · tag ${id.tag}` : '';
	return { key: `${idKey(id)}`, label: name, hint: hint ?? `${named} · ${kind} ${kind === 'other' ? id.otherId : number}${tag}`, icon, group, value: id };
}

// Names that more than one option shares, which then carry their rank, or their id where they have none.
function withRanks(options: PickerOption[], ranks: Map<string, number>): PickerOption[] {
	const counts = new Map<string, number>();
	options.forEach(option => counts.set(option.label, (counts.get(option.label) ?? 0) + 1));
	return options.map(option => {
		const rank = ranks.get(option.key);
		if ((counts.get(option.label) ?? 0) < 2) return option;
		return { ...option, label: rank ? `${option.label} (Rank ${rank})` : `${option.label} · ${option.hint}` };
	});
}

// What the field may hold: the names of sim/siml/names.json the spec registers in the namespace the field reads, and
// for the sets that take one, the items with a use. Each option writes its ActionID as the table holds it, tag and rank
// included, so the text names it.
export function ActionIdPicker({
	json,
	onPick,
	family,
	kind,
	field,
	set,
	unit,
}: {
	json: Json | undefined;
	onPick: (value: Json) => void;
	family: Family;
	kind: string;
	field: string;
	set: ACTION_ID_SET;
	unit: Json | undefined;
}) {
	const db = useDb();
	const registered = useRegistered();
	const { usedIds } = useEditor();
	const options = useMemo(
		() => () => {
			if (registered.status !== 'ready') return [];
			const used = new Set(usedIds);
			const first = (list: PickerOption[]) => [...list.filter(o => used.has(o.key)), ...list.filter(o => !used.has(o.key))];
			const offered = offeredFor(registered, { family, kind, name: field, set }, unit);
			const ranks = new Map(offered.map(entry => [`${idKey(entry.id)}`, entry.id.rank ?? 0]));
			const named = withRanks(
				offered.map(entry => dbOption(db, entry.id as JsonObject, entry.id.itemId === undefined ? registered.spec : 'item', entry.name)),
				ranks,
			);
			const placeholders =
				set === 'castable_spells'
					? [
							dbOption(
								db,
								{ otherId: 'OtherActionOffensivePotion' },
								'placeholder',
								'other.offensive_potion',
								'the offensive potion the player has',
							),
							dbOption(
								db,
								{ otherId: 'OtherActionDefensivePotion' },
								'placeholder',
								'other.defensive_potion',
								'the defensive potion the player has',
							),
							dbOption(
								db,
								{ otherId: 'OtherActionResourcePotion' },
								'placeholder',
								'other.resource_potion',
								'the resource potion the player has',
							),
						]
					: [];
			return first([...named, ...placeholders]);
		},
		[db, usedIds, registered, family, kind, field, set, unit],
	);
	const note =
		registered.status === 'ready' ? undefined : `Nothing to offer: ${registered.status === 'loading' ? 'the names are loading' : registered.reason}.`;
	return (
		<span className="id-picker">
			<SearchPicker
				trigger={<IdChip json={json} />}
				options={options}
				note={note}
				onPick={onPick}
				placeholder={registered.status === 'ready' ? `Search what ${registered.spec} registers` : 'Search names'}
			/>
		</span>
	);
}

// The kind list of an action or value, filed under the site's submenus and searchable by name and description.
export function KindPicker({
	family,
	current,
	onPick,
	isPrepull,
	isGroup,
	trigger,
}: {
	family: Family;
	current?: string;
	onPick: (kind: string) => void;
	isPrepull: boolean;
	isGroup: boolean;
	trigger: ReactNode;
}) {
	const options = useMemo(
		() => () =>
			kindsFor(family, isPrepull, isGroup).map((model): PickerOption => ({
				key: model.kind,
				label: model.label,
				group: model.submenu?.map(submenuLabel).join(' › '),
				hint: model.shortDescription,
				value: model.kind,
			})),
		[family, isPrepull, isGroup],
	);
	return (
		<SearchPicker trigger={trigger} options={options} onPick={kind => kind !== current && onPick(kind as string)} placeholder={`Search ${family} kinds`} />
	);
}

const INDEXED = new Set(['Target', 'Pet', 'Player']);

// A unit reference: the kind's default (no reference), or one of the units the field's set offers and, for the
// indexed types, which one. The index is written on Enter or when the box loses focus.
export function UnitPicker({ json, set, onPick }: { json: Json | undefined; set: UNIT_SET; onPick: (value: Json | undefined) => void }) {
	const type = isObject(json) && typeof json.type === 'string' ? json.type : '';
	const index = isObject(json) && typeof json.index === 'number' ? json.index : 0;
	const choices = useMemo(() => unitChoices(set), [set]);
	const offered = choices.some(choice => choice.type === type);
	return (
		<span className="unit-picker">
			<select
				value={type}
				onChange={e => {
					const choice = choices.find(c => c.type === e.target.value);
					onPick(choice ? { ...choice, ...(INDEXED.has(e.target.value) && index ? { index } : {}) } : undefined);
				}}>
				<option value="">default</option>
				{choices.map(choice => (
					<option key={String(choice.type)} value={String(choice.type)}>
						{unitText({ type: choice.type }) ?? String(choice.type)}
					</option>
				))}
				{type && !offered ? <option value={type}>{unitText(json) ?? type}</option> : null}
			</select>
			{INDEXED.has(type) ? (
				<CommitInput
					type="number"
					className="narrow"
					value={String(index + 1)}
					onCommit={v => {
						const n = Math.max(0, Math.trunc(Number(v)) - 1 || 0);
						const { index: _index, ...rest } = json as JsonObject;
						onPick(n ? { ...rest, index: n } : rest);
					}}
				/>
			) : null}
		</span>
	);
}

export function EnumSelect({
	info,
	json,
	table,
	onPick,
}: {
	info: ReturnType<typeof fieldInfo>;
	json: Json | undefined;
	table?: { options: Array<{ value: number; label: string; tooltip?: string }> };
	onPick: (value: Json | undefined) => void;
}) {
	const current = enumNumber(info, json) ?? 0;
	const options: Array<{ value: number; label: string; tooltip?: string }> = table?.options ?? enumChoices(info);
	return (
		<select
			value={current}
			onChange={e => {
				const value = Number(e.target.value);
				// Proto JSON leaves a zero enum out.
				onPick(value === 0 ? undefined : (enumJsonName(info, value) as Json));
			}}>
			{options.some(o => o.value === current) ? null : <option value={current}>{String(json ?? current)}</option>}
			{options.map(option => (
				<option key={option.value} value={option.value} title={option.tooltip}>
					{option.label}
				</option>
			))}
		</select>
	);
}
