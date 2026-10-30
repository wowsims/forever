import i18n from '@i18n/config';
import { useSimReady } from '@sim/hooks/useSimReady';
import { Button } from '@ui-kit/Button';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { Icon } from '@ui-kit/Icon';
import { useState } from 'react';

import { SIML_ROTATION_EXPORTER } from '../../exporters';
import { exporterDialog } from '../Exporter/ExporterDialog';
import { SimlRotationImporterDialog } from '../Importer';

// Module scope, because `exporterDialog` builds a component type: called in the render body it would build a new
// one each time and remount the open dialog.
const SimlRotationExporterDialog = exporterDialog(SIML_ROTATION_EXPORTER);

type Open = 'export' | 'import' | null;

/** The rotation as SIML text: written out, or read in over the current rotation alone. */
export const SimlRotationActions = () => {
	// Both dialogs read the player, which is whole only once the sim is ready.
	const ready = useSimReady();
	const [open, setOpen] = useState<Open>(null);
	const openChange = (dialog: Open) => (next: boolean) => setOpen(next ? dialog : null);

	return (
		<ContentBlock testId="siml-rotation" config={{ header: { title: i18n.t('rotation_tab.siml.title'), tooltip: i18n.t('rotation_tab.siml.tooltip') } }}>
			<div className="flex flex-wrap gap-2">
				<Button data-testid="siml-rotation-export" disabled={!ready} onClick={() => setOpen('export')}>
					<Icon name="right-from-bracket" />
					{i18n.t('rotation_tab.siml.export_button')}
				</Button>
				<Button data-testid="siml-rotation-import" disabled={!ready} onClick={() => setOpen('import')}>
					<Icon name="download" />
					{i18n.t('rotation_tab.siml.import_button')}
				</Button>
			</div>
			<SimlRotationExporterDialog open={open === 'export'} onOpenChange={openChange('export')} />
			<SimlRotationImporterDialog open={open === 'import'} onOpenChange={openChange('import')} />
		</ContentBlock>
	);
};
