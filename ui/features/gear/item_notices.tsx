import { Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateAreaType, translateStat } from '@i18n/localization';
import {
	MISSING_ENCHANT_EFFECTS,
	MISSING_ITEM_EFFECTS,
	UNVALIDATED_RATE_ENCHANTS,
	UNVALIDATED_RATE_ITEMS,
	UNVALIDATED_RATE_SET_BONUSES,
	type UnvalidatedRate,
} from '@sim/constants/missing_effects_auto_gen';
import type { Database } from '@sim/proto/database';
import { LocaleHtml } from '@ui-kit/Tooltip';
import type { ReactNode } from 'react';

export type ItemNoticeData = {
	// SpecUnknown is used as default and should always be present.
	// A falsy value disables the notice for that spec.
	[Spec.SpecUnknown]: ReactNode;
} & Record<number, ReactNode>;

// Keys are item counts for each set bonus (typically 2 and 4), values are the
// notice that should be displayed for each bonus. If null, will default to
// GENERIC_MISSING_SET_BONUS_NOTICE_DATA.
export type SetBonusNoticeData = Map<number, ReactNode> | null;

const WantToHelpMessage = () => <p className="mb-0">{i18n.t('gear_tab.item_notices.want_to_help')}</p>;

const MissingRandomSuffixWarning = () => <p className="mb-0">{i18n.t('gear_tab.item_notices.missing_random_suffix')}</p>;

export const MISSING_RANDOM_SUFFIX_WARNING = <MissingRandomSuffixWarning />;

const MissingEffectNotice = ({ effectKind, tooltips }: { effectKind: 'item' | 'enchant'; tooltips: string[] }) =>
	!tooltips.length ? (
		<>
			<p className="font-bold">
				{effectKind === 'item' ? i18n.t('gear_tab.item_notices.missing_effect.item') : i18n.t('gear_tab.item_notices.missing_effect.enchant')}
			</p>
			<p>{i18n.t('gear_tab.item_notices.missing_effect.working_on_it')}</p>
			<WantToHelpMessage />
		</>
	) : (
		<>
			<p className="font-bold">
				{effectKind === 'item'
					? i18n.t('gear_tab.item_notices.missing_effect.item_listed')
					: i18n.t('gear_tab.item_notices.missing_effect.enchant_listed')}
			</p>
			<ul>
				{tooltips
					.filter(tooltip => !!tooltip)
					.map(tooltip => (
						<li key={tooltip}>{tooltip}</li>
					))}
			</ul>
		</>
	);

const missingEffectNotice = (effectKind: 'item' | 'enchant', tooltips: string[]): ItemNoticeData => ({
	[Spec.SpecUnknown]: <MissingEffectNotice effectKind={effectKind} tooltips={tooltips} />,
});

const TentativeImplementationWarning = () => (
	<>
		<p>
			<LocaleHtml html={i18n.t('gear_tab.item_notices.tentative_implementation')} />
		</p>
		<WantToHelpMessage />
	</>
);

const TENTATIVE_IMPLEMENTATION_WARNING = <TentativeImplementationWarning />;

const WillNotBeImplementedWarning = () => <>{i18n.t('gear_tab.item_notices.will_not_be_implemented')}</>;

const WILL_NOT_BE_IMPLEMENTED_WARNING = <WillNotBeImplementedWarning />;

const unvalidatedRateText = ({ kind, origin }: UnvalidatedRate): string => {
	if (kind === 'ppm') {
		if (origin === 'vanilla') return i18n.t('gear_tab.item_notices.unvalidated_rate.ppm_vanilla');
		if (origin === 'unknown') return i18n.t('gear_tab.item_notices.unvalidated_rate.ppm_unknown');
		return i18n.t('gear_tab.item_notices.unvalidated_rate.ppm_tbc');
	}
	if (origin === 'vanilla') return i18n.t('gear_tab.item_notices.unvalidated_rate.chance_vanilla');
	if (origin === 'unknown') return i18n.t('gear_tab.item_notices.unvalidated_rate.chance_unknown');
	return i18n.t('gear_tab.item_notices.unvalidated_rate.chance_tbc');
};

const UnvalidatedRateText = ({ rate }: { rate: UnvalidatedRate }) => <>{unvalidatedRateText(rate)}</>;

// Gear with an unvalidated rate can also miss an effect, so the rate warning joins the notice it already has.
const addUnvalidatedRateNotices = (notices: Map<number, ItemNoticeData>, rates: Map<number, UnvalidatedRate>) => {
	for (const [id, rate] of rates) {
		const notice = notices.get(id);
		notices.set(id, {
			...notice,
			[Spec.SpecUnknown]: (
				<>
					{notice?.[Spec.SpecUnknown]}
					<p className="mb-0">
						<UnvalidatedRateText rate={rate} />
					</p>
				</>
			),
		});
	}
};

const WILL_NOT_BE_IMPLEMENTED_ITEMS: number[] = [];

const TENTATIVE_IMPLEMENTATION_ITEMS: number[] = [];

export const ITEM_NOTICES = new Map<number, ItemNoticeData>([
	...WILL_NOT_BE_IMPLEMENTED_ITEMS.map((itemID): [number, ItemNoticeData] => [
		itemID,
		{
			[Spec.SpecUnknown]: WILL_NOT_BE_IMPLEMENTED_WARNING,
		},
	]),
	...TENTATIVE_IMPLEMENTATION_ITEMS.map((itemID): [number, ItemNoticeData] => [
		itemID,
		{
			[Spec.SpecUnknown]: TENTATIVE_IMPLEMENTATION_WARNING,
		},
	]),
	...[...MISSING_ITEM_EFFECTS].map(([itemID, tooltips]): [number, ItemNoticeData] => [itemID, missingEffectNotice('item', tooltips)]),
]);

addUnvalidatedRateNotices(ITEM_NOTICES, UNVALIDATED_RATE_ITEMS);

export const ENCHANT_NOTICES = new Map<number, ItemNoticeData>(
	[...MISSING_ENCHANT_EFFECTS].map(([effectID, tooltips]): [number, ItemNoticeData] => [effectID, missingEffectNotice('enchant', tooltips)]),
);

addUnvalidatedRateNotices(ENCHANT_NOTICES, UNVALIDATED_RATE_ENCHANTS);

const NotYetImplemented = () => <>{i18n.t('gear_tab.item_notices.not_yet_implemented')}</>;

export const GENERIC_MISSING_SET_BONUS_NOTICE_DATA = new Map<number, ReactNode>([
	[2, <NotYetImplemented key={2} />],
	[4, <NotYetImplemented key={4} />],
]);

const SET_BONUS_NOTICES = new Map<number, SetBonusNoticeData>(
	[...UNVALIDATED_RATE_SET_BONUSES].map(([setID, bonuses]): [number, SetBonusNoticeData] => [
		setID,
		new Map([...bonuses].map(([pieceCount, rate]): [number, ReactNode] => [pieceCount, <UnvalidatedRateText key={pieceCount} rate={rate} />])),
	]),
);

export const registerSetBonusNotices = (db: Database) => {
	SET_BONUS_NOTICES.forEach((value: SetBonusNoticeData, key: number) => {
		const noticeData = value || GENERIC_MISSING_SET_BONUS_NOTICE_DATA;
		const noticeContent = (
			<>
				<p className="mb-1"> {i18n.t('gear_tab.item_notices.set_bonus_warnings')}</p>
				<ul className="mb-0">
					{Array.from(noticeData.keys()).map(pieceCount => (
						<li key={pieceCount}>
							{i18n.t('gear_tab.item_notices.set_bonus_piece', { pieces: pieceCount.toFixed(0) })} {noticeData.get(pieceCount)}
						</li>
					))}
				</ul>
			</>
		);

		for (const id of db.getItemIdsForSet(key)) {
			ITEM_NOTICES.set(id, { [Spec.SpecUnknown]: noticeContent });
		}
	});
};

// Informational notices, shown behind an info icon rather than a warning.
export const ITEM_INFO_NOTICES = new Map<number, ReactNode>();

// Wowhead's tooltip shows the folded number; this names the area-only part.
export const registerAreaStatsNotices = (db: Database) => {
	for (const item of db.getAllItems()) {
		const areaStats = item.scalingOptions?.[0]?.areaStats ?? [];
		if (!areaStats.length) continue;

		ITEM_INFO_NOTICES.set(
			item.id,
			<>
				<p className="mb-1">{i18n.t('gear_tab.item_notices.area_stats')}</p>
				<ul className="mb-0">
					{areaStats.map(bonus => (
						<li key={bonus.areaType}>
							{translateAreaType(bonus.areaType)}:{' '}
							{Object.entries(bonus.stats)
								.map(([stat, value]) => `+${value} ${translateStat(Number(stat))}`)
								.join(', ')}
						</li>
					))}
				</ul>
			</>,
		);
	}
};
