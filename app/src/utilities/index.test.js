// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {expect, test} from 'vitest';
import {extractEntities, filterDescriptors} from '.';
import {SELECTOR_TYPE_HEURISTIC, SELECTOR_STATUS_SUCCESS} from '@/constants/index.js';

test('reactor csv', () => {
	const csvExport = '"This file contains a list of all clusters within the graph with the following name:"\n'
		+ '"Russian Silk Road Market"\n'
		+ '"The graph URL is:"\n'
		+ '"https://reactor.chainalysis.com/graph/57cbe25391e584fe6f8fbc93c854801f"\n'
		+ '\n'
		+ '"Root address","Custom name","Shared name","Chainalysis name","Category","BTC Balance","Transfers","Addresses"\n'
		+ '"12xZoVcb1jrQrWCaN7uCwwHEmMPsYrtPfm","","","Silk Road Marketplace","Darknet market",603.2435549,1478248,830938\n'
		+ '"1P1ik9HkCNSs7Zmgc3LzEN72X1Egwm9Chb","","","","",0,14,3\n'
		+ '"13GdXePufLt9DipRa5AvG18xLcFNN7TrJY","","","","",0,2,1\n'
		+ '"1LAxD3rWsDeE1jBCxiRzNzkwhfQmqQpA8Z","","","","",0,1071,104\n'
		+ '"14j6jLececs66ZQ8ew6vTFNiEn2NupacWJ","","","","",0.000378,24,2\n'
		+ '"19Mz2o9RDABT74SA9njZqMtJXKEzj2qUoH","","","","",0,11,1\n'
		+ '"1PzGnXGvoGGtCcGpqzkJHebZVgM48VL2x4","","","","",0,3,1\n'
		+ '"39jzjt85tcRkqki6BeRzsD4FfF6BZKYQnR","","","","",0.00000666,7,1\n'
		+ '"bc1q9sh6544xls87x7skjzyfhkty4wq7z76vn7qzq9","","","","",0,22,1\n'
		+ '"bc1q5shngj24323nsrmxv99st02na6srekfctt30ch","","","","",0,34,1\n'
		+ '"bc1qmxjefnuy06v345v6vhwpwt05dztztmx4g3y7wp","","","Seized Assets - Silk Road related","Seized funds",0,21,2\n'
		+ '"bc1qfj4trvfnute2kkdfxssymq2p6ztj63r68j5d3t","","","","",0,83,9\n';

	expect(extractEntities(csvExport)).toStrictEqual([
		'7cbe25391e584fe6f8fbc93c8548',
		'12xZoVcb1jrQrWCaN7uCwwHEmMPsYrtPfm',
		'1P1ik9HkCNSs7Zmgc3LzEN72X1Egwm9Chb',
		'13GdXePufLt9DipRa5AvG18xLcFNN7TrJY',
		'1LAxD3rWsDeE1jBCxiRzNzkwhfQmqQpA8Z',
		'14j6jLececs66ZQ8ew6vTFNiEn2NupacWJ',
		'19Mz2o9RDABT74SA9njZqMtJXKEzj2qUoH',
		'1PzGnXGvoGGtCcGpqzkJHebZVgM48VL2x4',
		'39jzjt85tcRkqki6BeRzsD4FfF6BZKYQnR',
		'bc1q9sh6544xls87x7skjzyfhkty4wq7z76vn7qzq9',
		'bc1q5shngj24323nsrmxv99st02na6srekfctt30ch',
		'bc1qmxjefnuy06v345v6vhwpwt05dztztmx4g3y7wp',
		'bc1qfj4trvfnute2kkdfxssymq2p6ztj63r68j5d3t',
	]);
});

test('trm csv', () => {
	const csvExport = '"Type","Chain","Address","Entity URN","Name","Risk Score","Categories","Entity URN",'
		+ '"Txn Hash","Timestamp","From","To","Asset","Value","Value USD"\n'
		+ '"address","btc","bc1q0yvhvvud4nzq65yayd98upjqvz8xyep442kk0j","/entity/wallet_cluster/2147171178717118612,'
		+ '/entity/inet/85.106.119.102,/entity/inet/41.113.253.119","bc1q0yvhvvud4nzq65yayd98upjqvz8xyep442kk0j:btc,'
		+ '85.106.119.102,41.113.253.119","high","Wallet Cluster,Unhosted Wallet,IP Address","","","","","","","",""\n'
		+ '"address","btc","3AKuZXjMb9zBtdaMdTEDUVVP8pCphEuUWV","/entity/wallet_cluster/7462600734850833646",'
		+ '"3AKuZXjMb9zBtdaMdTEDUVVP8pCphEuUWV:btc","medium","Wallet Cluster","","","","","","","",""\n'
		+ '"address","btc","bc1qeff3wuh8y5xu45akx6ulns3mtrfrmaj9g06enj",'
		+ '"/entity/manual/cab8a388-67d5-4d7f-89bf-93c74ded56c2","Wasabi","high","Mixer","","","","","","","",""\n'
		+ '"entity","","","","bc1q0yvhvvud4nzq65yayd98upjqvz8xyep442kk0j:btc","medium","Wallet Cluster",'
		+ '"/entity/wallet_cluster/2147171178717118612","","","","","","",""\n'
		+ '"transfer","btc","","","","","","","7e3b58966f0f8746fa591d5730f74965f46c6af52980517c10b45a6d2a737c03",'
		+ '"2022-03-01 13:25:54.000Z","3AKuZXjMb9zBtdaMdTEDUVVP8pCphEuUWV","","btc","18.12453128","813485.149893368"\n'
		+ '"transfer","btc","","","","","","","8d9726c32b6d00127a52c0302669388db5ae1f9de4430eb7f09cbf5f9ede50a0",'
		+ '"2022-08-06 19:21:29.000Z","bc1q0yvhvvud4nzq65yayd98upjqvz8xyep442kk0j","","btc","18.12452732","420636.2049858384"\n'
		+ '"transfer","btc","","","","","","","7e3b58966f0f8746fa591d5730f74965f46c6af52980517c10b45a6d2a737c03",'
		+ '"2022-03-01 13:25:54.000Z","","bc1q0yvhvvud4nzq65yayd98upjqvz8xyep442kk0j","btc","18.12452732","813484.972156292"\n'
		+ '"transfer","btc","","","","","","","8d9726c32b6d00127a52c0302669388db5ae1f9de4430eb7f09cbf5f9ede50a0",'
		+ '"2022-08-06 19:21:29.000Z","","bc1qeff3wuh8y5xu45akx6ulns3mtrfrmaj9g06enj","btc","18.12450972","420635.7965229264"';

	expect(extractEntities(csvExport)).toStrictEqual([
		'bc1q0yvhvvud4nzq65yayd98upjqvz8xyep442kk0j',
		'3AKuZXjMb9zBtdaMdTEDUVVP8pCphEuUWV',
		'bc1qeff3wuh8y5xu45akx6ulns3mtrfrmaj9g06enj',
		'7e3b58966f0f8746fa591d5730f74965f46c6af52980517c10b45a6d2a737c03',
		'8d9726c32b6d00127a52c0302669388db5ae1f9de4430eb7f09cbf5f9ede50a0',
	]);
});

test('empty', () => {
	expect(extractEntities('asdf')).toStrictEqual([]);
});

test('trim', () => {
	expect(extractEntities('    ')).toStrictEqual([]);
	expect(extractEntities('')).toStrictEqual([]);
});

test('filterDescriptors', () => {
	const descriptors = [
		{allowedParents: ['whirlpool destination']},
		{allowedParents: ['wasabi 2.0 origin']},
		{allowedParents: ['dash_reverse_look']},
		{allowedParents: ['wasabi2_one_source_by_time']},
	];

	const singleNode = [
		{type: 'transaction', txtype: 'wasabi 2.0 origin'},
	];
	expect(filterDescriptors(descriptors, singleNode)).toHaveLength(1);

	const onlyTransactions = [
		{type: 'transaction', txtype: 'wasabi 2.0 origin'},
		{type: 'transaction', txtype: 'wasabi 2.0 origin'},
		{type: 'transaction', txtype: 'mixing'},
	];
	expect(filterDescriptors(descriptors, onlyTransactions)).toHaveLength(1);

	const onlyHeuristics = [
		{
			type: 'selector', heuristicOptions: {type: 'wasabi2_one_source_by_time'},
			selectorType: SELECTOR_TYPE_HEURISTIC, selectorStatus: SELECTOR_STATUS_SUCCESS,
		},
		{
			type: 'selector', heuristicOptions: {type: 'wasabi2_one_source_by_time'},
			selectorType: SELECTOR_TYPE_HEURISTIC, selectorStatus: SELECTOR_STATUS_SUCCESS,
		},
	];
	expect(filterDescriptors(descriptors, onlyHeuristics)).toHaveLength(1);

	const mixed = [
		{
			type: 'selector', heuristicOptions: {type: 'wasabi2_one_source_by_time'},
			selectorType: SELECTOR_TYPE_HEURISTIC, selectorStatus: SELECTOR_STATUS_SUCCESS,
		},
		{type: 'transaction', txtype: 'wasabi 2.0 origin'},
		{
			type: 'selector', heuristicOptions: {type: 'wasabi2_one_source_by_time'},
			selectorType: SELECTOR_TYPE_HEURISTIC, selectorStatus: SELECTOR_STATUS_SUCCESS,
		},
		{type: 'transaction', txtype: 'mixing'},
	];
	expect(filterDescriptors(descriptors, mixed)).toHaveLength(2);

	const fail = [
		{
			type: 'selector', heuristicOptions: {type: 'wasabi2_one_source_by_time'},
			selectorType: SELECTOR_TYPE_HEURISTIC, selectorStatus: SELECTOR_STATUS_SUCCESS,
		},
		{type: 'transaction', txtype: 'wasabi 2.0 origin'},
		{
			type: 'selector', heuristicOptions: {type: 'wasabi2_one_source_by_time'},
			selectorType: SELECTOR_TYPE_HEURISTIC, selectorStatus: SELECTOR_STATUS_SUCCESS,
		},
		{type: 'transaction', txtype: 'mixing'},
		{},
	];
	expect(() => filterDescriptors(descriptors, fail)).toThrowError('invalid node type');

	expect(filterDescriptors(descriptors, fail, false)).toHaveLength(2);
});
