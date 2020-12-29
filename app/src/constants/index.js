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

// search responses
export const RESPONSE_EMPTY = 'response_empty';
export const RESPONSE_TYPE_TRANSACTION = 'tx';
export const RESPONSE_TYPE_ADDRESS = 'addr';
export const RESPONSE_TYPE_BLOCK = 'block';

// frontend route names
export const ROUTE_NAME_ENTRY_PAGE = 'Entry Page';
export const ROUTE_NAME_404_PAGE = 'Page not found';
export const ROUTE_NAME_NO_RESULTS = 'No results found';
export const ROUTE_NAME_HEURISTIC_PAGE = 'Heuristic Editor';
export const ROUTE_NAME_BLOCK_PAGE = 'Block Page';
export const ROUTE_NAME_TRANSACTION_PAGE = 'Transaction Page';
export const ROUTE_NAME_ADDRESS_PAGE = 'Address Page';

// application
export const PAGE_TITLE = 'Dakar';
export const APPLICATION_NAME = 'Dakar';
export const CSV_DOWNLOAD_MAX_ORIGINS = 1000;
