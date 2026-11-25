import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { type ReactNode, useState } from 'react';

import { plannerText } from '../EncounterPlanner/utils/configs';

export interface PlannerSection {
	id: string;
	title: string;
	/** Follows the title while the section is closed. */
	summary: string;
	/** Closed until the player opens it, for a section whose values are all unset. */
	startsClosed?: boolean;
	/** Open whatever the player chose, while a field in it has an issue. */
	forcedOpen?: boolean;
	children: ReactNode;
}

/** The sections the player opened or closed, by id, so a choice carries over to the next event or target shown. */
export interface PlannerSectionState {
	toggles: Readonly<Record<string, boolean>>;
	onToggle: (id: string, open: boolean) => void;
}

export const usePlannerSectionState = (): PlannerSectionState => {
	const [toggles, setToggles] = useState<Readonly<Record<string, boolean>>>({});
	return { toggles, onToggle: (id, open) => setToggles(current => ({ ...current, [id]: open })) };
};

export interface PlannerSectionsProps {
	sections: ReadonlyArray<PlannerSection>;
	state: PlannerSectionState;
	testId: string;
}

export const PlannerSections = ({ sections, state, testId }: PlannerSectionsProps) => {
	const open = sections.filter(section => section.forcedOpen || (state.toggles[section.id] ?? !section.startsClosed)).map(section => section.id);
	const change = (next: string[]) =>
		sections.filter(({ id }) => next.includes(id) !== open.includes(id)).forEach(({ id }) => state.onToggle(id, next.includes(id)));
	return (
		<Accordion value={open} onValueChange={change} density="compact" testId={testId}>
			{sections.map(section => (
				<AccordionItem
					key={section.id}
					value={section.id}
					title={
						<span className="min-w-0 truncate">
							{open.includes(section.id)
								? section.title
								: plannerText('panel.section_summary', { title: section.title, summary: section.summary })}
						</span>
					}>
					{section.children}
				</AccordionItem>
			))}
		</Accordion>
	);
};
