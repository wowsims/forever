import { ItemQuality } from '@generated/proto/common';
import { ResourceType } from '@generated/proto/spell';
import { UIItem_FactionRestriction } from '@generated/proto/ui';
import { Unstyled } from '@storybook/addon-docs/blocks';
import {
	CLASS_BG,
	CLASS_BORDER,
	CLASS_TEXT,
	DANGER_TEXT,
	FACTION_TEXT,
	QUALITY_TEXT,
	RESOURCE_TEXT,
	SPELL_SCHOOL_BG,
	SPELL_SCHOOL_TEXT,
	TIMELINE_SERIES_TEXT,
	TONE_TEXT,
} from '@ui-kit/utils/colors';
import clsx from 'clsx';

interface ClassMap {
	name: string;
	map: Partial<Record<string | number, string>>;
	keyNames?: Record<number, string>;
	sampleClassName?: string;
}

const MAPS: ReadonlyArray<ClassMap> = [
	{ name: 'QUALITY_TEXT', map: QUALITY_TEXT, keyNames: ItemQuality },
	{ name: 'CLASS_TEXT', map: CLASS_TEXT },
	{ name: 'CLASS_BG', map: CLASS_BG, sampleClassName: 'text-black' },
	{ name: 'CLASS_BORDER', map: CLASS_BORDER, sampleClassName: 'border-2' },
	{ name: 'SPELL_SCHOOL_TEXT', map: SPELL_SCHOOL_TEXT },
	{ name: 'SPELL_SCHOOL_BG', map: SPELL_SCHOOL_BG, sampleClassName: 'text-black' },
	{ name: 'RESOURCE_TEXT', map: RESOURCE_TEXT, keyNames: ResourceType },
	{ name: 'TIMELINE_SERIES_TEXT', map: TIMELINE_SERIES_TEXT },
	{ name: 'FACTION_TEXT', map: FACTION_TEXT, keyNames: UIItem_FactionRestriction },
	{ name: 'TONE_TEXT', map: TONE_TEXT },
	{ name: 'DANGER_TEXT', map: DANGER_TEXT },
];

const ClassMapTable = ({ name, map, keyNames, sampleClassName }: ClassMap) => (
	<section className="mb-6">
		<h3>
			<code>{name}</code>
		</h3>
		<table className="w-full text-sm">
			<tbody>
				{Object.entries(map).map(([key, className]) => {
					const label = keyNames?.[Number(key)] ?? key;
					return (
						<tr key={key} className="border-b border-border-muted">
							<td className="w-56 p-2">
								<span className={clsx('inline-block px-2 py-1', sampleClassName, className)}>{label}</span>
							</td>
							<td className="p-2 font-mono text-ui text-link">{className}</td>
						</tr>
					);
				})}
			</tbody>
		</table>
	</section>
);

export const ClassMaps = () => (
	<Unstyled>
		{MAPS.map(classMap => (
			<ClassMapTable key={classMap.name} {...classMap} />
		))}
	</Unstyled>
);
