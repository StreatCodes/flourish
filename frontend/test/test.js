import assert from 'assert';
import mocha from 'mocha';

import API from '../api.js';

describe('Authentication', function() {
	const api = new API('https://localhost:8080');
	it('should throw when providing invalid credentials', async function() {
		await assert.rejects(
			api.login('FakeUser', 'FakePassword'),
			new Error("Invalid username or password")
		);
	});

	//Domains
	it('should require admin session before listing domains', async function() {
		await assert.rejects(
			api.listDomains(),
			new Error('API Key required')
		)
	});
	it('should require admin session before creating a domain', async function() {
		await assert.rejects(
			api.createDomain('example.com'),
			new Error('API Key required')
		);
	});
	it('should require admin session before deleting a domain', async function() {
		await assert.rejects(
			api.deleteDomain('example.com'),
			new Error('API Key required')
		);
	});

	//Users
	it('should require admin session before listing a domain\'s users', async function() {
		await assert.rejects(
			api.listUsers('example.com'),
			new Error('API Key required')
		)
	});

	it('should return new session as string', async function() {
		await assert.doesNotReject(
			api.login('mort@streats.dev', 'computer')
		);
	});
});

describe('Domains', async function() {
	const api = new API('https://localhost:8080');
	before(async () => {
		await api.login('mort@streats.dev', 'computer').catch(e => console.log(e));
	});

	it('should fail to create new domain with invalid name', async function() {
		await assert.rejects(
			api.createDomain('invalid(*&AS!!!SDasd'),
			new Error('Invalid domain')
		)
	});
	it('should return 201 HTTP status created', async function() {
		await assert.doesNotReject(
			api.createDomain('example.com')
		);
	});
	it('should show the newly created domain in the domain list', async function() {
		const domains = await api.listDomains(0, 999999);
		const found = domains.find(d => d.Name === 'example.com');
		assert.notStrictEqual(found, undefined);
	});
	it('should not include a non-existant domain', async function() {
		const domains = await api.listDomains(0, 999999);
		const notFound = domains.find(d => d.Name === 'non.existant.domain.name.com');
		assert.strictEqual(notFound, undefined);
	});
	it('should not allow multiple domains to be created with the same name', async function() {
		await assert.rejects(
			api.createDomain('example.com'),
			new Error('Domain already exists')
		);
	});
	it('should remove domain and all child data', async function() {
		await assert.doesNotReject(api.deleteDomain('example.com'))
		const domains = await api.listDomains(0, 999999);
		assert.strictEqual(domains.includes('example.com'), false);
	})
	it('should fail to delete a non-existant domain', async function() {
		await assert.rejects(
			api.deleteDomain('a.non.existant.domain.com'),
			new Error('Domain does not exist')
		);
	});
});

describe('Users', async function() {
	const api = new API('https://localhost:8080');
	before(async () => await api.login('mort@streats.dev', 'computer'));
	before(async () => await api.createDomain('example.com'))
	after(async () => await api.deleteDomain('example.com'))

	it('should throw when listing users for a non-existant domain', async function() {
		await assert.rejects(
			api.listUsers('non.existant.domain.example.com')
		)
	});
	it('should throw when creating a user with an invalid name', async function() {
		await assert.rejects(
			api.createUser('example.com', 'invalid(*&AS!@@@@!!@SDasd', 'aS2d@1dma('),
			new Error('Invalid username')
		)
	});
	it('should successfully create new user', async function() {
		await assert.doesNotReject(
			api.createUser('example.com', 'user', 'aS2d@1dma(')
		);
	});
	it('should include the new user in the list of domain users', async function() {
		const users = await api.listUsers('example.com', 0, 999999);
		const found = users.find(u => u.Username === 'user');
		assert.notStrictEqual(found, undefined);
	});
	it('should not be able to create a second user with the same username', async function() {
		await assert.rejects(
			api.createUser('example.com', 'user', 'aS2d@1dma('),
			new Error('User already exists')
		);
	});
	//TODO
	// it('should require the user to enter an adequate password', async function() {
	// 	await assert.rejects(
	// 		api.createUser('example.com', 'user', 'password'),
	// 		new Error('Inadequate password')
	// 	);
	// });
	// it('should require the username to be atleast 1 character long', async function() {
	// 	await assert.rejects(
	// 		api.createUser('example.com', '', 'password'),
	// 		new Error('Username must be atleast 1 character')
	// 	);
	// });
});