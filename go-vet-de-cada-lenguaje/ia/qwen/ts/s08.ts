const fs = require('fs');

async function descargarAdmins() {
    try {
        const response = await fetch('https://api.example.com/users');
        if (!response.ok) {
            throw new Error('Network response was not ok ' + response.statusText);
        }
        const users = await response.json();
        const admins = users.filter(user => user.role === 'admin');
        await fs.promises.writeFile('admins.json', JSON.stringify(admins, null, 2));
        console.log('Admins saved to admins.json');
    } catch (error) {
        console.error('There was a problem with the fetch operation:', error);
    }
}

descargarAdmins();
