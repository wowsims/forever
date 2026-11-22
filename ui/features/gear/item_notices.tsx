import { Spec } from '@generated/proto/common';
import { type ProcRate, ProcRateKind, ProcRateStatus } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { translateAreaType, translateStat } from '@i18n/localization';
import { MISSING_ENCHANT_EFFECTS, MISSING_ITEM_EFFECTS, MISSING_SET_BONUSES } from '@sim/constants/missing_effects_auto_gen';
import type { Database } from '@sim/proto/database';
import { signedStatAmount } from '@sim/proto/stats';
import { LocaleHtml } from '@ui-kit/Tooltip';
import { Fragment, type ReactNode } from 'react';

export type ItemNoticeData = {
	// SpecUnknown is used as default and should always be present.
	// A falsy value disables the notice for that spec.
	[Spec.SpecUnknown]: ReactNode;
} & Record<number, ReactNode>;

const WantToHelpMessage = () => <p className="mb-0">{i18n.t('gear_tab.item_notices.want_to_help')}</p>;

const MissingRandomSuffixWarning = () => <p className="mb-0">{i18n.t('gear_tab.item_notices.missing_random_suffix')}</p>;

export const MISSING_RANDOM_SUFFIX_WARNING = <MissingRandomSuffixWarning />;

const BestRandomSuffixInfo = () => <p className="mb-0">{i18n.t('gear_tab.item_notices.best_random_suffix')}</p>;

export const BEST_RANDOM_SUFFIX_INFO = <BestRandomSuffixInfo />;

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

// Up to 2 decimals without trailing zeros.
const formatRate = (rate: number): string => String(Math.round(rate * 100) / 100);

// A measured rate and one the game files state show as a line; any other is a warning.
const isKnownRate = ({ status }: ProcRate): boolean => status === ProcRateStatus.ProcRateStatusMeasured || status === ProcRateStatus.ProcRateStatusGameFiles;

// The sim's rate is stated where the gear has a single one: procs per minute for a ppm, a percent for a chance.
const unvalidatedRateText = ({ kind, rate, status }: ProcRate): string => {
	const ppm = kind === ProcRateKind.ProcRateKindPpm;
	const vanilla = status === ProcRateStatus.ProcRateStatusVanillaCarryover;
	const observed = status === ProcRateStatus.ProcRateStatusObserved;
	if (rate > 0) {
		const options = { rate: formatRate(rate) };
		if (ppm) {
			if (vanilla) return i18n.t('gear_tab.item_notices.unvalidated_rate.ppm_vanilla_rate', options);
			if (observed) return i18n.t('gear_tab.item_notices.unvalidated_rate.ppm_observed_rate', options);
			return i18n.t('gear_tab.item_notices.unvalidated_rate.ppm_tbc_rate', options);
		}
		if (vanilla) return i18n.t('gear_tab.item_notices.unvalidated_rate.chance_vanilla_rate', options);
		if (observed) return i18n.t('gear_tab.item_notices.unvalidated_rate.chance_observed_rate', options);
		return i18n.t('gear_tab.item_notices.unvalidated_rate.chance_tbc_rate', options);
	}
	if (ppm) {
		if (vanilla) return i18n.t('gear_tab.item_notices.unvalidated_rate.ppm_vanilla');
		if (observed) return i18n.t('gear_tab.item_notices.unvalidated_rate.ppm_observed');
		return i18n.t('gear_tab.item_notices.unvalidated_rate.ppm_tbc');
	}
	if (vanilla) return i18n.t('gear_tab.item_notices.unvalidated_rate.chance_vanilla');
	if (observed) return i18n.t('gear_tab.item_notices.unvalidated_rate.chance_observed');
	return i18n.t('gear_tab.item_notices.unvalidated_rate.chance_tbc');
};

const knownRateText = ({ kind, rate }: ProcRate): string =>
	kind === ProcRateKind.ProcRateKindPpm
		? i18n.t('gear_tab.item_notices.proc_rate', { rate: rate.toFixed(2) })
		: i18n.t('gear_tab.item_notices.proc_chance', { rate: formatRate(rate) });

const KnownRateLine = ({ rate, prefix }: { rate: ProcRate; prefix?: string }) => (
	<p className="mb-0">
		{prefix && `${prefix} `}
		{knownRateText(rate)}
	</p>
);

const UnvalidatedRateText = ({ rate }: { rate: ProcRate }) => <>{unvalidatedRateText(rate)}</>;

// Gear with an unvalidated rate can also miss an effect, so the rate warning joins the notice it already has.
const addUnvalidatedRateNotice = (notices: Map<number, ItemNoticeData>, id: number, rate: ProcRate) => {
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
};

// Each item's own rate line.
const OWN_ITEM_RATE_NOTICES = new Map<number, ReactNode>();

// Each item's rate lines: its own, then those of the set bonuses it belongs to.
export const ITEM_RATE_NOTICES = new Map<number, ReactNode>();

// By enchant effect id.
export const ENCHANT_RATE_NOTICES = new Map<number, ReactNode>();

const WILL_NOT_BE_IMPLEMENTED_ITEMS: number[] = [];

const TENTATIVE_IMPLEMENTATION_ITEMS: number[] = [];

const ZumrahsVexingCaneNotice = () => <p className="mb-0">{i18n.t('gear_tab.item_notices.partial_item.zumrahs_vexing_cane')}</p>;

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
	[18082, { [Spec.SpecUnknown]: <ZumrahsVexingCaneNotice /> }], // The cane's zombie adds below 1 DPS (EV-0124).
	...[...MISSING_ITEM_EFFECTS].map(([itemID, tooltips]): [number, ItemNoticeData] => [itemID, missingEffectNotice('item', tooltips)]),
]);

export const ENCHANT_NOTICES = new Map<number, ItemNoticeData>(
	[...MISSING_ENCHANT_EFFECTS].map(([effectID, tooltips]): [number, ItemNoticeData] => [effectID, missingEffectNotice('enchant', tooltips)]),
);

// The rates the database carries on items and enchants: a known rate as a line, an unvalidated one as a warning
// after the notice the gear already has. Runs before registerSetBonusNotices, which adds each piece's set rates.
export const registerProcRateNotices = (db: Database) => {
	for (const item of db.getAllItems()) {
		if (!item.procRate) continue;
		if (isKnownRate(item.procRate)) {
			const line = <KnownRateLine rate={item.procRate} />;
			OWN_ITEM_RATE_NOTICES.set(item.id, line);
			ITEM_RATE_NOTICES.set(item.id, line);
		} else {
			addUnvalidatedRateNotice(ITEM_NOTICES, item.id, item.procRate);
		}
	}

	// Every enchant sharing an effect id carries its rate.
	const enchantRates = new Map<number, ProcRate>();
	for (const enchant of db.getAllEnchants()) {
		if (enchant.procRate) enchantRates.set(enchant.effectId, enchant.procRate);
	}
	for (const [effectID, rate] of enchantRates) {
		if (isKnownRate(rate)) ENCHANT_RATE_NOTICES.set(effectID, <KnownRateLine rate={rate} />);
		else addUnvalidatedRateNotice(ENCHANT_NOTICES, effectID, rate);
	}
};

const NotYetImplementedBonus = ({ tooltips }: { tooltips: string[] }) => (
	<>
		{i18n.t('gear_tab.item_notices.not_yet_implemented')}
		<ul className="mb-0">
			{tooltips
				.filter(tooltip => !!tooltip)
				.map(tooltip => (
					<li key={tooltip}>{tooltip}</li>
				))}
		</ul>
	</>
);

const PartlyImplementedBonus = ({ text }: { text: () => string }) => (
	<>
		{i18n.t('gear_tab.item_notices.not_fully_supported')} {text()}
	</>
);

// Bonuses the sim registers but models only in part, by set id and piece count.
const PARTLY_IMPLEMENTED_SET_BONUSES = new Map<number, Map<number, () => string>>([
	[185, new Map([[5, () => i18n.t('gear_tab.item_notices.partial_set_bonus.wildheart_raiment_5')]])],
	[500, new Map([[3, () => i18n.t('gear_tab.item_notices.partial_set_bonus.implements_of_unspoken_names_3')]])],
	[
		523,
		new Map([
			[4, () => i18n.t('gear_tab.item_notices.partial_set_bonus.dreadnaughts_battlegear_4')],
			[8, () => i18n.t('gear_tab.item_notices.partial_set_bonus.dreadnaughts_battlegear_8')],
		]),
	],
	[2103, new Map([[3, () => i18n.t('gear_tab.item_notices.partial_set_bonus.battleplate_of_glory_3')]])],
	[2114, new Map([[4, () => i18n.t('gear_tab.item_notices.partial_set_bonus.grovekeeper_eclipse_4')]])],
	[2131, new Map([[2, () => i18n.t('gear_tab.item_notices.partial_set_bonus.partners_in_crime_2')]])],
]);

const UntestedBonus = ({ text }: { text: () => string }) => (
	<>
		<p className="mb-0">{i18n.t('gear_tab.item_notices.ingame_validation_needed')}</p>
		<p className="mb-0">{text()}</p>
	</>
);

// Bonuses built on a reading no in-game test has confirmed yet, by set id and piece count.
const UNTESTED_SET_BONUSES = new Map<number, Map<number, () => string>>([
	[65, new Map([[2, () => i18n.t('gear_tab.item_notices.untested_set_bonus.spiders_kiss_2')]])],
	[122, new Map([[3, () => i18n.t('gear_tab.item_notices.untested_set_bonus.necropile_raiment_3')]])],
	[123, new Map([[3, () => i18n.t('gear_tab.item_notices.untested_set_bonus.bloodmail_regalia_3')]])],
	[124, new Map([[5, () => i18n.t('gear_tab.item_notices.untested_set_bonus.deathbone_guardian_5')]])],
	[494, new Map([[3, () => i18n.t('gear_tab.item_notices.untested_set_bonus.symbols_of_unending_life_3')]])],
	[501, new Map([[3, () => i18n.t('gear_tab.item_notices.untested_set_bonus.stormcallers_garb_3')]])],
	[502, new Map([[3, () => i18n.t('gear_tab.item_notices.untested_set_bonus.gift_of_the_gathering_storm_3')]])],
	[506, new Map([[3, () => i18n.t('gear_tab.item_notices.untested_set_bonus.battlegear_of_eternal_justice_3')]])],
	[510, new Map([[3, () => i18n.t('gear_tab.item_notices.untested_set_bonus.trappings_of_the_unseen_path_3')]])],
	[
		524,
		new Map([
			[4, () => i18n.t('gear_tab.item_notices.untested_set_bonus.bonescythe_armor_4')],
			[8, () => i18n.t('gear_tab.item_notices.untested_set_bonus.bonescythe_armor_8')],
		]),
	],
	[527, new Map([[8, () => i18n.t('gear_tab.item_notices.untested_set_bonus.the_earthshatterer_8')]])],
	[2071, new Map([[2, () => i18n.t('gear_tab.item_notices.untested_set_bonus.teachings_of_the_furbolgs_2')]])],
	[2098, new Map([[5, () => i18n.t('gear_tab.item_notices.untested_set_bonus.manaflare_regalia_5')]])],
	[2100, new Map([[5, () => i18n.t('gear_tab.item_notices.untested_set_bonus.demonheart_raiment_5')]])],
	[2133, new Map([[4, () => i18n.t('gear_tab.item_notices.untested_set_bonus.rotmenders_raiment_4')]])],
	[2134, new Map([[5, () => i18n.t('gear_tab.item_notices.untested_set_bonus.violet_sorcerers_vestments_5')]])],
	[2136, new Map([[5, () => i18n.t('gear_tab.item_notices.untested_set_bonus.defias_enforcers_garb_5')]])],
]);

const addSetBonusNotices = <T,>(
	setBonusNotices: Map<number, Map<number, ReactNode[]>>,
	bonusesBySet: Map<number, Map<number, T>>,
	notice: (value: T) => ReactNode,
) => {
	for (const [setID, bonuses] of bonusesBySet) {
		const notices = setBonusNotices.get(setID) ?? new Map<number, ReactNode[]>();
		for (const [pieceCount, value] of bonuses) {
			notices.set(pieceCount, [...(notices.get(pieceCount) ?? []), notice(value)]);
		}
		setBonusNotices.set(setID, notices);
	}
};

// A set piece can also miss an item effect or carry a rate of its own, so the set notice joins the
// notice it already has. Each set bonus rate also goes after the piece's own rate line in ITEM_RATE_NOTICES.
export const registerSetBonusNotices = (db: Database) => {
	const itemIdsBySet = new Map<number, number[]>();
	for (const id of db.getAllItemIds()) {
		const setID = db.getItemById(id)!.setId;
		if (setID) itemIdsBySet.set(setID, [...(itemIdsBySet.get(setID) ?? []), id]);
	}

	// By set id, then by piece count.
	const knownRates = new Map<number, Map<number, ProcRate>>();
	const unvalidatedRates = new Map<number, Map<number, ProcRate>>();
	for (const { setId, pieces, rate } of db.getSetBonusProcRates()) {
		if (!rate) continue;
		const rates = isKnownRate(rate) ? knownRates : unvalidatedRates;
		rates.set(setId, (rates.get(setId) ?? new Map<number, ProcRate>()).set(pieces, rate));
	}

	// By set id, then by piece count: what each bonus leaves out, a missing effect, a partial one, an
	// untested reading or an unvalidated rate, in that order.
	const setBonusNotices = new Map<number, Map<number, ReactNode[]>>();
	addSetBonusNotices(setBonusNotices, MISSING_SET_BONUSES, tooltips => <NotYetImplementedBonus tooltips={tooltips} />);
	addSetBonusNotices(setBonusNotices, PARTLY_IMPLEMENTED_SET_BONUSES, text => <PartlyImplementedBonus text={text} />);
	addSetBonusNotices(setBonusNotices, UNTESTED_SET_BONUSES, text => <UntestedBonus text={text} />);
	addSetBonusNotices(setBonusNotices, unvalidatedRates, rate => <UnvalidatedRateText rate={rate} />);

	knownRates.forEach((setRates, setID) => {
		const lines = [...setRates]
			.sort(([a], [b]) => a - b)
			.map(([pieceCount, rate]) => (
				<KnownRateLine key={pieceCount} rate={rate} prefix={i18n.t('gear_tab.item_notices.set_bonus_piece', { pieces: pieceCount.toFixed(0) })} />
			));

		for (const id of itemIdsBySet.get(setID) ?? []) {
			ITEM_RATE_NOTICES.set(
				id,
				<>
					{OWN_ITEM_RATE_NOTICES.get(id)}
					{lines}
				</>,
			);
		}
	});

	setBonusNotices.forEach((bonuses, setID) => {
		const noticeContent = (
			<>
				<p className="mb-1"> {i18n.t('gear_tab.item_notices.set_bonus_warnings')}</p>
				<ul className="mb-0">
					{[...bonuses.keys()]
						.sort((a, b) => a - b)
						.map(pieceCount => (
							<li key={pieceCount}>
								{i18n.t('gear_tab.item_notices.set_bonus_piece', { pieces: pieceCount.toFixed(0) })}{' '}
								{bonuses.get(pieceCount)?.map((notice, index) => (
									<Fragment key={index}>
										{index > 0 && ' '}
										{notice}
									</Fragment>
								))}
							</li>
						))}
				</ul>
			</>
		);

		for (const id of itemIdsBySet.get(setID) ?? []) {
			const notice = ITEM_NOTICES.get(id);
			ITEM_NOTICES.set(id, {
				...notice,
				[Spec.SpecUnknown]: (
					<>
						{notice?.[Spec.SpecUnknown]}
						{noticeContent}
					</>
				),
			});
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
								.map(([stat, value]) => `${signedStatAmount(Number(stat), value)} ${translateStat(Number(stat))}`)
								.join(', ')}
						</li>
					))}
				</ul>
			</>,
		);
	}
};
