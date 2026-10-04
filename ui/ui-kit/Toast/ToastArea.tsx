import { Toast as BaseToast } from '@base-ui/react/toast';
import { usePortalContainer } from '@ui-kit/hooks/usePortalContainer';

import { DEFAULT_TOAST_DELAY, DEFAULT_TOAST_LIMIT, type ToastManager } from './manager';
import { ToastViewport } from './ToastViewport';

export interface ToastAreaProps {
	manager: ToastManager;
	/** Render in flow, where its caller placed it, instead of through a portal pinned to the bottom right of the viewport. */
	inline?: boolean;
	className?: string;
	limit?: number;
	testId?: string;
}

export const ToastArea = ({ manager, inline = false, className, limit = DEFAULT_TOAST_LIMIT, testId }: ToastAreaProps) => {
	const portalContainer = usePortalContainer();
	const viewport = <ToastViewport inline={inline} className={className} testId={testId} />;

	return (
		<BaseToast.Provider toastManager={manager.base} timeout={DEFAULT_TOAST_DELAY} limit={limit}>
			{inline ? (
				viewport
			) : (
				// Named, because with a `container` the portal renders a wrapper element of its own.
				<BaseToast.Portal className="contents" data-testid="sim-toast-portal" container={portalContainer ?? undefined}>
					{viewport}
				</BaseToast.Portal>
			)}
		</BaseToast.Provider>
	);
};
