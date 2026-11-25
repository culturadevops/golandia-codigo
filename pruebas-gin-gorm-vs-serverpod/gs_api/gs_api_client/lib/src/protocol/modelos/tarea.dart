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
import '../modelos/autor.dart' as _ifj65lf3;

abstract class Tarea
    implements _isc.SerializableModel, _isc.ProtocolSerialization {
  Tarea._({
    this.id,
    required this.titulo,
    required this.hecha,
    required this.autorId,
    this.autor,
  });

  factory Tarea({
    int? id,
    required String titulo,
    required bool hecha,
    required int autorId,
    _ifj65lf3.Autor? autor,
  }) = _TareaImpl;

  factory Tarea.fromJson(Map<String, dynamic> jsonSerialization) {
    return Tarea(
      id: jsonSerialization['id'] as int?,
      titulo: jsonSerialization['titulo'] as String,
      hecha: _isc.BoolJsonExtension.fromJson(jsonSerialization['hecha']),
      autorId: jsonSerialization['autorId'] as int,
      autor: jsonSerialization['autor'] == null
          ? null
          : _idw36flt.Protocol().deserialize<_ifj65lf3.Autor>(
              jsonSerialization['autor'],
            ),
    );
  }

  /// The database id, set if the object has been inserted into the
  /// database or if it has been fetched from the database. Otherwise,
  /// the id will be null.
  int? id;

  String titulo;

  bool hecha;

  int autorId;

  _ifj65lf3.Autor? autor;

  /// Returns a shallow copy of this [Tarea]
  /// with some or all fields replaced by the given arguments.
  @_isc.useResult
  Tarea copyWith({
    int? id,
    String? titulo,
    bool? hecha,
    int? autorId,
    _ifj65lf3.Autor? autor = const _UndefinedTarea$autor(),
  });
  @override
  Map<String, dynamic> toJson() {
    return {
      '__className__': 'Tarea',
      if (id != null) 'id': id,
      'titulo': titulo,
      'hecha': hecha,
      'autorId': autorId,
      if (autor != null) 'autor': autor?.toJson(),
    };
  }

  @override
  Map<String, dynamic> toJsonForProtocol() {
    return {
      '__className__': 'Tarea',
      if (id != null) 'id': id,
      'titulo': titulo,
      'hecha': hecha,
      'autorId': autorId,
      if (autor != null) 'autor': autor?.toJsonForProtocol(),
    };
  }

  @override
  String toString() {
    return _isc.SerializationManager.encode(this);
  }
}

class _Undefined {}

class _UndefinedTarea$autor extends _isc.UndefinedSentinel
    implements _ifj65lf3.Autor {
  const _UndefinedTarea$autor();
}

class _TareaImpl extends Tarea {
  _TareaImpl({
    int? id,
    required String titulo,
    required bool hecha,
    required int autorId,
    _ifj65lf3.Autor? autor,
  }) : super._(
         id: id,
         titulo: titulo,
         hecha: hecha,
         autorId: autorId,
         autor: autor,
       );

  /// Returns a shallow copy of this [Tarea]
  /// with some or all fields replaced by the given arguments.
  @_isc.useResult
  @override
  Tarea copyWith({
    Object? id = _Undefined,
    String? titulo,
    bool? hecha,
    int? autorId,
    _ifj65lf3.Autor? autor = const _UndefinedTarea$autor(),
  }) {
    return Tarea(
      id: id is int? ? id : this.id,
      titulo: titulo ?? this.titulo,
      hecha: hecha ?? this.hecha,
      autorId: autorId ?? this.autorId,
      autor: autor is _isc.UndefinedSentinel ? this.autor?.copyWith() : autor,
    );
  }
}
