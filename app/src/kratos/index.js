// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {ROUTE_NAME_ENTRY_PAGE, ROUTE_NAME_LOGIN_PAGE, ROUTE_NAME_OAUTH_ERROR_PAGE} from '@/constants';
import {isFunction} from '@/utilities';

async function refreshFlow(onRefreshFlow) {
	if (!isFunction(onRefreshFlow)) {
		return;
	}

	await onRefreshFlow();
}

// Returns true if the error was handled
async function handleErrorCodeAndID(context, error, onRefreshFlow, isOAuth) {
	switch (error.response.error.id) {
		case 'session_already_available':
			context.$router.push({name: ROUTE_NAME_ENTRY_PAGE});
			return true;
		case 'session_aal2_required':
		case 'session_refresh_required':
		case 'browser_location_change_required':
			window.location.href = error.response.redirect_browser_to;
			return true;
		case 'self_service_flow_expired':
		case 'self_service_flow_return_to_forbidden':
		case 'security_identity_mismatch':
			await refreshFlow(onRefreshFlow);
			return true;
		case 'security_csrf_violation':
			context.localStore.setSession(null);

			if (isOAuth) {
				context.$router.push({name: ROUTE_NAME_OAUTH_ERROR_PAGE});
				return true;
			}

			context.navStore.setFailedRoute(context.$route);
			context.$router.push({name: ROUTE_NAME_LOGIN_PAGE});
			return true;
		case 'session_inactive':
			await context.navStore.setFailedRoute(context.$route);
			context.localStore.setSession(null);
			context.$router.push({name: ROUTE_NAME_LOGIN_PAGE});
			return true;
		default:
			break;
	}

	switch (error.response.error.code) {
		case 410: // Expired
			await refreshFlow(onRefreshFlow);
			return true;
		case 401: // Unauthorized access
			// return false so error message can be displayed
			return false;
		default:
			break;
	}

	return false;
}

// HandleGetFlowError tries to handle possible ory kratos error scenarios.
// onRefreshFlow is called when a flow has expired.
export default async function handleGetFlowError(context, error, onRefreshFlow, isOAuth) {
	if (error.response?.error && await handleErrorCodeAndID(context, error, onRefreshFlow, isOAuth)) {
		return;
	}

	if (error.response?.error?.reason || error.response?.error?.status) {
		let msg = '';

		if (error.response.error.status) {
			msg = error.response.error.status;
		}

		if (error.response.error.reason) {
			if (msg !== '') {
				msg += ': ';
			}

			msg += error.response.error.reason;
		}

		return msg;
	}

	if (error.message) {
		return error.message;
	}

	// Return error if it was not possible to handle it
	throw error;
}
