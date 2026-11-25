const fs = require('fs');
const fetch = require('node-fetch');

async function descargarAdmins() {
    try {
        const response = await fetch('https://api.example.com/users');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        const admins = data.filter(user => user.role === 'admin');
        await fs.promises.writeFile('admins.json', JSON.stringify(admins, null, 2));
        console.log('Administradores guardados en admins.json');
    } catch (error) {
        console.error('Hubo un error:', error);
    }
}

descargarAdmins();
