// A tree is ONE css grid and every talent is placed by gridRow/gridColumn, because a
// prerequisite arrow is itself a grid item spanning from the parent's cell to the child's. The
// z-index ladder exists so a deeper chain paints under the one above it.

import type { TalentConfig, TalentLocation, TalentsConfig, TalentTreeConfig } from '@sim/talents/config';
import { locationKey } from '@sim/talents/config';

export type ReqDir = 'down' | 'right' | 'left' | 'rightdown' | 'leftdown';

export interface TalentArrow {
	parentIdx: number;
	childIdx: number;
	dir: ReqDir;
	rowSize?: number;
	colSize?: number;
	gridRow: number;
	gridColumn: number;
	gridRowEnd: number;
	gridColumnEnd: number;
	zIndex?: number;
}

export interface TreeGraph {
	prereqIdx: Array<number>;
	childIdxs: Array<Array<number>>;
	zIndex: Array<number | undefined>;
	arrows: Array<TalentArrow>;
}

export interface TalentGraph {
	numRows: number;
	numCols: number;
	trees: Array<TreeGraph>;
}

const ROOT_Z_INDEX = 20;

const arrowFor = (parent: TalentLocation, child: TalentLocation, parentIdx: number, childIdx: number): TalentArrow => {
	let gridRowEnd = Math.max(parent.rowIdx, child.rowIdx) + 1;
	let gridColumnEnd = Math.max(parent.colIdx, child.colIdx) + 1;
	let dir: ReqDir;
	let rowSize: number | undefined;
	let colSize: number | undefined;

	if (parent.rowIdx === child.rowIdx) {
		dir = parent.colIdx < child.colIdx ? 'right' : 'left';
		colSize = Math.abs(parent.colIdx - child.colIdx);
		gridColumnEnd = dir === 'left' ? gridColumnEnd + 1 : gridColumnEnd - 1;
	} else if (parent.colIdx === child.colIdx) {
		dir = 'down';
		rowSize = Math.abs(parent.rowIdx - child.rowIdx);
		gridRowEnd += 1;
	} else {
		dir = parent.colIdx < child.colIdx ? 'rightdown' : 'leftdown';
		colSize = Math.abs(parent.colIdx - child.colIdx);
		rowSize = Math.abs(parent.rowIdx - child.rowIdx);
		gridRowEnd += 1;
		gridColumnEnd = dir === 'rightdown' ? gridColumnEnd + 1 : gridColumnEnd - 1;
	}

	return {
		parentIdx,
		childIdx,
		dir,
		rowSize,
		colSize,
		gridRow: parent.rowIdx + 1,
		gridColumn: parent.colIdx + 1,
		gridRowEnd,
		gridColumnEnd,
	};
};

const buildTreeGraph = <TalentsProto>(treeConfig: TalentTreeConfig<TalentsProto>): TreeGraph => {
	const talents: Array<TalentConfig<TalentsProto>> = treeConfig.talents;
	const idxByLocation = new Map(talents.map((talent, idx) => [locationKey(talent.location), idx]));

	const prereqIdx = talents.map(talent => (talent.prereqLocation ? (idxByLocation.get(locationKey(talent.prereqLocation)) ?? -1) : -1));
	const childIdxs: Array<Array<number>> = talents.map(() => []);
	prereqIdx.forEach((parentIdx, childIdx) => {
		if (parentIdx >= 0) childIdxs[parentIdx].push(childIdx);
	});

	const zIndex: Array<number | undefined> = talents.map(() => undefined);
	const arrows: Array<TalentArrow> = [];
	const assign = (idx: number, z: number) => {
		zIndex[idx] = z;
		for (const childIdx of childIdxs[idx]) {
			arrows.push({ ...arrowFor(talents[idx].location, talents[childIdx].location, idx, childIdx), zIndex: z - 1 });
			assign(childIdx, z - 2);
		}
	};
	talents.forEach((talent, idx) => {
		if (childIdxs[idx].length === 0 || prereqIdx[idx] >= 0) return;
		assign(idx, ROOT_Z_INDEX);
	});

	return { prereqIdx, childIdxs, zIndex, arrows };
};

export const buildTalentGraph = <TalentsProto>(config: TalentsConfig<TalentsProto>): TalentGraph => {
	const locations = config.flatMap(tree => tree.talents.map(talent => talent.location));
	return {
		numRows: Math.max(...locations.map(location => location.rowIdx)) + 1,
		numCols: Math.max(...locations.map(location => location.colIdx)) + 1,
		trees: config.map(buildTreeGraph),
	};
};
