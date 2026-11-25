import fs from 'fs';
import path from 'path';

async function descargarUsuariosAdmin() {
    const response = await fetch('https://api.example.com/users');
    const usuarios = await response.json();

    const usuariosAdmin = usuarios.filter(usuario => usuario.rol === 'admin');

    if (usuariosAdmin.length > 0) {
        const archivoSalida = path.join(__dirname, 'usuarios_admin.json');
        fs.writeFileSync(archivoSalida, JSON.stringify(usuariosAdmin, null, 2));
        console.log(`Se han guardado ${usuariosAdmin.length} usuarios admin en el archivo ${archivoSalida}`);
    } else {
        console.log('No se encontraron usuarios con el rol admin.');
    }
}

descargarUsuariosAdmin().catch(error => console.error('Error al descargar o procesar los usuarios:', error));
