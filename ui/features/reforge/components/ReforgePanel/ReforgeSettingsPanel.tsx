import { ConnectedBooleanPicker } from '@features/inputs/components/ConnectedBooleanPicker';
import type { ReforgeOptimizerModel, ReforgeOptimizerOptions } from '@features/reforge/model/reforge_optimizer';
import { SavedEpWeights } from '@features/stat-weights/components/SavedEpWeights';
import { useOpenEpWeights } from '@features/stat-weights/hooks/useEpWeightsDialog';
import i18n from '@i18n/config';
import { useSimHost, useSpecConfig } from '@sim/context/SimHostContext';
import { Button } from '@ui-kit/Button';
import { type ReactNode, useMemo } from 'react';

import { trackEvent } from '../../../../tracking/utils';
import { useReforgeField } from '../../hooks/useReforgeField';
import { DEFAULT_REFORGE_ID_PREFIX, ReforgeIdPrefixContext } from './context/ReforgeIdPrefixContext';
import { ReforgeBreakpointLimits } from './ReforgeBreakpointLimits';
import { ReforgeFrozenSlots } from './ReforgeFrozenSlots';
import { ReforgeStatCaps } from './ReforgeStatCaps';
import { buildStatTooltips } from './utils';

export interface ReforgeSettingsPanelViewProps {
	customEpToggle: ReactNode;
	/** How to edit the weights, for while the custom EP values are off. */
	showEpHelp: boolean;
	statCaps: ReactNode;
	softCaps: ReactNode;
	freezeSlots: ReactNode;
	savedEpWeights: ReactNode;
	onEditWeights: () => void;
}

export const ReforgeSettingsPanelView = ({
	customEpToggle,
	showEpHelp,
	statCaps,
	softCaps,
	freezeSlots,
	savedEpWeights,
	onEditWeights,
}: ReforgeSettingsPanelViewProps) => (
	<>
		{customEpToggle}
		{showEpHelp && (
			<div className="mb-0">
				<p>{i18n.t('sidebar.buttons.suggest_reforges.enable_modification')}</p>
				<p>{i18n.t('sidebar.buttons.suggest_reforges.modify_in_editor')}</p>
				<p>{i18n.t('sidebar.buttons.suggest_reforges.hard_cap_info')}</p>
			</div>
		)}
		{statCaps}
		{softCaps}
		{freezeSlots}
		{savedEpWeights}
		<Button variant="outline-primary" className="mt-2" data-testid="reforge-edit-weights" onClick={onEditWeights}>
			{i18n.t('sidebar.buttons.suggest_reforges.edit_weights')}
		</Button>
	</>
);

export interface ReforgeSettingsPanelProps {
	model: ReforgeOptimizerModel;
	options?: ReforgeOptimizerOptions;
	/** Closes the popover when the EP weights dialog opens over it. Omitted where the panel is inline and has nothing to close. */
	onClose?: () => void;
	/** Scopes every DOM id, so a second mount does not collide with the sidebar popover's. */
	idPrefix?: string;
}

/** The reforge settings body, shared by the sidebar popover and the Batch Sim sidebar. In the popover it is mounted only while that is open, so every section is built per open. */
export const ReforgeSettingsPanel = ({ model, options, onClose, idPrefix = DEFAULT_REFORGE_ID_PREFIX }: ReforgeSettingsPanelProps) => {
	const host = useSimHost();
	const player = host.player;
	const individualConfig = useSpecConfig();
	const openEpWeights = useOpenEpWeights();
	const settings = model.settings;

	const useCustomEPValues = useReforgeField(settings, 'useCustomEPValues', () => settings.useCustomEPValues);
	const useSoftCapBreakpoints = useReforgeField(settings, 'useSoftCapBreakpoints', () => settings.useSoftCapBreakpoints);
	const freezeItemSlots = useReforgeField(settings, 'freezeItemSlots', () => settings.freezeItemSlots);

	const statTooltips = useMemo(() => buildStatTooltips(options?.statTooltips), [options]);
	const softCapsConfig = model.softCapsConfig;
	const hasSoftCaps = !!softCapsConfig?.length;

	return (
		<ReforgeIdPrefixContext.Provider value={idPrefix}>
			<ReforgeSettingsPanelView
				customEpToggle={
					<ConnectedBooleanPicker
						modObject={player}
						config={{
							extraClassNames: ['mb-2'],
							id: `${idPrefix}-enable-custom-ep-weights`,
							label: i18n.t('sidebar.buttons.suggest_reforges.use_custom'),
							layout: 'inline',
							storeField: 'reforge:useCustomEPValues',
							getValue: () => settings.useCustomEPValues,
							setValue: (_player, newValue) => {
								trackEvent({ action: 'settings', category: 'reforging', label: 'use_custom_ep', value: newValue });
								settings.setUseCustomEPValues(newValue);
							},
						}}
					/>
				}
				showEpHelp={!useCustomEPValues}
				statCaps={
					<ReforgeStatCaps
						model={model}
						player={player}
						displayStats={individualConfig.displayStats}
						statTooltips={statTooltips}
						useCustomEPValues={useCustomEPValues}
					/>
				}
				softCaps={
					hasSoftCaps && (
						<>
							<ConnectedBooleanPicker
								modObject={player}
								config={{
									extraClassNames: ['mb-2'],
									id: `${idPrefix}-enable-soft-cap-breakpoints`,
									label: i18n.t('sidebar.buttons.suggest_reforges.use_soft_cap_breakpoints'),
									layout: 'inline',
									storeField: 'reforge:useSoftCapBreakpoints',
									getValue: () => settings.useSoftCapBreakpoints,
									setValue: (_player, newValue) => {
										trackEvent({ action: 'settings', category: 'reforging', label: 'softcap_breakpoints', value: newValue });
										settings.setUseSoftCapBreakpoints(newValue);
									},
								}}
							/>
							{model.enableBreakpointLimits && (
								<ReforgeBreakpointLimits
									settings={settings}
									softCapsConfig={softCapsConfig}
									player={player}
									useSoftCapBreakpoints={useSoftCapBreakpoints}
								/>
							)}
						</>
					)
				}
				freezeSlots={
					<>
						<ConnectedBooleanPicker
							modObject={player}
							config={{
								extraClassNames: ['mb-2'],
								id: `${idPrefix}-freeze-item-slots`,
								label: i18n.t('sidebar.buttons.suggest_reforges.freeze_item_slots'),
								labelTooltip: i18n.t('sidebar.buttons.suggest_reforges.freeze_item_slots_tooltip'),
								layout: 'inline',
								storeField: 'reforge:freezeItemSlots',
								getValue: () => settings.freezeItemSlots,
								setValue: (_player, newValue) => {
									trackEvent({ action: 'settings', category: 'reforging', label: 'freeze_item_slots', value: newValue });
									settings.setFreezeItemSlots(newValue);
								},
							}}
						/>
						<ReforgeFrozenSlots settings={settings} player={player} freezeItemSlots={freezeItemSlots} />
					</>
				}
				savedEpWeights={<SavedEpWeights className="mt-4" loadOnly presetsOnly={!useCustomEPValues} />}
				onEditWeights={() => {
					openEpWeights();
					onClose?.();
				}}
			/>
		</ReforgeIdPrefixContext.Provider>
	);
};
