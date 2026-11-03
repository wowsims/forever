import type { TalentsConfig } from '@sim/talents/config';
import { newTalentsConfig } from '@sim/talents/config';
import { mageTalentsConfig } from '@sim/talents/mage';
import { rogueTalentsConfig } from '@sim/talents/rogue';
import { describe, expect, it } from 'vitest';

import { buildTalentGraph } from './tree_graph';

type Fake = Record<string, never>;

const talent = (rowIdx: number, colIdx: number, prereqLocation?: { rowIdx: number; colIdx: number }) => ({
	fieldName: `t${rowIdx}${colIdx}`,
	fancyName: `Talent ${rowIdx}/${colIdx}`,
	location: { rowIdx, colIdx },
	spellId: 1,
	maxPoints: 1,
	...(prereqLocation ? { prereqLocation } : {}),
});

const treeOf = (talents: ReturnType<typeof talent>[]) => ({ name: 'A', backgroundUrl: 'background.jpg', talents });

const graphOf = (talents: ReturnType<typeof talent>[]) => buildTalentGraph(newTalentsConfig([treeOf(talents)] as TalentsConfig<Fake>)).trees[0];

describe('buildTalentGraph arrows', () => {
	it('spans a straight vertical prerequisite as a down arrow', () => {
		const [arrow] = graphOf([talent(0, 0), talent(2, 0, { rowIdx: 0, colIdx: 0 })]).arrows;
		expect(arrow).toMatchObject({ dir: 'down', rowSize: 2, gridRow: 1, gridColumn: 1, gridRowEnd: 4, gridColumnEnd: 1 });
		expect(arrow.colSize).toBeUndefined();
	});

	it('points right when the child sits further along the same row', () => {
		const [arrow] = graphOf([talent(1, 0), talent(1, 1, { rowIdx: 1, colIdx: 0 })]).arrows;
		expect(arrow).toMatchObject({ dir: 'right', colSize: 1, gridRow: 2, gridColumn: 1, gridRowEnd: 2, gridColumnEnd: 1 });
	});

	it('points left when the child sits back along the same row', () => {
		const [arrow] = graphOf([talent(1, 0, { rowIdx: 1, colIdx: 1 }), talent(1, 1)]).arrows;
		expect(arrow).toMatchObject({ dir: 'left', colSize: 1, gridRow: 2, gridColumn: 2, gridRowEnd: 2, gridColumnEnd: 3 });
	});

	it('uses the two-segment diagonals for a child that is both down and across', () => {
		const rightdown = graphOf([talent(0, 0), talent(1, 2, { rowIdx: 0, colIdx: 0 })]).arrows[0];
		expect(rightdown).toMatchObject({ dir: 'rightdown', colSize: 2, rowSize: 1, gridRowEnd: 3, gridColumnEnd: 4 });

		const leftdown = graphOf([talent(0, 2), talent(1, 0, { rowIdx: 0, colIdx: 2 })]).arrows[0];
		expect(leftdown).toMatchObject({ dir: 'leftdown', colSize: 2, rowSize: 1, gridRowEnd: 3, gridColumnEnd: 2 });
	});

	it('lowers each generation by two and paints an arrow one below its parent', () => {
		const graph = graphOf([talent(0, 0), talent(1, 0, { rowIdx: 0, colIdx: 0 }), talent(2, 0, { rowIdx: 1, colIdx: 0 })]);
		expect(graph.zIndex).toEqual([20, 18, 16]);
		expect(graph.arrows.map(arrow => arrow.zIndex)).toEqual([19, 17]);
	});

	it('leaves a talent outside every chain without a z-index', () => {
		const graph = graphOf([talent(0, 0), talent(0, 1)]);
		expect(graph.zIndex).toEqual([undefined, undefined]);
		expect(graph.arrows).toHaveLength(0);
	});
});

describe('buildTalentGraph dimensions', () => {
	it('keeps the Rogue tree to four columns', () => {
		const graph = buildTalentGraph(rogueTalentsConfig);

		expect(graph.numCols).toBe(4);
		expect(rogueTalentsConfig[2].talents.find(talent => talent.fieldName === 'setup')?.location).toEqual({ rowIdx: 0, colIdx: 0 });
	});

	it('sizes every tree of a class off the largest tree', () => {
		const graph = buildTalentGraph(mageTalentsConfig);
		const locations = mageTalentsConfig.flatMap(tree => tree.talents.map(t => t.location));
		expect(graph.numRows).toBe(Math.max(...locations.map(l => l.rowIdx)) + 1);
		expect(graph.numCols).toBe(Math.max(...locations.map(l => l.colIdx)) + 1);
		expect(graph.trees).toHaveLength(3);
	});

	it('resolves every prereqLocation in the real mage config to a talent index', () => {
		const graph = buildTalentGraph(mageTalentsConfig);
		graph.trees.forEach((tree, treeIdx) => {
			mageTalentsConfig[treeIdx].talents.forEach((t, idx) => {
				if (t.prereqLocation) expect(tree.prereqIdx[idx]).toBeGreaterThanOrEqual(0);
				else expect(tree.prereqIdx[idx]).toBe(-1);
			});
		});
	});
});
