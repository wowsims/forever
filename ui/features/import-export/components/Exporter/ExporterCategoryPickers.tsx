import { BooleanPicker } from '@ui-kit/BooleanPicker';

import { EXPORT_CATEGORY_OPTIONS } from '../../exporters/categories';
import type { ExportCategories } from '../../exporters/types';

export interface ExporterCategoryPickersProps {
	/** Written in place, so its identity has to hold. */
	categories: ExportCategories;
	/** Rung after a write, because the record is not a store and nothing else says it moved. */
	onChange: () => void;
}

export const ExporterCategoryPickers = ({ categories, onChange }: ExporterCategoryPickersProps) => (
	<div className="grid grid-cols-4 gap-x-4 gap-y-1 max-md:grid-cols-3">
		{EXPORT_CATEGORY_OPTIONS.map(option => (
			<BooleanPicker
				key={option.category}
				id={`link-exporter-${option.category}`}
				label={option.label}
				labelTooltip={option.labelTooltip}
				layout="inline"
				value={categories[option.category]}
				onChange={next => {
					categories[option.category] = next;
					onChange();
				}}
			/>
		))}
	</div>
);
