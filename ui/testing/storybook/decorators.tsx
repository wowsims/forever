import type { Decorator } from '@storybook/react-vite';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import clsx from 'clsx';
import { type ReactNode, useState } from 'react';

interface SimUiFrameProps {
	spec: string;
	children: ReactNode;
}

// A component, not a closure with hooks: react-hooks/rules-of-hooks is an error at the lint root.
// The portal container sits INSIDE the frame so overlays keep the theme variables.
const SimUiFrame = ({ spec, children }: SimUiFrameProps) => {
	const [portal, setPortal] = useState<HTMLDivElement | null>(null);
	const themed = spec !== 'none';
	const specClass = themed && spec !== 'default' ? spec : undefined;
	return (
		<div
			className={clsx(
				themed && 'sim-ui group/sim',
				specClass,
				themed ? 'bg-sim bg-cover bg-no-repeat' : 'bg-background',
				'min-h-screen p-page text-foreground',
			)}
			data-testid="storybook-sim-ui">
			<PortalContainerContext value={portal}>{children}</PortalContainerContext>
			<div ref={setPortal} data-testid="storybook-portal" />
		</div>
	);
};

export const withSimUi: Decorator = (Story, { globals, parameters }) => (
	<SimUiFrame spec={(parameters.spec as string | undefined) ?? (globals.spec as string | undefined) ?? 'default'}>
		<Story />
	</SimUiFrame>
);

/** Renders a tooltip body on the real tooltip surface, so its font matches what `Tooltip` shows it at. */
export const withTooltipSurface: Decorator = Story => (
	<div className="ui-tooltip inline-block p-2">
		<Story />
	</div>
);

/** Stories share one origin; clear and seed localStorage per story. */
export const withStorage =
	(seed: Record<string, string> = {}): Decorator =>
	Story => {
		window.localStorage.clear();
		for (const [key, value] of Object.entries(seed)) window.localStorage.setItem(key, value);
		return <Story />;
	};
