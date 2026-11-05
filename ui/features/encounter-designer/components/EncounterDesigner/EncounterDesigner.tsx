import { durationConfigs, executeConfigs } from '@features/encounter/components/EncounterPicker/utils/configs';
import { useEncounterChoice } from '@features/encounter/hooks/useEncounterChoice';
import { useMyEncounters } from '@features/encounter/hooks/useMyEncounters';
import { uniqueEncounterName } from '@features/encounter/model/my_encounters';
import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { ConnectedNumberField, ConnectedStringField } from '@features/inputs/components/ConnectedTextField';
import { useListItems } from '@features/inputs/hooks/useListItems';
import { type AreaType, type Encounter as EncounterProto, EncounterDesign, type PresetTarget, Target } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Encounter, remapDesignTargets } from '@sim/raid/encounter';
import { batch } from '@sim/state/batch';
import { subscribeEncounterField } from '@sim/state/subscriptions';
import { Button } from '@ui-kit/Button';
import { ConfirmPopover } from '@ui-kit/ConfirmPopover';
import { Dialog } from '@ui-kit/Dialog';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { type ReactNode, type Ref, useMemo, useRef, useState } from 'react';

import type { DesignerSelection, DesignEventKind } from '../../model/lanes';
import { addEvent } from '../../model/lanes';
import { type DesignIssue, validateDesign } from '../../model/validate';
import { encounterTargetContext, npcConfig, presetTargetConfig, targetsListConfig } from '../DesignerAddStats/utils/configs';
import { DesignerEventPanel, type DesignerTargetEditor } from '../DesignerEventPanel';
import { DesignerIssues } from '../DesignerIssues';
import { DesignerTimeline } from '../DesignerTimeline';
import { defaultWaveTemplate, designerPresetTargets, designerText, nameConfig } from './utils/configs';

export interface EncounterDesignerViewProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** The encounter's duration fields; the timeline is drawn against the base duration. */
	duration: ReactNode;
	/** The encounter choice: Custom, the presets and My encounters. */
	preset?: ReactNode;
	/** The execute proportions, for the specs that read them. */
	execute?: ReactNode;
	design: EncounterDesign;
	baseSeconds: number;
	selected?: DesignerSelection;
	issues: ReadonlyArray<DesignIssue>;
	targetNames: ReadonlyArray<string>;
	presetTargets: ReadonlyArray<PresetTarget>;
	/** A new add of a wave: the default template's copy. */
	newAdd: () => Target;
	targetEditor?: DesignerTargetEditor;
	panelRef?: Ref<HTMLElement>;
	onSelect: (selection: DesignerSelection | undefined) => void;
	onActivate: (selection: DesignerSelection) => void;
	onChange: (design: EncounterDesign) => void;
	onAdd: (kind: DesignEventKind) => void;
	onAddTarget: () => void;
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
	preset,
	execute,
	design,
	baseSeconds,
	selected,
	issues,
	targetNames,
	presetTargets,
	newAdd,
	targetEditor,
	panelRef,
	onSelect,
	onActivate,
	onChange,
	onAdd,
	onAddTarget,
	onSave,
	name,
	onSaveCopy,
	onDelete,
	deleteName,
}: EncounterDesignerViewProps) => {
	const [deleteButton, setDeleteButton] = useState<HTMLButtonElement | null>(null);
	const [confirmingDelete, setConfirmingDelete] = useState(false);
	// Opening focuses the first event lane's add button: the first tabbable element would be a field in the header.
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
				<div className="mx-4 flex w-full max-w-2xl flex-col gap-2">
					{/* The encounter choice shares its row with the name only, so a preset path such as Default/Dynamic Adds fits. */}
					<FieldGrid>
						{preset}
						{name}
					</FieldGrid>
					<FieldGrid>{duration}</FieldGrid>
					{execute && <FieldGrid>{execute}</FieldGrid>}
				</div>
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
						onAddTarget={onAddTarget}
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
						targetEditor={targetEditor}
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
	/** Shows the execute proportions in the header, as the Encounter panel does. */
	showExecuteProportion: boolean;
}

interface EncounterSnapshot {
	name: string;
	myEncounter: boolean;
	duration: number;
	durationVariation: number;
	executeProportion20: number;
	executeProportion35: number;
	areaTypes: Array<AreaType>;
	zoneId: number;
	targets: Array<Target>;
	design: EncounterDesign;
}

const CHOICE_FIELD = { id: 'encounter-designer-preset' };
const TARGET_LIST = targetsListConfig();

export const EncounterDesigner = ({ open, onOpenChange, replaced, showExecuteProportion }: EncounterDesignerProps) => {
	const host = useSimHost();
	const encounter = host.sim.encounter;
	const myEncounters = useMyEncounters();
	const choice = useEncounterChoice(encounter, CHOICE_FIELD);
	// The boss follows its Boss health lane when that has keyframes; every other target follows these.
	const execute = useMemo(
		() =>
			executeConfigs().map(config => ({
				...config,
				id: `encounter-designer-${config.id}`,
				labelTooltip: `${config.labelTooltip} ${designerText('execute_boss_health')}`,
			})),
		[],
	);
	const baseSeconds = useStoreSubscribe(subscribeEncounterField(encounter, 'duration'), () => encounter.getDuration());
	const targets = useStoreSubscribe(subscribeEncounterField(encounter, 'targets'), () => encounter.getTargets());
	const targetNames = useMemo(() => targets.map((target, index) => target.name || designerText('target_name', { number: index + 1 })), [targets]);
	const presetTargets = useMemo(() => designerPresetTargets(encounter.sim.db.getAllPresetEncounters(), encounter.sim.db.getAllPresetTargets()), [encounter]);
	const duration = useMemo(() => durationConfigs().map(config => ({ ...config, id: `encounter-designer-${config.id}` })), []);
	const nameField = useMemo(() => nameConfig(), []);
	const targetList = useListItems(encounter, TARGET_LIST);
	const { threat: threatMetrics } = useDisplayMetrics(encounter.sim);
	const panelRef = useRef<HTMLElement>(null);

	// Design edits stay in the draft until Save. The header's fields and the target fields write to the store as they
	// change, so each opening snapshots them, and every close but Save puts them back.
	const [draft, setDraft] = useState(() => ({ design: encounter.getDesign(), targets }));
	const [selected, setSelected] = useState<DesignerSelection>();
	const [snapshot, setSnapshot] = useState<EncounterSnapshot>();
	// What the designer edits: the encounter it opened on, or the one picked last in the header.
	const [origin, setOrigin] = useState<{ name: string; myEncounter: boolean }>();
	const [openedWith, setOpenedWith] = useState(false);
	if (open !== openedWith) {
		setOpenedWith(open);
		if (open) {
			setSnapshot({
				name: encounter.getName(),
				myEncounter: encounter.isMyEncounter(),
				duration: encounter.getDuration(),
				durationVariation: encounter.getDurationVariation(),
				executeProportion20: encounter.getExecuteProportion20(),
				executeProportion35: encounter.getExecuteProportion35(),
				areaTypes: encounter.getAreaTypes(),
				zoneId: encounter.getZoneId(),
				targets,
				design: encounter.getDesign(),
			});
			setOrigin({ name: encounter.getName(), myEncounter: encounter.isMyEncounter() });
			setDraft({ design: encounter.getDesign(), targets: encounter.getTargets() });
			setSelected(undefined);
		}
	} else if (draft.targets !== targets) {
		// The store remaps its own design when targets move; the draft follows the same way.
		setDraft({ design: remapDesignTargets(draft.design, draft.targets, targets), targets });
	}
	const issues = useMemo(() => validateDesign(draft.design, targets.length), [draft.design, targets.length]);

	// A preset or an entry picked here writes its own targets and design to the store, and the draft takes up that
	// design in place of its own, against the new targets. The pick is what Save and Delete act on from then. Custom
	// writes nothing, so the draft stays.
	const presetField = {
		...choice,
		config: {
			...choice.config,
			setValue: (subject: Encounter, value: string) => {
				const [design, targets] = [encounter.getDesign(), encounter.getTargets()];
				choice.config.setValue(subject, value);
				if (encounter.getDesign() === design && encounter.getTargets() === targets) return;
				setDraft({ design: encounter.getDesign(), targets: encounter.getTargets() });
				setOrigin({ name: encounter.getName(), myEncounter: encounter.isMyEncounter() });
				setSelected(undefined);
			},
		},
	};

	const close = (nextOpen: boolean) => {
		if (!nextOpen && replaced) {
			encounter.fromProto(replaced);
		} else if (!nextOpen && snapshot) {
			batch(() => {
				encounter.setName(snapshot.name);
				encounter.setMyEncounter(snapshot.myEncounter);
				encounter.setDuration(snapshot.duration);
				encounter.setDurationVariation(snapshot.durationVariation);
				encounter.setExecuteProportion20(snapshot.executeProportion20);
				encounter.setExecuteProportion35(snapshot.executeProportion35);
				encounter.setAreaTypes(snapshot.areaTypes, snapshot.zoneId);
				if (encounter.getTargets() !== snapshot.targets) encounter.setTargets(snapshot.targets);
				encounter.setDesign(snapshot.design);
			});
		}
		onOpenChange(nextOpen);
	};

	const newAdd = () => Target.clone(defaultWaveTemplate(presetTargets, encounter.sim.raid.getTanks().length) ?? Encounter.defaultTargetProto());

	const add = (kind: DesignEventKind) => {
		const next = EncounterDesign.clone(draft.design);
		const ref = addEvent(next, kind);
		// The sim spawns a wave's adds from their templates, so a new wave starts with one add from the default template.
		if (kind === 'wave') next.addWaves[ref.index].adds = [newAdd()];
		setDraft(d => ({ ...d, design: next }));
		setSelected(ref);
	};

	const selectTarget = (index: number) => setSelected({ kind: 'target', index });
	const addTarget = () => {
		targetList.create();
		selectTarget(targets.length);
	};
	const selectedIndex = selected?.kind === 'target' ? selected.index : undefined;
	const targetFields = useMemo(() => {
		if (selectedIndex === undefined) return undefined;
		const context = encounterTargetContext(encounter, selectedIndex, `encounter-designer-target-${selectedIndex}`);
		return { context, identity: [npcConfig(context), presetTargetConfig(context)] };
	}, [encounter, selectedIndex]);
	const editTarget = (index: number, target: Target, { context, identity }: NonNullable<typeof targetFields>): DesignerTargetEditor => ({
		target,
		idPrefix: context.idPrefix,
		identity,
		threatMetrics,
		canRemove: targetList.canRemove(index),
		onModify: context.modifyTarget,
		onMove: direction => {
			targetList.move(index, direction < 0 ? index - 1 : index + 2);
			selectTarget(index + direction);
		},
		onCopy: () => {
			targetList.append(Target.clone(target));
			selectTarget(targets.length);
		},
		onRemove: () => {
			targetList.remove(index);
			setSelected(undefined);
		},
	});
	const selectedTarget = selectedIndex === undefined ? undefined : targets[selectedIndex];

	// The entry the designer edits: the encounter was marked as one when it opened or was picked in the header, under
	// the name it had then.
	const entry = origin?.myEncounter && !replaced ? myEncounters.entries.find(entry => entry.name === origin.name) : undefined;

	// The list holds an entry without the mark, which the encounter takes once it is stored.
	const store = (replacing?: string) => {
		batch(() => {
			encounter.setDesign(draft.design);
			const taken = myEncounters.entries.map(entry => entry.name).filter(name => name !== replacing);
			encounter.setName(uniqueEncounterName(encounter.getName() || i18n.t('settings_tab.encounter.new_encounter'), taken));
			encounter.setMyEncounter(true);
		});
		myEncounters.save(encounter.getName(), { ...encounter.toProto(), myEncounter: false }, replacing);
		onOpenChange(false);
	};

	// Save updates the entry the designer edits and adds the fight to My encounters when it was started from the +, has a
	// name (a typed one counts) or holds an event in any lane, under a free name when an entry has it already. A simple
	// fight without a name (a preset, or one edited by hand) only goes to the sim.
	const save = () => {
		encounter.setDesign(draft.design);
		if (entry || replaced || encounter.getName() || encounter.hasDesign()) return store(entry?.name);
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
			preset={<ConnectedChoiceField {...presetField} />}
			execute={showExecuteProportion && execute.map(config => <ConnectedNumberField key={config.id} modObject={encounter} config={config} />)}
			newAdd={newAdd}
			design={draft.design}
			baseSeconds={baseSeconds}
			selected={selected}
			issues={issues}
			targetNames={targetNames}
			presetTargets={presetTargets}
			targetEditor={selected?.kind === 'target' && selectedTarget && targetFields ? editTarget(selected.index, selectedTarget, targetFields) : undefined}
			panelRef={panelRef}
			onSelect={setSelected}
			onActivate={() => panelRef.current?.focus()}
			onChange={design => setDraft(d => ({ ...d, design }))}
			onAdd={add}
			onAddTarget={addTarget}
			onSave={save}
			onSaveCopy={replaced ? undefined : () => store()}
			onDelete={entry && deleteEntry}
			deleteName={entry?.name}
		/>
	);
};
