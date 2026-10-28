import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import { EncounterDesign, type PresetTarget, Target } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Encounter, remapDesignTargets } from '@sim/raid/encounter';
import { batch } from '@sim/state/batch';
import { subscribeEncounterField } from '@sim/state/subscriptions';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { type ReactNode, type Ref, useMemo, useRef, useState } from 'react';

import type { DesignEventKind, DesignEventRef } from '../../model/design/lanes';
import { addEvent } from '../../model/design/lanes';
import { type DesignIssue, validateDesign } from '../../model/design/validate';
import { durationConfigs } from '../EncounterPicker/utils/configs';
import { TargetsPicker } from '../TargetsPicker';
import { DesignerEventPanel } from './DesignerEventPanel';
import { DesignerIssues } from './DesignerIssues';
import { DesignerTimeline } from './DesignerTimeline';
import { defaultWaveTemplate, designerPresetTargets, designerText } from './utils/configs';

export interface EncounterDesignerViewProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** The encounter's duration fields; the timeline is drawn against the base duration. */
	duration: ReactNode;
	targets: ReactNode;
	design: EncounterDesign;
	baseSeconds: number;
	selected?: DesignEventRef;
	issues: ReadonlyArray<DesignIssue>;
	targetNames: ReadonlyArray<string>;
	presetTargets: ReadonlyArray<PresetTarget>;
	/** A new add of a wave: the default template's copy. */
	newAdd: () => Target;
	panelRef?: Ref<HTMLElement>;
	onSelect: (ref: DesignEventRef | undefined) => void;
	onActivate: (ref: DesignEventRef) => void;
	onChange: (design: EncounterDesign) => void;
	onAdd: (kind: DesignEventKind) => void;
	onSave: () => void;
}

export const EncounterDesignerView = ({
	open,
	onOpenChange,
	duration,
	targets,
	design,
	baseSeconds,
	selected,
	issues,
	targetNames,
	presetTargets,
	newAdd,
	panelRef,
	onSelect,
	onActivate,
	onChange,
	onAdd,
	onSave,
}: EncounterDesignerViewProps) => {
	const [sections, setSections] = useState<string[]>([]);
	// Opening focuses the first add button: the first tabbable element would be the Duration field in the header.
	const initialFocus = useRef<HTMLElement | null>(null);
	const findInitialFocus = (timeline: HTMLElement | null) => {
		initialFocus.current = timeline?.querySelector<HTMLElement>('[data-lane] button') ?? null;
	};
	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			initialFocus={initialFocus}
			testId="encounter-designer"
			size="full"
			scrollContents
			title={designerText('title')}
			headerChildren={<FieldGrid className="mx-4 max-w-lg">{duration}</FieldGrid>}
			footer={
				<div className="flex gap-2">
					<Button variant="outline-cancel" onClick={() => onOpenChange(false)}>
						{i18n.t('common.cancel')}
					</Button>
					<Button variant="primary" onClick={onSave}>
						{designerText('save')}
					</Button>
				</div>
			}>
			<Accordion value={sections} onValueChange={setSections} density="compact">
				<AccordionItem value="targets" title={designerText('targets')} testId="encounter-designer-targets">
					{targets}
				</AccordionItem>
			</Accordion>
			<div className="flex gap-4 max-md:flex-col">
				<div ref={findInitialFocus} className="min-w-0 flex-1">
					<DesignerTimeline
						design={design}
						baseSeconds={baseSeconds}
						selected={selected}
						issues={issues}
						targetNames={targetNames}
						onSelect={onSelect}
						onActivate={onActivate}
						onChange={onChange}
						onAdd={onAdd}
					/>
				</div>
				<div className="w-80 shrink-0 max-md:w-full">
					<DesignerEventPanel
						ref={panelRef}
						design={design}
						selected={selected}
						baseSeconds={baseSeconds}
						targetNames={targetNames}
						presetTargets={presetTargets}
						newAdd={newAdd}
						onChange={onChange}
						onSelect={onSelect}
					/>
				</div>
			</div>
			<DesignerIssues design={design} issues={issues} baseSeconds={baseSeconds} onSelect={onSelect} />
		</Dialog>
	);
};

export interface EncounterDesignerProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

interface EncounterSnapshot {
	duration: number;
	durationVariation: number;
	targets: Array<Target>;
	design: EncounterDesign;
}

export const EncounterDesigner = ({ open, onOpenChange }: EncounterDesignerProps) => {
	const encounter = useSimHost().sim.encounter;
	const baseSeconds = useStoreSubscribe(subscribeEncounterField(encounter, 'duration'), () => encounter.getDuration());
	const targets = useStoreSubscribe(subscribeEncounterField(encounter, 'targets'), () => encounter.getTargets());
	const targetNames = useMemo(() => targets.map((target, index) => target.name || designerText('target_name', { number: index + 1 })), [targets]);
	const presetTargets = useMemo(() => designerPresetTargets(encounter.sim.db.getAllPresetEncounters(), encounter.sim.db.getAllPresetTargets()), [encounter]);
	const duration = useMemo(() => durationConfigs(), []);
	const panelRef = useRef<HTMLElement>(null);

	// Design edits stay in the draft until Save. The duration and target fields write to the store as they change,
	// so each opening snapshots them, and every close but Save puts them back.
	const [draft, setDraft] = useState(() => encounter.getDesign());
	const [selected, setSelected] = useState<DesignEventRef>();
	const [snapshot, setSnapshot] = useState<EncounterSnapshot>();
	const [openedWith, setOpenedWith] = useState(false);
	const [seenTargets, setSeenTargets] = useState(targets);
	if (open !== openedWith) {
		setOpenedWith(open);
		setSeenTargets(targets);
		if (open) {
			setSnapshot({
				duration: encounter.getDuration(),
				durationVariation: encounter.getDurationVariation(),
				targets,
				design: encounter.getDesign(),
			});
			setDraft(encounter.getDesign());
			setSelected(undefined);
		}
	} else if (targets !== seenTargets) {
		// The store remaps its own design when targets move; the draft follows the same way.
		setSeenTargets(targets);
		setDraft(current => remapDesignTargets(current, seenTargets, targets));
	}
	const issues = useMemo(() => validateDesign(draft, targets.length), [draft, targets.length]);

	const close = (nextOpen: boolean) => {
		if (!nextOpen && snapshot) {
			batch(() => {
				encounter.setDuration(snapshot.duration);
				encounter.setDurationVariation(snapshot.durationVariation);
				if (encounter.getTargets() !== snapshot.targets) encounter.setTargets(snapshot.targets);
				encounter.setDesign(snapshot.design);
			});
		}
		onOpenChange(nextOpen);
	};

	const newAdd = () => Target.clone(defaultWaveTemplate(presetTargets, encounter.sim.raid.getTanks().length) ?? Encounter.defaultTargetProto());

	const add = (kind: DesignEventKind) => {
		const next = EncounterDesign.clone(draft);
		const ref = addEvent(next, kind);
		// The sim spawns a wave's adds from their templates, so a new wave starts with one add from the default template.
		if (kind === 'wave') next.addWaves[ref.index].adds = [newAdd()];
		setDraft(next);
		setSelected(ref);
	};

	const save = () => {
		encounter.setDesign(draft);
		onOpenChange(false);
	};

	return (
		<EncounterDesignerView
			open={open}
			onOpenChange={close}
			duration={duration.map(config => (
				<ConnectedNumberField key={config.id} modObject={encounter} config={config} />
			))}
			targets={<TargetsPicker encounter={encounter} />}
			newAdd={newAdd}
			design={draft}
			baseSeconds={baseSeconds}
			selected={selected}
			issues={issues}
			targetNames={targetNames}
			presetTargets={presetTargets}
			panelRef={panelRef}
			onSelect={setSelected}
			onActivate={() => panelRef.current?.focus()}
			onChange={setDraft}
			onAdd={add}
			onSave={save}
		/>
	);
};
