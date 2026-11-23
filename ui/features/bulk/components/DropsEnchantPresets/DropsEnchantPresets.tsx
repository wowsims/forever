import { EnchantPresetDialog } from '@features/enchant-presets/components/EnchantPresets/EnchantPresetDialog';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { type ReactNode, type Ref, useMemo, useRef, useState } from 'react';

import { useDropsEnchantPicks } from '../../hooks/useDropsEnchantPicks';
import { bulkEnchantPickList, formatEnchantPickSummary } from '../../model/builds/enchant_picks';
import { setDropsEnchantPresetChecked } from '../../model/drops/settings';
import { BulkCheckList, type BulkCheckRow } from '../BulkCheckList';

export interface DropsEnchantPresetsViewProps {
	rows: readonly BulkCheckRow[];
	onCheckedChange(name: string, checked: boolean): void;
	onNew(): void;
	dialog: ReactNode;
	listRef?: Ref<HTMLDivElement>;
}

const pickText = (key: string) => i18n.t(`bulk_tab.settings.enchant_picks.${key}`);

// Drops keeps its own pick, apart from Gear's ticks. Presets are the user's own, so they all list under "Saved". Saving the
// first preset unmounts New preset, so the dialog hands focus to the list, whose first tabbable element is the new preset's
// box, or New preset after a cancel.
export const DropsEnchantPresetsView = ({ rows, onCheckedChange, onNew, dialog, listRef }: DropsEnchantPresetsViewProps) => (
	<div ref={listRef} className="grid gap-2" data-testid="drops-enchant-presets">
		<BulkCheckList
			strings={{ saved: pickText('saved'), empty: pickText('empty'), duplicate: pickText('duplicate') }}
			current={null}
			presets={[]}
			saved={rows}
			onCheckedChange={(ref, checked) => onCheckedChange(ref.name, checked)}
			testIds={{ list: 'drops-enchant-pick-list', row: 'drops-enchant-pick' }}
		/>
		{!rows.length && (
			<Button variant="outline-primary" size="sm" className="justify-self-start" onClick={onNew} data-testid="drops-enchant-pick-new">
				<Icon name="plus" />
				{i18n.t('bulk_tab.settings.enchant_presets.new')}
			</Button>
		)}
		{dialog}
	</div>
);

export const DropsEnchantPresets = () => {
	const player = usePlayer();
	const { entries, checked } = useDropsEnchantPicks();
	const [creating, setCreating] = useState(false);
	const listRef = useRef<HTMLDivElement>(null);

	const rows = useMemo(
		() =>
			bulkEnchantPickList(entries, checked).map((item): BulkCheckRow => ({
				ref: item.ref,
				name: item.name,
				summary: formatEnchantPickSummary(item.preset),
				checked: item.checked,
				duplicate: item.duplicate,
			})),
		[entries, checked],
	);

	return (
		<DropsEnchantPresetsView
			rows={rows}
			onCheckedChange={(name, isChecked) =>
				setDropsEnchantPresetChecked(
					player,
					name,
					isChecked,
					entries.map(entry => entry.name),
				)
			}
			onNew={() => setCreating(true)}
			dialog={creating && <EnchantPresetDialog open preset={null} onClose={() => setCreating(false)} finalFocus={listRef} />}
			listRef={listRef}
		/>
	);
};
