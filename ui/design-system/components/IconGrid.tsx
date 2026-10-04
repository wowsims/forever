import { Unstyled } from '@storybook/addon-docs/blocks';
import { hasRegularDefinition, Icon, type IconName, type IconSize } from '@ui-kit/Icon';
import { DEFINITIONS } from '@ui-kit/Icon/definitions';

const ICON_NAMES = Object.keys(DEFINITIONS) as IconName[];

const BRANDS = new Set<IconName>(['discord', 'github', 'patreon']);

const SIZES = Object.keys({
	'2xs': true,
	xs: true,
	sm: true,
	lg: true,
	xl: true,
	'2xl': true,
	'1x': true,
	'2x': true,
	'3x': true,
	'4x': true,
	'5x': true,
} satisfies Record<IconSize, true>) as IconSize[];

export const IconGrid = () => (
	<Unstyled>
		<table className="mb-6 text-sm">
			<thead>
				<tr className="border-b border-border">
					<th className="p-2 text-left">IconName</th>
					<th className="p-2">solid</th>
					<th className="p-2">regular</th>
					<th className="p-2">brands</th>
				</tr>
			</thead>
			<tbody>
				{ICON_NAMES.map(name => (
					<tr key={name} className="border-b border-border-muted">
						<td className="p-2">
							<code>{name}</code>
						</td>
						<td className="p-2 text-center text-lg">
							<Icon name={name} />
						</td>
						<td className="p-2 text-center text-lg">{hasRegularDefinition(name) && <Icon name={name} style="regular" />}</td>
						<td className="p-2 text-center text-lg">{BRANDS.has(name) && <Icon name={name} style="brands" />}</td>
					</tr>
				))}
			</tbody>
		</table>
	</Unstyled>
);

export const IconSizes = () => (
	<Unstyled>
		<div className="mb-6 flex flex-wrap items-end gap-6">
			{SIZES.map(size => (
				<div key={size} className="flex flex-col items-center gap-1">
					<Icon name="star" size={size} />
					<code className="text-ui">{size}</code>
				</div>
			))}
		</div>
	</Unstyled>
);
