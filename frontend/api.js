import fetch from 'node-fetch';

export default class API {
	constructor(baseUrl = '') {
		this.session = {
			Token: ''
		};
		this.baseUrl = baseUrl;
	}

	login = async function(username, password) {
		const data = {
			Username: username,
			Password: password
		}

		let res = await fetch(`${this.baseUrl}/login`, {
			method: 'post',
			body: JSON.stringify(data)
		});

		if(res.status === 401) {
			throw new Error('Invalid username or password');
		} else if(res.status !== 200) {
			throw new Error(`Unexpected response status code: ${res.status}`);
		}

		this.session = await res.json();
	}

	//Returns array of domains
	listDomains = async function(offset = 0, limit = 200) {
		let res = await fetch(`${this.baseUrl}/domain?offset=${offset}&limit=${limit}`, {
			headers: {
				'API-Token': this.session.Token
			}
		});

		if(res.status !== 200) {
			const responseBody = await res.text();
			throw new Error(responseBody.trim());
		}
		
		const domains = await res.json();
		return domains;
	}

	//Returns newly created domain
	createDomain = async function(domainName) {
		const data = {
			Name: domainName
		}

		let res = await fetch(`${this.baseUrl}/domain`, {
			method: 'post',
			headers: {
				'API-Token': this.session.Token
			},
			body: JSON.stringify(data)
		});

		if(res.status !== 201) {
			const responseBody = await res.text();
			throw new Error(responseBody.trim());
		}
	}

	//Returns newly created domain
	deleteDomain = async function(domainName) {
		let res = await fetch(`${this.baseUrl}/domain/${domainName}`, {
			method: 'delete',
			headers: {
				'API-Token': this.session.Token
			}
		});

		if(res.status !== 200) {
			const responseBody = await res.text();
			throw new Error(responseBody.trim());
		}
	}
}