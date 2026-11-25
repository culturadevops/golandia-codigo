/* AUTOMATICALLY GENERATED CODE DO NOT MODIFY */
/*   To generate run: "serverpod generate"    */

// ignore_for_file: implementation_imports
// ignore_for_file: library_private_types_in_public_api
// ignore_for_file: non_constant_identifier_names
// ignore_for_file: public_member_api_docs
// ignore_for_file: type_literal_in_constant_pattern
// ignore_for_file: use_super_parameters
// ignore_for_file: invalid_use_of_internal_member
// ignore_for_file: dead_code, unnecessary_type_check

// ignore_for_file: no_leading_underscores_for_library_prefixes
import 'package:gs_api_client/src/protocol/modelos/autor.dart' as _i91hbr78;
import 'package:serverpod_client/serverpod_client.dart' as _isc;
import 'modelos/autor.dart' as _icetbgw5;
import 'modelos/tarea.dart' as _ib2abh2r;
export 'modelos/autor.dart';
export 'modelos/tarea.dart';
export 'client.dart';

class Protocol extends _isc.SerializationManager
    implements _isc.ProtocolDeserializationProvider {
  Protocol._();

  factory Protocol() => _instance;

  static final Protocol _instance = Protocol._();

  @override
  _isc.ProtocolDeserialization get deserializationMetadata =>
      _isc.ProtocolDeserialization.cached(
        this,
        types: const [
          _isc.getType<_icetbgw5.Autor>,
          _isc.getType<_ib2abh2r.Tarea>,
          _isc.getType<_icetbgw5.Autor?>,
          _isc.getType<_ib2abh2r.Tarea?>,
          _isc.getType<List<_ib2abh2r.Tarea>>,
          _isc.getType<List<_ib2abh2r.Tarea>?>,
          _isc.getType<List<_i91hbr78.Autor>>,
        ],
        modules: const [],
        classNames: const [
          'Autor',
          'Tarea',
        ],
      );

  static String? getClassNameFromObjectJson(dynamic data) =>
      _isc.ProtocolDeserialization.readClassName(data);

  @override
  T deserialize<T>(
    dynamic data, [
    Type? t,
  ]) {
    t ??= T;
    final rawDataClassName = _isc.ProtocolDeserialization.readClassName(data);
    final dataClassName = rawDataClassName;
    if (dataClassName != null && dataClassName != getClassNameForType(t)) {
      try {
        return deserializeByClassName({
          'className': dataClassName,
          'data': data,
        });
      } on _isc.DeserializationClassNameNotFoundException catch (_) {
        // If the className is not recognized (e.g., older client receiving
        // data with a new subtype), fall back to deserializing without the
        // className, using the expected type T.
      }
    }

    if (t == _icetbgw5.Autor) {
      return _icetbgw5.Autor.fromJson(data) as T;
    }
    if (t == _ib2abh2r.Tarea) {
      return _ib2abh2r.Tarea.fromJson(data) as T;
    }
    if (t == _isc.getType<_icetbgw5.Autor?>()) {
      return (data != null ? _icetbgw5.Autor.fromJson(data) : null) as T;
    }
    if (t == _isc.getType<_ib2abh2r.Tarea?>()) {
      return (data != null ? _ib2abh2r.Tarea.fromJson(data) : null) as T;
    }
    if (t == List<_ib2abh2r.Tarea>) {
      return (data as List).map((e) => deserialize<_ib2abh2r.Tarea>(e)).toList()
          as T;
    }
    if (t == _isc.getType<List<_ib2abh2r.Tarea>?>()) {
      return (data != null
              ? (data as List)
                    .map((e) => deserialize<_ib2abh2r.Tarea>(e))
                    .toList()
              : null)
          as T;
    }
    if (t == List<_i91hbr78.Autor>) {
      return (data as List).map((e) => deserialize<_i91hbr78.Autor>(e)).toList()
          as T;
    }
    return super.deserialize<T>(data, t);
  }

  static String? getClassNameForType(Type type) {
    return switch (type) {
      _icetbgw5.Autor => 'Autor',
      _ib2abh2r.Tarea => 'Tarea',
      _ => null,
    };
  }

  @override
  String? getClassNameForObject(Object? data) {
    String? className = super.getClassNameForObject(data);
    if (className != null) return className;

    if (data is Map<String, dynamic> && data['__className__'] is String) {
      return (data['__className__'] as String).replaceFirst('gs_api.', '');
    }

    switch (data) {
      case _icetbgw5.Autor():
        return 'Autor';
      case _ib2abh2r.Tarea():
        return 'Tarea';
    }
    return null;
  }

  @override
  dynamic deserializeByClassName(Map<String, dynamic> data) {
    var dataClassName = data['className'];
    if (dataClassName is! String) {
      return super.deserializeByClassName(data);
    }
    if (dataClassName == 'Autor') {
      return deserialize<_icetbgw5.Autor>(data['data']);
    }
    if (dataClassName == 'Tarea') {
      return deserialize<_ib2abh2r.Tarea>(data['data']);
    }
    return super.deserializeByClassName(data);
  }

  @override
  String getModuleName() => 'gs_api';

  /// Maps any `Record`s known to this [Protocol] to their JSON representation
  ///
  /// Throws in case the record type is not known.
  ///
  /// This method will return `null` (only) for `null` inputs.
  Map<String, dynamic>? mapRecordToJson(Record? record) {
    if (record == null) {
      return null;
    }
    throw Exception('Unsupported record type ${record.runtimeType}');
  }
}
