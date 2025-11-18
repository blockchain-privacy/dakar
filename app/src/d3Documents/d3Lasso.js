// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later OR BSD-3-Clause

// Copied from https://github.com/skokenes/d3-lasso and modernized
// Copyright 2016, Speros Kokenes
// All rights reserved.
//
//   Redistribution and use in source and binary forms, with or without modification,
//   are permitted provided that the following conditions are met:
//
//   * Redistributions of source code must retain the above copyright notice, this
// list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// * Neither the name of the author nor the names of contributors may be used to
// endorse or promote products derived from this software without specific prior
// written permission.
//
//   THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
// ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

import {drag} from 'd3-drag';
import {pointers} from 'd3-selection';

function isPointInside(vs, point) {
	// Ray-casting algorithm based on
	// https://wrf.ecse.rpi.edu/Research/Short_Notes/pnpoly.html

	const x = point[0];
	const y = point[1];

	let inside = false;
	for (let i = 0, j = vs.length - 1; i < vs.length; j = i++) {
		const xi = vs[i][0];
		const	yi = vs[i][1];
		const xj = vs[j][0];
		const	yj = vs[j][1];

		const intersect = ((yi > y) !== (yj > y))
			&& (x < ((xj - xi) * ((y - yi) / (yj - yi))) + xi);
		if (intersect) {
			inside = !inside;
		}
	}

	return inside;
}

export default function lasso() {
	let items = [];
	let closePathDistance = 75;
	let closePathSelect = true;
	let isPathClosed = false;
	let hoverSelect = true;
	let dragFilter = null;
	let targetArea;
	const on = {
		start() {
		}, draw() {
		}, end() {
		},
	};

	// Function to execute on call
	function lasso(_this) {
		// Add a new group for the lasso
		const g = _this.append('g').attr('class', 'lasso');

		// Add the drawn path for the lasso
		const dynPath = g.append('path').attr('class', 'drawn');

		// Add a closed path
		const closePath = g.append('path').attr('class', 'loop_close');

		// Add an origin node
		const originNode = g.append('circle').attr('class', 'origin');

		// The transformed lasso path for rendering
		let tPath;

		// The lasso origin for calculations
		let origin;

		// The transformed lasso origin for rendering
		let tOrigin;

		// Store off coordinates drawn
		let drawnCoords;

		// Apply drag behaviors
		const dragAction = drag()
			.filter(e => dragFilter(e))
			.on('start', dragstart)
			.on('drag', dragMove)
			.on('end', dragend);

		// Call drag
		targetArea.call(dragAction);

		function dragstart() {
			// Init coordinates
			drawnCoords = [];

			// Initialize paths
			tPath = '';
			dynPath.attr('d', null);
			closePath.attr('d', null);

			// Set every item to have a false selection and reset their center point and counters
			items.nodes().forEach(e => {
				e.__lasso.possible = false;
				e.__lasso.selected = false;
				e.__lasso.hoverSelect = false;
				e.__lasso.loopSelect = false;

				const box = e.getBoundingClientRect();
				e.__lasso.lassoPoint = [Math.round(box.left + (box.width / 2)), Math.round(box.top + (box.height / 2))];
			});

			// If hover is on, add hover function
			if (hoverSelect) {
				items.on('mouseover.lasso', function () {
					// If hovered, change lasso selection attribute to true
					this.__lasso.hoverSelect = true;
				});
			}

			// Run user defined start function
			on.start();
		}

		function dragMove(e) {
			// Get mouse position within body, used for calculations
			let x;
			let y;

			if (e.sourceEvent.type === 'touchmove') {
				x = e.sourceEvent.touches[0].clientX;
				y = e.sourceEvent.touches[0].clientY;
			} else {
				x = e.sourceEvent.clientX;
				y = e.sourceEvent.clientY;
			}

			// Get mouse position within drawing area
			const pointerEvent = pointers(e, targetArea.node())[0];
			const tx = pointerEvent[0];
			const ty = pointerEvent[1];

			// Initialize the path or add the latest point to it
			if (tPath === '') {
				tPath = tPath + 'M ' + tx + ' ' + ty;
				origin = [x, y];
				tOrigin = [tx, ty];
				// Draw origin node
				originNode
					.attr('cx', tx)
					.attr('cy', ty)
					.attr('r', 4)
					.attr('display', null);
			} else {
				tPath = tPath + ' L ' + tx + ' ' + ty;
			}

			drawnCoords.push([x, y]);

			// Calculate the current distance from the lasso origin
			const distance = Math.sqrt(((x - origin[0]) ** 2) + ((y - origin[1]) ** 2));

			// Set the closed path line
			const closeDrawPath = 'M ' + tx + ' ' + ty + ' L ' + tOrigin[0] + ' ' + tOrigin[1];

			// Draw the lines
			dynPath.attr('d', tPath);

			closePath.attr('d', closeDrawPath);

			// Check if the path is closed
			isPathClosed = distance <= closePathDistance;

			// If within the closed path distance parameter, show the closed path. otherwise, hide it
			if (isPathClosed && closePathSelect) {
				closePath.attr('display', null);
			} else {
				closePath.attr('display', 'none');
			}

			items.nodes().forEach(n => {
				n.__lasso.loopSelect = (isPathClosed && closePathSelect) ? (isPointInside(drawnCoords, n.__lasso.lassoPoint)) : false;
				n.__lasso.possible = n.__lasso.hoverSelect || n.__lasso.loopSelect;
			});

			on.draw();
		}

		function dragend() {
			// Remove mouseover tagging function
			items.on('mouseover.lasso', null);

			items.nodes().forEach(n => {
				n.__lasso.selected = n.__lasso.possible;
				n.__lasso.possible = false;
			});

			// Clear lasso
			dynPath.attr('d', null);
			closePath.attr('d', null);
			originNode.attr('display', 'none');

			// Run user defined end function
			on.end();
		}
	}

	// Set or get list of items for lasso to select
	lasso.items = function (_) {
		if (!arguments.length) {
			return items;
		}

		items = _;
		const nodes = items.nodes();
		nodes.forEach(n => {
			n.__lasso = {
				possible: false,
				selected: false,
			};
		});
		return lasso;
	};

	// Return possible items
	lasso.possibleItems = function () {
		return items.filter(function () {
			return this.__lasso.possible;
		});
	};

	// Return selected items
	lasso.selectedItems = function () {
		return items.filter(function () {
			return this.__lasso.selected;
		});
	};

	// Return not possible items
	lasso.notPossibleItems = function () {
		return items.filter(function () {
			return !this.__lasso.possible;
		});
	};

	// Return not selected items
	lasso.notSelectedItems = function () {
		return items.filter(function () {
			return !this.__lasso.selected;
		});
	};

	// Distance required before path auto closes loop
	lasso.dragFilter = function (_) {
		if (!arguments.length) {
			return dragFilter;
		}

		dragFilter = _;
		return lasso;
	};

	// Distance required before path auto closes loop
	lasso.closePathDistance = function (_) {
		if (!arguments.length) {
			return closePathDistance;
		}

		closePathDistance = _;
		return lasso;
	};

	// Option to loop select or not
	lasso.closePathSelect = function (_) {
		if (!arguments.length) {
			return closePathSelect;
		}

		closePathSelect = _ === true;
		return lasso;
	};

	// Not sure what this is for
	lasso.isPathClosed = function (_) {
		if (!arguments.length) {
			return isPathClosed;
		}

		isPathClosed = _ === true;
		return lasso;
	};

	// Option to select on hover or not
	lasso.hoverSelect = function (_) {
		if (!arguments.length) {
			return hoverSelect;
		}

		hoverSelect = _ === true;
		return lasso;
	};

	// Events
	lasso.on = function (type, _) {
		if (!arguments.length) {
			return on;
		}

		if (arguments.length === 1) {
			return on[type];
		}

		const types = ['start', 'draw', 'end'];
		if (types.indexOf(type) > -1) {
			on[type] = _;
		}

		return lasso;
	};

	// Area where lasso can be triggered from
	lasso.targetArea = function (_) {
		if (!arguments.length) {
			return targetArea;
		}

		targetArea = _;
		return lasso;
	};

	return lasso;
}
