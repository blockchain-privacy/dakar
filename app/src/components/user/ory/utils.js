// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

// GetNodeName returns either the name or id of the given node.
// If neither are available returns a string containing a random number.
export function getNodeName(node) {
	if (node.attributes?.name) {
		return node.attributes.name;
	}

	if (node.attributes?.id) {
		return node.attributes.id;
	}

	return `${Math.random() * 1e6}`;
}

export function isUiNodeInputAttributes(attrs) {
	return attrs.node_type === 'input';
}

export function isUiNodeImageAttributes(attrs) {
	return attrs.node_type === 'img';
}

export function isUiNodeAnchorAttributes(attrs) {
	return attrs.node_type === 'a';
}

export function isUiNodeScriptAttributes(attrs) {
	return attrs.node_type === 'script';
}

export function isUiNodeTextAttributes(attrs) {
	return attrs.node_type === 'text';
}
