const fs = require('fs').promises;

async function descargarUsuariosAdmin() {
    const response = await fetch('https://api.example.com/users');
    const usuarios = await response.json();
    const usuariosAdmin = usuarios.filter(usuario => usuario.rol === 'admin');

    if (usuariosAdmin.length > 0) {
        const contenido = JSON.stringify(usuariosAdmin, null, 2);
        await fs.writeFile('usuariosAdmin.json', contenido, 'utf8');
        console.log('Usuarios admin guardados en usuariosAdmin.json');
    } else {
        console.log('No se encontraron usuarios admin.');
    }
}

descargarUsuariosAdmin();
