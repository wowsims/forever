import type { SimSettingCategories } from '@sim/constants/sim_settings';
import type { IndividualSimHost } from '@sim/sim_host';

export type ExportCategories = Record<SimSettingCategories, boolean>;

export interface ExporterDefinition {
	title: string;
	allowDownload?: boolean;
	downloadFileName?: string;
	downloadMimeType?: string;
	selectCategories?: boolean;
	/** The exported text, or a promise of it for an exporter that loads what it prints with. */
	getData: (host: IndividualSimHost<any>, categories: ExportCategories) => string | Promise<string>;
}
