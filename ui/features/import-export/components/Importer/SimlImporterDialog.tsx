import i18n from '@i18n/config';

import { SIML_IMPORTER, SIML_ROTATION_IMPORTER } from '../../importers';
import { Importer } from './Importer';
import type { ImporterDialogProps } from './types';

/** Any SIML file: a full setup, or the part of one that an uploaded file's name says it holds. */
export const SimlImporterDialog = ({ open, onOpenChange }: ImporterDialogProps) => (
	<Importer open={open} onOpenChange={onOpenChange} {...SIML_IMPORTER}>
		<>
			<p>{i18n.t('import.siml.description')}</p>
			<p>{i18n.t('import.siml.instructions')}</p>
		</>
	</Importer>
);

/** A rotation, read as an x.apl.siml whatever the uploaded file is named. */
export const SimlRotationImporterDialog = ({ open, onOpenChange }: ImporterDialogProps) => (
	<Importer open={open} onOpenChange={onOpenChange} {...SIML_ROTATION_IMPORTER}>
		<>
			<p>{i18n.t('import.siml_rotation.description')}</p>
			<p>{i18n.t('import.siml_rotation.instructions')}</p>
		</>
	</Importer>
);
