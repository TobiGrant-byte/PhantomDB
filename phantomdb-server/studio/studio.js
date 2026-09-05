let authHeader = null;

function apiFetch(url) {
	return fetch(url, { headers: { Authorization: authHeader } });
}

document.getElementById('login-form').addEventListener('submit', async (e) => {
	e.preventDefault();
	const username = document.getElementById('login-username').value;
	const password = document.getElementById('login-password').value;
	const candidate = 'Basic ' + btoa(`${username}:${password}`);

	const res = await fetch('/api/collections', { headers: { Authorization: candidate } });

	if (res.status === 401) {
		document.getElementById('login-error').style.display = 'block';
		return;
	}

	authHeader = candidate;
	document.getElementById('login-screen').style.display = 'none';
	document.getElementById('app').classList.add('visible');
	loadCollections();
});

async function loadCollections() {
	const res = await apiFetch('/api/collections');
	const collections = await res.json();
	const list = document.getElementById('collection-list');
	list.innerHTML = '';
	Object.keys(collections).sort().forEach(name => {
		const div = document.createElement('div');
		div.className = 'collection';
		div.textContent = `${name}  (${collections[name]})`;
		div.onclick = () => loadCollection(name, div);
		list.appendChild(div);
	});
}

async function loadCollection(name, clickedEl) {
	document.querySelectorAll('.collection').forEach(el => el.classList.remove('active'));
	clickedEl.classList.add('active');

	const res = await apiFetch(`/api/scan?start=${encodeURIComponent(name + ':')}&end=${encodeURIComponent(name + ':\xff')}`);
	const rows = await res.json();

	document.getElementById('empty').style.display = 'none';
	const table = document.getElementById('data-table');
	table.style.display = 'table';
	const body = document.getElementById('data-body');
	body.innerHTML = '';

	rows.forEach(row => {
		const tr = document.createElement('tr');
		const keyTd = document.createElement('td');
		keyTd.textContent = row.Key;
		const valTd = document.createElement('td');
		valTd.textContent = row.Value.length > 200 ? row.Value.slice(0, 200) + '...' : row.Value;
		tr.appendChild(keyTd);
		tr.appendChild(valTd);
		body.appendChild(tr);
	});
}