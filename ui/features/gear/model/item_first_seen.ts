import i18n from '@i18n/config';

const HOUR_MS = 60 * 60 * 1000;
const DAY_MS = 24 * HOUR_MS;

export { isNewlyDiscovered } from '@sim/proto/item_first_seen';

// Local calendar days, so a stamp from late last night reads as yesterday.
const calendarDaysAgo = (date: Date): number => {
	const today = new Date(Date.now());
	today.setHours(0, 0, 0, 0);
	const day = new Date(date);
	day.setHours(0, 0, 0, 0);
	return Math.round((today.getTime() - day.getTime()) / DAY_MS);
};

const ordinalSuffix = (day: number): string => {
	switch (new Intl.PluralRules(i18n.language, { type: 'ordinal' }).select(day)) {
		case 'one':
			return i18n.t('gear_tab.item_notices.ordinal.one');
		case 'two':
			return i18n.t('gear_tab.item_notices.ordinal.two');
		case 'few':
			return i18n.t('gear_tab.item_notices.ordinal.few');
		default:
			return i18n.t('gear_tab.item_notices.ordinal.other');
	}
};

// Whole hours and no minutes: a stamp is when the sim's database first held the item, which can trail the game by hours.
export const discoveredLabel = (firstSeenSeconds: number): string => {
	const date = new Date(firstSeenSeconds * 1000);
	const hoursAgo = Math.floor((Date.now() - date.getTime()) / HOUR_MS);
	if (hoursAgo < 1) return i18n.t('gear_tab.item_notices.discovered_within_hour');
	if (hoursAgo < 24) {
		const when = new Intl.RelativeTimeFormat(i18n.language, { numeric: 'always' }).format(-hoursAgo, 'hour');
		return i18n.t('gear_tab.item_notices.discovered_ago', { when });
	}
	const daysAgo = calendarDaysAgo(date);
	if (daysAgo === 1) return i18n.t('gear_tab.item_notices.discovered_yesterday');
	const month = new Intl.DateTimeFormat(i18n.language, { month: 'long' }).format(date);
	return i18n.t('gear_tab.item_notices.discovered_on', { date: `${month} ${date.getDate()}${ordinalSuffix(date.getDate())}` });
};
