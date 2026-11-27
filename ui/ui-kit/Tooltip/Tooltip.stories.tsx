import type { Meta, StoryObj } from '@storybook/react-vite';
import { GameTooltip } from '@ui-kit/GameTooltip';
import { expect, userEvent, within } from 'storybook/test';

import { LocaleHtml as LocaleHtmlContent } from './LocaleHtml';
import { Tooltip, type TooltipPlace } from './Tooltip';
import { tooltipAnchorProps } from './utils';

const PLACES: Array<TooltipPlace> = ['top', 'right', 'bottom', 'left'];
const PADDINGS = [true, false];

const ANCHOR = 'border border-primary px-4 py-2 text-white';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/large';
const LINKED_ANCHORS = [
	{ name: 'Summon Hawk', icon: 'spell_nature_ravenform', href: 'https://wowhead.com/forever/spell=1293241' },
	{ name: 'Improved Revenge', icon: 'ability_warrior_revenge', href: 'https://wowhead.com/forever/spell=12797' },
	{ name: 'Devilsaur Gauntlets', icon: 'inv_gauntlets_26', href: 'https://wowhead.com/forever/item=15063' },
];

const meta = {
	title: 'Kit/Tooltip',
	component: Tooltip,
	tags: ['autodocs'],
	args: {
		id: 'tooltip-story',
		content: 'Reforge to hit cap.',
	},
	argTypes: {
		place: { control: 'select', options: PLACES },
		maxWidth: { control: 'text' },
		content: { control: 'text' },
		render: { control: false },
	},
	render: args => (
		<div className="p-24">
			<button type="button" className={ANCHOR} {...tooltipAnchorProps(args.id)}>
				Hover me
			</button>
			<Tooltip {...args} />
		</div>
	),
} satisfies Meta<typeof Tooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Places: Story = {
	render: args => (
		<div className="grid grid-cols-4 gap-24 p-24">
			{PADDINGS.flatMap(padded =>
				PLACES.map(place => {
					const id = `tooltip-story-${place}-${padded ? 'padded' : 'unpadded'}`;
					return (
						<div key={id}>
							<button type="button" className={ANCHOR} {...tooltipAnchorProps(id)}>
								{place} {padded ? 'padded' : 'unpadded'}
							</button>
							<Tooltip {...args} id={id} place={place} padded={padded} />
						</div>
					);
				}),
			)}
		</div>
	),
};

export const Clickable: Story = {
	args: {
		clickable: true,
		content: (
			<span>
				Stays open under the pointer, so this{' '}
				<a href="#tooltip-link" className="text-link">
					link
				</a>{' '}
				can be reached.
			</span>
		),
	},
	play: async ({ canvasElement }) => {
		await userEvent.hover(within(canvasElement).getByRole('button', { name: 'Hover me' }));
		const tooltip = await within(canvasElement).findByRole('tooltip');
		await expect(canvasElement.querySelector('[data-testid="storybook-portal"]')?.contains(tooltip)).toBe(true);
		await expect(within(tooltip).getByRole('link', { name: 'link' })).toBeTruthy();
	},
};

export const Padded: Story = {
	args: { padded: true },
};

export const Unpadded: Story = {
	args: { padded: false },
};

export const LocaleHtml: Story = {
	args: {
		content: <LocaleHtmlContent html="Increases <strong>haste</strong> by <em>5%</em> for 10 sec." />,
	},
};

/** Two rows of icons that link to their pages: the pointer reaches each tooltip's link, and moving on to the next icon switches to its tooltip. */
export const CarriesLink: Story = {
	args: { id: 'tooltip-story-linked', anchorSelect: 'a[href^="https://wowhead.com/forever/"]', carriesLink: true, content: undefined },
	render: args => (
		<div className="flex flex-col items-start gap-2 p-24">
			{[0, 1].map(row => (
				<div key={row} className="flex gap-2">
					{LINKED_ANCHORS.map(anchor => (
						<a
							key={anchor.name}
							href={anchor.href}
							aria-label={anchor.name}
							className="size-10 border border-border bg-cover"
							style={{ backgroundImage: `url('${ICONS}/${anchor.icon}.jpg')` }}
						/>
					))}
				</div>
			))}
			<Tooltip
				{...args}
				maxWidth="max-w-96"
				render={({ activeAnchor }) => {
					const anchor = LINKED_ANCHORS.find(({ href }) => href === activeAnchor?.getAttribute('href'));
					return anchor ? (
						<GameTooltip name={anchor.name} iconUrl={`${ICONS}/${anchor.icon}.jpg`} link={{ href: anchor.href, label: 'Wowhead' }} />
					) : null;
				}}
			/>
		</div>
	),
};
