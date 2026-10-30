import type { InputConfig } from '@features/inputs/input';
import type { APLValue } from '@generated/proto/apl';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';

type DoAtConfig = InputConfig<Player<any>, APLValue | undefined>;

/** The Do At picker a pre-pull row leads with, in the Rotation tab and in a drawn APL entry. */
export const doAtConfig = (getValue: DoAtConfig['getValue'], setValue: DoAtConfig['setValue']): DoAtConfig => ({
	label: i18n.t('rotation_tab.apl.prepull_actions.do_at.label'),
	labelTooltip: i18n.t('rotation_tab.apl.prepull_actions.do_at.tooltip'),
	extraClassNames: ['w-auto'],
	layout: 'inline',
	getValue,
	setValue,
});
