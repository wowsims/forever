import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { useId } from 'react';

export interface SavedDataTickRow {
	name: string;
	note?: string;
	checked: boolean;
}

export interface SavedDataTickListProps {
	rows: readonly SavedDataTickRow[];
	onChange: (index: number, checked: boolean) => void;
}

export const SavedDataTickList = ({ rows, onChange }: SavedDataTickListProps) => {
	const idPrefix = useId();
	return (
		<div className="mt-2 mb-2 grid max-h-[30vh] gap-1 overflow-y-auto" data-testid="saved-data-tick-list">
			{rows.map(({ name, note, checked }, index) => (
				<div key={index} className="flex items-center gap-2" data-testid="saved-data-tick-row">
					<BooleanPicker
						id={`${idPrefix}-${index}`}
						className="wrap-anywhere"
						label={name}
						layout="inline"
						size="xs"
						value={checked}
						aria-describedby={note ? `${idPrefix}-${index}-note` : undefined}
						onChange={next => onChange(index, next)}
					/>
					{note && (
						<span id={`${idPrefix}-${index}-note`} className="min-w-0 text-xs wrap-anywhere text-gray-400" data-testid="saved-data-tick-note">
							{note}
						</span>
					)}
				</div>
			))}
		</div>
	);
};
