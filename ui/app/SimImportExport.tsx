import {
	AddonImporterDialog,
	ArmoryImporterDialog,
	CLI_EXPORTER,
	exporterDialog,
	JSON_EXPORTER,
	JsonImporterDialog,
	LINK_EXPORTER,
	PAWN_EP_EXPORTER,
	SIXTY_UPGRADES_EP_EXPORTER,
	SixtyUpgradesImporterDialog,
	WOWHEAD_GEAR_PLANNER_EXPORTER,
	WowheadImporterDialog,
} from '@features/import-export';
import i18n from '@i18n/config';
import { useSimReady } from '@sim/hooks/useSimReady';
import { isDevMode } from '@sim/utils/env';

import { ImportExportKind } from './header/import_export';
import { ImportExportItem, ImportExportMenu } from './header/ImportExportMenu';

// Module scope, because `exporterDialog` builds a component type: called in the render body it would
// build a new one each time and remount the open dialog.
const LinkExporterDialog = exporterDialog(LINK_EXPORTER);
const JsonExporterDialog = exporterDialog(JSON_EXPORTER);
const WowheadExporterDialog = exporterDialog(WOWHEAD_GEAR_PLANNER_EXPORTER);
const SixtyUpgradesExporterDialog = exporterDialog(SIXTY_UPGRADES_EP_EXPORTER);
const PawnExporterDialog = exporterDialog(PAWN_EP_EXPORTER);
const CliExporterDialog = exporterDialog(CLI_EXPORTER);

export const SimImportExport = () => {
	// The menus open from the first frame but every dialog reads the item database, so the entries
	// wait for init exactly as the imperative registration did.
	const ready = useSimReady();

	return (
		<div className="flex flex-nowrap items-end font-bold" data-testid="import-export">
			<ImportExportMenu kind={ImportExportKind.Import} icon="download" title={i18n.t('import.title')}>
				{ready && (
					<>
						<ImportExportItem label="JSON" dialog={JsonImporterDialog} />
						<ImportExportItem label="60U" dialog={SixtyUpgradesImporterDialog} />
						<ImportExportItem label="WoWHead" dialog={WowheadImporterDialog} />
						<ImportExportItem label={i18n.t('import.menu.addon')} dialog={AddonImporterDialog} />
						{/* Dev-only until the WoW Forever Profile API namespace is published; see workers/armory. */}
						{isDevMode() && <ImportExportItem label={i18n.t('import.menu.armory')} dialog={ArmoryImporterDialog} />}
					</>
				)}
			</ImportExportMenu>
			<ImportExportMenu kind={ImportExportKind.Export} icon="right-from-bracket" title={i18n.t('export.title')}>
				{ready && (
					<>
						<ImportExportItem label={i18n.t('export.menu.link')} dialog={LinkExporterDialog} />
						<ImportExportItem label="JSON" dialog={JsonExporterDialog} />
						<ImportExportItem label="WoWHead" dialog={WowheadExporterDialog} />
						<ImportExportItem label={i18n.t('export.menu.sixty_upgrades_ep')} dialog={SixtyUpgradesExporterDialog} />
						<ImportExportItem label={i18n.t('export.menu.pawn_ep')} dialog={PawnExporterDialog} />
						<ImportExportItem label="CLI" dialog={CliExporterDialog} />
					</>
				)}
			</ImportExportMenu>
		</div>
	);
};
