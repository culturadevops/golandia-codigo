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
import 'package:gs_api_server/src/generated/modelos/autor.dart' as _imvges7y;
import 'package:serverpod/protocol.dart' as _isp;
import 'package:serverpod/serverpod.dart' as _is;
import 'modelos/autor.dart' as _icetbgw5;
import 'modelos/tarea.dart' as _ib2abh2r;
export 'modelos/autor.dart';
export 'modelos/tarea.dart';

class Protocol extends _is.DatabaseSerializationManager
    implements _is.DatabaseProtocolDeserializationProvider {
  Protocol._();

  factory Protocol() => _instance;

  static final Protocol _instance = Protocol._();

  @override
  _is.ProtocolDeserialization get deserializationMetadata =>
      _is.ProtocolDeserialization.cached(
        this,
        types: const [
          _is.getType<_icetbgw5.Autor>,
          _is.getType<_ib2abh2r.Tarea>,
          _is.getType<_icetbgw5.Autor?>,
          _is.getType<_ib2abh2r.Tarea?>,
          _is.getType<List<_ib2abh2r.Tarea>>,
          _is.getType<List<_ib2abh2r.Tarea>?>,
          _is.getType<List<_imvges7y.Autor>>,
        ],
        modules: const [_isp.Protocol.new],
        classNames: const [
          'Autor',
          'Tarea',
        ],
      );

  static List<_isp.TableDefinition> get targetTableDefinitions => [
    ..._ownTableDefinitions,
    ..._isp.Protocol.targetTableDefinitions,
  ];

  static List<_isp.TableDefinition> get _ownTableDefinitions => [
    _isp.TableDefinition(
      name: 'autor',
      dartName: 'Autor',
      schema: 'public',
      module: 'gs_api',
      columns: [
        _isp.ColumnDefinition(
          name: 'id',
          columnType: _isp.ColumnType.bigint,
          isNullable: false,
          dartType: 'int?',
          columnDefault: 'serial',
        ),
        _isp.ColumnDefinition(
          name: 'nombre',
          columnType: _isp.ColumnType.text,
          isNullable: false,
          dartType: 'String',
        ),
      ],
      foreignKeys: [],
      indexes: [],
      managed: true,
    ),
    _isp.TableDefinition(
      name: 'tarea',
      dartName: 'Tarea',
      schema: 'public',
      module: 'gs_api',
      columns: [
        _isp.ColumnDefinition(
          name: 'id',
          columnType: _isp.ColumnType.bigint,
          isNullable: false,
          dartType: 'int?',
          columnDefault: 'serial',
        ),
        _isp.ColumnDefinition(
          name: 'titulo',
          columnType: _isp.ColumnType.text,
          isNullable: false,
          dartType: 'String',
        ),
        _isp.ColumnDefinition(
          name: 'hecha',
          columnType: _isp.ColumnType.boolean,
          isNullable: false,
          dartType: 'bool',
        ),
        _isp.ColumnDefinition(
          name: 'autorId',
          columnType: _isp.ColumnType.bigint,
          isNullable: false,
          dartType: 'int',
        ),
      ],
      foreignKeys: [
        _isp.ForeignKeyDefinition(
          constraintName: 'tarea_fk_0',
          columns: ['autorId'],
          referenceTable: 'autor',
          referenceTableSchema: 'public',
          referenceColumns: ['id'],
          onUpdate: _isp.ForeignKeyAction.noAction,
          onDelete: _isp.ForeignKeyAction.noAction,
          matchType: null,
        ),
      ],
      indexes: [],
      managed: true,
    ),
  ];

  static String? getClassNameFromObjectJson(dynamic data) =>
      _is.ProtocolDeserialization.readClassName(data);

  @override
  T deserialize<T>(
    dynamic data, [
    Type? t,
  ]) {
    t ??= T;
    final rawDataClassName = _is.ProtocolDeserialization.readClassName(data);
    final dataClassName = rawDataClassName;
    if (dataClassName != null && dataClassName != getClassNameForType(t)) {
      try {
        return deserializeByClassName({
          'className': dataClassName,
          'data': data,
        });
      } on _is.DeserializationClassNameNotFoundException catch (_) {
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
    if (t == _is.getType<_icetbgw5.Autor?>()) {
      return (data != null ? _icetbgw5.Autor.fromJson(data) : null) as T;
    }
    if (t == _is.getType<_ib2abh2r.Tarea?>()) {
      return (data != null ? _ib2abh2r.Tarea.fromJson(data) : null) as T;
    }
    if (t == List<_ib2abh2r.Tarea>) {
      return (data as List).map((e) => deserialize<_ib2abh2r.Tarea>(e)).toList()
          as T;
    }
    if (t == _is.getType<List<_ib2abh2r.Tarea>?>()) {
      return (data != null
              ? (data as List)
                    .map((e) => deserialize<_ib2abh2r.Tarea>(e))
                    .toList()
              : null)
          as T;
    }
    if (t == List<_imvges7y.Autor>) {
      return (data as List).map((e) => deserialize<_imvges7y.Autor>(e)).toList()
          as T;
    }
    final modules = deserializationMetadata.modulesForDeserialization(
      t,
      className: rawDataClassName,
    );
    for (final module in modules) {
      try {
        return module.deserialize<T>(data, t);
      } on _is.DeserializationTypeNotFoundException catch (_) {}
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
    className = _isp.Protocol().getClassNameForObject(data);
    if (className != null) {
      return className.contains('.') ? className : 'serverpod.$className';
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
    if (dataClassName.startsWith('serverpod.')) {
      return _isp.Protocol().deserializeByClassName({
        ...data,
        'className': dataClassName.substring(10),
      });
    }
    return super.deserializeByClassName(data);
  }

  @override
  _is.Table? getTableForType(Type t) {
    {
      var table = _isp.Protocol().getTableForType(t);
      if (table != null) {
        return table;
      }
    }
    switch (t) {
      case _icetbgw5.Autor:
        return _icetbgw5.Autor.t;
      case _ib2abh2r.Tarea:
        return _ib2abh2r.Tarea.t;
    }
    return null;
  }

  @override
  List<_isp.TableDefinition> getTargetTableDefinitions() =>
      targetTableDefinitions;

  @override
  List<_isp.TableDefinition> getOwnTableDefinitions() => _ownTableDefinitions;

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
    try {
      return _isp.Protocol().mapRecordToJson(record);
    } catch (_) {}
    throw Exception('Unsupported record type ${record.runtimeType}');
  }
}
