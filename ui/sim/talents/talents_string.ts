// The talent-string codec, as a pure function over the config.
//
// A talents string is one run of per-talent digits per tree, joined with '-': digit `i` of tree `t` is
// `config[t].talents[i]`, in the order the config lists them, NOT a row index. Both the per-tree
// trailing zeroes and the trailing dashes are trimmed, which is what lets a string carry an empty
// middle tree ('2500052300030150330125--053500031003001').

import type { TalentsConfig } from '@sim/talents/config';

/** Points spent per talent, shaped exactly like the config: `[treeIdx][talentIdx]`. */
export type TalentPoints = Array<Array<number>>;

const clampPoints = (raw: number, maxPoints: number): number => {
	if (!Number.isFinite(raw)) return 0;
	return Math.min(maxPoints, Math.max(0, raw));
};

export const parseTalentsString = <TalentsProto>(config: TalentsConfig<TalentsProto>, talentsString: string): TalentPoints => {
	const treeStrings = talentsString.split('-');
	return config.map((treeConfig, treeIdx) => {
		const treeString = treeStrings[treeIdx] ?? '';
		return treeConfig.talents.map((talent, talentIdx) => clampPoints(Number(treeString.charAt(talentIdx)), talent.maxPoints));
	});
};

/**
 * No more trees than the config has, no tree longer than its talents and no digit above its talent's maximum. A shorter
 * tree is fine, since trailing zeroes are trimmed.
 */
export const talentsStringFits = <TalentsProto>(config: TalentsConfig<TalentsProto>, talentsString: string): boolean => {
	const treeStrings = talentsString.split('-');
	return (
		treeStrings.length <= config.length &&
		treeStrings.every(
			(treeString, treeIdx) =>
				treeString.length <= config[treeIdx].talents.length &&
				[...treeString].every((digit, talentIdx) => /^\d$/.test(digit) && Number(digit) <= config[treeIdx].talents[talentIdx].maxPoints),
		)
	);
};

export const serializeTalentsString = (points: TalentPoints): string =>
	points
		.map(tree => tree.join('').replace(/0+$/g, ''))
		.join('-')
		.replace(/-+$/g, '');

export const withTalentPoints = (points: TalentPoints, treeIdx: number, talentIdx: number, newPoints: number): TalentPoints =>
	points.map((tree, idx) => (idx === treeIdx ? tree.map((value, i) => (i === talentIdx ? newPoints : value)) : tree));

export const withTreeCleared = (points: TalentPoints, treeIdx: number): TalentPoints => points.map((tree, idx) => (idx === treeIdx ? tree.map(() => 0) : tree));

export const treePointTotal = (points: TalentPoints, treeIdx: number): number => points[treeIdx].reduce((total, value) => total + value, 0);

export const totalPointsSpent = (points: TalentPoints): number => points.reduce((total, tree) => total + tree.reduce((sub, value) => sub + value, 0), 0);
