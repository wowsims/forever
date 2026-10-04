import { Unstyled } from '@storybook/addon-docs/blocks';
import type { ReactNode } from 'react';

import baseCss from '../../styles/base.css?raw';
import breakpointsCss from '../../styles/theme/breakpoints.css?raw';
import { resolvePxValue } from '../tokens/parse_theme';
import { parseRootVars, THEME_TOKENS } from '../tokens/theme_tokens';

const PLAIN_LENGTH = /^([\d.]+(px|rem)|--spacing\([\d.]+\))$/;
const ROOT_FONT_SIZE = Number(baseCss.match(/--font-size-root:\s*([\d.]+)px/)?.[1]);
const BASE_STEPS = [1, 2, 4, 6, 8, 10, 12, 16, 20, 24];

const formatPx = (px: number) => `${Math.round(px * 100) / 100}px`;

interface ScaleRowProps {
	name: string;
	value: string;
	px: number | null;
	children: ReactNode;
}

const ScaleRow = ({ name, value, px, children }: ScaleRowProps) => (
	<div className="flex items-start gap-6 py-2">
		<div className="w-56 shrink-0">
			<div className="font-bold">{name}</div>
			<code className="text-ui text-link">
				({value}
				{px !== null && value !== formatPx(px) && ` = ${formatPx(px)}`})
			</code>
		</div>
		{children}
	</div>
);

export const BaseSpacingScale = () => (
	<Unstyled>
		<div className="mb-6">
			{BASE_STEPS.map(step => {
				const size = `calc(var(--spacing) * ${step})`;
				return (
					<ScaleRow key={step} name={`${step}`} value={`--spacing(${step})`} px={step * ROOT_FONT_SIZE * 0.25}>
						<span className="block shrink-0 rounded-md bg-primary" style={{ width: size, height: size }} />
					</ScaleRow>
				);
			})}
		</div>
	</Unstyled>
);

interface SpacingScaleProps {
	prefixes: ReadonlyArray<string>;
	bar?: boolean;
}

export const SpacingScale = ({ prefixes, bar }: SpacingScaleProps) => {
	const tokens = THEME_TOKENS.filter(token => prefixes.includes(token.prefix) && PLAIN_LENGTH.test(token.value))
		.map(token => ({ ...token, px: resolvePxValue(token.value, ROOT_FONT_SIZE) }))
		.sort((a, b) => (a.px ?? 0) - (b.px ?? 0));
	return (
		<Unstyled>
			<div className="mb-6">
				{tokens.map(token => (
					<ScaleRow key={token.name} name={token.name.slice(token.prefix.length + 3)} value={token.value} px={token.px}>
						{bar ? (
							<span className="block h-2 max-w-full rounded-md bg-primary" style={{ width: token.px ?? 0 }} />
						) : (
							<span className="block shrink-0 rounded-md bg-primary" style={{ width: token.px ?? 0, height: token.px ?? 0 }} />
						)}
					</ScaleRow>
				))}
			</div>
		</Unstyled>
	);
};

const VARIANT_RE = /@variant\s+([a-z-]+)\s*\{([^}]*)\}/g;

export const ResponsiveSpacing = () => (
	<Unstyled>
		<table className="mb-6 w-full text-sm">
			<thead>
				<tr className="border-b border-border">
					<th className="p-2">Variant</th>
					<th className="p-2">Variable</th>
					<th className="p-2">Value</th>
				</tr>
			</thead>
			<tbody>
				{[...breakpointsCss.matchAll(VARIANT_RE)].flatMap(([, variant, body]) =>
					parseRootVars(body).map(token => (
						<tr key={`${variant}${token.name}`} className="border-b border-border-muted">
							<td className="p-2">
								<code>{variant}</code>
							</td>
							<td className="p-2">
								<code>{token.name}</code>
							</td>
							<td className="p-2 font-mono text-link">{token.value}</td>
						</tr>
					)),
				)}
			</tbody>
		</table>
	</Unstyled>
);
