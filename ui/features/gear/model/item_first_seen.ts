import i18n from '@i18n/config';

const DAY_MS = 24 * 60 * 60 * 1000;
const NEWLY_DISCOVERED_WINDOW_MS = 7 * DAY_MS;

export const isNewlyDiscovered = (firstSeenSeconds: number): boolean => Date.now() - firstSeenSeconds * 1000 <= NEWLY_DISCOVERED_WINDOW_MS;

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

export const discoveredLabel = (firstSeenSeconds: number): string => {
	const date = new Date(firstSeenSeconds * 1000);
	const daysAgo = calendarDaysAgo(date);
	if (daysAgo === 0) return i18n.t('gear_tab.item_notices.discovered_today');
	if (daysAgo === 1) return i18n.t('gear_tab.item_notices.discovered_yesterday');
	const month = new Intl.DateTimeFormat(i18n.language, { month: 'long' }).format(date);
	return i18n.t('gear_tab.item_notices.discovered_on', { date: `${month} ${date.getDate()}${ordinalSuffix(date.getDate())}` });
};
