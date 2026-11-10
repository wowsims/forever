import i18n from '@i18n/config';
import { createElement, type ReactElement } from 'react';
import { Trans } from 'react-i18next';

export interface LocaleHtmlProps {
	html: string;
}

const OPENING_TAG = /<([a-zA-Z][a-zA-Z0-9-]*)/g;

const componentsIn = (html: string): Record<string, ReactElement> => {
	const components: Record<string, ReactElement> = {};
	for (const [, name] of html.matchAll(OPENING_TAG)) components[name] ??= createElement(name);
	return components;
};

/**
 * Markup that came out of a locale file: pass `i18n.t(...)` output or a frozen spec config string, never
 * anything else. Each tag becomes a React element carrying the locale's own attributes; a tag that cannot
 * be resolved renders as text. Interpolation runs before the tags are read and is not escaped, so a value that
 * holds a tag becomes markup; keep user-chosen text out of the string. The markup is already
 * translated, so `cimode` hands it back as it is instead of looking it up as a key that is never there.
 */
export const LocaleHtml = ({ html }: LocaleHtmlProps) => (
	<Trans i18n={i18n} defaults={html} components={componentsIn(html)} tOptions={{ lng: 'cimode', nsSeparator: false, keySeparator: false }} />
);
