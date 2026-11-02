import { Tabs } from '@base-ui/react/tabs';
import i18n from '@i18n/config';
import type { Phase } from '@sim/constants/other';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { TabNav } from '@ui-kit/TabNav';
import { useMemo } from 'react';

import type { PresetGroupItem, PresetGroupSection } from './types';

export interface PresetGroupPickerProps {
	sections: ReadonlyArray<PresetGroupSection>;
	phases: ReadonlyArray<Phase>;
	phase: Phase;
	onSelectPhase: (phase: Phase) => void;
}

const UNGROUPED = '';

const groupNamesOf = (sections: ReadonlyArray<PresetGroupSection>): Array<string> => [
	...new Set(sections.flatMap(section => section.items.map(item => item.group).filter((group): group is string => !!group))),
];

const rowsOf = (items: ReadonlyArray<PresetGroupItem>, phase: Phase, groupNames: ReadonlyArray<string>): Array<[string, Array<PresetGroupItem>]> => {
	const byGroup = new Map<string, Array<PresetGroupItem>>();
	for (const item of items) {
		if (item.phase !== undefined && item.phase !== phase) continue;
		const group = item.group || UNGROUPED;
		if (!byGroup.has(group)) byGroup.set(group, []);
		byGroup.get(group)!.push(item);
	}

	// Ungrouped first (e.g. "Pre-Raid"), then the named groups in the order they were declared.
	return [UNGROUPED, ...groupNames].flatMap(group => {
		const groupItems = byGroup.get(group);
		return groupItems?.length ? [[group, groupItems] as [string, Array<PresetGroupItem>]] : [];
	});
};

export const PresetGroupPicker = ({ sections, phases, phase, onSelectPhase }: PresetGroupPickerProps) => {
	const groupNames = useMemo(() => groupNamesOf(sections), [sections]);
	const showGroupLabels = groupNames.length > 1;

	return (
		<div className="flex flex-col gap-section" data-testid="preset-group-picker">
			{phases.length > 1 && (
				<Tabs.Root value={String(phase)} onValueChange={next => onSelectPhase(Number(next) as Phase)}>
					<TabNav
						testId="preset-group-phase-tabs"
						tabs={phases.map(value => ({ id: String(value), label: i18n.t(`common.phase_names.${value}`) }))}
					/>
				</Tabs.Root>
			)}
			{sections.map(section => (
				<ContentBlock key={section.title} config={{ header: { title: section.title, tooltip: section.tooltip } }} headerChildren={section.headerChildren}>
					<div className="flex flex-col gap-2" data-testid="preset-group-section-body">
						{rowsOf(section.items, phase, groupNames).map(([group, items]) => (
							<div key={group || UNGROUPED} className="contents">
								{showGroupLabels && group !== UNGROUPED && (
									<div className="text-xs font-semibold tracking-wide text-foreground uppercase" data-testid="preset-group-label">
										{group}
									</div>
								)}
								<div className="ui-saved-data-presets" data-testid="saved-data-presets">
									{items.map(item => (
										<div key={item.key} className="contents">
											{item.node}
										</div>
									))}
								</div>
							</div>
						))}
					</div>
					{section.footer}
				</ContentBlock>
			))}
		</div>
	);
};
