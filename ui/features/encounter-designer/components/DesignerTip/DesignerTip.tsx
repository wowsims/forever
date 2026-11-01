import { parseTip } from '../EncounterDesigner/utils/tips';

export interface DesignerTipProps {
	/** The anchor's `data-tip` text: the title, then one rule per line. */
	text: string | null | undefined;
}

export const DesignerTip = ({ text }: DesignerTipProps) => {
	const tip = parseTip(text);
	if (!tip) return null;
	return (
		<div className="flex flex-col">
			<span className="font-bold">{tip.title}</span>
			{tip.lines.map((line, index) => (
				<span key={index}>{line}</span>
			))}
		</div>
	);
};
