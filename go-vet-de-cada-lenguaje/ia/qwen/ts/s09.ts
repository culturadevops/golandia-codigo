async function descargarAdmins() {
    const url = 'https://api.example.com/users';
    try {
        const response = await fetch(url);
        const usuarios = await response.json();
        const admins = usuarios.filter(usuario => usuario.rol === 'admin');
        await writeAdminsToFile(admins);
        console.log('Admins descargados y guardados con éxito.');
    } catch (error) {
        console.error('Error al descargar usuarios:', error);
    }
}

function writeAdminsToFile(admins: any[]) {
    const data = JSON.stringify(admins, null, 2);
    const fs = require('fs');
    fs.writeFile('admins.json', data, (err) => {
        if (err) {
            console.error('Error al guardar los admins en el archivo:', err);
        } else {
            console.log('Archivo admins.json creado con éxito.');
        }
    });
}

// Ejecutar la función
descargarAdmins();
