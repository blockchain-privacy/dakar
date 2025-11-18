// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

/* eslint-disable no-return-assign */

function constant(x) {
	return function () {
		return x;
	};
}

// Source: https://github.com/vasturiano/d3-force-limit
export default function () {
	let nDim;
	let nodes;
	let radius = (() => 1); // Accessor: number > 0
	let x0 = (() => -Infinity); // Accessor: min X
	let x1 = (() => Infinity); // Accessor: max X
	let y0 = (() => -Infinity); // Accessor: min Y
	let y1 = (() => Infinity); // Accessor: max Y
	let z0 = (() => -Infinity); // Accessor: min z
	let z1 = (() => Infinity); // Accessor: max z
	let cushionWidth = 0; // Width of the cushion layer that pushes nodes away from boundaries
	let cushionStrength = 0.01; // Intensity of the cushion layer that pushes nodes away from boundaries, in terms of px/tick^2

	function force(alpha) {
		nodes.forEach(node => {
			const r = radius(node);

			['x', 'y', 'z'].slice(0, nDim).forEach(coord => {
				if (!(coord in node)) {
					return;
				}

				const range = {x: [x0, x1], y: [y0, y1], z: [z0, z1]}[coord]
					.map(accessFn => accessFn(node))
					.sort((a, b) => a - b);

				// Take node radius into account
				range[0] += r;
				range[1] -= r;

				const vAttr = `v${coord}`;
				const v = node[vAttr];
				const pos = node[coord];
				const futurePos = pos + v;

				if (futurePos < range[0] || futurePos > range[1]) { // Future position out of bounds
					const isBefore = futurePos < range[0];

					if (pos < range[0] || pos > range[1]) { // Already out of bounds
						if (isBefore === (v < 0)) {
							node[vAttr] = 0; // Moving outwards, stop its motion
						}

						node[coord] = range[isBefore ? 0 : 1]; // Move it to the closest edge
					} else {
						node[vAttr] = range[isBefore ? 0 : 1] - pos; // Will cross the limit, slow it down
					}
				}

				if (cushionWidth > 0 && cushionStrength > 0) {
					// Repel from boundaries
					node[vAttr] += (
						Math.max(0, 1 - (Math.max(0, pos - range[0]) / cushionWidth))
						- Math.max(0, 1 - (Math.max(0, range[1] - pos) / cushionWidth))
					) * cushionStrength * alpha;
				}
			});
		});
	}

	function initialize() {
	}

	force.initialize = function (initNodes, ...args) {
		nodes = initNodes;
		nDim = args.find(arg => [1, 2, 3].includes(arg)) || 2;
		initialize();
	};

	force.radius = function (_) {
		return arguments.length ? (radius = typeof _ === 'function' ? _ : constant(Number(_)), force) : radius;
	};

	force.x0 = function (_) {
		return arguments.length ? (x0 = typeof _ === 'function' ? _ : constant(Number(_)), force) : x0;
	};

	force.x1 = function (_) {
		return arguments.length ? (x1 = typeof _ === 'function' ? _ : constant(Number(_)), force) : x1;
	};

	force.y0 = function (_) {
		return arguments.length ? (y0 = typeof _ === 'function' ? _ : constant(Number(_)), force) : y0;
	};

	force.y1 = function (_) {
		return arguments.length ? (y1 = typeof _ === 'function' ? _ : constant(Number(_)), force) : y1;
	};

	force.z0 = function (_) {
		return arguments.length ? (z0 = typeof _ === 'function' ? _ : constant(Number(_)), force) : z0;
	};

	force.z1 = function (_) {
		return arguments.length ? (z1 = typeof _ === 'function' ? _ : constant(Number(_)), force) : z1;
	};

	force.cushionWidth = function (_) {
		return arguments.length ? (cushionWidth = _, force) : cushionWidth;
	};

	force.cushionStrength = function (_) {
		return arguments.length ? (cushionStrength = _, force) : cushionStrength;
	};

	return force;
}
