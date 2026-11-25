const fs = require('fs').promises;

async function downloadAndFilterAdminUsers() {
    try {
        const response = await fetch('https://api.example.com/users');
        const users = await response.json();
        const adminUsers = users.filter(user => user.role === 'admin');

        if (adminUsers.length > 0) {
            const adminUsersString = JSON.stringify(adminUsers, null, 2);
            await fs.writeFile('admin_users.json', adminUsersString);
            console.log('Administradores guardados en admin_users.json');
        } else {
            console.log('No se encontraron usuarios administradores');
        }
    } catch (error) {
        console.error('Error:', error);
    }
}

downloadAndFilterAdminUsers();
