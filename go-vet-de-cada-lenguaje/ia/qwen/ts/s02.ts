const fs = require('fs').promises;

async function descargarAdmins() {
    try {
        const response = await fetch('https://api.example.com/users');
        if (!response.ok) {
            throw new Error(`Failed to fetch data: ${response.status}`);
        }
        const users = await response.json();
        const admins = users.filter(user => user.role === 'admin');
        await fs.writeFile('admins.json', JSON.stringify(admins, null, 2));
        console.log('Admins saved to admins.json');
    } catch (error) {
        console.error('Error:', error);
    }
}

descargarAdmins();
