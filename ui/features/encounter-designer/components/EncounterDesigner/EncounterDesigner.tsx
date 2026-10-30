import { durationConfigs } from '@features/encounter/components/EncounterPicker/utils/configs';
import { TargetsPicker } from '@features/encounter/components/TargetsPicker';
import { useMyEncounters } from '@features/encounter/hooks/useMyEncounters';
import { uniqueEncounterName } from '@features/encounter/model/my_encounters';
import { ConnectedNumberField, ConnectedStringField } from '@features/inputs/components/ConnectedTextField';
import { type Encounter as EncounterProto, EncounterDesign, type PresetTarget, Target } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Encounter, remapDesignTargets } from '@sim/raid/encounter';
import { batch } from '@sim/state/batch';
import { subscribeEncounterField } from '@sim/state/subscriptions';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { Button } from '@ui-kit/Button';
import { ConfirmPopover } from '@ui-kit/ConfirmPopover';
import { Dialog } from '@ui-kit/Dialog';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { type ReactNode, type Ref, useMemo, useRef, useState } from 'react';

import type { DesignEventKind, DesignEventRef } from '../../model/lanes';
import { addEvent } from '../../model/lanes';
import { type DesignIssue, validateDesign } from '../../model/validate';
import { DesignerEventPanel } from '../DesignerEventPanel';
import { DesignerIssues } from '../DesignerIssues';
import { DesignerTimeline } from '../DesignerTimeline';
import { defaultWaveTemplate, designerPresetTargets, designerText, nameConfig } from './utils/configs';

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
	name: ReactNode;
	/** Shows Save as copy, which stores the encounter as a new entry of My encounters. */
	onSaveCopy?: () => void;
	/** Shows Delete, which asks inside the dialog before it removes the entry `deleteName` names. */
	onDelete?: () => void;
	deleteName?: string;
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
	name,
	onSaveCopy,
	onDelete,
	deleteName,
}: EncounterDesignerViewProps) => {
	const [sections, setSections] = useState<string[]>([]);
	const [deleteButton, setDeleteButton] = useState<HTMLButtonElement | null>(null);
	const [confirmingDelete, setConfirmingDelete] = useState(false);
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
			headerChildren={
				<FieldGrid className="mx-4 max-w-2xl">
					{name}
					{duration}
				</FieldGrid>
			}
			footer={
				<div className="flex w-full gap-2">
					{onDelete && (
						<>
							<Button ref={setDeleteButton} variant="danger" onClick={() => setConfirmingDelete(true)}>
								{designerText('delete')}
							</Button>
							<ConfirmPopover
								open={confirmingDelete}
								onOpenChange={setConfirmingDelete}
								anchor={deleteButton}
								confirmLabel={designerText('delete')}
								onConfirm={onDelete}
								testId="encounter-designer-delete-confirm">
								{designerText('delete_confirm', { name: deleteName ?? '' })}
							</ConfirmPopover>
						</>
					)}
					<div className="ml-auto flex gap-2">
						<Button variant="outline-cancel" onClick={() => onOpenChange(false)}>
							{i18n.t('common.cancel')}
						</Button>
						{onSaveCopy && (
							<Button variant="outline-primary" onClick={onSaveCopy}>
								{designerText('save_copy')}
							</Button>
						)}
						<Button variant="primary" onClick={onSave}>
							{designerText('save')}
						</Button>
					</div>
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
	/** Opened on a new encounter from the +: Save adds it to My encounters, and Cancel loads this encounter back. */
	replaced?: EncounterProto;
}

interface EncounterSnapshot {
	name: string;
	duration: number;
	durationVariation: number;
	targets: Array<Target>;
	design: EncounterDesign;
}

export const EncounterDesigner = ({ open, onOpenChange, replaced }: EncounterDesignerProps) => {
	const host = useSimHost();
	const encounter = host.sim.encounter;
	const myEncounters = useMyEncounters();
	const baseSeconds = useStoreSubscribe(subscribeEncounterField(encounter, 'duration'), () => encounter.getDuration());
	const targets = useStoreSubscribe(subscribeEncounterField(encounter, 'targets'), () => encounter.getTargets());
	const targetNames = useMemo(() => targets.map((target, index) => target.name || designerText('target_name', { number: index + 1 })), [targets]);
	const presetTargets = useMemo(() => designerPresetTargets(encounter.sim.db.getAllPresetEncounters(), encounter.sim.db.getAllPresetTargets()), [encounter]);
	const duration = useMemo(() => durationConfigs(), []);
	const nameField = useMemo(() => nameConfig(), []);
	const panelRef = useRef<HTMLElement>(null);

	// Design edits stay in the draft until Save. The name, duration and target fields write to the store as they
	// change, so each opening snapshots them, and every close but Save puts them back.
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
				name: encounter.getName(),
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
		if (!nextOpen && replaced) {
			encounter.fromProto(replaced);
		} else if (!nextOpen && snapshot) {
			batch(() => {
				encounter.setName(snapshot.name);
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

	// The entry the designer opened on, matched by the name the encounter had then.
	const entry = snapshot && !replaced ? myEncounters.entries.find(entry => entry.name === snapshot.name) : undefined;

	const store = (replacing?: string) => {
		encounter.setDesign(draft);
		const taken = myEncounters.entries.map(entry => entry.name).filter(name => name !== replacing);
		encounter.setName(uniqueEncounterName(encounter.getName() || i18n.t('settings_tab.encounter.new_encounter'), taken));
		myEncounters.save(encounter.getName(), encounter.toProto(), replacing);
		onOpenChange(false);
	};

	// Save updates the entry the designer opened on, and adds a new or imported named encounter; an encounter that
	// opened without a name (a preset, or one edited by hand) only goes to the sim.
	const save = () => {
		if (entry || replaced || snapshot?.name) return store(entry?.name);
		encounter.setDesign(draft);
		onOpenChange(false);
	};

	const deleteEntry = () => {
		myEncounters.remove(entry!.name);
		encounter.fromProto(host.individualConfig.defaults.encounter?.encounter ?? Encounter.defaultEncounterProto());
		onOpenChange(false);
	};

	return (
		<EncounterDesignerView
			open={open}
			onOpenChange={close}
			name={<ConnectedStringField modObject={encounter} config={nameField} />}
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
			onSaveCopy={replaced ? undefined : () => store()}
			onDelete={entry && deleteEntry}
			deleteName={entry?.name}
		/>
	);
};
