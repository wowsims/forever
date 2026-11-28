import { createContext, type ReactNode, use, useMemo, useState } from 'react';

// The Upgrades side's run panel is built in its tab body, under the providers it reads, and portalled into these
// elements of the shell: the sidebar from `lg` up, the dock's sheet and bar below it.
export interface UpgradesSlots {
	panel: HTMLElement | null;
	/** The dock bar's count and run button; null with no dock, where the panel keeps its own run button. */
	summary: HTMLElement | null;
	action: HTMLElement | null;
}

type UpgradesSlotRefs = { [K in keyof UpgradesSlots]: (element: HTMLElement | null) => void };

const UpgradesSlotsContext = createContext<UpgradesSlots>({ panel: null, summary: null, action: null });
const UpgradesSlotRefsContext = createContext<UpgradesSlotRefs | null>(null);

export const useUpgradesSlots = () => use(UpgradesSlotsContext);

/** Ref callbacks for the shell's elements that host the run panel. */
export const useUpgradesSlotRefs = () => use(UpgradesSlotRefsContext)!;

export const UpgradesSlotsProvider = ({ children }: { children: ReactNode }) => {
	const [panel, setPanel] = useState<HTMLElement | null>(null);
	const [summary, setSummary] = useState<HTMLElement | null>(null);
	const [action, setAction] = useState<HTMLElement | null>(null);
	const slots = useMemo(() => ({ panel, summary, action }), [panel, summary, action]);
	const refs = useMemo(() => ({ panel: setPanel, summary: setSummary, action: setAction }), []);
	return (
		<UpgradesSlotRefsContext value={refs}>
			<UpgradesSlotsContext value={slots}>{children}</UpgradesSlotsContext>
		</UpgradesSlotRefsContext>
	);
};
