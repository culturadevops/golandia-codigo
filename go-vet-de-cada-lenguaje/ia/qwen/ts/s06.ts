async function descargarAdmins() {
    const response = await fetch('https://api.example.com/users');
    if (!response.ok) {
        throw new Error('Error al obtener los usuarios');
    }
    const users = await response.json();
    const admins = users.filter(user => user.role === 'admin');
    const adminsJson = JSON.stringify(admins, null, 2);
    await Deno.writeTextFile('admins.json', adminsJson);
}

// Ejecutar la función
descargarAdmins().catch(err => console.error('Error:', err));
