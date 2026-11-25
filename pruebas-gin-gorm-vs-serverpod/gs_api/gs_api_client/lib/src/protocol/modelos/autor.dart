/* AUTOMATICALLY GENERATED CODE DO NOT MODIFY */
/*   To generate run: "serverpod generate"    */

// ignore_for_file: implementation_imports
// ignore_for_file: library_private_types_in_public_api
// ignore_for_file: non_constant_identifier_names
// ignore_for_file: public_member_api_docs
// ignore_for_file: type_literal_in_constant_pattern
// ignore_for_file: use_super_parameters
// ignore_for_file: invalid_use_of_internal_member

// ignore_for_file: no_leading_underscores_for_library_prefixes
import 'package:gs_api_client/src/protocol/protocol.dart' as _idw36flt;
import 'package:serverpod_client/serverpod_client.dart' as _isc;
import '../modelos/tarea.dart' as _i8g1zmuz;

abstract class Autor
    implements _isc.SerializableModel, _isc.ProtocolSerialization {
  Autor._({
    this.id,
    required this.nombre,
    this.tareas,
  });

  factory Autor({
    int? id,
    required String nombre,
    List<_i8g1zmuz.Tarea>? tareas,
  }) = _AutorImpl;

  factory Autor.fromJson(Map<String, dynamic> jsonSerialization) {
    return Autor(
      id: jsonSerialization['id'] as int?,
      nombre: jsonSerialization['nombre'] as String,
      tareas: jsonSerialization['tareas'] == null
          ? null
          : _idw36flt.Protocol().deserialize<List<_i8g1zmuz.Tarea>>(
              jsonSerialization['tareas'],
            ),
    );
  }

  /// The database id, set if the object has been inserted into the
  /// database or if it has been fetched from the database. Otherwise,
  /// the id will be null.
  int? id;

  String nombre;

  List<_i8g1zmuz.Tarea>? tareas;

  /// Returns a shallow copy of this [Autor]
  /// with some or all fields replaced by the given arguments.
  @_isc.useResult
  Autor copyWith({
    int? id,
    String? nombre,
    List<_i8g1zmuz.Tarea>? tareas =
        const _isc.$UndefinedList<_i8g1zmuz.Tarea>(),
  });
  @override
  Map<String, dynamic> toJson() {
    return {
      '__className__': 'Autor',
      if (id != null) 'id': id,
      'nombre': nombre,
      if (tareas != null)
        'tareas': tareas?.toJson(valueToJson: (v) => v.toJson()),
    };
  }

  @override
  Map<String, dynamic> toJsonForProtocol() {
    return {
      '__className__': 'Autor',
      if (id != null) 'id': id,
      'nombre': nombre,
      if (tareas != null)
        'tareas': tareas?.toJson(valueToJson: (v) => v.toJsonForProtocol()),
    };
  }

  @override
  String toString() {
    return _isc.SerializationManager.encode(this);
  }
}

class _Undefined {}

class _AutorImpl extends Autor {
  _AutorImpl({
    int? id,
    required String nombre,
    List<_i8g1zmuz.Tarea>? tareas,
  }) : super._(
         id: id,
         nombre: nombre,
         tareas: tareas,
       );

  /// Returns a shallow copy of this [Autor]
  /// with some or all fields replaced by the given arguments.
  @_isc.useResult
  @override
  Autor copyWith({
    Object? id = _Undefined,
    String? nombre,
    List<_i8g1zmuz.Tarea>? tareas =
        const _isc.$UndefinedList<_i8g1zmuz.Tarea>(),
  }) {
    return Autor(
      id: id is int? ? id : this.id,
      nombre: nombre ?? this.nombre,
      tareas: tareas is _isc.UndefinedSentinel
          ? this.tareas?.map((e0) => e0.copyWith()).toList()
          : tareas,
    );
  }
}
