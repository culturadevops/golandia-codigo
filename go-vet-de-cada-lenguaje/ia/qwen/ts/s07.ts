async function descargarAdmins() {
    const response = await fetch('https://api.example.com/users');
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    const users = await response.json();
    const admins = users.filter(user => user.role === 'admin');
    await new Promise(resolve => {
        const fs = require('fs');
        fs.writeFile('admins.json', JSON.stringify(admins, null, 2), err => {
            if (err) {
                console.error('Error writing to file:', err);
                return;
            }
            console.log('Admins saved to admins.json');
            resolve();
        });
    });
}

descargarAdmins();
