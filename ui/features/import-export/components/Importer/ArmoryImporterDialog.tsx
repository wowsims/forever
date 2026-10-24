import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { Button } from '@ui-kit/Button';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { Dialog } from '@ui-kit/Dialog';
import { Icon } from '@ui-kit/Icon';
import { StringField } from '@ui-kit/TextField';
import { toastManager } from '@ui-kit/Toast';
import { useEffect, useState } from 'react';

import { trackPageView } from '../../../../tracking/utils';
import type { ArmoryRegion } from '../../importers/armory';
import { ARMORY_REGIONS, importArmoryCharacter, parseCharacterName } from '../../importers/armory';
import type { ImporterDialogProps } from './types';

// The last region and realm looked up, so the next lookup only needs a name.
const LAST_LOOKUP_KEY = 'armory-importer-last-lookup';

const readLastLookup = (): { region: ArmoryRegion; realm: string } => {
	try {
		const saved = JSON.parse(localStorage.getItem(LAST_LOOKUP_KEY) ?? 'null');
		if (saved && ARMORY_REGIONS.includes(saved.region) && typeof saved.realm == 'string') return saved;
	} catch {
		// Storage is unavailable or holds something else; start from the defaults.
	}
	return { region: 'us', realm: '' };
};

const REGION_OPTIONS = ARMORY_REGIONS.map(region => ({ value: region, label: region.toUpperCase() }));

export const ArmoryImporterDialog = ({ open, onOpenChange }: ImporterDialogProps) => {
	const host = useSimHost();
	const [region, setRegion] = useState<ArmoryRegion>(() => readLastLookup().region);
	const [realm, setRealm] = useState(() => readLastLookup().realm);
	const [name, setName] = useState('');
	const [loading, setLoading] = useState(false);
	const title = i18n.t('import.armory.title');

	useEffect(() => {
		if (!open) return;
		trackPageView(title, '/import/armory');
	}, [open, title]);

	// A name pasted as the game writes it in logs (`Erlyn-Farthing-ClassicBetaPvE2`) carries the realm too.
	const onNameChange = (entered: string) => {
		const parsed = parseCharacterName(entered);
		if (parsed?.realm) {
			setRealm(parsed.realm);
			setName(parsed.name);
		} else {
			setName(entered);
		}
	};

	const runImport = async () => {
		const parsed = parseCharacterName(name);
		if (!parsed || !realm.trim()) {
			toastManager.add({ variant: 'error', body: i18n.t('import.armory.errors.missing_fields') });
			return;
		}
		setLoading(true);
		try {
			await importArmoryCharacter(host, { region, realm, name: parsed.name });
			try {
				localStorage.setItem(LAST_LOOKUP_KEY, JSON.stringify({ region, realm }));
			} catch {
				// Remembering the lookup is a convenience; the import already succeeded.
			}
			onOpenChange(false);
		} catch (error: any) {
			toastManager.add({ variant: 'error', body: `${i18n.t('import.error')} ${error?.message || error}` });
		} finally {
			setLoading(false);
		}
	};

	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			testId="armory-importer"
			size="md"
			verticalAlign="center"
			title={
				<span className="flex items-center gap-2">
					<Icon name="battle-net" style="brands" className="text-blizzard" />
					{title}
				</span>
			}
			footer={
				<Button data-testid="import-button" className="w-48" onClick={runImport} loading={loading}>
					<Icon name="download" style="base" />
					{i18n.t('import.json.import_button')}
				</Button>
			}>
			<p>{i18n.t('import.armory.description')}</p>
			<div className="flex flex-wrap items-end gap-2">
				<ChoiceField<ArmoryRegion>
					id="armory-importer-region"
					label={i18n.t('import.armory.region')}
					options={REGION_OPTIONS}
					value={region}
					onChange={setRegion}
					className="w-28 shrink-0"
					testId="armory-importer-region"
				/>
				<StringField
					id="armory-importer-realm"
					label={i18n.t('import.armory.realm')}
					placeholder="ClassicBetaPvE2"
					value={realm}
					onChange={setRealm}
					className="min-w-40 flex-1"
					testId="armory-importer-realm"
				/>
				<StringField
					id="armory-importer-name"
					label={i18n.t('import.armory.name')}
					placeholder="Erlyn Farthing"
					value={name}
					onChange={onNameChange}
					className="min-w-40 flex-1"
					testId="armory-importer-name"
				/>
			</div>
		</Dialog>
	);
};
