import i18n from '@i18n/config';
import { ISSUES_REPO, REPO_CHOOSE_NEW_ISSUE_URL, REPO_RELEASES_URL, SOCIALS } from '@sim/constants/other';
import type { Sim } from '@sim/sim';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { SocialLink } from '@ui-kit/SocialLink';
import { Spinner } from '@ui-kit/Spinner';
import { toastManager } from '@ui-kit/Toast';
import { isNative } from '@ui-kit/utils/dom';
import { type ReactNode, useEffect, useRef, useState } from 'react';

import { ToolbarItem } from './ToolbarItem';
import { type NativeSimUpdateStatus, useNativeSimUpdate } from './useNativeSimUpdate';
import { useOutdatedNativeSim } from './useOutdatedNativeSim';

/** What the toolbar needs to offer the update in place of the link to the releases page. */
export interface NativeSimUpdate {
	latest: string;
	status: NativeSimUpdateStatus;
	onUpdate: () => void;
}

export interface SimToolbarViewProps {
	knownIssues: ReadonlyArray<ReactNode>;
	outdatedNativeSim: boolean;
	/** Set when the host can update itself. Without it the outdated item links to the releases page. */
	nativeSimUpdate?: NativeSimUpdate;
	onOpenSettings: () => void;
}

const NativeSimUpdateItem = ({ latest, status, onUpdate }: NativeSimUpdate) => {
	const [open, setOpen] = useState(false);
	const toastId = useRef<string | undefined>(undefined);
	const updating = status.state === 'updating';

	const openDialog = () => {
		toastManager.close(toastId.current);
		setOpen(true);
	};

	useEffect(() => {
		const id = toastManager.add({
			variant: 'info',
			title: i18n.t('sim.new_build.title'),
			autohide: false,
			body: (
				<>
					<p>{i18n.t('info.native_sim_update_toast', { version: latest })}</p>
					<Button variant="outline-light" onClick={openDialog}>
						{i18n.t('info.native_sim_update_title')}
					</Button>
				</>
			),
		});
		toastId.current = id;
		return () => toastManager.close(id);
	}, [latest]);

	return (
		<>
			<ToolbarItem
				className="text-link-danger"
				testId="native-sim-update"
				icon="gauge-high"
				tooltip={i18n.t('info.native_sim_update')}
				onClick={openDialog}
			/>
			<Dialog
				open={open}
				onOpenChange={setOpen}
				size="sm"
				preventClose={updating}
				title={i18n.t('info.native_sim_update_title')}
				testId="native-sim-update-dialog"
				footer={
					<div className="flex gap-2">
						<Button variant="outline-cancel" disabled={updating} onClick={() => setOpen(false)}>
							{i18n.t('common.cancel')}
						</Button>
						<Button variant="primary" disabled={updating} onClick={onUpdate}>
							{i18n.t('info.native_sim_update_action')}
						</Button>
					</div>
				}>
				{updating ? (
					<div className="flex items-center gap-2" role="status" data-testid="native-sim-updating">
						<Spinner size="sm" />
						{i18n.t('info.native_sim_updating')}
					</div>
				) : (
					<>
						<p className="mb-0">{i18n.t('info.native_sim_update_confirm', { version: latest })}</p>
						{status.state === 'error' && (
							<p className="mt-2 mb-0 text-link-danger" role="alert" data-testid="native-sim-update-error">
								{status.error}
							</p>
						)}
					</>
				)}
			</Dialog>
		</>
	);
};

export const SimToolbarView = ({ knownIssues, outdatedNativeSim, nativeSimUpdate, onOpenSettings }: SimToolbarViewProps) => (
	<>
		<ToolbarItem
			className="text-link-danger"
			testId="known-issues"
			hidden={knownIssues.length === 0}
			tooltip={
				<ul className="mb-0 pl-4 text-left">
					{knownIssues.map((issue, index) => (
						<li key={index}>{issue}</li>
					))}
				</ul>
			}>
			{i18n.t('info.known_issues')}
		</ToolbarItem>

		{ISSUES_REPO && <ToolbarItem href={REPO_CHOOSE_NEW_ISSUE_URL} icon="bug" tooltip={i18n.t('info.bug_report')} />}

		{!isNative() && <ToolbarItem href={REPO_RELEASES_URL} icon="gauge-high" tooltip={i18n.t('info.download_native_sim')} />}

		<ToolbarItem testId="sim-options" icon="gear" tooltip={i18n.t('info.sim_options')} onClick={onOpenSettings} />

		<div className="mt-2 mb-2 ml-4 flex border-l border-l-border [&_a]:my-2 [&_button]:my-2" data-testid="sim-toolbar-socials">
			{SOCIALS.map(social => (
				<div key={social.key} className="ml-4 flex transition-colors duration-150 ease-in-out" data-testid="sim-toolbar-item">
					<SocialLink social={social} />
				</div>
			))}
		</div>

		{nativeSimUpdate ? (
			<NativeSimUpdateItem {...nativeSimUpdate} />
		) : (
			outdatedNativeSim && (
				<ToolbarItem href={REPO_RELEASES_URL} icon="gauge-high" className="text-link-danger" tooltip={i18n.t('info.native_sim_outdated')} />
			)
		)}
	</>
);

export interface SimToolbarProps {
	// The toolbar renders inside the shell, which is built before `SimHostObject` adopts it, so
	// there is no `SimHostProvider` above it to read the sim from.
	sim: Sim;
	knownIssues: ReadonlyArray<ReactNode>;
	onOpenSettings: () => void;
}

export const SimToolbar = ({ sim, knownIssues, onOpenSettings }: SimToolbarProps) => {
	const outdatedNativeSim = useOutdatedNativeSim(sim);
	const { status, update } = useNativeSimUpdate(outdatedNativeSim);
	const nativeSimUpdate = outdatedNativeSim?.canUpdate ? { latest: outdatedNativeSim.latest, status, onUpdate: update } : undefined;
	return (
		<div className="ml-auto flex flex-nowrap items-end text-(length:--text-ui) font-bold" data-testid="sim-toolbar">
			<SimToolbarView
				knownIssues={knownIssues}
				outdatedNativeSim={outdatedNativeSim != null}
				nativeSimUpdate={nativeSimUpdate}
				onOpenSettings={onOpenSettings}
			/>
		</div>
	);
};
