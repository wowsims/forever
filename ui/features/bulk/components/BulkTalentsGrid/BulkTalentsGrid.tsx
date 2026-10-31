import i18n from '@i18n/config';
import { useActionId } from '@sim/hooks/useActionId';
import { ActionId } from '@sim/proto/action_id';
import type { TalentConfig, TalentsConfig } from '@sim/talents/config';
import { parseTalentsString, totalPointsSpent, treePointTotal } from '@sim/talents/talents_string';
import { useMemo } from 'react';

export interface BulkTalentsGridProps {
	config: TalentsConfig<any>;
	talentsString: string;
}

// The icon lookup leaves the rank out: a talent is one spell with one icon.
const TalentCell = ({ talent, points }: { talent: TalentConfig<any>; points: number }) => {
	const { iconUrl } = useActionId(ActionId.fromTalent(talent.spellId, 1, talent.definitionId ?? 0));
	return (
		<li data-testid="bulk-talents-grid-talent">
			<div
				className="ui-talent-picker-root m-0 cursor-default"
				data-points={String(points)}
				data-full={String(points >= talent.maxPoints)}
				style={{ backgroundImage: iconUrl ? `url('${iconUrl}')` : undefined }}>
				<span className="sr-only">{talent.fancyName}</span>
				<span className="ui-talent-picker-points">
					{points}/{talent.maxPoints}
				</span>
			</div>
		</li>
	);
};

/** The grid as text, for a description: a sentence per tree with its points and each talent it takes. */
export const bulkTalentsText = ({ config, talentsString }: BulkTalentsGridProps): string => {
	const points = parseTalentsString(config, talentsString);
	if (totalPointsSpent(points) === 0) return i18n.t('bulk_tab.settings.talents.no_points');
	return config
		.flatMap((tree, treeIdx) => {
			const spent = treePointTotal(points, treeIdx);
			if (spent === 0) return [];
			const talents = tree.talents.flatMap((talent, talentIdx) =>
				points[treeIdx][talentIdx] > 0 ? [`${talent.fancyName} ${points[treeIdx][talentIdx]}/${talent.maxPoints}`] : [],
			);
			return [`${tree.name} ${spent}: ${talents.join(', ')}.`];
		})
		.join(' ');
};

export const BulkTalentsGrid = ({ config, talentsString }: BulkTalentsGridProps) => {
	const points = useMemo(() => parseTalentsString(config, talentsString), [config, talentsString]);

	if (totalPointsSpent(points) === 0) return <span>{i18n.t('bulk_tab.settings.talents.no_points')}</span>;

	return (
		<div className="flex flex-col gap-4 text-left" data-testid="bulk-talents-grid">
			{config.map(
				(tree, treeIdx) =>
					treePointTotal(points, treeIdx) > 0 && (
						<div key={tree.name} className="flex flex-col gap-1" data-testid="bulk-talents-grid-tree">
							<div className="flex gap-2 text-sm font-bold">
								<span>{tree.name}</span>
								<span>{treePointTotal(points, treeIdx)}</span>
							</div>
							<ul className="m-0 flex list-none flex-wrap gap-2 p-0">
								{tree.talents.map((talent, talentIdx) =>
									points[treeIdx][talentIdx] > 0 ? <TalentCell key={talentIdx} talent={talent} points={points[treeIdx][talentIdx]} /> : null,
								)}
							</ul>
						</div>
					),
			)}
		</div>
	);
};
