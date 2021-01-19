const apiVersion = 'v1';
const routePrefix = `/api/${apiVersion}/`;

// backend routes
export const ROUTE_SEARCH = `${routePrefix}search/`;
export const ROUTE_TRANSACTION = `${routePrefix}tx/`;
export const ROUTE_BLOCK = `${routePrefix}blk/`;
export const ROUTE_ADDRESS = `${routePrefix}address/`;
export const ROUTE_ADDRESS_OUTPUT_RANGE = `${routePrefix}addressOutputRange/`;
export const ROUTE_META = `${routePrefix}meta/`;
export const ROUTE_PATHS = `${routePrefix}paths/`;
export const ROUTE_HEURISTICS = `${routePrefix}heuristics/`;
export const ROUTE_HEURISTICS_SUMMARY = `${routePrefix}heuristicsSummary/`;
export const ROUTE_EXECUTE_HEURISTICS = `${routePrefix}executeHeuristics/`;
export const ROUTE_HEURISTIC_DETAILS = `${routePrefix}heuristicDetails/`;
export const ROUTE_HEURISTIC_STATUS = `${routePrefix}heuristicStatus/`;
export const ROUTE_USER_LIST = `${routePrefix}getUsers/`;
export const ROUTE_USER_CREATE = `${routePrefix}createUser/`;
export const ROUTE_USER_DELETE = `${routePrefix}deleteUser/`;
export const ROUTE_USER_LOGIN = `${routePrefix}login/`;
export const ROUTE_USER_LOGOUT = `${routePrefix}logout/`;

// search responses
export const RESPONSE_EMPTY = 'response_empty';
export const RESPONSE_TYPE_TRANSACTION = 'tx';
export const RESPONSE_TYPE_ADDRESS = 'addr';
export const RESPONSE_TYPE_BLOCK = 'block';

// frontend route names
export const ROUTE_NAME_ENTRY_PAGE = 'Entry Page';
export const ROUTE_NAME_404_PAGE = 'Page not found';
export const ROUTE_NAME_NO_RESULTS = 'No results found';
export const ROUTE_NAME_LOGIN_PAGE = 'Login Page';
export const ROUTE_NAME_USER_ADMIN_PAGE = 'User Administration Page';
export const ROUTE_NAME_HEURISTIC_PAGE = 'Heuristic Editor';
export const ROUTE_NAME_BLOCK_PAGE = 'Block Page';
export const ROUTE_NAME_TRANSACTION_PAGE = 'Transaction Page';
export const ROUTE_NAME_ADDRESS_PAGE = 'Address Page';

// application
export const PAGE_TITLE = 'Dakar';
export const APPLICATION_NAME = 'Dakar';
export const CSV_DOWNLOAD_MAX_ORIGINS = 1000;
export const LOCALSTORAGE_FIELD_USER_EMAIL = 'user_email';
export const LOCALSTORAGE_FIELD_USER_ROLES = 'user_roles';

// blockchain
export const COIN_UNIT_DASH = 'Dash';
export const COIN_UNIT_BTC = 'BTC';
export const COIN_UNIT_DOGE = 'Doge';
export const COIN_UNIT = COIN_UNIT_DASH;

// user management
// PASSWORD_MIN_CHARACTERS is the number of character a password must have at least
export const PASSWORD_MIN_CHARACTERS = 10;
export const PASSWORD_MAX_CHARACTERS = 250;
