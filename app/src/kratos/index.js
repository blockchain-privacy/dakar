import {ROUTE_NAME_ENTRY_PAGE, ROUTE_NAME_LOGIN_PAGE} from '@/constants';
import {isFunction} from '@/utilities';

async function refreshFlow(onRefreshFlow) {
	if (!isFunction(onRefreshFlow)) {
		return;
	}

	await onRefreshFlow();
}

// Returns true if error has been handled
async function checkErrorID(context, error, onRefreshFlow) {
	switch (error.response.error.id) {
		case 'session_already_available': // User is already signed in, let's redirect them home!
			context.$router.push({name: ROUTE_NAME_ENTRY_PAGE});
			return true;
		case 'session_aal2_required': // 2FA is enabled and enforced, but user did not perform 2fa yet!
		case 'session_refresh_required': // We need to re-authenticate to perform this action
		case 'browser_location_change_required': // Ory Kratos asked us to point the user to this URL.
			window.location.href = error.response.redirect_browser_to;
			return true;
		case 'self_service_flow_expired': // The flow expired, let's request a new one.
		case 'self_service_flow_return_to_forbidden': // The return is invalid, we need a new flow
		case 'security_identity_mismatch': // The requested item was intended for someone else. Let's request a new flow...
			await refreshFlow(onRefreshFlow);
			return true;
		case 'security_csrf_violation': // A CSRF violation occurred, remove session and let user login anew
			context.navStore.setFailedRoute(context.$route);
			context.localStore.setSession(null);
			context.$router.push({name: ROUTE_NAME_LOGIN_PAGE});
			if (error.response.error.message) {
				context.msgStore.addMessage({
					text: error.response.error.reason,
					type: 'error',
					temporary: true,
					category: context.$route.name,
				});
			}

			return true;
		case 'session_inactive':
			await context.navStore.setFailedRoute(context.$route);
			context.localStore.setSession(null);
			context.$router.push({name: ROUTE_NAME_LOGIN_PAGE});
			return true;
		default:
			return false;
	}
}

async function checkErrorStatus(context, error, onRefreshFlow) {
	switch (error.response.status) {
		case 410: // The flow expired, let's request a new one.
			await refreshFlow(onRefreshFlow);
			return true;
		case 401: { // Unauthorized access -> show error
			await refreshFlow(onRefreshFlow);
			let msg = '';
			if (error.response?.error?.reason) {
				msg = error.response.error.reason;
			} else {
				msg = error.response.statusText;
			}

			context.msgStore.addMessage({text: msg, type: 'error', temporary: true});
			return true;
		}

		default:
			return false;
	}
}

// HandleGetFlowError tries to handle possible ory kratos error scenarios.
// onRefreshFlow is called when a flow has expired.
export default async function handleGetFlowError(context, error, onRefreshFlow) {
	if (error.response?.error?.id) {
		if (await checkErrorID(context, error, onRefreshFlow)) {
			return Promise.resolve();
		}
	}

	if (error.response?.status) {
		if (await checkErrorStatus(context, error, onRefreshFlow)) {
			return Promise.resolve();
		}
	}

	if (error.message) {
		context.msgStore.addMessage({
			text: error.message, type: 'error', temporary: true, category: context.$route.name,
		});
	}

	// Return error if it was not possible to handle it
	return Promise.reject(error);
}
