// Shape of a class's talent trees, plus the builder that validates a generated
// tree JSON. Pure data, kept out of the talents picker so ui/sim's import path
// stays free of a view module.

export type TalentsConfig<TalentsProto> = Array<TalentTreeConfig<TalentsProto>>;

export type TalentTreeConfig<TalentsProto> = {
	name: string;
	backgroundUrl: string;
	talents: Array<TalentConfig<TalentsProto>>;
};

export type TalentLocation = {
	// 0-indexed row in the tree
	rowIdx: number;
	// 0-indexed column in the tree
	colIdx: number;
};

export const locationKey = (location: TalentLocation): string => `${location.rowIdx},${location.colIdx}`;

export type TalentConfig<TalentsProto> = {
	fieldName?: keyof TalentsProto | string;

	// Display name, as it appears in game.
	fancyName: string;

	location: TalentLocation;

	// Location of a prerequisite talent, if any
	prereqLocation?: TalentLocation;

	// The one spell every rank of this talent reports; the rank rides along as ?rank=N.
	spellId: number;

	definitionId?: number;

	maxPoints: number;
};

export function newTalentsConfig<TalentsProto>(talents: TalentsConfig<TalentsProto>): TalentsConfig<TalentsProto> {
	talents.forEach(tree => {
		tree.talents.forEach((talent, i) => {
			// Validate that talents are given in the correct order (left-to-right top-to-bottom).
			if (i != 0) {
				const prevTalent = tree.talents[i - 1];
				if (
					talent.location.rowIdx < prevTalent.location.rowIdx ||
					(talent.location.rowIdx == prevTalent.location.rowIdx && talent.location.colIdx <= prevTalent.location.colIdx)
				) {
					throw new Error(`Out-of-order talent: ${String(talent.fieldName)}`);
				}
			}
		});
	});
	return talents;
}
