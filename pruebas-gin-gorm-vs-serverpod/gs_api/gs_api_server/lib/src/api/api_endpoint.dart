import 'package:serverpod/serverpod.dart';

import '../generated/protocol.dart';

/// La misma API que el servidor de Go (go/main.go): una tarea por id, los autores con sus tareas y crear una tarea.
class ApiEndpoint extends Endpoint {
  /// Una tarea por id.
  Future<Tarea?> tarea(Session session, int id) async {
    return Tarea.db.findById(session, id);
  }

  /// Los primeros [limite] autores con sus tareas, con include: lo que recomienda Serverpod.
  Future<List<Autor>> autores(Session session, int limite) async {
    return Autor.db.find(
      session,
      limit: limite,
      orderBy: (a) => a.id,
      include: Autor.include(tareas: Tarea.includeList(orderBy: (t) => t.id)),
    );
  }

  /// Lo mismo, a mano: una consulta por autor (el N+1).
  Future<List<Autor>> autoresN1(Session session, int limite) async {
    final lista = await Autor.db.find(session, limit: limite, orderBy: (a) => a.id);
    for (final a in lista) {
      a.tareas = await Tarea.db.find(session, where: (t) => t.autorId.equals(a.id!), orderBy: (t) => t.id);
    }
    return lista;
  }

  /// Crea una tarea y devuelve su id.
  Future<int> crear(Session session, String titulo, int autorId) async {
    final t = await Tarea.db.insertRow(session, Tarea(titulo: titulo, hecha: false, autorId: autorId));
    return t.id!;
  }
}
