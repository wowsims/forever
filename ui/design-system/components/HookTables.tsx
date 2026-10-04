import { Unstyled } from '@storybook/addon-docs/blocks';

import allowlist from '../../class_hook_allowlist.json';
import retired from '../../retired_class_names.json';

interface AllowlistEntry {
	token?: string;
	pattern?: string;
	reason: string;
}

export const ClassHookAllowlist = () => (
	<Unstyled>
		<table className="mb-6 w-full text-sm">
			<thead>
				<tr className="border-b border-border">
					<th className="p-2">Token or pattern</th>
					<th className="p-2">Reason</th>
				</tr>
			</thead>
			<tbody>
				{(allowlist as AllowlistEntry[]).map(entry => (
					<tr key={entry.token ?? entry.pattern} className="border-b border-border-muted">
						<td className="p-2 align-top whitespace-nowrap">
							<code>{entry.token ?? `/${entry.pattern}/`}</code>
						</td>
						<td className="p-2">{entry.reason}</td>
					</tr>
				))}
			</tbody>
		</table>
	</Unstyled>
);

export const RetiredClassNames = () => (
	<Unstyled>
		<p className="mb-2">{retired.length} names.</p>
		<div className="mb-6 flex flex-wrap gap-x-2 gap-y-1 font-mono text-ui">
			{retired.map(name => (
				<span key={name}>{name}</span>
			))}
		</div>
	</Unstyled>
);
