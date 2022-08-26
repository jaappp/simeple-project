/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');
const path = require('path');
const { buildCAClient, registerAndEnrollUser, enrollAdmin } = require('./javascript/CAUtil.js');
const { buildCCPOrg1, buildWallet } = require('./javascript/AppUtil.js');

const channelName = 'mychannel';
const chaincodeName = 'ago';
const mspOrg1 = 'Org1MSP';
const walletPath = path.join(__dirname, 'wallet');
const org1UserId = 'appUser';

function prettyJSONString(inputString) {
	return JSON.stringify(JSON.parse(inputString), null, 2);
}

async function main() {
	try {
		const ccp = buildCCPOrg1();
		const caClient = buildCAClient(FabricCAServices, ccp, 'ca.org1.example.com');
		const wallet = await buildWallet(Wallets, walletPath);
		await enrollAdmin(caClient, wallet, mspOrg1);
		await registerAndEnrollUser(caClient, wallet, mspOrg1, org1UserId, 'org1.department1');

		const gateway = new Gateway();

		try {
			  await gateway.connect(ccp, {
				wallet,
				identity: org1UserId,
				discovery: { enabled: true, asLocalhost: true } 
			});

			const network = await gateway.getNetwork(channelName);

			const contract = network.getContract(chaincodeName);
			
			let result;

			console.log('\n--> Submit Transaction: Apply');
			await contract.submitTransaction('Apply','DIS10001','bstudent');
			console.log('*** Result: committed');

			console.log('\n--> Submit Transaction: Supply');
			await contract.submitTransaction('Supply','DIS10001','bstudent');
			console.log('*** Result: committed');

			console.log('\n--> Evaluate Transaction: GetDisplay');
			 result = await contract.evaluateTransaction('GetDisplay','DIS10001');
			console.log(`*** Result: ${prettyJSONString(result.toString())}`);

			
			console.log('\n--> Evaluate Transaction: GetAssetHistory');
			result = await contract.evaluateTransaction('GetAssetHistory','DIS10001');
			console.log(`*** Result: ${prettyJSONString(result.toString())}`);


		} finally {
			gateway.disconnect();
		}
	} catch (error) {
		console.error(`******** FAILED to run the application: ${error}`);
	}
}

main();
