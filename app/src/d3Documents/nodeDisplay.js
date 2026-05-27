// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {
	mdiArrowCollapseDown,
	mdiArrowLeft,
	mdiArrowRight,
	mdiClockAlertOutline,
	mdiIncognitoOff,
	mdiMerge,
	mdiTune,
} from '@mdi/js';
import {
	SELECTOR_STATUS_SUCCESS,
	SELECTOR_TYPE_HEURISTIC,
	SELECTOR_TYPE_TX_GRAPH,
	SELECTOR_TYPE_TX_PROP,
	WORKSPACE_NODE_TYPE_CLUSTER,
	WORKSPACE_NODE_TYPE_SELECTOR,
	WORKSPACE_NODE_TYPE_TRANSACTION,
} from '@/constants/index.js';
import {
	cashLeft,
	cashRight,
	incognitoFilter,
	sigmaLeft,
	sigmaRight,
} from '@/customIcons/index.js';
import {abbreviateNumber} from '@/d3Documents/util.js';

// Sets
// - result count (number in center of node)
// - node title
// - node icons
export function setNodesDisplayAttributes(nodes, heuristicTypeMap) {
	return nodes.map(d => {
		d.nodeDisplayTitle = getNodeTitle(d, heuristicTypeMap);
		d.nodeDisplayResultCount = getResultCount(d);
		d.nodeDisplayIconObject = getNodeIconObject(d);
		return d;
	});
}

// Returns an object containing
// - icons: a string array with the svg paths of the icons
// - parameter: a heuristic parameter if any
// or null if not applicable
// eslint-disable-next-line complexity
function getNodeIconObject(d) {
	if (d.type !== WORKSPACE_NODE_TYPE_SELECTOR) {
		return null;
	}

	const icons = [];
	let parameter;

	if (d.selectorType === SELECTOR_TYPE_HEURISTIC && d.heuristicOptions) {
		if (d.heuristicOptions.clusterTypes?.length > 0) {
			icons.push(mdiMerge);
		}

		if (d.heuristicOptions.excludeSpendingGaps) {
			icons.push(mdiClockAlertOutline);
		}

		if (d.heuristicOptions.parameter) {
			icons.push(mdiTune);
		}

		parameter = d.heuristicOptions.parameter;
	} else if (d.selectorType === SELECTOR_TYPE_TX_PROP && d.txPropOptions) {
		if (d.txPropOptions.inputRange) {
			icons.push(cashLeft);
		}

		if (d.txPropOptions.outputRange) {
			icons.push(cashRight);
		}

		if (d.txPropOptions.inputSum) {
			icons.push(sigmaLeft);
		}

		if (d.txPropOptions.outputSum) {
			icons.push(sigmaRight);
		}

		if (d.txPropOptions.excludePrivacyTransactions) {
			icons.push(mdiIncognitoOff);
		}

		if (d.txPropOptions.txTypes) {
			icons.push(incognitoFilter);
		}
	} else if (d.selectorType === SELECTOR_TYPE_TX_GRAPH && d.txGraphOptions) {
		if (d.txGraphOptions.excludePrivacyTransactions) {
			icons.push(mdiIncognitoOff);
		}

		if (d.txGraphOptions.isForward) {
			icons.push(mdiArrowRight);
		} else {
			icons.push(mdiArrowLeft);
		}

		if (d.txGraphOptions.depth) {
			icons.push(mdiArrowCollapseDown);
		}

		parameter = d.txGraphOptions.depth;
	}

	return {icons, parameter};
}

// Returns the result count which is displayed centered on the node
function getResultCount(d) {
	if (d.type !== WORKSPACE_NODE_TYPE_SELECTOR || d.selectorStatus !== SELECTOR_STATUS_SUCCESS) {
		return '';
	}

	return abbreviateNumber(d.selectorTotalResultCount);
}

// Returns the display title of the node
function getNodeTitle(d, heuristicTypeMap) {
	if (d.type === WORKSPACE_NODE_TYPE_CLUSTER) {
		return d.addressHash;
	}

	if (d.type === WORKSPACE_NODE_TYPE_TRANSACTION) {
		return d.transactionHash;
	}

	if (d.type === WORKSPACE_NODE_TYPE_SELECTOR) {
		if (d.selectorType === SELECTOR_TYPE_HEURISTIC && d.heuristicOptions && heuristicTypeMap) {
			const title = heuristicTypeMap.get(d.heuristicOptions.type);
			if (title !== undefined) {
				return title;
			}
		} else if (d.selectorType === SELECTOR_TYPE_TX_PROP) {
			if (!d.txPropOptions.startDate || !d.txPropOptions.endDate) {
				return '';
			}

			const dateOptions = {day: 'numeric', month: 'numeric', year: 'numeric'};
			const startDateStr = new Date(d.txPropOptions.startDate).toLocaleDateString(undefined, dateOptions);
			const endDateStr = new Date(d.txPropOptions.endDate).toLocaleDateString(undefined, dateOptions);

			return `${startDateStr} - ${endDateStr}`;
		} else if (d.selectorType === SELECTOR_TYPE_TX_GRAPH) {
			// Only show icons
			return '';
		}
	}

	return d.uid;
}
