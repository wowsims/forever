import type { ComponentType } from 'react';

import type { ImportExportDialogProps } from '../import_export';

export interface ImportExportItemProps {
	label: string;
	dialog: ComponentType<ImportExportDialogProps>;
}

// A declaration, never rendered: `ImportExportMenu` reads these props off its children.
export const ImportExportItem = (_props: ImportExportItemProps): null => null;
