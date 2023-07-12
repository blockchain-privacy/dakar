import { ROUTE_NAME_ENTRY_PAGE, ROUTE_NAME_LOGIN_PAGE } from '../constants';

function refreshFlow(router) {
  router.push({
    // for our use case, removing the flow
    // parameter from the search query and
    // reloading the page are sufficient
    // to refresh the flow
    query: {},
  });

  window.location.reload();
}

export default function handleGetFlowError(router, store, error) {
  if (error.response && error.response.data && error.response.data.error) {
    switch (error.response.data.error.id) {
      case 'session_already_available': // User is already signed in, let's redirect them home!
        router.push({ name: ROUTE_NAME_ENTRY_PAGE });
        return Promise.resolve();
      case 'session_aal2_required': // 2FA is enabled and enforced, but user did not perform 2fa yet!
      case 'session_refresh_required': // We need to re-authenticate to perform this action
      case 'browser_location_change_required': // Ory Kratos asked us to point the user to this URL.
        window.location.href = error.response.data.redirect_browser_to;
        return Promise.resolve();
      case 'self_service_flow_expired': // The flow expired, let's request a new one.
      case 'self_service_flow_return_to_forbidden': // the return is invalid, we need a new flow
      case 'security_csrf_violation': // A CSRF violation occurred. Best to just refresh the flow!
      case 'security_identity_mismatch': // The requested item was intended for someone else. Let's request a new flow...
        refreshFlow(router);
        return Promise.resolve();
      case 'session_inactive':
        store.dispatch('setFailedRoute', router.history.current);
        store.dispatch('setSession', null);
        router.push({ name: ROUTE_NAME_LOGIN_PAGE });
        return Promise.resolve();
      default:
    }
  }

  if (error.response && error.response.status) {
    switch (error.response.status) {
      case 410: // The flow expired, let's request a new one.
        refreshFlow(router);
        return Promise.resolve();
      case 401: { // Unauthorized access -> show error
        let msg = '';
        if (error.response.data && error.response.data.error && error.response.data.error.reason) {
          msg = error.response.data.error.reason;
        } else {
          msg = error.response.statusText;
        }
        store.dispatch('addMessage', { text: msg, type: 'error', temporary: true });
        return Promise.resolve();
      }
      default:
    }
  }

  if (error.message) {
    store.dispatch('addMessage', { text: error.message, type: 'error', temporary: true });
  }

  // We are not able to handle the error? Return it.
  return Promise.reject(error);
}
