import { Unstyled } from '@storybook/addon-docs/blocks';
import type { ReactNode } from 'react';

import { THEME_TOKENS, type ThemeToken, tokensNamed } from '../tokens/theme_tokens';

type Preview = 'radius' | 'shadow' | 'drop-shadow' | 'text-shadow' | 'animation' | 'background-image' | 'font';

const PREVIEWS: Record<Preview, (value: string) => ReactNode> = {
	radius: value => <span className="block size-10 border border-border bg-surface-raised" style={{ borderRadius: value }} />,
	shadow: value => <span className="block size-10 bg-surface" style={{ boxShadow: value }} />,
	'drop-shadow': value => <span className="block size-10 rounded-full bg-surface" style={{ filter: `drop-shadow(${value})` }} />,
	'text-shadow': value => (
		<span className="text-lg font-bold" style={{ textShadow: value }}>
			Aa 1234
		</span>
	),
	animation: value => (
		<span className="block h-4 w-24" style={{ animation: value, backgroundImage: 'var(--background-image-skeleton)', backgroundSize: '200% 100%' }} />
	),
	'background-image': value => <span className="block h-8 w-24 border border-border bg-cover" style={{ backgroundImage: value }} />,
	font: value => <span style={{ fontFamily: value }}>The quick brown fox, 0123456789</span>,
};

interface TokenTableProps {
	prefix?: string;
	exclude?: string;
	file?: string;
	preview?: Preview;
	numeric?: boolean;
}

const sourceOf = (token: ThemeToken) => `${token.file}${token.runtime ? ' (runtime)' : token.inline ? ' (inline)' : ''}`;

export const TokenTable = ({ prefix, exclude, file, preview, numeric }: TokenTableProps) => {
	const tokens = (prefix ? tokensNamed(prefix, exclude) : [...THEME_TOKENS]).filter(token => !file || token.file === file);
	if (numeric) tokens.sort((a, b) => Number(a.value) - Number(b.value));
	return (
		<Unstyled>
			<table className="mb-6 w-full text-sm">
				<thead>
					<tr className="border-b border-border">
						<th className="p-2">Token</th>
						<th className="p-2">Value</th>
						<th className="p-2">Source</th>
						{preview && <th className="p-2">Preview</th>}
					</tr>
				</thead>
				<tbody>
					{tokens.map(token => (
						<tr key={token.name} className="border-b border-border-muted">
							<td className="p-2 align-top">
								<code>{token.name}</code>
							</td>
							<td className="p-2 align-top font-mono text-ui break-all text-link">{token.value}</td>
							<td className="p-2 align-top text-ui whitespace-nowrap">{sourceOf(token)}</td>
							{preview && <td className="p-2 align-top">{PREVIEWS[preview](`var(${token.name})`)}</td>}
						</tr>
					))}
				</tbody>
			</table>
		</Unstyled>
	);
};
