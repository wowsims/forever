import { Unstyled } from '@storybook/addon-docs/blocks';
import { Box } from '@ui-kit/Box';
import { Spinner } from '@ui-kit/Spinner';
import clsx from 'clsx';
import type { ReactNode } from 'react';

import { UI_CLASS_FILES } from '../tokens/ui_classes';
import { UTILITIES } from '../tokens/utilities';

const uiClassSample = (name: string): ReactNode => {
	if (name === 'ui-button' || name.startsWith('ui-button-'))
		return (
			<button type="button" data-size="md" className={clsx('ui-button', name === 'ui-button-icon' && 'ui-button-primary', name)}>
				Button
			</button>
		);
	if (name === 'ui-chip')
		return (
			<div className="ui-chip self-start">
				<span className="p-2">Chip</span>
			</div>
		);
	if (name === 'ui-spinner') return <Spinner />;
	if (name === 'ui-skeleton') return <span className="ui-skeleton" />;
	return undefined;
};

const UTILITY_SAMPLES: Record<string, ReactNode> = {
	'text-fluid-xl': <span className="text-fluid-xl">Fluid heading</span>,
	'text-fluid-5xl': <span className="text-fluid-5xl">Fluid</span>,
	interactive: <span className="interactive">Pointer cursor, no text selection</span>,
	'icon-sm': <span className="icon-sm bg-primary" />,
	'focus-ring': (
		<button type="button" data-size="md" className="ui-button ui-button-secondary focus-visible:focus-ring">
			Tab to focus
		</button>
	),
	'focus-ring-inset': (
		<button type="button" data-size="md" className="ui-button ui-button-secondary focus-visible:focus-ring-inset">
			Tab to focus
		</button>
	),
	'active-underline': <span className="active-underline inline-block px-page py-1">Underlined tab</span>,
};

interface UiClassCardProps {
	name: string;
	rules: ReadonlyArray<string>;
	sample: ReactNode;
}

const UiClassCard = ({ name, rules, sample }: UiClassCardProps) => (
	<Box border="border-muted" className="flex flex-col gap-2" testId="ui-class-card">
		<code className="font-bold">{name}</code>
		{rules.map((rule, i) => (
			<pre key={i} className="overflow-x-auto bg-surface p-2 text-ui whitespace-pre text-link">
				{rule}
			</pre>
		))}
		{sample ?? <span className="text-ui text-foreground-disabled">No standalone sample: needs its component.</span>}
	</Box>
);

export const UiClassGroups = () => (
	<Unstyled>
		{UI_CLASS_FILES.map(({ file, classes }) => (
			<section key={file} className="mb-6">
				<h3>
					<code>{file}</code> ({classes.length})
				</h3>
				<div className="grid grid-cols-auto-fit-220 gap-2">
					{classes.map(uiClass => (
						<UiClassCard key={uiClass.name} name={uiClass.name} rules={uiClass.rules} sample={uiClassSample(uiClass.name)} />
					))}
				</div>
			</section>
		))}
	</Unstyled>
);

export const UtilityCards = ({ names }: { names?: ReadonlyArray<string> }) => (
	<Unstyled>
		<div className="mb-6 grid grid-cols-auto-fit-220 gap-2">
			{UTILITIES.filter(utility => !names || names.includes(utility.name)).map(utility => (
				<UiClassCard key={utility.name} name={utility.name} rules={[utility.text]} sample={UTILITY_SAMPLES[utility.name]} />
			))}
		</div>
	</Unstyled>
);
