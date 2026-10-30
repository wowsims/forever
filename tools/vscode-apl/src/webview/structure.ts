import { createContext, useContext } from 'react';

import type { AplStructure, WhenBlock } from '../protocol';
import { isObject, type Json, type JsonPath } from './model';

// The when blocks, def calls and defs the server reads from the text (sim/siml/structure.go), which the rotation's
// JSON does not hold.

export const emptyStructure: AplStructure = { whens: [], rows: [], calls: [], defs: [], whenRanges: [], defRanges: [] };

export const StructureContext = createContext<AplStructure>(emptyStructure);

export const useStructure = () => useContext(StructureContext);

export const key = (path: JsonPath) => path.join('/');

export type DefCall = AplStructure['calls'][number];
export type RowInfo = AplStructure['rows'][number];

// The rows and def calls of each server reply by their joined paths, built on first use: every row and value of a
// render looks itself up.
const indexes = new WeakMap<AplStructure, { rows: Map<string, RowInfo>; calls: Map<string, DefCall> }>();
function indexOf(structure: AplStructure) {
	let index = indexes.get(structure);
	if (!index) {
		index = { rows: new Map(structure.rows.map(row => [key(row.path), row])), calls: new Map(structure.calls.map(call => [key(call.path), call])) };
		indexes.set(structure, index);
	}
	return index;
}

// The def call a value at `path` is written as.
export function callAt(structure: AplStructure, path: JsonPath | undefined): DefCall | undefined {
	return path ? indexOf(structure).calls.get(key(path)) : undefined;
}

// What the server says of the row at `path`: the conjuncts its blocks join to it.
export const rowAt = (structure: AplStructure, path: JsonPath): RowInfo | undefined => indexOf(structure).rows.get(key(path));

export const callText = (name: string, args: string[]) => (args.length > 0 ? `${name}(${args.join(', ')})` : name);

// The blocks of the list at `list`, each with its depth: 0 for one in no other block.
export function blocksIn(structure: AplStructure, list: JsonPath): Array<WhenBlock & { depth: number }> {
	const at = key(list);
	const depth = (block: WhenBlock): number => (block.parent < 0 ? 0 : 1 + depth(structure.whens[block.parent]));
	return structure.whens.filter(block => key(block.list) === at).map(block => ({ ...block, depth: depth(block) }));
}

// The blocks row `index` of a list is in, the outermost first.
export function blocksAt(blocks: Array<WhenBlock & { depth: number }>, index: number): Array<WhenBlock & { depth: number }> {
	return blocks.filter(block => block.first <= index && index <= block.last).sort((a, b) => a.depth - b.depth);
}

// What a row's blocks join to its condition, in order.
export const joinedBy = (blocks: WhenBlock[]): Json[] => blocks.flatMap(block => block.conjuncts);

// A condition of conjuncts: none, the one alone, or their and.
export function conditionOf(conjuncts: Json[]): Json | undefined {
	if (conjuncts.length === 0) return undefined;
	if (conjuncts.length === 1) return conjuncts[0];
	return { and: { vals: conjuncts } };
}

// A row's own conjuncts after the `inherited` its blocks join, with the index of each in its condition's and. A row
// in no block has its condition's and values as its own, or the condition alone.
export function ownConjuncts(condition: Json | undefined, inherited: number, own = true): Json[] {
	if (condition === undefined || !own) return [];
	const vals = isObject(condition) && isObject(condition.and) && Array.isArray(condition.and.vals) ? condition.and.vals : undefined;
	if (inherited === 0) return vals ?? [condition];
	return vals ? vals.slice(inherited) : [];
}
