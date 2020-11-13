const apiVersion = 'v1'
const routePrefix = `/api/${apiVersion}/`;

// backend routes
export const ROUTE_TRANSACTION = routePrefix + 'tx/';
export const ROUTE_BLOCK = routePrefix + 'blk/';
export const ROUTE_ADDRESS = routePrefix + 'address/';
export const ROUTE_META = routePrefix + 'meta/';
export const ROUTE_PATHS = routePrefix + 'paths/';

// frontend route names
export const ROUTE_NAME_SEARCH_PAGE = 'Search Page';
export const ROUTE_NAME_ENTRY_PAGE = 'Entry Page';
export const ROUTE_NAME_404_PAGE = 'Page not found';
export const ROUTE_NAME_HEURISTIC_PAGE = 'Heuristic Editor';

// application
export const PAGE_TITLE = 'Dakar';
export const APPLICATION_NAME = 'Dakar';
export const CSV_DOWNLOAD_MAX_ORIGINS = 1000;
