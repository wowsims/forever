import { Unstyled } from '@storybook/addon-docs/blocks';

import baseCss from '../../styles/base.css?raw';
import { tokensNamed } from '../tokens/theme_tokens';

const FONT_FILES = import.meta.glob<string>('../../../assets/fonts/*.woff2', { query: '?url', import: 'default', eager: true });
const fileName = (path: string) => path.slice(path.lastIndexOf('/') + 1);

const FONT_FACE_RE = /@font-face\s*\{([^}]*)\}/g;
const declaredFaces = new Map(
	[...baseCss.matchAll(FONT_FACE_RE)].map(([, body]) => [
		fileName(/src:\s*url\(([^)]+)\)/.exec(body)?.[1] ?? ''),
		{ family: /font-family:\s*([^;]+);/.exec(body)?.[1], weight: /font-weight:\s*([^;]+);/.exec(body)?.[1] ?? 'normal' },
	]),
);

export const FontFaces = () => (
	<Unstyled>
		<table className="mb-6 w-full text-sm">
			<thead>
				<tr className="border-b border-border">
					<th className="p-2">File in assets/fonts</th>
					<th className="p-2">@font-face in base.css</th>
				</tr>
			</thead>
			<tbody>
				{Object.keys(FONT_FILES)
					.map(fileName)
					.sort()
					.map(file => {
						const face = declaredFaces.get(file);
						return (
							<tr key={file} className="border-b border-border-muted">
								<td className="p-2">
									<code>{file}</code>
								</td>
								<td className="p-2">{face ? `${face.family}, weight ${face.weight}` : 'not declared'}</td>
							</tr>
						);
					})}
			</tbody>
		</table>
	</Unstyled>
);

export const TypeScale = () => (
	<Unstyled>
		<div className="mb-6 flex flex-col">
			{tokensNamed('text', 'text-shadow').map(token => (
				<div key={token.name} className="flex items-baseline gap-4 border-b border-border-muted py-2">
					<code className="w-40 shrink-0 text-ui">{token.name}</code>
					<code className="w-16 shrink-0 text-ui text-link">{token.value}</code>
					<span style={{ fontSize: `var(${token.name})` }}>The quick brown fox, 0123456789</span>
				</div>
			))}
		</div>
	</Unstyled>
);

export const Headings = () => (
	<Unstyled>
		<div className="mb-6">
			<h1>h1 heading</h1>
			<h2>h2 heading</h2>
			<h3>h3 heading</h3>
			<h4>h4 heading</h4>
			<h5>h5 heading</h5>
			<h6>h6 heading</h6>
			<p className="text-fluid-xl">text-fluid-xl</p>
			<p className="text-fluid-5xl">text-fluid-5xl</p>
		</div>
	</Unstyled>
);
