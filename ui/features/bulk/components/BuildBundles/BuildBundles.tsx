import { EnchantPresetDialog } from '@features/enchant-presets/components/EnchantPresets/EnchantPresetDialog';
import type { BulkBuildBundle } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { type ReactNode, type Ref, useEffect, useMemo, useRef, useState } from 'react';

import { useBulkBuildsPlan } from '../../context/BulkBuildsContext';
import { useBulkEnchantPicks } from '../../hooks/useBulkEnchantPicks';
import {
	addBulkBundle,
	bulkBundleList,
	bundleOptions,
	pruneBulkBundles,
	removeBulkBundle,
	replaceBulkBundle,
	setBulkBundleChecked,
	setBulkCurrentBundleChecked,
} from '../../model/builds/bundles';
import { bulkEnchantPickList, formatEnchantPickSummary, setBulkEnchantPickChecked } from '../../model/builds/enchant_picks';
import { BulkBundleDialog } from '../BulkBundleDialog';
import { BulkBundleList, type BulkBundleListProps } from '../BulkBundleList';
import { BulkCheckList, type BulkCheckRow } from '../BulkCheckList';

export interface BuildBundlesViewProps {
	list: BulkBundleListProps;
	onNew(): void;
	dialog: ReactNode;
	enchantPicks: readonly BulkCheckRow[];
	onEnchantPickCheckedChange(name: string, checked: boolean): void;
	onNewEnchantPreset(): void;
	enchantPresetDialog: ReactNode;
	enchantPicksRef?: Ref<HTMLDivElement>;
}

const enchantPicksText = (key: string) => i18n.t(`bulk_tab.settings.enchant_picks.${key}`);

// Enchant presets are the user's own, so they all list under "Saved"; the presets heading never shows. Deleting a bundle
// unmounts the × that held focus, so focus moves to New bundle. Saving the first enchant preset unmounts New preset, so
// the dialog hands focus to the enchant picks column, whose first tabbable element is the new preset's box, or New preset
// after a cancel.
export const BuildBundlesView = ({
	list,
	onNew,
	dialog,
	enchantPicks,
	onEnchantPickCheckedChange,
	onNewEnchantPreset,
	enchantPresetDialog,
	enchantPicksRef,
}: BuildBundlesViewProps) => {
	const newBundleRef = useRef<HTMLButtonElement>(null);
	return (
		<>
			<div className="grid gap-2">
				<BulkBundleList
					{...list}
					onDelete={key => {
						list.onDelete(key);
						newBundleRef.current?.focus();
					}}
				/>
				<Button ref={newBundleRef} variant="outline-primary" size="sm" className="justify-self-start" onClick={onNew} data-testid="bulk-bundle-new">
					<Icon name="plus" />
					{i18n.t('bulk_tab.builds.bundles.new')}
				</Button>
				{dialog}
			</div>
			<div ref={enchantPicksRef} className="grid gap-2">
				<h6 className="mb-0">{enchantPicksText('heading')}</h6>
				<BulkCheckList
					strings={{ saved: enchantPicksText('saved'), empty: enchantPicksText('empty'), duplicate: enchantPicksText('duplicate') }}
					current={null}
					presets={[]}
					saved={enchantPicks}
					onCheckedChange={(ref, checked) => onEnchantPickCheckedChange(ref.name, checked)}
					testIds={{ list: 'bulk-enchant-picks', row: 'bulk-enchant-pick' }}
				/>
				{!enchantPicks.length && (
					<Button variant="outline-primary" size="sm" className="justify-self-start" onClick={onNewEnchantPreset} data-testid="bulk-enchant-pick-new">
						<Icon name="plus" />
						{i18n.t('bulk_tab.settings.enchant_presets.new')}
					</Button>
				)}
				{enchantPresetDialog}
			</div>
		</>
	);
};

interface DialogState {
	open: boolean;
	/** The stored bundle being edited, null for a new one. */
	bundle: BulkBuildBundle | null;
	// A fresh key per opening starts the dialog's picks from the bundle again.
	key: number;
}

// Renders under a `BulkBuildsProvider` in Builds mode, the only place it shows.
export const BuildBundles = () => {
	const player = usePlayer();
	const { sources, sameAs } = useBulkBuildsPlan()!;
	const { ready, current, parts, bundles, currentChecked } = sources;
	const { entries: enchantPickEntries, checked: checkedEnchantPicks } = useBulkEnchantPicks();
	const [dialog, setDialog] = useState<DialogState>({ open: false, bundle: null, key: 0 });
	const [creatingEnchantPreset, setCreatingEnchantPreset] = useState(false);
	const enchantPicksRef = useRef<HTMLDivElement>(null);

	// Dropping a bundle whose saved part is gone writes the store, so it runs in an effect after render. Until the sim is
	// ready no gear set is listed, and every bundle holding a saved one would read as gone. It runs again on each change to
	// the bundles, so one saved with a pick deleted while the dialog was open drops right away.
	useEffect(() => {
		if (ready) pruneBulkBundles(player, parts);
	}, [player, ready, parts, bundles]);

	const rows = useMemo(() => {
		const list = bulkBundleList(current, parts, bundles, currentChecked);
		const labels = new Map<number, string>([[-1, i18n.t('bulk_tab.builds.bundles.current')], ...list.map(row => [row.index, row.label] as const)]);
		return list.map(row => {
			const earlier = sameAs.get(row.index);
			const sameAsLabel = row.duplicateOf ?? (earlier === undefined ? undefined : labels.get(earlier));
			return {
				key: String(row.index),
				label: row.label,
				checked: row.checked,
				disabledReason: sameAsLabel === undefined ? undefined : i18n.t('bulk_tab.builds.bundles.duplicate', { bundle: sameAsLabel }),
			};
		});
	}, [current, parts, bundles, currentChecked, sameAs]);
	const options = useMemo(() => bundleOptions(parts), [parts]);

	const enchantPickRows = useMemo(
		() =>
			bulkEnchantPickList(enchantPickEntries, checkedEnchantPicks).map((item): BulkCheckRow => ({
				ref: item.ref,
				name: item.name,
				summary: formatEnchantPickSummary(item.preset),
				checked: item.checked,
				duplicate: item.duplicate,
			})),
		[enchantPickEntries, checkedEnchantPicks],
	);

	const openDialog = (bundle: BulkBuildBundle | null) => setDialog(state => ({ open: true, bundle, key: state.key + 1 }));
	// Found by identity, so a prune of an earlier bundle moves the edit along; a prune of this one closes the dialog unsaved.
	const editing = dialog.bundle === null ? null : bundles.indexOf(dialog.bundle);
	const save = (bundle: BulkBuildBundle) => (editing === null ? addBulkBundle(player, bundle) : replaceBulkBundle(player, editing, bundle));

	return (
		<BuildBundlesView
			list={{
				current: {
					label: i18n.t('bulk_tab.builds.bundles.current'),
					checked: currentChecked,
					onCheckedChange: checked => setBulkCurrentBundleChecked(player, checked),
				},
				rows,
				onCheckedChange: (key, checked) => setBulkBundleChecked(player, Number(key), checked),
				onEdit: key => openDialog(bundles[Number(key)]),
				onDelete: key => removeBulkBundle(player, Number(key)),
				testIds: { list: 'bulk-bundles', row: 'bulk-bundle', current: 'bulk-bundle-current' },
			}}
			onNew={() => openDialog(null)}
			dialog={
				dialog.open &&
				editing !== -1 && (
					<BulkBundleDialog
						key={dialog.key}
						open
						bundle={dialog.bundle}
						options={options}
						onSave={save}
						onClose={() => setDialog(state => ({ ...state, open: false }))}
					/>
				)
			}
			enchantPicks={enchantPickRows}
			onEnchantPickCheckedChange={(name, checked) =>
				setBulkEnchantPickChecked(
					player,
					name,
					checked,
					enchantPickEntries.map(entry => entry.name),
				)
			}
			onNewEnchantPreset={() => setCreatingEnchantPreset(true)}
			enchantPresetDialog={
				creatingEnchantPreset && <EnchantPresetDialog open preset={null} onClose={() => setCreatingEnchantPreset(false)} finalFocus={enchantPicksRef} />
			}
			enchantPicksRef={enchantPicksRef}
		/>
	);
};
