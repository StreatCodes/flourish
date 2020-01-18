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
	it('should return new session as string', async function() {
		await assert.doesNotReject(
			api.login('mort@streats.dev', 'computer')
		);
	});
});

describe('Domains', async function() {
	const api = new API('https://localhost:8080');

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


	it('should return 201 HTTP status created', async function() {
		await api.login('mort@streats.dev', 'computer');
		await assert.doesNotReject(
			api.createDomain('example.com')
		);
	});
	it('should not allow multiple domains to be created with the same name', async function() {
		await api.login('mort@streats.dev', 'computer');
		await assert.rejects(
			api.createDomain('example.com'),
			new Error('Domain already exists')
		);
	});
	it('should remove domain and all child data', async function() {
		await api.login('mort@streats.dev', 'computer');
		await assert.doesNotReject(api.deleteDomain('example.com'))
		const domains = await api.listDomains(0, 999999);
		assert.strictEqual(domains.includes('example.com'), false);
	})
	it('should fail to delete a non-existant domain', async function() {
		await api.login('mort@streats.dev', 'computer');
		await assert.rejects(
			api.deleteDomain('a.non.existant.domain.com'),
			new Error('Domain does not exist')
		);
	});
});