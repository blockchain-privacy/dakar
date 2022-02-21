const apiVersion = 'v1';
const routePrefix = `/api/${apiVersion}/`;

// backend routes
export const ROUTE_SEARCH = `${routePrefix}search/`;
export const ROUTE_TRANSACTION = `${routePrefix}tx/`;
export const ROUTE_BLOCK = `${routePrefix}blk/`;
export const ROUTE_BLOCK_RANGE = `${routePrefix}blkRange/`;
export const ROUTE_ADDRESS = `${routePrefix}address/`;
export const ROUTE_ADDRESS_OUTPUT_RANGE = `${routePrefix}addressOutputRange/`;
export const ROUTE_META = `${routePrefix}meta/`;
export const ROUTE_HEURISTICS = `${routePrefix}heuristics/`;
export const ROUTE_HEURISTICS_SUMMARY = `${routePrefix}heuristicsSummary/`;
export const ROUTE_EXECUTE_HEURISTICS = `${routePrefix}executeHeuristics/`;
export const ROUTE_HEURISTIC_DETAILS = `${routePrefix}heuristicDetails/`;
export const ROUTE_HEURISTIC_STATUS = `${routePrefix}heuristicStatus/`;
export const ROUTE_HEURISTIC_LIST = `${routePrefix}heuristicList/`;
export const ROUTE_HEURISTIC_DESCRIPTORS = `${routePrefix}heuristicDescriptors/`;
export const ROUTE_DELETE_HEURISTIC = `${routePrefix}deleteHeuristic/`;
export const ROUTE_USER_LIST = `${routePrefix}getUsers/`;
export const ROUTE_USER_CREATE = `${routePrefix}createUser/`;
export const ROUTE_USER_DELETE = `${routePrefix}deleteUser/`;
export const ROUTE_USER_LOGIN = `${routePrefix}login/`;
export const ROUTE_USER_LOGOUT = `${routePrefix}logout/`;
export const ROUTE_USER_MODIFY = `${routePrefix}modifyUser/`;
export const ROUTE_SHORTEST_TRANSACTION_PATH = `${routePrefix}shortestTransactionPath/`;
export const ROUTE_CONNECTION_LOOKUP = `${routePrefix}connectionLookup/`;
export const ROUTE_CLUSTER_LOOKUP = `${routePrefix}clusterLookup/`;
export const ROUTE_CLUSTER_SUMMARY = `${routePrefix}clusterSummary/`;
export const ROUTE_CLUSTER_HMI_LOOKUP = `${routePrefix}hmiLookup/`;
export const ROUTE_ADD_CLUSTER = `${routePrefix}addCluster/`;
export const ROUTE_DELETE_CLUSTER = `${routePrefix}deleteCluster/`;
export const ROUTE_DELETE_ALL_CLUSTERS = `${routePrefix}deleteAllClusters/`;
export const ROUTE_CLUSTER_OVERVIEW = `${routePrefix}clusterOverview/`;
export const ROUTE_MIXING_ACTIVITY = `${routePrefix}mixingActivity/`;
export const ROUTE_ADD_PRIVATE_ATTRIBUTION = `${routePrefix}addPrivateAttribution/`;
export const ROUTE_ADD_PUBLIC_ATTRIBUTION = `${routePrefix}addPublicAttribution/`;
export const ROUTE_ATTRIBUTION_OVERVIEW = `${routePrefix}attributionOverview/`;
export const ROUTE_DELETE_PRIVATE_ATTRIBUTION = `${routePrefix}deletePrivateAttribution/`;
export const ROUTE_DELETE_PUBLIC_ATTRIBUTION = `${routePrefix}deletePublicAttribution/`;
export const ROUTE_DELETE_ALL_PRIVATE_ATTRIBUTIONS = `${routePrefix}deleteAllPrivateAttributions/`;
export const ROUTE_SEARCH_ATTRIBUTIONS = `${routePrefix}searchAttributions/`;

// search responses
export const RESPONSE_EMPTY = 'response_empty';
export const RESPONSE_TYPE_TRANSACTION = 'tx';
export const RESPONSE_TYPE_ADDRESS = 'addr';
export const RESPONSE_TYPE_BLOCK = 'block';

// frontend route names
export const ROUTE_NAME_ENTRY_PAGE = 'Entry Page';
export const ROUTE_NAME_STATUS_PAGE = 'Status Page';
export const ROUTE_NAME_404_PAGE = 'Page not found';
export const ROUTE_NAME_NO_RESULTS = 'No results found';
export const ROUTE_NAME_LOGIN_PAGE = 'Login Page';
export const ROUTE_NAME_USER_ADMIN_PAGE = 'User Administration Page';
export const ROUTE_NAME_USER_PROFILE_PAGE = 'User Profile Page';
export const ROUTE_NAME_USER_MISC_PAGE = 'User Misc Page';
export const ROUTE_NAME_HEURISTIC_PAGE = 'Heuristic Editor';
export const ROUTE_NAME_BLOCK_PAGE = 'Block Page';
export const ROUTE_NAME_TRANSACTION_PAGE = 'Transaction Page';
export const ROUTE_NAME_ADDRESS_PAGE = 'Address Page';
export const ROUTE_NAME_USER_HEURISTIC_PAGE = 'User Heuristic Page';
export const ROUTE_NAME_SHORTEST_PATH_PAGE = 'User Shortest Path Page';
export const ROUTE_NAME_CONNECTION_LOOKUP_PAGE = 'User Connection Lookup Page';
export const ROUTE_NAME_CLUSTER_OVERVIEW = 'Cluster Overview Page';
export const ROUTE_NAME_CLUSTER_VIEW_PAGE = 'Cluster View Page';
export const ROUTE_NAME_ATTRIBUTIONS = 'Attributions Page';

// application
export const PAGE_TITLE = 'Dakar';
export const APPLICATION_NAME = 'Dakar';
export const APPLICATION_SUBTITLE = 'Dash Blockchain Analytics';
export const LOCALSTORAGE_FIELD_USER = 'user';
export const LOCALSTORAGE_FIELD_SETTINGS = 'settings';
export const DEFAULT_SETTINGS = { dark: false };

// blockchain
export const COIN_UNIT_DASH = 'Dash';
export const COIN_UNIT_BTC = 'BTC';
export const COIN_UNIT_DOGE = 'Doge';
export const COIN_UNIT = COIN_UNIT_DASH;

// user management
// PASSWORD_MIN_CHARACTERS is the number of characters a password must have at least
export const PASSWORD_MIN_CHARACTERS = 10;
// PASSWORD_MAX_CHARACTERS is the number of characters a password can have at most
export const PASSWORD_MAX_CHARACTERS = 250;
// TOKEN_TIMEOUT is the number of milliseconds a token is valid for
// time is 48 hours:  1000 * 60 * 60 * 48 = 172800000
export const TOKEN_TIMEOUT = 172800000;

// cluster
export const CLUSTER_TYPE_HMI = 'hmi';
export const CLUSTER_TYPE_FMI = 'fmi';
export const CLUSTER_TYPE_CUSTOM = 'custom';
