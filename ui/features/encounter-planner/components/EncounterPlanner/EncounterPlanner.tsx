import { durationConfigs, executeConfigs } from '@features/encounter/components/EncounterPicker/utils/configs';
import { useEncounterChoice } from '@features/encounter/hooks/useEncounterChoice';
import { useMyEncounters } from '@features/encounter/hooks/useMyEncounters';
import { uniqueEncounterName } from '@features/encounter/model/my_encounters';
import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { ConnectedNumberField, ConnectedStringField } from '@features/inputs/components/ConnectedTextField';
import { useListItems } from '@features/inputs/hooks/useListItems';
import { type AreaType, type Encounter as EncounterProto, EncounterPlan, type PresetTarget, Target } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Encounter, remapPlanTargets } from '@sim/raid/encounter';
import { batch } from '@sim/state/batch';
import { subscribeEncounterField } from '@sim/state/subscriptions';
import { Button } from '@ui-kit/Button';
import { ConfirmPopover } from '@ui-kit/ConfirmPopover';
import { Dialog } from '@ui-kit/Dialog';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { type ReactNode, type Ref, useMemo, useRef, useState } from 'react';

import type { PlanEventKind, PlannerSelection } from '../../model/lanes';
import { addEvent } from '../../model/lanes';
import { type PlanIssue, validatePlan } from '../../model/validate';
import { encounterTargetContext, npcConfig, presetTargetConfig, targetsListConfig } from '../PlannerAddStats/utils/configs';
import { PlannerEventPanel, type PlannerTargetEditor } from '../PlannerEventPanel';
import { PlannerIssues } from '../PlannerIssues';
import { PlannerTimeline } from '../PlannerTimeline';
import { defaultWaveTemplate, nameConfig, plannerPresetTargets, plannerText } from './utils/configs';

export interface EncounterPlannerViewProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** The encounter's duration fields; the timeline is drawn against the base duration. */
	duration: ReactNode;
	/** The encounter choice: Custom, the presets and My encounters. */
	preset?: ReactNode;
	/** The execute proportions, for the specs that read them. */
	execute?: ReactNode;
	plan: EncounterPlan;
	baseSeconds: number;
	selected?: PlannerSelection;
	issues: ReadonlyArray<PlanIssue>;
	targetNames: ReadonlyArray<string>;
	presetTargets: ReadonlyArray<PresetTarget>;
	/** A new add of a wave: the default template's copy. */
	newAdd: () => Target;
	targetEditor?: PlannerTargetEditor;
	panelRef?: Ref<HTMLElement>;
	onSelect: (selection: PlannerSelection | undefined) => void;
	onActivate: (selection: PlannerSelection) => void;
	onChange: (plan: EncounterPlan) => void;
	onAdd: (kind: PlanEventKind, targetIndex?: number) => void;
	onAddTarget: () => void;
	onSave: () => void;
	name: ReactNode;
	/** Shows Save as copy, which stores the encounter as a new entry of My encounters. */
	onSaveCopy?: () => void;
	/** Shows Delete, which asks inside the dialog before it removes the entry `deleteName` names. */
	onDelete?: () => void;
	deleteName?: string;
}

export const EncounterPlannerView = ({
	open,
	onOpenChange,
	duration,
	preset,
	execute,
	plan,
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
}: EncounterPlannerViewProps) => {
	const [deleteButton, setDeleteButton] = useState<HTMLButtonElement | null>(null);
	const [confirmingDelete, setConfirmingDelete] = useState(false);
	// Opening focuses the first add button, the first target's + Downtime: the first tabbable element would be a field in the header.
	const initialFocus = useRef<HTMLElement | null>(null);
	const findInitialFocus = (timeline: HTMLElement | null) => {
		initialFocus.current = timeline?.querySelector<HTMLElement>('[data-lane-add] button') ?? null;
	};
	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			initialFocus={initialFocus}
			testId="encounter-planner"
			size="full"
			scrollContents
			title={plannerText('title')}
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
								{plannerText('delete')}
							</Button>
							<ConfirmPopover
								open={confirmingDelete}
								onOpenChange={setConfirmingDelete}
								anchor={deleteButton}
								confirmLabel={plannerText('delete')}
								onConfirm={onDelete}
								testId="encounter-planner-delete-confirm">
								{plannerText('delete_confirm', { name: deleteName ?? '' })}
							</ConfirmPopover>
						</>
					)}
					<div className="ml-auto flex gap-2">
						<Button variant="outline-cancel" onClick={() => onOpenChange(false)}>
							{i18n.t('common.cancel')}
						</Button>
						{onSaveCopy && (
							<Button variant="outline-primary" onClick={onSaveCopy}>
								{plannerText('save_copy')}
							</Button>
						)}
						<Button variant="primary" onClick={onSave}>
							{plannerText('save')}
						</Button>
					</div>
				</div>
			}>
			<div className="flex gap-4 max-md:flex-col">
				<div ref={findInitialFocus} className="min-w-0 flex-1">
					<PlannerTimeline
						plan={plan}
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
					<PlannerEventPanel
						ref={panelRef}
						plan={plan}
						selected={selected}
						issues={issues}
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
			<PlannerIssues plan={plan} issues={issues} baseSeconds={baseSeconds} onSelect={onSelect} />
		</Dialog>
	);
};

export interface EncounterPlannerProps {
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
	plan: EncounterPlan;
}

const CHOICE_FIELD = { id: 'encounter-planner-preset' };
const TARGET_LIST = targetsListConfig();

export const EncounterPlanner = ({ open, onOpenChange, replaced, showExecuteProportion }: EncounterPlannerProps) => {
	const host = useSimHost();
	const encounter = host.sim.encounter;
	const myEncounters = useMyEncounters();
	const choice = useEncounterChoice(encounter, CHOICE_FIELD);
	// The first target follows the boss health keyframes on its row when it has some; every other target follows these.
	const execute = useMemo(
		() =>
			executeConfigs().map(config => ({
				...config,
				id: `encounter-planner-${config.id}`,
				labelTooltip: `${config.labelTooltip} ${plannerText('execute_boss_health')}`,
			})),
		[],
	);
	const baseSeconds = useStoreSubscribe(subscribeEncounterField(encounter, 'duration'), () => encounter.getDuration());
	const targets = useStoreSubscribe(subscribeEncounterField(encounter, 'targets'), () => encounter.getTargets());
	const targetNames = useMemo(() => targets.map((target, index) => target.name || plannerText('target_name', { number: index + 1 })), [targets]);
	const presetTargets = useMemo(() => plannerPresetTargets(encounter.sim.db.getAllPresetEncounters(), encounter.sim.db.getAllPresetTargets()), [encounter]);
	const duration = useMemo(() => durationConfigs().map(config => ({ ...config, id: `encounter-planner-${config.id}` })), []);
	const nameField = useMemo(() => nameConfig(), []);
	const targetList = useListItems(encounter, TARGET_LIST);
	const { threat: threatMetrics } = useDisplayMetrics(encounter.sim);
	const panelRef = useRef<HTMLElement>(null);

	// Edits to the plan stay in the draft until Save. The header's fields and the target fields write to the store as they
	// change, so each opening snapshots them, and every close but Save puts them back.
	const [draft, setDraft] = useState(() => ({ plan: encounter.getPlan(), targets }));
	const [selected, setSelected] = useState<PlannerSelection>();
	const [snapshot, setSnapshot] = useState<EncounterSnapshot>();
	// What the planner edits: the encounter it opened on, or the one picked last in the header.
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
				plan: encounter.getPlan(),
			});
			setOrigin({ name: encounter.getName(), myEncounter: encounter.isMyEncounter() });
			setDraft({ plan: encounter.getPlan(), targets: encounter.getTargets() });
			setSelected(undefined);
		}
	} else if (draft.targets !== targets) {
		// The store remaps its own plan when targets move; the draft follows the same way.
		setDraft({ plan: remapPlanTargets(draft.plan, draft.targets, targets), targets });
	}
	const issues = useMemo(() => validatePlan(draft.plan, targets.length), [draft.plan, targets.length]);

	// A preset or an entry picked here writes its own targets and plan to the store, and the draft takes up that
	// plan in place of its own, against the new targets. The pick is what Save and Delete act on from then. Custom
	// writes nothing, so the draft stays.
	const presetField = {
		...choice,
		config: {
			...choice.config,
			setValue: (subject: Encounter, value: string) => {
				const [plan, targets] = [encounter.getPlan(), encounter.getTargets()];
				choice.config.setValue(subject, value);
				if (encounter.getPlan() === plan && encounter.getTargets() === targets) return;
				setDraft({ plan: encounter.getPlan(), targets: encounter.getTargets() });
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
				encounter.setPlan(snapshot.plan);
			});
		}
		onOpenChange(nextOpen);
	};

	const newAdd = () => Target.clone(defaultWaveTemplate(presetTargets, encounter.sim.raid.getTanks().length) ?? Encounter.defaultTargetProto());

	const add = (kind: PlanEventKind, targetIndex?: number) => {
		const next = EncounterPlan.clone(draft.plan);
		const ref = addEvent(next, kind, targetIndex);
		// The sim spawns a wave's adds from their templates, so a new wave starts with one add from the default template.
		if (kind === 'wave') next.addWaves[ref.index].adds = [newAdd()];
		setDraft(d => ({ ...d, plan: next }));
		setSelected(ref);
	};

	const selectTarget = (index: number) => setSelected({ kind: 'target', index });
	const appendTarget = (target: Target) => {
		targetList.append({
			...target,
			name: uniqueEncounterName(
				target.name,
				targets.map(({ name }) => name),
			),
		});
		selectTarget(targets.length);
	};
	const addTarget = () => appendTarget(TARGET_LIST.newItem());
	const selectedIndex = selected?.kind === 'target' ? selected.index : undefined;
	const targetFields = useMemo(() => {
		if (selectedIndex === undefined) return undefined;
		const context = encounterTargetContext(encounter, selectedIndex, `encounter-planner-target-${selectedIndex}`);
		return { context, identity: [npcConfig(context), presetTargetConfig(context)] };
	}, [encounter, selectedIndex]);
	const editTarget = (index: number, target: Target, { context, identity }: NonNullable<typeof targetFields>): PlannerTargetEditor => ({
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
		onCopy: () => appendTarget(Target.clone(target)),
		onRemove: () => {
			targetList.remove(index);
			setSelected(undefined);
		},
	});
	const selectedTarget = selectedIndex === undefined ? undefined : targets[selectedIndex];

	// The entry the planner edits: the encounter was marked as one when it opened or was picked in the header, under
	// the name it had then.
	const entry = origin?.myEncounter && !replaced ? myEncounters.entries.find(entry => entry.name === origin.name) : undefined;

	// The list holds an entry without the mark, which the encounter takes once it is stored.
	const store = (replacing?: string) => {
		batch(() => {
			encounter.setPlan(draft.plan);
			const taken = myEncounters.entries.map(entry => entry.name).filter(name => name !== replacing);
			encounter.setName(uniqueEncounterName(encounter.getName() || i18n.t('settings_tab.encounter.new_encounter'), taken));
			encounter.setMyEncounter(true);
		});
		myEncounters.save(encounter.getName(), { ...encounter.toProto(), myEncounter: false }, replacing);
		onOpenChange(false);
	};

	// Save updates the entry the planner edits and adds the fight to My encounters when it was started from the +, has a
	// name (a typed one counts) or holds an event in any lane, under a free name when an entry has it already. A simple
	// fight without a name (a preset, or one edited by hand) only goes to the sim.
	const save = () => {
		encounter.setPlan(draft.plan);
		if (entry || replaced || encounter.getName() || encounter.hasPlan()) return store(entry?.name);
		onOpenChange(false);
	};

	const deleteEntry = () => {
		myEncounters.remove(entry!.name);
		encounter.fromProto(host.individualConfig.defaults.encounter?.encounter ?? Encounter.defaultEncounterProto());
		onOpenChange(false);
	};

	return (
		<EncounterPlannerView
			open={open}
			onOpenChange={close}
			name={<ConnectedStringField modObject={encounter} config={nameField} />}
			duration={duration.map(config => (
				<ConnectedNumberField key={config.id} modObject={encounter} config={config} />
			))}
			preset={<ConnectedChoiceField {...presetField} />}
			execute={showExecuteProportion && execute.map(config => <ConnectedNumberField key={config.id} modObject={encounter} config={config} />)}
			newAdd={newAdd}
			plan={draft.plan}
			baseSeconds={baseSeconds}
			selected={selected}
			issues={issues}
			targetNames={targetNames}
			presetTargets={presetTargets}
			targetEditor={selected?.kind === 'target' && selectedTarget && targetFields ? editTarget(selected.index, selectedTarget, targetFields) : undefined}
			panelRef={panelRef}
			onSelect={setSelected}
			onActivate={() => panelRef.current?.focus()}
			onChange={plan => setDraft(d => ({ ...d, plan }))}
			onAdd={add}
			onAddTarget={addTarget}
			onSave={save}
			onSaveCopy={replaced ? undefined : () => store()}
			onDelete={entry && deleteEntry}
			deleteName={entry?.name}
		/>
	);
};
