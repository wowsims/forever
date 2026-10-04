import { Unstyled } from '@storybook/addon-docs/blocks';
import clsx from 'clsx';
import { type CSSProperties, type ReactNode, useLayoutEffect, useRef, useState } from 'react';

import { COLOR_FAMILIES, COLOR_TOKENS, familyOf } from '../tokens/color_families';
import type { ThemeVar } from '../tokens/parse_theme';

const byName = new Map(COLOR_TOKENS.map(token => [token.name, token]));

const CHANNELS_RE = /^(rgba?|color|oklch|oklab)\((?:srgb\s+)?([\d.]+%?)[,\s]+([\d.-]+)[,\s]+([\d.-]+)(?:\s*[,/]\s*([\d.%]+))?\)$/;

const fraction = (text: string) => (text.endsWith('%') ? Number(text.slice(0, -1)) / 100 : Number(text));

// WCAG relative luminance of a computed colour, composited over a backdrop when it has alpha; an oklch or
// oklab colour's lightness cubed approximates it.
const luminance = (color: string, backdrop: number): number | undefined => {
	const m = color.match(CHANNELS_RE);
	if (!m) return undefined;
	const alpha = m[5] === undefined ? 1 : fraction(m[5]);
	if (m[1].startsWith('ok')) return alpha * fraction(m[2]) ** 3 + (1 - alpha) * backdrop;
	const scale = m[1] === 'color' ? 1 : 255;
	const linear = (channel: string) => {
		const c = Number(channel) / scale;
		return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
	};
	const own = 0.2126 * linear(m[2]) + 0.7152 * linear(m[3]) + 0.0722 * linear(m[4]);
	return alpha * own + (1 - alpha) * backdrop;
};

interface PaintedProps {
	background: string;
	color?: string;
	className?: string;
	children: ReactNode;
}

const Painted = ({ background, color, className, children }: PaintedProps) => {
	const ref = useRef<HTMLDivElement>(null);
	const [light, setLight] = useState(false);
	useLayoutEffect(() => {
		const el = ref.current;
		if (!el || color) return;
		const parent = el.parentElement ? (luminance(getComputedStyle(el.parentElement).backgroundColor, 0) ?? 0) : 0;
		setLight((luminance(getComputedStyle(el).backgroundColor, parent) ?? 0) > 0.35);
	}, [background, color]);
	const style: CSSProperties = { backgroundColor: background, color: color ?? (light ? 'var(--color-black)' : 'var(--color-white)') };
	return (
		<div ref={ref} className={className} style={style}>
			{children}
		</div>
	);
};

const FamilyCard = ({ name, tokens }: { name: string; tokens: ThemeVar[] }) => {
	const lead = byName.get(`--color-${name.toLowerCase()}`);
	return (
		<section className="overflow-hidden rounded-md border border-border-muted">
			<Painted background={lead ? `var(${lead.name})` : 'var(--color-surface-raised)'} className="px-4 py-2 font-bold">
				{name}
			</Painted>
			{tokens.map(token => {
				const foreground = byName.get(`${token.name}-foreground`);
				return (
					<Painted key={token.name} background={`var(${token.name})`} color={foreground && `var(${foreground.name})`} className="px-4 py-2 text-ui">
						<span className={clsx(token.inline && 'italic')} title={token.inline ? 'inline' : undefined}>
							{token.name.slice('--color-'.length)}
						</span>
						<code className="block break-all">{token.value}</code>
					</Painted>
				);
			})}
		</section>
	);
};

export const ColorFamilies = () => (
	<Unstyled>
		<div className="grid grid-cols-auto-fit-220 items-start gap-4">
			{COLOR_FAMILIES.map(family => (
				<FamilyCard
					key={family.name}
					name={family.name}
					tokens={COLOR_TOKENS.filter(token => familyOf(token) === family && !(family.name === 'Class' && token.name.endsWith('-foreground')))}
				/>
			))}
		</div>
	</Unstyled>
);
